//go:build integration

package deliveryworker_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/persistence"
	deliveryworker "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/worker"
	"github.com/jcsoftdev/pulzifi-back/shared/crypto"
)

// ---------------------------------------------------------------------------
// DB helpers
// ---------------------------------------------------------------------------

func openWorkerTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		host := os.Getenv("DB_HOST")
		if host == "" {
			t.Skip("DATABASE_URL (or DB_HOST) not set; skipping integration test")
		}
		port := os.Getenv("DB_PORT")
		if port == "" {
			port = "5432"
		}
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), host, port, os.Getenv("DB_NAME"))
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatalf("db.Ping: %v", err)
	}
	return db
}

const workerTestTenant = "jcsoftdev_inc"

// insertWorkerTestOrg seeds a minimal org row that the worker's listTenants query will return.
// Returns the org ID and a cleanup func.
func insertWorkerTestOrg(t *testing.T, ctx context.Context, db *sql.DB) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	orgID := uuid.New()

	_, err := db.ExecContext(ctx, `
		INSERT INTO public.users (id, email, password_hash, status)
		VALUES ($1, $2, 'testhash', 'approved')
		ON CONFLICT DO NOTHING`,
		userID, fmt.Sprintf("workertest-%s@example.com", userID),
	)
	if err != nil {
		t.Fatalf("insertWorkerTestOrg: insert user: %v", err)
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO public.organizations (id, name, subdomain, schema_name, owner_user_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (schema_name) DO NOTHING`,
		orgID,
		"Worker Test Org",
		"workertest-"+orgID.String()[:8],
		workerTestTenant,
		userID,
	)
	if err != nil {
		t.Fatalf("insertWorkerTestOrg: insert org: %v", err)
	}

	t.Cleanup(func() {
		db.ExecContext(ctx, `DELETE FROM public.organizations WHERE id = $1`, orgID)
		db.ExecContext(ctx, `DELETE FROM public.users WHERE id = $1`, userID)
	})
	return orgID
}

// insertWorkerTestDest inserts a minimal destination row directly into the tenant schema.
func insertWorkerTestDest(t *testing.T, ctx context.Context, db *sql.DB, integID *uuid.UUID) uuid.UUID {
	t.Helper()
	destID := uuid.New()
	scopeID := uuid.New()

	if _, err := db.ExecContext(ctx, "SET search_path TO "+workerTestTenant+", public"); err != nil {
		t.Fatalf("insertWorkerTestDest: set search path: %v", err)
	}

	var integIDVal interface{} = nil
	if integID != nil {
		integIDVal = *integID
	}

	_, err := db.ExecContext(ctx, `
		INSERT INTO integration_destinations
			(id, integration_id, service_type, scope_type, scope_id, target, events, enabled, created_at, updated_at)
		VALUES ($1, $2, 'slack', 'org', $3, '{"channel_id":"C123"}', ARRAY['test.event'], true, NOW(), NOW())`,
		destID, integIDVal, scopeID,
	)
	if err != nil {
		t.Fatalf("insertWorkerTestDest: %v", err)
	}

	t.Cleanup(func() {
		db.ExecContext(ctx, "SET search_path TO "+workerTestTenant+", public")
		db.ExecContext(ctx, `DELETE FROM integration_deliveries WHERE destination_id = $1`, destID)
		db.ExecContext(ctx, `DELETE FROM integration_destinations WHERE id = $1`, destID)
	})
	return destID
}

// insertWorkerTestDelivery inserts a pending delivery row directly.
func insertWorkerTestDelivery(t *testing.T, ctx context.Context, db *sql.DB, destID uuid.UUID) uuid.UUID {
	t.Helper()
	delID := uuid.New()

	if _, err := db.ExecContext(ctx, "SET search_path TO "+workerTestTenant+", public"); err != nil {
		t.Fatalf("insertWorkerTestDelivery: set search path: %v", err)
	}

	_, err := db.ExecContext(ctx, `
		INSERT INTO integration_deliveries
			(id, destination_id, event_type, event_payload, status, attempts, next_attempt_at, created_at)
		VALUES ($1, $2, 'test.event', '{}', 'pending', 0, NOW(), NOW())`,
		delID, destID,
	)
	if err != nil {
		t.Fatalf("insertWorkerTestDelivery: %v", err)
	}

	t.Cleanup(func() {
		db.ExecContext(ctx, "SET search_path TO "+workerTestTenant+", public")
		db.ExecContext(ctx, `DELETE FROM integration_deliveries WHERE id = $1`, delID)
	})
	return delID
}

// fetchWorkerDelivery re-reads a delivery row for assertions.
func fetchWorkerDelivery(t *testing.T, ctx context.Context, db *sql.DB, id uuid.UUID) *entities.Delivery {
	t.Helper()
	if _, err := db.ExecContext(ctx, "SET search_path TO "+workerTestTenant+", public"); err != nil {
		t.Fatalf("fetchWorkerDelivery: set search path: %v", err)
	}
	row := db.QueryRowContext(ctx, `
		SELECT id, destination_id, event_type, event_payload, status, attempts,
		       last_attempt_at, next_attempt_at, response_code, response_body,
		       error_message, delivered_at, created_at
		FROM integration_deliveries WHERE id = $1`, id)

	var d entities.Delivery
	var payloadJSON []byte
	var lastAttemptAt sql.NullTime
	var nextAttemptAt time.Time
	var responseCode sql.NullInt32
	var responseBody sql.NullString
	var errorMessage sql.NullString
	var deliveredAt sql.NullTime

	err := row.Scan(
		&d.ID, &d.DestinationID, &d.EventType, &payloadJSON,
		&d.Status, &d.Attempts,
		&lastAttemptAt, &nextAttemptAt,
		&responseCode, &responseBody, &errorMessage, &deliveredAt,
		&d.CreatedAt,
	)
	if err != nil {
		t.Fatalf("fetchWorkerDelivery scan %s: %v", id, err)
	}
	if lastAttemptAt.Valid {
		v := lastAttemptAt.Time
		d.LastAttemptAt = &v
	}
	d.NextAttemptAt = &nextAttemptAt
	if responseCode.Valid {
		v := int(responseCode.Int32)
		d.ResponseCode = &v
	}
	if responseBody.Valid {
		v := responseBody.String
		d.ResponseBody = &v
	}
	if errorMessage.Valid {
		v := errorMessage.String
		d.ErrorMessage = &v
	}
	if deliveredAt.Valid {
		v := deliveredAt.Time
		d.DeliveredAt = &v
	}
	return &d
}

// ---------------------------------------------------------------------------
// Mock provider client
// ---------------------------------------------------------------------------

// mockResult encodes what the mock should return for one Send call.
type mockResult struct {
	result *entities.DeliveryResult
	err    error
}

// mockProviderClient returns pre-configured results in order.
type mockProviderClient struct {
	results []mockResult
	calls   int
}

func (m *mockProviderClient) ServiceType() string { return "slack" }

func (m *mockProviderClient) OAuthAuthorizeURL(_, _ string) (string, error) { return "", nil }

func (m *mockProviderClient) HandleCallback(_ context.Context, _, _ string) (*entities.OAuthResult, error) {
	return nil, nil
}

func (m *mockProviderClient) RefreshAccessToken(_ context.Context, _ string) (*entities.OAuthResult, error) {
	return nil, nil
}

func (m *mockProviderClient) ListTargets(_ context.Context, _ *entities.Integration) ([]entities.Target, error) {
	return nil, nil
}

func (m *mockProviderClient) Send(_ context.Context, _ *entities.Integration, _ *entities.Destination, _ *entities.NotificationPayload) (*entities.DeliveryResult, error) {
	if m.calls >= len(m.results) {
		return &entities.DeliveryResult{Code: 200}, nil
	}
	r := m.results[m.calls]
	m.calls++
	return r.result, r.err
}

// ---------------------------------------------------------------------------
// Mock provider registry
// ---------------------------------------------------------------------------

type mockRegistry struct {
	client services.ProviderClient
}

func (r *mockRegistry) Get(serviceType string) (services.ProviderClient, bool) {
	if serviceType == "slack" {
		return r.client, true
	}
	return nil, false
}

// ---------------------------------------------------------------------------
// Mock repo factory (wraps real postgres repos)
// ---------------------------------------------------------------------------

type postgresRepoFactory struct{ db *sql.DB }

func (f *postgresRepoFactory) DestinationRepoForTenant(tenant string) repositories.DestinationRepository {
	return persistence.NewDestinationPostgresRepository(f.db, tenant)
}

func (f *postgresRepoFactory) DeliveryRepoForTenant(tenant string) repositories.DeliveryRepository {
	return persistence.NewDeliveryPostgresRepository(f.db, tenant)
}

// ---------------------------------------------------------------------------
// TestDeliveryWorker_TickProcessesAllOutcomes
// ---------------------------------------------------------------------------

// TestDeliveryWorker_TickProcessesAllOutcomes inserts 3 pending deliveries and a
// mock client that returns (success / 5xx error / 401 error) for the three deliveries.
// It calls worker.tick directly with Synchronous=true to avoid timing issues.
func TestDeliveryWorker_TickProcessesAllOutcomes(t *testing.T) {
	ctx := context.Background()
	db := openWorkerTestDB(t)
	t.Cleanup(func() { db.Close() })

	// Ensure the jcsoftdev_inc schema is registered in public.organizations.
	insertWorkerTestOrg(t, ctx, db)

	// Set up encryption for IntegrationRepo.
	enc, err := crypto.NewAESGCM(make([]byte, 32))
	if err != nil {
		t.Fatalf("NewAESGCM: %v", err)
	}
	intRepo := persistence.NewIntegrationPostgresRepository(db, enc)

	// Insert 3 destinations (no integration required for the test — nil integration_id is fine).
	dest1 := insertWorkerTestDest(t, ctx, db, nil)
	dest2 := insertWorkerTestDest(t, ctx, db, nil)
	dest3 := insertWorkerTestDest(t, ctx, db, nil)

	// Insert 3 deliveries.
	del1 := insertWorkerTestDelivery(t, ctx, db, dest1) // → success (200)
	del2 := insertWorkerTestDelivery(t, ctx, db, dest2) // → 5xx transient error
	del3 := insertWorkerTestDelivery(t, ctx, db, dest3) // → 401 auth error

	// Mock client: call 1 = ok, call 2 = 5xx, call 3 = 401
	mock := &mockProviderClient{
		results: []mockResult{
			{result: &entities.DeliveryResult{Code: 200, BodySnip: "ok"}, err: nil},
			{result: &entities.DeliveryResult{Code: 503, BodySnip: "unavailable"}, err: errors.New("service unavailable")},
			{result: &entities.DeliveryResult{Code: 401, BodySnip: "unauthorized"}, err: errors.New("unauthorized")},
		},
	}

	registry := &mockRegistry{client: mock}
	factory := &postgresRepoFactory{db: db}

	worker := deliveryworker.New(
		db,
		factory,
		intRepo,
		registry,
		services.NewPayloadBuilder(),
		deliveryworker.Config{
			PoolSize:       5,
			ClaimBatchSize: 50,
			PollInterval:   time.Hour, // unused — we call tick directly
			MaxAttempts:    5,
			Synchronous:    true,
		},
	)

	// Run one tick synchronously.
	sem := make(chan struct{}, 5)
	worker.TickForTest(ctx, sem)

	// Assert delivery 1 → delivered
	d1 := fetchWorkerDelivery(t, ctx, db, del1)
	if d1.Status != entities.DeliveryDelivered {
		t.Errorf("delivery 1: want delivered, got %q", d1.Status)
	}
	if d1.ResponseCode == nil || *d1.ResponseCode != 200 {
		t.Errorf("delivery 1: response_code want 200, got %v", d1.ResponseCode)
	}

	// Assert delivery 2 → pending (rescheduled) with attempts=1
	d2 := fetchWorkerDelivery(t, ctx, db, del2)
	if d2.Status != entities.DeliveryPending {
		t.Errorf("delivery 2: want pending (rescheduled), got %q", d2.Status)
	}
	if d2.Attempts != 1 {
		t.Errorf("delivery 2: want attempts=1, got %d", d2.Attempts)
	}
	if d2.NextAttemptAt == nil || !d2.NextAttemptAt.After(time.Now()) {
		t.Errorf("delivery 2: next_attempt_at should be in the future, got %v", d2.NextAttemptAt)
	}

	// Re-fetch delivery 2 via repo to get AttemptHistory populated.
	delRepo2 := factory.DeliveryRepoForTenant(workerTestTenant)
	d2Full, err := delRepo2.GetByID(ctx, del2)
	if err != nil {
		t.Fatalf("GetByID delivery 2: %v", err)
	}
	if d2Full == nil {
		t.Fatal("GetByID delivery 2 returned nil")
	}
	if len(d2Full.AttemptHistory) != 1 {
		t.Errorf("delivery 2: want AttemptHistory len=1, got %d", len(d2Full.AttemptHistory))
	} else {
		h := d2Full.AttemptHistory[0]
		want503 := 503
		if h.Code == nil || *h.Code != want503 {
			t.Errorf("delivery 2: AttemptHistory[0].Code want 503, got %v", h.Code)
		}
	}

	// Assert delivery 3 → dead (401 auth failure)
	d3 := fetchWorkerDelivery(t, ctx, db, del3)
	if d3.Status != entities.DeliveryDead {
		t.Errorf("delivery 3: want dead, got %q", d3.Status)
	}
	if d3.ErrorMessage == nil || *d3.ErrorMessage == "" {
		t.Errorf("delivery 3: error_message should be set")
	}

	// Re-fetch delivery 3 via repo to get AttemptHistory populated.
	delRepo3 := factory.DeliveryRepoForTenant(workerTestTenant)
	d3Full, err := delRepo3.GetByID(ctx, del3)
	if err != nil {
		t.Fatalf("GetByID delivery 3: %v", err)
	}
	if d3Full == nil {
		t.Fatal("GetByID delivery 3 returned nil")
	}
	if len(d3Full.AttemptHistory) != 1 {
		t.Errorf("delivery 3: want AttemptHistory len=1, got %d", len(d3Full.AttemptHistory))
	} else {
		h := d3Full.AttemptHistory[0]
		want401 := 401
		if h.Code == nil || *h.Code != want401 {
			t.Errorf("delivery 3: AttemptHistory[0].Code want 401, got %v", h.Code)
		}
	}
}
