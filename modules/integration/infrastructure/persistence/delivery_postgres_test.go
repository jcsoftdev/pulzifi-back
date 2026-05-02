//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/persistence"
)

// newDelivRepo opens a fresh DB connection and returns a DeliveryRepository
// scoped to testTenant. The connection is closed in t.Cleanup.
func newDelivRepo(t *testing.T) (interface {
	Create(ctx context.Context, d *entities.Delivery) error
	ClaimPending(ctx context.Context, limit int, now time.Time) ([]*entities.Delivery, error)
	MarkDelivered(ctx context.Context, id uuid.UUID, code int, bodySnip string) error
	MarkFailed(ctx context.Context, id uuid.UUID, code *int, body, errMsg string, nextAttempt time.Time) error
	MarkDead(ctx context.Context, id uuid.UUID, errMsg string) error
	ListByDestination(ctx context.Context, destinationID uuid.UUID, limit, offset int) ([]*entities.Delivery, error)
	Retry(ctx context.Context, id uuid.UUID) error
}, *sql.DB) {
	t.Helper()
	db := openTestDB(t)
	t.Cleanup(func() { db.Close() })
	return persistence.NewDeliveryPostgresRepository(db, testTenant), db
}

// insertDestForDelivery inserts a minimal integration_destinations row directly,
// keeping delivery tests independent of DestinationRepo. Registers cleanup.
func insertDestForDelivery(t *testing.T, ctx context.Context, db *sql.DB) uuid.UUID {
	t.Helper()
	destID := uuid.New()
	scopeID := uuid.New()

	if _, err := db.ExecContext(ctx, "SET search_path TO "+testTenant+", public"); err != nil {
		t.Fatalf("insertDestForDelivery: set search path: %v", err)
	}
	_, err := db.ExecContext(ctx, `
		INSERT INTO integration_destinations
			(id, service_type, scope_type, scope_id, target, events, enabled, created_at, updated_at)
		VALUES ($1, 'webhook', 'org', $2, '{}', ARRAY['test.event'], true, NOW(), NOW())`,
		destID, scopeID,
	)
	if err != nil {
		t.Fatalf("insertDestForDelivery: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(ctx,
			"DELETE FROM "+testTenant+".integration_deliveries WHERE destination_id = $1", destID); err != nil {
			t.Logf("cleanup deliveries for dest %s: %v", destID, err)
		}
		if _, err := db.ExecContext(ctx,
			"DELETE FROM "+testTenant+".integration_destinations WHERE id = $1", destID); err != nil {
			t.Logf("cleanup dest %s: %v", destID, err)
		}
	})
	return destID
}

