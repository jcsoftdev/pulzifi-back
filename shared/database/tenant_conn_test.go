package database_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	_ "github.com/lib/pq"

	"github.com/jcsoftdev/pulzifi-back/shared/database"
	"github.com/jcsoftdev/pulzifi-back/shared/testguard"
)

func TestWithTenant_RejectsInvalidSchema(t *testing.T) {
	// Validation happens before any DB access, so a nil *sql.DB is fine here:
	// fn must never run and BeginTx must never be reached.
	cases := []string{"", "bad schema", "1leading", "has;semi", "drop--table'"}
	for _, name := range cases {
		called := false
		err := database.WithTenant(context.Background(), nil, name, func(_ *sql.Tx) error {
			called = true
			return nil
		})
		if err == nil {
			t.Errorf("WithTenant(%q) = nil error, want rejection", name)
		}
		if called {
			t.Errorf("WithTenant(%q) ran fn despite invalid schema", name)
		}
	}
}

// Integration tests below require a real PostgreSQL instance.
func openWithTenantTestDB(t *testing.T) *sql.DB {
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

func TestWithTenant_CommitsAndPinsSearchPath(t *testing.T) {
	db := openWithTenantTestDB(t)
	ctx := context.Background()

	var current string
	err := database.WithTenant(ctx, db, "public", func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, "SHOW search_path").Scan(&current)
	})
	if err != nil {
		t.Fatalf("WithTenant: %v", err)
	}
	if current == "" {
		t.Error("search_path was empty inside WithTenant")
	}
}

func TestWithTenant_RollsBackOnFnError(t *testing.T) {
	db := openWithTenantTestDB(t)
	ctx := context.Background()

	sentinel := errors.New("boom")
	if _, err := db.ExecContext(ctx, "CREATE TEMP TABLE IF NOT EXISTS withtenant_rb (n int)"); err != nil {
		t.Fatalf("setup temp table: %v", err)
	}

	err := database.WithTenant(ctx, db, "public", func(tx *sql.Tx) error {
		if _, e := tx.ExecContext(ctx, "INSERT INTO withtenant_rb (n) VALUES (1)"); e != nil {
			return e
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("WithTenant err = %v, want sentinel", err)
	}

	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM withtenant_rb").Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Errorf("rows after rollback = %d, want 0 (transaction not rolled back)", count)
	}
}
