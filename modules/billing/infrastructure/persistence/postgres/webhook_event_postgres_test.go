//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/infrastructure/persistence/postgres"
	"github.com/jcsoftdev/pulzifi-back/shared/testguard"
)

// TestMain cleans up any leftover test rows from interrupted runs.
//
// testguard.RequireLocalDB is not called here because TestMain has no
// *testing.T to call t.Skip on — it can only os.Exit. Instead this mirrors
// testguard's local-DSN policy inline before running the sweep. Every actual
// test in this file also calls testguard.RequireLocalDB via
// openBillingTestDB before touching the database, so this is belt-and-braces
// for the pre-test cleanup sweep only.
func TestMain(m *testing.M) {
	dsn := billingTestDSN()
	if dsn != "" && isLocalOrAllowedRemoteDSN(dsn) {
		if db, err := sql.Open("postgres", dsn); err == nil {
			_, _ = db.Exec(`DELETE FROM public.stripe_webhook_events WHERE event_id LIKE 'evt_test_%'`)
			db.Close()
		}
	}
	os.Exit(m.Run())
}

// isLocalOrAllowedRemoteDSN mirrors testguard's local-DSN policy for use
// outside of a *testing.T context.
func isLocalOrAllowedRemoteDSN(dsn string) bool {
	if os.Getenv("INTEGRATION_DB_ALLOW_REMOTE") == "1" {
		return true
	}
	u, err := url.Parse(dsn)
	if err != nil {
		return false
	}
	switch u.Hostname() {
	case "localhost", "127.0.0.1", "::1", "postgres", "host.docker.internal":
		return true
	}
	return false
}

// openBillingTestDB returns a connected *sql.DB or calls t.Skip when
// DATABASE_URL (or DB_HOST) is not set.  This keeps the integration tests
// opt-in — `go test ./...` (without -tags=integration) never touches Postgres.
func openBillingTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := billingTestDSN()
	if dsn == "" {
		t.Skip("DATABASE_URL (or DB_HOST) not set; skipping billing integration test")
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

func billingTestDSN() string {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}
	host := os.Getenv("DB_HOST")
	if host == "" {
		return ""
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		host, port, os.Getenv("DB_NAME"),
	)
}

// TestWebhookEventPostgres_SaveNew verifies that inserting a new event returns
// (true, nil) and the row can be queried back.
func TestWebhookEventPostgres_SaveNew(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := openBillingTestDB(t)
	defer db.Close()

	repo := postgres.NewWebhookEventPostgresRepository(db)
	ctx := context.Background()

	eventID := fmt.Sprintf("evt_test_%d", time.Now().UnixNano())
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM public.stripe_webhook_events WHERE event_id = $1`, eventID)
	})

	ev := &entities.WebhookEvent{
		EventID:    eventID,
		EventType:  "checkout.session.completed",
		ReceivedAt: time.Now().UTC(),
		Status:     entities.WebhookEventReceived,
	}

	inserted, err := repo.Save(ctx, ev)
	if err != nil {
		t.Fatalf("Save: unexpected error: %v", err)
	}
	if !inserted {
		t.Error("Save: expected true (first insert), got false")
	}

	// Verify the row exists in the DB.
	var count int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM public.stripe_webhook_events WHERE event_id = $1`, eventID,
	).Scan(&count); err != nil {
		t.Fatalf("count query: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 row in stripe_webhook_events, got %d", count)
	}
}

// TestWebhookEventPostgres_SaveDuplicate_ReturnsIdempotent verifies the UNIQUE
// constraint: inserting the same event_id twice returns (false, nil) without
// error — the idempotency contract Stripe relies on.
func TestWebhookEventPostgres_SaveDuplicate_ReturnsIdempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := openBillingTestDB(t)
	defer db.Close()

	repo := postgres.NewWebhookEventPostgresRepository(db)
	ctx := context.Background()

	eventID := fmt.Sprintf("evt_test_dup_%d", time.Now().UnixNano())
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM public.stripe_webhook_events WHERE event_id = $1`, eventID)
	})

	ev := &entities.WebhookEvent{
		EventID:    eventID,
		EventType:  "invoice.paid",
		ReceivedAt: time.Now().UTC(),
		Status:     entities.WebhookEventReceived,
	}

	// First insert — should succeed.
	inserted1, err := repo.Save(ctx, ev)
	if err != nil {
		t.Fatalf("Save (first): unexpected error: %v", err)
	}
	if !inserted1 {
		t.Error("Save (first): expected true, got false")
	}

	// Second insert of identical event_id — should be a no-op.
	inserted2, err := repo.Save(ctx, ev)
	if err != nil {
		t.Fatalf("Save (duplicate): unexpected error: %v", err)
	}
	if inserted2 {
		t.Error("Save (duplicate): expected false (idempotent), got true")
	}

	// Exactly one row should exist.
	var count int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM public.stripe_webhook_events WHERE event_id = $1`, eventID,
	).Scan(&count); err != nil {
		t.Fatalf("count query: %v", err)
	}
	if count != 1 {
		t.Errorf("expected exactly 1 row after duplicate save, got %d", count)
	}
}

// TestWebhookEventPostgres_MarkProcessed verifies that MarkProcessed sets
// processed_at (non-null) and updates the status column.
func TestWebhookEventPostgres_MarkProcessed(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := openBillingTestDB(t)
	defer db.Close()

	repo := postgres.NewWebhookEventPostgresRepository(db)
	ctx := context.Background()

	eventID := fmt.Sprintf("evt_test_mark_%d", time.Now().UnixNano())
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM public.stripe_webhook_events WHERE event_id = $1`, eventID)
	})

	// Save the event first.
	ev := &entities.WebhookEvent{
		EventID:    eventID,
		EventType:  "customer.subscription.updated",
		ReceivedAt: time.Now().UTC(),
		Status:     entities.WebhookEventReceived,
	}
	if _, err := repo.Save(ctx, ev); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Mark as processed.
	if err := repo.MarkProcessed(ctx, eventID, entities.WebhookEventProcessed); err != nil {
		t.Fatalf("MarkProcessed: %v", err)
	}

	// Verify the row was updated correctly.
	var status string
	var processedAt sql.NullTime
	if err := db.QueryRowContext(ctx,
		`SELECT status, processed_at FROM public.stripe_webhook_events WHERE event_id = $1`, eventID,
	).Scan(&status, &processedAt); err != nil {
		t.Fatalf("verify query: %v", err)
	}

	if status != string(entities.WebhookEventProcessed) {
		t.Errorf("status: want %q, got %q", entities.WebhookEventProcessed, status)
	}
	if !processedAt.Valid {
		t.Error("processed_at should be non-null after MarkProcessed")
	}
}
