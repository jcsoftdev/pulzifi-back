package authwiring_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	authwiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/auth"
	"github.com/jcsoftdev/pulzifi-back/shared/testguard"
	_ "github.com/lib/pq"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func membershipTestDSN() string {
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

// openMembershipTestDB opens a connection for the membership checker
// integration test. The test seeds and deletes real rows in public.users and
// public.organizations, so it must never run against a shared/production
// database picked up from ambient env vars (this has happened: leftover
// membershiptest-* rows in prod broke the admin users page).
func openMembershipTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := membershipTestDSN()
	if dsn == "" {
		t.Skip("DATABASE_URL (or DB_HOST) not set; skipping membership checker integration test")
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

// seedMembershipUser inserts a minimal user row and registers cleanup.
// Returns the userID.
func seedMembershipUser(t *testing.T, db *sql.DB, ctx context.Context) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM public.users WHERE id = $1`, userID)
	})
	_, err := db.ExecContext(ctx, `
		INSERT INTO public.users (id, email, password_hash, status)
		VALUES ($1, $2, 'testhash', 'approved')
		ON CONFLICT DO NOTHING`,
		userID, fmt.Sprintf("membershiptest-%s@example.com", userID),
	)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return userID
}

// seedOrganization inserts a minimal organization row and registers cleanup.
// Returns the orgID.
func seedOrganization(t *testing.T, db *sql.DB, ctx context.Context, ownerUserID uuid.UUID) uuid.UUID {
	t.Helper()
	orgID := uuid.New()
	suffix := orgID.String()[:8]
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM public.organization_members WHERE organization_id = $1`, orgID)
		_, _ = db.ExecContext(ctx, `DELETE FROM public.organizations WHERE id = $1`, orgID)
	})
	_, err := db.ExecContext(ctx, `
		INSERT INTO public.organizations (id, name, subdomain, schema_name, owner_user_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT DO NOTHING`,
		orgID,
		"MembershipTest Org "+suffix,
		"membershiptest-"+suffix,
		"membershiptest_"+suffix,
		ownerUserID,
	)
	if err != nil {
		t.Fatalf("seed org: %v", err)
	}
	return orgID
}

// seedMember inserts an organization_members row. deletedAt = true inserts
// a soft-deleted row (deleted_at = NOW()); false inserts an active row.
func seedMember(t *testing.T, db *sql.DB, ctx context.Context, orgID, userID uuid.UUID, softDeleted bool) {
	t.Helper()
	memberID := uuid.New()
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM public.organization_members WHERE id = $1`, memberID)
	})
	var err error
	if softDeleted {
		_, err = db.ExecContext(ctx, `
			INSERT INTO public.organization_members (id, organization_id, user_id, role, joined_at, deleted_at)
			VALUES ($1, $2, $3, 'member', NOW(), NOW())`,
			memberID, orgID, userID,
		)
	} else {
		_, err = db.ExecContext(ctx, `
			INSERT INTO public.organization_members (id, organization_id, user_id, role, joined_at)
			VALUES ($1, $2, $3, 'member', NOW())`,
			memberID, orgID, userID,
		)
	}
	if err != nil {
		t.Fatalf("seed member (softDeleted=%v): %v", softDeleted, err)
	}
}

// ── Integration tests ──────────────────────────────────────────────────────────

// TestMembershipChecker_HasAnyMembership_True verifies that a user with an
// active membership row returns true.
func TestMembershipChecker_HasAnyMembership_True(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	db := openMembershipTestDB(t)
	defer db.Close()

	ctx := context.Background()
	userID := seedMembershipUser(t, db, ctx)
	orgID := seedOrganization(t, db, ctx, userID)
	seedMember(t, db, ctx, orgID, userID, false)

	checker := authwiring.NewMembershipChecker(db)
	got, err := checker.HasAnyMembership(ctx, userID)
	if err != nil {
		t.Fatalf("HasAnyMembership: unexpected error: %v", err)
	}
	if !got {
		t.Error("HasAnyMembership: want true for user with active membership, got false")
	}
}

// TestMembershipChecker_HasAnyMembership_False verifies that a user with no
// membership rows at all returns false.
func TestMembershipChecker_HasAnyMembership_False(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	db := openMembershipTestDB(t)
	defer db.Close()

	ctx := context.Background()
	userID := seedMembershipUser(t, db, ctx)
	// No org, no member row seeded.

	checker := authwiring.NewMembershipChecker(db)
	got, err := checker.HasAnyMembership(ctx, userID)
	if err != nil {
		t.Fatalf("HasAnyMembership: unexpected error: %v", err)
	}
	if got {
		t.Error("HasAnyMembership: want false for user with no membership, got true")
	}
}

// TestMembershipChecker_HasAnyMembership_SoftDeleted verifies that a user
// whose only membership row has deleted_at set is treated as having no membership.
func TestMembershipChecker_HasAnyMembership_SoftDeleted(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	db := openMembershipTestDB(t)
	defer db.Close()

	ctx := context.Background()
	userID := seedMembershipUser(t, db, ctx)
	orgID := seedOrganization(t, db, ctx, userID)
	seedMember(t, db, ctx, orgID, userID, true) // soft-deleted

	checker := authwiring.NewMembershipChecker(db)
	got, err := checker.HasAnyMembership(ctx, userID)
	if err != nil {
		t.Fatalf("HasAnyMembership: unexpected error: %v", err)
	}
	if got {
		t.Error("HasAnyMembership: want false for user with only soft-deleted membership, got true")
	}
}

// TestMembershipChecker_HasAnyMembership_DBError verifies that a query error
// is propagated correctly. This test uses a closed DB to trigger an error
// without needing a real connection.
func TestMembershipChecker_HasAnyMembership_DBError(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://invalid:invalid@localhost:1/nodb?sslmode=disable")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	// Close immediately so all queries fail.
	db.Close()

	checker := authwiring.NewMembershipChecker(db)
	_, gotErr := checker.HasAnyMembership(context.Background(), uuid.New())
	if gotErr == nil {
		t.Fatal("HasAnyMembership: expected error from closed DB, got nil")
	}
	if errors.Is(gotErr, nil) {
		t.Fatal("HasAnyMembership: error should not be nil")
	}
}
