//go:build integration

package deliveryworker_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/persistence"
	providers "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/providers"
	sheetsprovider "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/providers/sheets"
	teamsprovider "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/providers/teams"
	deliveryworker "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/worker"
	"github.com/jcsoftdev/pulzifi-back/shared/crypto"
	"github.com/jcsoftdev/pulzifi-back/shared/testguard"
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
	testguard.RequireLocalDB(t, dsn)
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

// ---------------------------------------------------------------------------
// Helper: resolve the actual org ID for workerTestTenant (handles ON CONFLICT DO NOTHING).
// ---------------------------------------------------------------------------

func resolveWorkerOrgID(t *testing.T, ctx context.Context, db *sql.DB) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := db.QueryRowContext(ctx,
		`SELECT id FROM public.organizations WHERE schema_name = $1 LIMIT 1`,
		workerTestTenant,
	).Scan(&id)
	if err != nil {
		t.Fatalf("resolveWorkerOrgID: %v", err)
	}
	return id
}

// ---------------------------------------------------------------------------
// Helper: insert integration row in the public schema.
// ---------------------------------------------------------------------------

func insertWorkerIntegration(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	enc *crypto.AESGCM,
	orgID uuid.UUID,
	serviceType string,
	accessToken string,
	refreshToken string,
	expiresAt *time.Time,
) uuid.UUID {
	t.Helper()
	intRepo := persistence.NewIntegrationPostgresRepository(db, enc)
	integ := &entities.Integration{
		ID:             uuid.New(),
		OrgID:         orgID,
		ServiceType:   serviceType,
		Status:        entities.IntegrationActive,
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		TokenExpiresAt: expiresAt,
		ProviderMeta:  map[string]any{},
		CreatedBy:     orgID, // reuse orgID as a stand-in for created_by user
	}
	if err := intRepo.Create(ctx, integ); err != nil {
		t.Fatalf("insertWorkerIntegration: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, `DELETE FROM public.integrations WHERE id = $1`, integ.ID)
	})
	return integ.ID
}

// ---------------------------------------------------------------------------
// Helper: insert a destination with arbitrary service_type and target JSON.
// ---------------------------------------------------------------------------

func insertWorkerTestDestFull(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	integID *uuid.UUID,
	serviceType string,
	targetJSON string,
) uuid.UUID {
	t.Helper()
	destID := uuid.New()
	scopeID := uuid.New()

	if _, err := db.ExecContext(ctx, "SET search_path TO "+workerTestTenant+", public"); err != nil {
		t.Fatalf("insertWorkerTestDestFull: set search path: %v", err)
	}

	var integIDVal interface{} = nil
	if integID != nil {
		integIDVal = *integID
	}

	_, err := db.ExecContext(ctx, `
		INSERT INTO integration_destinations
			(id, integration_id, service_type, scope_type, scope_id, target, events, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, 'org', $4, $5::jsonb, ARRAY['test.event'], true, NOW(), NOW())`,
		destID, integIDVal, serviceType, scopeID, targetJSON,
	)
	if err != nil {
		t.Fatalf("insertWorkerTestDestFull: %v", err)
	}

	t.Cleanup(func() {
		db.ExecContext(ctx, "SET search_path TO "+workerTestTenant+", public")
		db.ExecContext(ctx, `DELETE FROM integration_deliveries WHERE destination_id = $1`, destID)
		db.ExecContext(ctx, `DELETE FROM integration_destinations WHERE id = $1`, destID)
	})
	return destID
}

// ---------------------------------------------------------------------------
// TestDeliveryProcessor_RefreshesSheetsToken
// ---------------------------------------------------------------------------

