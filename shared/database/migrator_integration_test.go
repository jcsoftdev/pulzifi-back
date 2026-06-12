//go:build integration

package database_test

import (
	"database/sql"
	"fmt"
	"math/rand"
	"testing"

	"github.com/jcsoftdev/pulzifi-back/shared/database"
)

// TestDeprovisionTenantSchema_RoundTrip provisions a throwaway schema, asserts it
// exists in information_schema.schemata, deprovisions it, and asserts it is gone.
// Requires DATABASE_URL to be set and a live PostgreSQL instance.
func TestDeprovisionTenantSchema_RoundTrip(t *testing.T) {
	db := openTestDB(t)

	// Use a random suffix to avoid clashes across parallel runs.
	schema := fmt.Sprintf("testdeprovision_%d", rand.Intn(1_000_000))

	// Safety cleanup in case an earlier run left the schema behind.
	dropTestSchema(t, db, schema)

	// Step 1: Provision the schema (creates it + runs tenant migrations).
	if err := database.ProvisionTenantSchema(db, schema); err != nil {
		t.Fatalf("ProvisionTenantSchema(%q): %v", schema, err)
	}

	// Step 2: Assert the schema exists.
	if !schemaExists(t, db, schema) {
		t.Fatalf("schema %q not found in information_schema.schemata after provisioning", schema)
	}

	// Step 3: Deprovision the schema.
	if err := database.DeprovisionTenantSchema(db, schema); err != nil {
		t.Fatalf("DeprovisionTenantSchema(%q): %v", schema, err)
	}

	// Step 4: Assert the schema is gone.
	if schemaExists(t, db, schema) {
		t.Errorf("schema %q still exists in information_schema.schemata after deprovisioning", schema)
	}
}

// TestDeprovisionTenantSchema_IdempotentOnAbsentSchema verifies that calling
// DeprovisionTenantSchema on a schema that does not exist returns nil (IF EXISTS).
func TestDeprovisionTenantSchema_IdempotentOnAbsentSchema(t *testing.T) {
	db := openTestDB(t)

	// Use a schema name that is extremely unlikely to exist.
	schema := fmt.Sprintf("testdeprovision_absent_%d", rand.Intn(1_000_000))

	// Should not error — DROP SCHEMA IF EXISTS is idempotent.
	if err := database.DeprovisionTenantSchema(db, schema); err != nil {
		t.Errorf("DeprovisionTenantSchema on absent schema: unexpected error: %v", err)
	}
}

// schemaExists reports whether the named schema appears in information_schema.schemata.
func schemaExists(t *testing.T, db *sql.DB, schema string) bool {
	t.Helper()
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM information_schema.schemata WHERE schema_name = $1
		)`, schema).Scan(&exists)
	if err != nil {
		t.Fatalf("schemaExists query: %v", err)
	}
	return exists
}
