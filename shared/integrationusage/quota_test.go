package integrationusage_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/jcsoftdev/pulzifi-back/shared/integrationusage"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

// TestMain sweeps leftover orgs/users from prior interrupted runs before any
// test executes. Belt-and-suspenders: each test also registers t.Cleanup, but
// that doesn't run when the suite is killed mid-run.
func TestMain(m *testing.M) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn != "" {
		if db, err := sql.Open("postgres", dsn); err == nil {
			_, _ = db.Exec(`DELETE FROM organizations WHERE subdomain LIKE 'iq-test-%'`)
			_, _ = db.Exec(`DELETE FROM users WHERE email LIKE 'iq-test-%@example.com'`)
			db.Close()
		}
	}
	os.Exit(m.Run())
}

// seedOrg inserts a fresh user + org and registers cleanup.
// Cleanup is registered BEFORE the inserts so an interrupted test
// (panic, Ctrl+C, timeout) still triggers the deferred deletion path
// where possible — and the rows are no-ops if nothing was inserted.
// The DELETE on organizations cascades to integrations, integration_send_quotas,
// org_data_keys, organization_members, and organization_plans via FK ON DELETE CASCADE.
func seedOrg(t *testing.T, db *sql.DB) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	orgID := uuid.New()
	sub := "iq-test-" + orgID.String()[:8]
	email := "iq-test-" + orgID.String()[:8] + "@example.com"

	t.Cleanup(func() {
		if _, err := db.Exec(`DELETE FROM organizations WHERE id = $1`, orgID); err != nil {
			t.Logf("cleanup: DELETE organizations %s: %v", orgID, err)
		}
		if _, err := db.Exec(`DELETE FROM users WHERE id = $1`, userID); err != nil {
			t.Logf("cleanup: DELETE users %s: %v", userID, err)
		}
	})

	if _, err := db.Exec(`INSERT INTO users (id, email, password_hash, created_at, updated_at)
                       VALUES ($1::uuid, $2, 'x', NOW(), NOW())`,
		userID, email); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Exec(`INSERT INTO organizations (id, name, subdomain, schema_name, owner_user_id, feature_flags, created_at, updated_at)
                       VALUES ($1::uuid, $2, $3, $3, $4::uuid, '{}'::jsonb, NOW(), NOW())`,
		orgID, orgID.String(), sub, userID); err != nil {
		t.Fatal(err)
	}

	return orgID
}

func fixedAllowed(n int) integrationusage.AllowedFunc {
	return func(_ context.Context, _ uuid.UUID) (int, error) {
		return n, nil
	}
}

// TestTracker_FirstSendCreatesRow — call once, verify row exists with count_used=1, count_allowed=5.
func TestTracker_FirstSendCreatesRow(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	orgID := seedOrg(t, db)
	tracker := integrationusage.NewTracker(db)

	err := tracker.CheckAndIncrement(context.Background(), orgID, "slack", fixedAllowed(5))
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	var countUsed, countAllowed int
	err = db.QueryRow(`
		SELECT count_used, count_allowed FROM integration_send_quotas
		WHERE org_id = $1 AND service_type = 'slack'
	`, orgID).Scan(&countUsed, &countAllowed)
	if err != nil {
		t.Fatalf("failed to query row: %v", err)
	}
	if countUsed != 1 {
		t.Errorf("expected count_used=1, got %d", countUsed)
	}
	if countAllowed != 5 {
		t.Errorf("expected count_allowed=5, got %d", countAllowed)
	}
}

// TestTracker_IncrementsUntilLimit — call 3 times with allowedFor(3). All 3 succeed. Final count_used=3.
func TestTracker_IncrementsUntilLimit(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	orgID := seedOrg(t, db)
	tracker := integrationusage.NewTracker(db)

	for i := 0; i < 3; i++ {
		err := tracker.CheckAndIncrement(context.Background(), orgID, "teams", fixedAllowed(3))
		if err != nil {
			t.Fatalf("call %d: expected nil error, got: %v", i+1, err)
		}
	}

	var countUsed int
	err := db.QueryRow(`
		SELECT count_used FROM integration_send_quotas
		WHERE org_id = $1 AND service_type = 'teams'
	`, orgID).Scan(&countUsed)
	if err != nil {
		t.Fatalf("failed to query row: %v", err)
	}
	if countUsed != 3 {
		t.Errorf("expected count_used=3, got %d", countUsed)
	}
}

// TestTracker_ReturnsExceededWhenAtLimit — call 3 times then 4th. 4th returns ErrQuotaExceeded.
// count_used stays at 3 (didn't increment past allowed).
func TestTracker_ReturnsExceededWhenAtLimit(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	orgID := seedOrg(t, db)
	tracker := integrationusage.NewTracker(db)

	for i := 0; i < 3; i++ {
		err := tracker.CheckAndIncrement(context.Background(), orgID, "discord", fixedAllowed(3))
		if err != nil {
			t.Fatalf("call %d: expected nil error, got: %v", i+1, err)
		}
	}

	// 4th call should be rejected
	err := tracker.CheckAndIncrement(context.Background(), orgID, "discord", fixedAllowed(3))
	if !errors.Is(err, integrationusage.ErrQuotaExceeded) {
		t.Fatalf("expected ErrQuotaExceeded, got: %v", err)
	}

	var countUsed int
	err = db.QueryRow(`
		SELECT count_used FROM integration_send_quotas
		WHERE org_id = $1 AND service_type = 'discord'
	`, orgID).Scan(&countUsed)
	if err != nil {
		t.Fatalf("failed to query row: %v", err)
	}
	if countUsed != 3 {
		t.Errorf("expected count_used to remain 3, got %d", countUsed)
	}
}

