package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

const migrationsBaseDir = "shared/database/migrations"

var validSchemaName = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// ErrInvalidSchemaName is returned by DeprovisionTenantSchema when the supplied
// schema name does not match the allowed pattern ^[a-zA-Z_][a-zA-Z0-9_]*$.
var ErrInvalidSchemaName = errors.New("invalid schema name")

// ProvisionTenantSchema ensures the tenant schema exists and runs all pending tenant migrations.
// It is safe to call multiple times (idempotent).
func ProvisionTenantSchema(db *sql.DB, schemaName string) error {
	if !validSchemaName.MatchString(schemaName) {
		return fmt.Errorf("invalid schema name: %q", schemaName)
	}

	ctx := context.Background()

	// Acquire a dedicated connection: the advisory lock, CREATE SCHEMA, search_path,
	// and migrations must all run on the SAME backend connection so the lock actually
	// serializes the provisioning sequence (a pool may otherwise scatter these across
	// connections).
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db connection: %w", err)
	}
	defer func() { _ = conn.Close() }()

	// Serialize concurrent provisioning of the SAME schema. Two callers racing to
	// create + migrate the same tenant would otherwise interleave CREATE SCHEMA and
	// migration steps, leaving a dirty/partial schema_migrations state. hashtext maps
	// the schema name to an int key, so different tenants take different locks and
	// still provision in parallel. Session-level lock (not xact) because golang-migrate
	// manages its own per-migration transactions on this conn; it auto-releases when the
	// pinned conn is closed, and we release explicitly below.
	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock(hashtext($1))", schemaName); err != nil {
		return fmt.Errorf("failed to acquire provisioning lock for tenant %q: %w", schemaName, err)
	}
	defer func() {
		if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_unlock(hashtext($1))", schemaName); err != nil {
			logger.Warn("failed to release tenant provisioning lock", zap.String("schema", schemaName), zap.Error(err))
		}
	}()

	// Create schema if not already created (PG trigger may have done this, but ensure it)
	if _, err := conn.ExecContext(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", pq.QuoteIdentifier(schemaName))); err != nil {
		return fmt.Errorf("failed to create schema %q: %w", schemaName, err)
	}

	if _, err := conn.ExecContext(ctx,
		fmt.Sprintf("SET search_path TO %s, public", pq.QuoteIdentifier(schemaName)),
	); err != nil {
		return fmt.Errorf("failed to set search_path for tenant %q: %w", schemaName, err)
	}

	driver, err := postgres.WithConnection(context.Background(), conn, &postgres.Config{
		SchemaName:      schemaName,
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		return fmt.Errorf("failed to create migration driver for tenant %q: %w", schemaName, err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	sourceURL := fmt.Sprintf("file://%s/%s/tenant", cwd, migrationsBaseDir)

	m, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance for tenant %q: %w", schemaName, err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("tenant migrations failed for %q: %w", schemaName, err)
	}

	logger.Info("Tenant schema provisioned", zap.String("schema", schemaName))
	return nil
}

// DeprovisionTenantSchema validates the schema name, terminates any lingering
// backends associated with that schema (best-effort, non-fatal per OQ-2), then
// executes DROP SCHEMA IF EXISTS … CASCADE. Idempotent: dropping an absent schema
// is a no-op due to IF EXISTS. This is the terminal, irreversible step of org
// deletion — callers must have soft-deleted the org (releasing workers) and
// hard-deleted public rows before calling this function.
//
// Signature mirrors ProvisionTenantSchema: no ctx param (OQ-3); uses
// context.Background() internally. Timeout belongs in the SchemaDropper adapter.
func DeprovisionTenantSchema(db *sql.DB, schemaName string) error {
	if !validSchemaName.MatchString(schemaName) {
		return fmt.Errorf("%w: %q", ErrInvalidSchemaName, schemaName)
	}

	ctx := context.Background()

	// Terminate backends whose last statement pinned search_path to this schema
	// (the pattern every tenant-scoped repo emits via WithTenant/GetSetSearchPathSQL).
	// Heuristic: pg_stat_activity cannot tell which schema a backend is "in", so this
	// narrows on the SET statement instead of any query mentioning the name, avoiding
	// killing unrelated backends. Best-effort: ignore errors (OQ-2); the soft-delete +
	// worker-release window is the primary protection.
	_, _ = db.ExecContext(ctx, `
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE pid <> pg_backend_pid()
		  AND datname = current_database()
		  AND query ILIKE 'SET%search_path%' || $1 || '%'`, schemaName)

	// Drop the schema and all objects inside it. The regex gate above is the
	// authoritative injection guard; pq.QuoteIdentifier is defense-in-depth in
	// case the validation is ever loosened.
	if _, err := db.ExecContext(ctx,
		fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", pq.QuoteIdentifier(schemaName)),
	); err != nil {
		return fmt.Errorf("failed to drop schema %q: %w", schemaName, err)
	}

	logger.Info("Tenant schema deprovisioned", zap.String("schema", schemaName))
	return nil
}
