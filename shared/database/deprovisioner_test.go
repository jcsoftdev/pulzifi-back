package database_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/jcsoftdev/pulzifi-back/shared/database"
)

// TestDeprovisionTenantSchema_InvalidNames verifies that DeprovisionTenantSchema
// returns ErrInvalidSchemaName for bad schema names before making any DB calls.
// Passing nil as db proves no DB call happens on invalid input.
func TestDeprovisionTenantSchema_InvalidNames(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		schemaName string
	}{
		{"sql injection", "'; DROP TABLE users; --"},
		{"empty string", ""},
		{"starts with digit", "123bad"},
		{"contains space", "bad schema"},
		{"contains hyphen", "bad-schema"},
		{"dot-path attempt", "public.users"},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// nil db: if validation fires before any DB call, no panic occurs.
			err := database.DeprovisionTenantSchema(nil, tt.schemaName)

			if !errors.Is(err, database.ErrInvalidSchemaName) {
				t.Errorf("DeprovisionTenantSchema(%q): want ErrInvalidSchemaName, got %v", tt.schemaName, err)
			}
		})
	}
}

// TestDeprovisionTenantSchema_ValidNamesPassValidation verifies that valid schema
// names pass the regex gate. The db points at an unreachable address (lib/pq
// connects lazily), so the function proceeds past validation and fails with a
// connection error on the DROP step — proving the only rejection path exercised
// is the regex, never a panic.
func TestDeprovisionTenantSchema_ValidNamesPassValidation(t *testing.T) {
	t.Parallel()

	db, err := sql.Open("postgres", "host=127.0.0.1 port=1 user=x dbname=x sslmode=disable connect_timeout=1")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	cases := []struct {
		name       string
		schemaName string
	}{
		{"simple lowercase", "valid_name"},
		{"uppercase", "ValidName"},
		{"leading underscore", "_private"},
		{"alphanumeric", "tenant123"},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := database.DeprovisionTenantSchema(db, tt.schemaName)
			if err == nil {
				t.Fatalf("DeprovisionTenantSchema(%q): expected connection error, got nil", tt.schemaName)
			}
			if errors.Is(err, database.ErrInvalidSchemaName) {
				t.Errorf("DeprovisionTenantSchema(%q): valid name rejected by validation", tt.schemaName)
			}
		})
	}
}