// TestTracker_ZeroAllowedExceedsImmediately — allowedFor(0) → first call returns ErrQuotaExceeded
// without inserting any row.
func TestTracker_ZeroAllowedExceedsImmediately(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	orgID := seedOrg(t, db)
	tracker := integrationusage.NewTracker(db)

	err := tracker.CheckAndIncrement(context.Background(), orgID, "twilio", fixedAllowed(0))
	if !errors.Is(err, integrationusage.ErrQuotaExceeded) {
		t.Fatalf("expected ErrQuotaExceeded, got: %v", err)
	}

	var count int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM integration_send_quotas WHERE org_id = $1
	`, orgID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count rows: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 rows for org, got %d", count)
	}
}

// TestTracker_ConcurrentCallsRaceSafe — 50 goroutines call with allowedFor(20).
// Exactly 20 return nil; exactly 30 return ErrQuotaExceeded. Verify final count_used == 20.
func TestTracker_ConcurrentCallsRaceSafe(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	orgID := seedOrg(t, db)
	tracker := integrationusage.NewTracker(db)

	const total = 50
	const limit = 20

	var wg sync.WaitGroup
	var successes atomic.Int32
	var exceeded atomic.Int32

	wg.Add(total)
	for i := 0; i < total; i++ {
		go func() {
			defer wg.Done()
			err := tracker.CheckAndIncrement(context.Background(), orgID, "webhook", fixedAllowed(limit))
			switch {
			case err == nil:
				successes.Add(1)
			case errors.Is(err, integrationusage.ErrQuotaExceeded):
				exceeded.Add(1)
			default:
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	if successes.Load() != limit {
		t.Errorf("expected %d successes, got %d", limit, successes.Load())
	}
	if exceeded.Load() != total-limit {
		t.Errorf("expected %d exceeded, got %d", total-limit, exceeded.Load())
	}

	var countUsed int
	err := db.QueryRow(`
		SELECT count_used FROM integration_send_quotas
		WHERE org_id = $1 AND service_type = 'webhook'
	`, orgID).Scan(&countUsed)
	if err != nil {
		t.Fatalf("failed to query row: %v", err)
	}
	if countUsed != limit {
		t.Errorf("expected count_used=%d, got %d", limit, countUsed)
	}
}

// TestTracker_NewMonthCreatesFreshRow — old exhausted row doesn't block new month.
func TestTracker_NewMonthCreatesFreshRow(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	orgID := seedOrg(t, db)
	tracker := integrationusage.NewTracker(db)

	// Insert an exhausted row for an old period
	oldPeriod := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := db.Exec(`
		INSERT INTO integration_send_quotas (org_id, service_type, period_start, count_allowed, count_used)
		VALUES ($1, 'email', $2, 5, 5)
	`, orgID, oldPeriod)
	if err != nil {
		t.Fatalf("failed to insert old-period row: %v", err)
	}

	// Call should INSERT a new row for the current month
	err = tracker.CheckAndIncrement(context.Background(), orgID, "email", fixedAllowed(10))
	if err != nil {
		t.Fatalf("expected nil error for current month, got: %v", err)
	}

	// Verify 2 rows exist for this org+service_type
	var rowCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM integration_send_quotas
		WHERE org_id = $1 AND service_type = 'email'
	`, orgID).Scan(&rowCount)
	if err != nil {
		t.Fatalf("failed to count rows: %v", err)
	}
	if rowCount != 2 {
		t.Errorf("expected 2 rows (old + new month), got %d", rowCount)
	}

	// Verify the new month row has count_used=1 and count_allowed=10
	currentPeriod := time.Date(time.Now().UTC().Year(), time.Now().UTC().Month(), 1, 0, 0, 0, 0, time.UTC)
	var countUsed, countAllowed int
	err = db.QueryRow(`
		SELECT count_used, count_allowed FROM integration_send_quotas
		WHERE org_id = $1 AND service_type = 'email' AND period_start = $2
	`, orgID, currentPeriod).Scan(&countUsed, &countAllowed)
	if err != nil {
		t.Fatalf("failed to query current period row: %v", err)
	}
	if countUsed != 1 {
		t.Errorf("expected count_used=1 for current month, got %d", countUsed)
	}
	if countAllowed != 10 {
		t.Errorf("expected count_allowed=10 for current month, got %d", countAllowed)
	}

	// Verify old row is untouched
	var oldCountUsed int
	err = db.QueryRow(`
		SELECT count_used FROM integration_send_quotas
		WHERE org_id = $1 AND service_type = 'email' AND period_start = $2
	`, orgID, oldPeriod).Scan(&oldCountUsed)
	if err != nil {
		t.Fatalf("failed to query old period row: %v", err)
	}
	if oldCountUsed != 5 {
		t.Errorf("expected old row count_used=5 (unchanged), got %d", oldCountUsed)
	}
}
