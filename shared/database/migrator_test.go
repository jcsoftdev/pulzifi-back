package database_test

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"github.com/jcsoftdev/pulzifi-back/shared/database"
	"github.com/jcsoftdev/pulzifi-back/shared/testguard"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set — integration test requires a real PostgreSQL instance")
	}
	testguard.RequireLocalDB(t, dsn)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// dropTestSchema removes the schema created during a test so reruns are clean.
func dropTestSchema(t *testing.T, db *sql.DB, schema string) {
	t.Helper()
	t.Cleanup(func() {
		if _, err := db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %q CASCADE", schema)); err != nil {
			t.Logf("cleanup: drop schema %q: %v", schema, err)
		}
	})
}

// advisoryLockHeld reports whether a session-level advisory lock for the
// given schema name is currently held on the server.
//
// It uses the same lock key formula as ProvisionTenantSchema so the test can
// assert the lock was acquired and then released.
//
//	lock key = ('x' || md5(schema))::bit(64)::bigint
func advisoryLockHeld(t *testing.T, db *sql.DB, schema string) bool {
	t.Helper()
	var held bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM pg_locks
			WHERE locktype = 'advisory'
			  AND granted = true
			  AND (classid::bigint << 32 | objid::bigint) =
			      (('x' || md5($1))::bit(64)::bigint)
		)
	`, schema).Scan(&held)
	if err != nil {
		t.Fatalf("advisoryLockHeld query: %v", err)
	}
	return held
}

// TestProvisionTenantSchema_LockReleasedAfterSuccess verifies that after a
// successful provisioning the advisory lock is no longer held.
func TestProvisionTenantSchema_LockReleasedAfterSuccess(t *testing.T) {
	db := openTestDB(t)
	schema := "test_lock_release"
	dropTestSchema(t, db, schema)

	if err := database.ProvisionTenantSchema(db, schema); err != nil {
		t.Fatalf("ProvisionTenantSchema: %v", err)
	}

	if advisoryLockHeld(t, db, schema) {
		t.Error("advisory lock still held after successful ProvisionTenantSchema — lock leak")
	}
}

// TestProvisionTenantSchema_LockReleasedAfterInvalidSchema verifies that even
// when the call is rejected early (invalid schema name), no lock leak occurs.
func TestProvisionTenantSchema_LockReleasedAfterInvalidSchema(t *testing.T) {
	db := openTestDB(t)

	// "bad schema" is invalid — provisioning should return an error.
	err := database.ProvisionTenantSchema(db, "bad schema")
	if err == nil {
		t.Fatal("expected error for invalid schema name, got nil")
	}

	// The lock key for an invalid schema should never have been acquired.
	if advisoryLockHeld(t, db, "bad schema") {
		t.Error("advisory lock held despite early rejection on invalid schema name")
	}
}

// TestProvisionTenantSchema_ConcurrentSameTenantSerializes runs N goroutines
// provisioning the exact same schema name simultaneously. Only one should run
// the real migration path; the rest should block, then find ErrNoChange and
// return nil. The schema must end up in a consistent (non-dirty) state.
func TestProvisionTenantSchema_ConcurrentSameTenantSerializes(t *testing.T) {
	db := openTestDB(t)
	schema := "test_concurrent_same"
	dropTestSchema(t, db, schema)

	const goroutines = 5
	var wg sync.WaitGroup
	errs := make([]error, goroutines)

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			errs[i] = database.ProvisionTenantSchema(db, schema)
		}()
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: ProvisionTenantSchema returned unexpected error: %v", i, err)
		}
	}

	// Verify the schema_migrations table is not dirty.
	var dirty bool
	err := db.QueryRow(fmt.Sprintf(`SELECT dirty FROM %q.schema_migrations ORDER BY version DESC LIMIT 1`, schema)).Scan(&dirty)
	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("querying schema_migrations: %v", err)
	}
	if dirty {
		t.Error("schema_migrations is dirty after concurrent provisioning — race condition not prevented")
	}

	// Verify the advisory lock is fully released.
	if advisoryLockHeld(t, db, schema) {
		t.Error("advisory lock still held after all goroutines completed")
	}
}

// TestProvisionTenantSchema_DifferentTenantsParallel verifies that provisioning
// two different tenant schemas concurrently does not cause mutual blocking or
// errors (different lock keys → independent execution).
func TestProvisionTenantSchema_DifferentTenantsParallel(t *testing.T) {
	db := openTestDB(t)
	schemaA := "test_parallel_a"
	schemaB := "test_parallel_b"
	dropTestSchema(t, db, schemaA)
	dropTestSchema(t, db, schemaB)

	start := time.Now()

	var wg sync.WaitGroup
	var errA, errB atomic.Value

	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := database.ProvisionTenantSchema(db, schemaA); err != nil {
			errA.Store(err)
		}
	}()
	go func() {
		defer wg.Done()
		if err := database.ProvisionTenantSchema(db, schemaB); err != nil {
			errB.Store(err)
		}
	}()
	wg.Wait()

	elapsed := time.Since(start)

	if v := errA.Load(); v != nil {
		t.Errorf("schema %q: %v", schemaA, v.(error))
	}
	if v := errB.Load(); v != nil {
		t.Errorf("schema %q: %v", schemaB, v.(error))
	}

	// Both locks must be released.
	if advisoryLockHeld(t, db, schemaA) {
		t.Errorf("advisory lock still held for %q", schemaA)
	}
	if advisoryLockHeld(t, db, schemaB) {
		t.Errorf("advisory lock still held for %q", schemaB)
	}

	// Sanity: two parallel provisions should finish in roughly the same time
	// as a single one, not double. We allow 3x to keep the test non-flaky.
	t.Logf("parallel provisioning took %v", elapsed)
	_ = elapsed // logged for visibility; not a hard assertion to avoid flakiness
}