// fetchDelivery queries a single delivery row by ID for assertions.
func fetchDelivery(t *testing.T, ctx context.Context, db *sql.DB, id uuid.UUID) *entities.Delivery {
	t.Helper()
	if _, err := db.ExecContext(ctx, "SET search_path TO "+testTenant+", public"); err != nil {
		t.Fatalf("fetchDelivery: set search path: %v", err)
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
		t.Fatalf("fetchDelivery scan %s: %v", id, err)
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

// --------------------------------------------------------------------------
// TestDeliveryRepo_RoundTrip
// --------------------------------------------------------------------------

func TestDeliveryRepo_RoundTrip(t *testing.T) {
	ctx := context.Background()
	repo, db := newDelivRepo(t)
	destID := insertDestForDelivery(t, ctx, db)

	d := &entities.Delivery{
		ID:            uuid.New(),
		DestinationID: destID,
		EventType:     "test.event",
		EventPayload:  map[string]any{"foo": "bar"},
	}

	// Create
	if err := repo.Create(ctx, d); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// ClaimPending — should return the delivery with status=in_flight, attempts=1
	now := time.Now().UTC()
	claimed, err := repo.ClaimPending(ctx, 10, now)
	if err != nil {
		t.Fatalf("ClaimPending: %v", err)
	}

	var found *entities.Delivery
	for _, c := range claimed {
		if c.ID == d.ID {
			found = c
			break
		}
	}
	if found == nil {
		t.Fatal("ClaimPending did not return the created delivery")
	}
	if found.Status != entities.DeliveryInFlight {
		t.Errorf("status: want in_flight, got %q", found.Status)
	}
	if found.Attempts != 1 {
		t.Errorf("attempts: want 1, got %d", found.Attempts)
	}
	if found.LastAttemptAt == nil {
		t.Error("last_attempt_at should be set after claim")
	}

	// MarkDelivered
	code := 200
	if err := repo.MarkDelivered(ctx, d.ID, code, "OK"); err != nil {
		t.Fatalf("MarkDelivered: %v", err)
	}

	// Re-query and assert
	got := fetchDelivery(t, ctx, db, d.ID)
	if got.Status != entities.DeliveryDelivered {
		t.Errorf("status after delivered: want delivered, got %q", got.Status)
	}
	if got.ResponseCode == nil || *got.ResponseCode != 200 {
		t.Errorf("response_code: want 200, got %v", got.ResponseCode)
	}
	if got.ResponseBody == nil || *got.ResponseBody != "OK" {
		t.Errorf("response_body: want OK, got %v", got.ResponseBody)
	}
	if got.DeliveredAt == nil {
		t.Error("delivered_at should be set")
	}
}

// --------------------------------------------------------------------------
// TestDeliveryRepo_MarkFailedReschedules
// --------------------------------------------------------------------------

func TestDeliveryRepo_MarkFailedReschedules(t *testing.T) {
	ctx := context.Background()
	repo, db := newDelivRepo(t)
	destID := insertDestForDelivery(t, ctx, db)

	d := &entities.Delivery{
		ID:            uuid.New(),
		DestinationID: destID,
		EventType:     "test.event",
		EventPayload:  map[string]any{},
	}
	if err := repo.Create(ctx, d); err != nil {
		t.Fatalf("Create: %v", err)
	}

	now := time.Now().UTC()

	// Claim it
	claimed, err := repo.ClaimPending(ctx, 10, now)
	if err != nil {
		t.Fatalf("ClaimPending: %v", err)
	}
	var found bool
	for _, c := range claimed {
		if c.ID == d.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected delivery to be claimed")
	}

	// MarkFailed with nextAttempt = now + 1 minute
	code := 503
	nextAttempt := now.Add(time.Minute)
	if err := repo.MarkFailed(ctx, d.ID, &code, "Service Unavailable", "endpoint down", nextAttempt); err != nil {
		t.Fatalf("MarkFailed: %v", err)
	}

	// Re-query: status=pending, attempts=1, next_attempt_at moved
	got := fetchDelivery(t, ctx, db, d.ID)
	if got.Status != entities.DeliveryPending {
		t.Errorf("status: want pending, got %q", got.Status)
	}
	if got.Attempts != 1 {
		t.Errorf("attempts: want 1, got %d", got.Attempts)
	}
	if got.NextAttemptAt == nil || !got.NextAttemptAt.After(now) {
		t.Errorf("next_attempt_at should be in the future, got %v", got.NextAttemptAt)
	}

	// ClaimPending with now → returns nothing (next_attempt_at in future)
	claimedNow, err := repo.ClaimPending(ctx, 10, now)
	if err != nil {
		t.Fatalf("ClaimPending(now): %v", err)
	}
	for _, c := range claimedNow {
		if c.ID == d.ID {
			t.Error("delivery should not be claimable yet (next_attempt_at in future)")
		}
	}

	// ClaimPending with now+2min → returns it, attempts=2
	later := now.Add(2 * time.Minute)
	claimedLater, err := repo.ClaimPending(ctx, 10, later)
	if err != nil {
		t.Fatalf("ClaimPending(now+2min): %v", err)
	}
	var foundLater *entities.Delivery
	for _, c := range claimedLater {
		if c.ID == d.ID {
			foundLater = c
			break
		}
	}
	if foundLater == nil {
		t.Fatal("expected delivery to be claimable at now+2min")
	}
	if foundLater.Attempts != 2 {
		t.Errorf("attempts after second claim: want 2, got %d", foundLater.Attempts)
	}
}

// --------------------------------------------------------------------------
// TestDeliveryRepo_MarkDead
// --------------------------------------------------------------------------

func TestDeliveryRepo_MarkDead(t *testing.T) {
	ctx := context.Background()
	repo, db := newDelivRepo(t)
	destID := insertDestForDelivery(t, ctx, db)

	d := &entities.Delivery{
		ID:            uuid.New(),
		DestinationID: destID,
		EventType:     "test.event",
		EventPayload:  map[string]any{},
	}
	if err := repo.Create(ctx, d); err != nil {
		t.Fatalf("Create: %v", err)
	}

	now := time.Now().UTC()

	// Claim
	claimed, err := repo.ClaimPending(ctx, 10, now)
	if err != nil {
		t.Fatalf("ClaimPending: %v", err)
	}
	var found bool
	for _, c := range claimed {
		if c.ID == d.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected delivery to be claimed")
	}

	// MarkDead
	if err := repo.MarkDead(ctx, d.ID, "max retries exceeded"); err != nil {
		t.Fatalf("MarkDead: %v", err)
	}

	got := fetchDelivery(t, ctx, db, d.ID)
	if got.Status != entities.DeliveryDead {
		t.Errorf("status: want dead, got %q", got.Status)
	}
	if got.ErrorMessage == nil || *got.ErrorMessage != "max retries exceeded" {
		t.Errorf("error_message: want 'max retries exceeded', got %v", got.ErrorMessage)
	}
}

// --------------------------------------------------------------------------
// TestDeliveryRepo_ParallelClaim_NoDuplicates
// --------------------------------------------------------------------------

func TestDeliveryRepo_ParallelClaim_NoDuplicates(t *testing.T) {
	ctx := context.Background()

	// Use a shared DB so both repos see the same data.
	db := openTestDB(t)
	t.Cleanup(func() { db.Close() })

	destID := insertDestForDelivery(t, ctx, db)

	repo1 := persistence.NewDeliveryPostgresRepository(db, testTenant)
	repo2 := persistence.NewDeliveryPostgresRepository(db, testTenant)

	// Insert 50 pending deliveries.
	if _, err := db.ExecContext(ctx, "SET search_path TO "+testTenant+", public"); err != nil {
		t.Fatalf("set search path: %v", err)
	}

	ids := make([]uuid.UUID, 50)
	for i := range ids {
		ids[i] = uuid.New()
		_, err := db.ExecContext(ctx, `
			INSERT INTO integration_deliveries
				(id, destination_id, event_type, event_payload, status, attempts, next_attempt_at, created_at)
			VALUES ($1, $2, 'parallel.test', '{}', 'pending', 0, NOW(), NOW())`,
			ids[i], destID,
		)
		if err != nil {
			t.Fatalf("insert delivery %d: %v", i, err)
		}
	}

	now := time.Now().UTC()

	// Launch two goroutines that simultaneously call ClaimPending(20, now).
	type result struct {
		claimed []*entities.Delivery
		err     error
	}
	ch := make(chan result, 2)
	var wg sync.WaitGroup
	wg.Add(2)

	// Use a barrier so both goroutines fire the query at the same time.
	var barrier sync.WaitGroup
	barrier.Add(1)

	launch := func(repo interface {
		ClaimPending(ctx context.Context, limit int, now time.Time) ([]*entities.Delivery, error)
	}) {
		defer wg.Done()
		barrier.Wait() // wait for the start signal
		claimed, err := repo.ClaimPending(ctx, 20, now)
		ch <- result{claimed: claimed, err: err}
	}

	go launch(repo1)
	go launch(repo2)
	barrier.Done() // release both goroutines simultaneously

	wg.Wait()
	close(ch)

	results := make([][]*entities.Delivery, 0, 2)
	for res := range ch {
		if res.err != nil {
			t.Fatalf("ClaimPending goroutine error: %v", res.err)
		}
		results = append(results, res.claimed)
	}

	// Assert: no ID appears twice across both result sets.
	seen := make(map[uuid.UUID]int)
	for _, batch := range results {
		for _, d := range batch {
			seen[d.ID]++
		}
	}
	for id, count := range seen {
		if count > 1 {
			t.Errorf("delivery %s claimed %d times (want 1) — FOR UPDATE SKIP LOCKED not working", id, count)
		}
	}

	// Assert total claimed ≤ 40 (50 inserted, 20 limit per goroutine).
	if len(seen) > 40 {
		t.Errorf("total unique claimed: want ≤40, got %d", len(seen))
	}

	counts := make([]int, len(results))
	for i, batch := range results {
		counts[i] = len(batch)
	}
	t.Logf("parallel claim: goroutine counts %v, unique total %d", counts, len(seen))
}

// --------------------------------------------------------------------------
// TestDeliveryRepo_ListByDestination
// --------------------------------------------------------------------------

func TestDeliveryRepo_ListByDestination(t *testing.T) {
	ctx := context.Background()
	repo, db := newDelivRepo(t)

	dest1 := insertDestForDelivery(t, ctx, db)
	dest2 := insertDestForDelivery(t, ctx, db)

	// Create 3 deliveries for dest1, 1 for dest2.
	// Small sleep between inserts to ensure distinct created_at ordering.
	for i := 0; i < 3; i++ {
		d := &entities.Delivery{
			ID:            uuid.New(),
			DestinationID: dest1,
			EventType:     "list.test",
			EventPayload:  map[string]any{"i": i},
		}
		if err := repo.Create(ctx, d); err != nil {
			t.Fatalf("Create dest1[%d]: %v", i, err)
		}
		time.Sleep(2 * time.Millisecond) // ensure created_at differs
	}
	d2 := &entities.Delivery{
		ID:            uuid.New(),
		DestinationID: dest2,
		EventType:     "list.test",
		EventPayload:  map[string]any{},
	}
	if err := repo.Create(ctx, d2); err != nil {
		t.Fatalf("Create dest2: %v", err)
	}

	// ListByDestination(dest1, 10, 0) → 3 rows.
	list, err := repo.ListByDestination(ctx, dest1, 10, 0)
	if err != nil {
		t.Fatalf("ListByDestination: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("want 3 rows for dest1, got %d", len(list))
	}
	// Verify newest-first ordering.
	for i := 1; i < len(list); i++ {
		if list[i-1].CreatedAt.Before(list[i].CreatedAt) {
			t.Errorf("rows not in DESC order: [%d].created_at=%v < [%d].created_at=%v",
				i-1, list[i-1].CreatedAt, i, list[i].CreatedAt)
		}
	}

	// ListByDestination(dest1, 2, 1) → 2 rows (skip the newest).
	page, err := repo.ListByDestination(ctx, dest1, 2, 1)
	if err != nil {
		t.Fatalf("ListByDestination(limit=2, offset=1): %v", err)
	}
	if len(page) != 2 {
		t.Errorf("want 2 rows with limit=2 offset=1, got %d", len(page))
	}

	// dest2 should have exactly 1 row.
	list2, err := repo.ListByDestination(ctx, dest2, 10, 0)
	if err != nil {
		t.Fatalf("ListByDestination dest2: %v", err)
	}
	if len(list2) != 1 {
		t.Errorf("want 1 row for dest2, got %d", len(list2))
	}
}

// --------------------------------------------------------------------------
// TestDeliveryRepo_Retry
// --------------------------------------------------------------------------

func TestDeliveryRepo_Retry(t *testing.T) {
	ctx := context.Background()
	repo, db := newDelivRepo(t)
	destID := insertDestForDelivery(t, ctx, db)

	d := &entities.Delivery{
		ID:            uuid.New(),
		DestinationID: destID,
		EventType:     "retry.test",
		EventPayload:  map[string]any{"x": 1},
	}
	if err := repo.Create(ctx, d); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Claim so we can call MarkDead (requires in_flight status implied by the worker flow;
	// MarkDead updates by id with no status check in the SQL, so we can call it directly).
	now := time.Now().UTC()
	_, err := repo.ClaimPending(ctx, 10, now)
	if err != nil {
		t.Fatalf("ClaimPending: %v", err)
	}

	// Mark dead.
	if err := repo.MarkDead(ctx, d.ID, "max retries exceeded"); err != nil {
		t.Fatalf("MarkDead: %v", err)
	}

	got := fetchDelivery(t, ctx, db, d.ID)
	if got.Status != entities.DeliveryDead {
		t.Fatalf("pre-condition: want dead, got %q", got.Status)
	}

	// Retry → no error.
	if err := repo.Retry(ctx, d.ID); err != nil {
		t.Fatalf("Retry: %v", err)
	}

	// Re-query: status='pending', attempts=0, nullable fields cleared.
	got = fetchDelivery(t, ctx, db, d.ID)
	if got.Status != entities.DeliveryPending {
		t.Errorf("status after Retry: want pending, got %q", got.Status)
	}
	if got.Attempts != 0 {
		t.Errorf("attempts after Retry: want 0, got %d", got.Attempts)
	}
	if got.ErrorMessage != nil {
		t.Errorf("error_message should be NULL after Retry, got %q", *got.ErrorMessage)
	}
	if got.LastAttemptAt != nil {
		t.Errorf("last_attempt_at should be NULL after Retry, got %v", got.LastAttemptAt)
	}

	// Second Retry on now-pending row → "not found or not in dead state" error.
	err = repo.Retry(ctx, d.ID)
	if err == nil {
		t.Fatal("expected error retrying non-dead delivery, got nil")
	}
	if !strings.Contains(err.Error(), "not found or not in dead state") {
		t.Errorf("unexpected error message: %v", err)
	}
}