func TestDeliveryProcessor_RefreshesSheetsToken(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	db := openWorkerTestDB(t)
	t.Cleanup(func() { db.Close() })

	enc, err := crypto.NewAESGCM(make([]byte, 32))
	if err != nil {
		t.Fatalf("NewAESGCM: %v", err)
	}

	// 1. Mock token endpoint — returns new access token.
	var refreshHits int32
	refreshSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&refreshHits, 1)
		if err := r.ParseForm(); err != nil {
			t.Errorf("token endpoint: ParseForm: %v", err)
			return
		}
		if r.FormValue("grant_type") != "refresh_token" {
			t.Errorf("token endpoint: want grant_type=refresh_token, got %s", r.FormValue("grant_type"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token":"new-sheets-token","expires_in":3600,"scope":"sheets","token_type":"Bearer"}`))
	}))
	defer refreshSrv.Close()

	// 2. Mock Sheets API append endpoint — verifies the new token is used.
	var sendHits int32
	sheetsSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&sendHits, 1)
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer new-sheets-token") {
			t.Errorf("sheets send: want Bearer new-sheets-token, got %s", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer sheetsSrv.Close()

	// Swap provider package-level URL vars; restore on cleanup.
	origTokenURL := sheetsprovider.TokenURL
	origSheetsBase := sheetsprovider.SheetsBase
	sheetsprovider.TokenURL = refreshSrv.URL
	sheetsprovider.SheetsBase = sheetsSrv.URL
	t.Cleanup(func() {
		sheetsprovider.TokenURL = origTokenURL
		sheetsprovider.SheetsBase = origSheetsBase
	})

	// 3. Seed org (the worker's listTenants query reads public.organizations).
	//    insertWorkerTestOrg may do ON CONFLICT DO NOTHING, so resolve the actual org ID after.
	insertWorkerTestOrg(t, ctx, db)
	orgID := resolveWorkerOrgID(t, ctx, db)

	// 4. Insert integration with token expiring in 30s (within the <1m refresh window).
	expiresAt := time.Now().UTC().Add(30 * time.Second)
	integID := insertWorkerIntegration(t, ctx, db, enc, orgID, "sheets",
		"old-sheets-token", "rt-sheets-xxx", &expiresAt)

	// 5. Insert destination (sheets, pointing to a spreadsheet).
	destID := insertWorkerTestDestFull(t, ctx, db, &integID, "sheets",
		`{"spreadsheet_id":"SHEET123","tab_name":"Alerts"}`)

	// 6. Insert pending delivery.
	delID := insertWorkerTestDelivery(t, ctx, db, destID)

	// 7. Build worker with sheets provider in the registry.
	intRepo := persistence.NewIntegrationPostgresRepository(db, enc)
	registry := providers.NewRegistry(
		sheetsprovider.New("test-cid", "test-secret"),
	)
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
			PollInterval:   time.Hour,
			MaxAttempts:    5,
			Synchronous:    true,
		},
	)

	// 8. Run one tick synchronously.
	sem := make(chan struct{}, 5)
	worker.TickForTest(ctx, sem)

	// 9. Assert: token endpoint was called exactly once.
	if got := atomic.LoadInt32(&refreshHits); got != 1 {
		t.Errorf("refreshHits: want 1, got %d", got)
	}

	// 10. Assert: Sheets send endpoint was called exactly once.
	if got := atomic.LoadInt32(&sendHits); got != 1 {
		t.Errorf("sendHits: want 1, got %d", got)
	}

	// 11. Assert: integration row has the new access token persisted.
	updatedInteg, err := intRepo.GetByID(ctx, integID)
	if err != nil {
		t.Fatalf("GetByID after refresh: %v", err)
	}
	if updatedInteg == nil {
		t.Fatal("integration not found after refresh")
	}
	if updatedInteg.AccessToken != "new-sheets-token" {
		t.Errorf("integration AccessToken: want new-sheets-token, got %q", updatedInteg.AccessToken)
	}
	if updatedInteg.TokenExpiresAt == nil {
		t.Error("integration TokenExpiresAt should be non-nil after refresh")
	} else {
		advance := time.Until(*updatedInteg.TokenExpiresAt)
		if advance < 50*time.Minute || advance > 70*time.Minute {
			t.Errorf("integration TokenExpiresAt advanced %v, want ~1h", advance)
		}
	}

	// 12. Assert: delivery is marked delivered.
	d := fetchWorkerDelivery(t, ctx, db, delID)
	if d.Status != entities.DeliveryDelivered {
		t.Errorf("delivery status: want delivered, got %q", d.Status)
	}
}

// ---------------------------------------------------------------------------
// TestDeliveryProcessor_RefreshesTeamsTokenAndRotatesRefresh
// ---------------------------------------------------------------------------

