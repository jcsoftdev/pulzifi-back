package database_test

import (
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
// names pass the regex check. We cannot use nil db here because valid names proceed
// to make DB calls; instead we verify by wrapping in recover — a panic from a nil
// db confirms the validation gate was passed (the code reached DB call logic).
func TestDeprovisionTenantSchema_ValidNamesPassValidation(t *testing.T) {
	t.Parallel()

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

			// We expect a panic (nil pointer on the db) because the schema name is valid
			// and the code proceeds past the validation gate to make a DB call.
			// A panic proves the validation passed (no ErrInvalidSchemaName was returned).
			didPanic := func() (panicked bool) {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
					}
				}()
				err := database.DeprovisionTenantSchema(nil, tt.schemaName)
				// If we reach here without panic, it must not be ErrInvalidSchemaName.
				if errors.Is(err, database.ErrInvalidSchemaName) {
					t.Errorf("DeprovisionTenantSchema(%q): unexpected ErrInvalidSchemaName", tt.schemaName)
				}
				return false
			}()

			// Either we got a panic (nil db → DB call reached) or a non-validation error.
			// Both confirm the schema name passed validation.
			_ = didPanic // either path is acceptable; the test goal is no ErrInvalidSchemaName
		})
	}
}