func TestDeliveryProcessor_RefreshesTeamsTokenAndRotatesRefresh(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	db := openWorkerTestDB(t)
	t.Cleanup(func() { db.Close() })

	enc, err := crypto.NewAESGCM(make([]byte, 32))
	if err != nil {
		t.Fatalf("NewAESGCM: %v", err)
	}

	// 1. Mock token endpoint — Microsoft rotates refresh_token on every refresh.
	var refreshHits int32
	refreshSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&refreshHits, 1)
		if err := r.ParseForm(); err != nil {
			t.Errorf("token endpoint: ParseForm: %v", err)
			return
		}
		if r.FormValue("grant_type") != "refresh_token" {
			t.Errorf("token endpoint: want grant_type=refresh_token, got %s", r.FormValue("grant_type"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token":"new-teams-token","refresh_token":"rt-new","expires_in":3600,"scope":"teams","token_type":"Bearer"}`))
	}))
	defer refreshSrv.Close()

	// 2. Mock Teams Graph send endpoint.
	var sendHits int32
	graphSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&sendHits, 1)
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer new-teams-token") {
			t.Errorf("graph send: want Bearer new-teams-token, got %s", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer graphSrv.Close()

	// Swap provider package-level URL vars; restore on cleanup.
	origTokenURL := teamsprovider.TokenURL
	origGraphBase := teamsprovider.GraphBase
	teamsprovider.TokenURL = refreshSrv.URL
	teamsprovider.GraphBase = graphSrv.URL
	t.Cleanup(func() {
		teamsprovider.TokenURL = origTokenURL
		teamsprovider.GraphBase = origGraphBase
	})

	// 3. Seed org.
	//    insertWorkerTestOrg may do ON CONFLICT DO NOTHING, so resolve the actual org ID after.
	insertWorkerTestOrg(t, ctx, db)
	orgID := resolveWorkerOrgID(t, ctx, db)

	// 4. Insert integration with token expiring in 30s (within the <1m refresh window).
	expiresAt := time.Now().UTC().Add(30 * time.Second)
	integID := insertWorkerIntegration(t, ctx, db, enc, orgID, "teams",
		"old-teams-token", "rt-old", &expiresAt)

	// 5. Insert destination — Teams, channel_target in "teamID|channelID" composite form.
	destID := insertWorkerTestDestFull(t, ctx, db, &integID, "teams",
		`{"channel_target":"teamA|chanB"}`)

	// 6. Insert pending delivery.
	delID := insertWorkerTestDelivery(t, ctx, db, destID)

	// 7. Build worker with teams provider in registry.
	intRepo := persistence.NewIntegrationPostgresRepository(db, enc)
	registry := providers.NewRegistry(
		teamsprovider.New("test-cid", "test-secret"),
	)
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
			PollInterval:   time.Hour,
			MaxAttempts:    5,
			Synchronous:    true,
		},
	)

	// 8. Run one tick synchronously.
	sem := make(chan struct{}, 5)
	worker.TickForTest(ctx, sem)

	// 9. Assert: token endpoint was called exactly once.
	if got := atomic.LoadInt32(&refreshHits); got != 1 {
		t.Errorf("refreshHits: want 1, got %d", got)
	}

	// 10. Assert: Teams send endpoint was called exactly once.
	if got := atomic.LoadInt32(&sendHits); got != 1 {
		t.Errorf("sendHits: want 1, got %d", got)
	}

	// 11. Assert: integration row has new access token AND rotated refresh token persisted.
	updatedInteg, err := intRepo.GetByID(ctx, integID)
	if err != nil {
		t.Fatalf("GetByID after refresh: %v", err)
	}
	if updatedInteg == nil {
		t.Fatal("integration not found after refresh")
	}
	if updatedInteg.AccessToken != "new-teams-token" {
		t.Errorf("integration AccessToken: want new-teams-token, got %q", updatedInteg.AccessToken)
	}
	if updatedInteg.RefreshToken != "rt-new" {
		t.Errorf("integration RefreshToken: want rt-new (rotated), got %q", updatedInteg.RefreshToken)
	}
	if updatedInteg.TokenExpiresAt == nil {
		t.Error("integration TokenExpiresAt should be non-nil after refresh")
	} else {
		advance := time.Until(*updatedInteg.TokenExpiresAt)
		if advance < 50*time.Minute || advance > 70*time.Minute {
			t.Errorf("integration TokenExpiresAt advanced %v, want ~1h", advance)
		}
	}

	// 12. Assert: delivery is marked delivered.
	d := fetchWorkerDelivery(t, ctx, db, delID)
	if d.Status != entities.DeliveryDelivered {
		t.Errorf("delivery status: want delivered, got %q", d.Status)
	}
}
