package database

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
)

// validTenantSchema matches only safe identifier characters. It mirrors the
// validation used by middleware.GetSetSearchPathSQL so WithTenant accepts every
// tenant name that the rest of the system already routes through.
var validTenantSchema = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*$`)

// WithTenant runs fn inside a single transaction whose search_path is pinned to
// the tenant schema via SET LOCAL.
//
// Why this exists: every tenant-scoped repository historically issued TWO
// decoupled pool round-trips — `SET search_path` (session-scoped) followed by a
// SEPARATE query. Under PgBouncer transaction-pooling those two statements can
// land on DIFFERENT backend connections, so the query runs against the wrong
// schema (or only `public`) → silent cross-tenant exposure. Binding both the
// SET LOCAL and the query to ONE transaction guarantees they share a backend
// connection. SET LOCAL is scoped to the transaction and auto-reverts on
// COMMIT/ROLLBACK, so nothing leaks onto a pooled connection afterwards.
//
// SET LOCAL is valid on direct (non-pooled) connections too, so call sites
// converted to WithTenant are safe to deploy BEFORE PgBouncer is introduced.
//
// fn MUST perform all of its work — including scanning any rows — against tx,
// because the transaction is committed (and its rows closed) as soon as fn
// returns. Returning an error from fn rolls the transaction back.
func WithTenant(ctx context.Context, db *sql.DB, tenant string, fn func(tx *sql.Tx) error) error {
	if !validTenantSchema.MatchString(tenant) {
		return fmt.Errorf("invalid tenant schema name: %q", tenant)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tenant tx for %q: %w", tenant, err)
	}

	// SET LOCAL cannot be parameterized; the tenant name is validated above and
	// quoted so it is safe to interpolate.
	if _, err := tx.ExecContext(ctx, `SET LOCAL search_path TO "`+tenant+`", public`); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("set local search_path for %q: %w", tenant, err)
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tenant tx for %q: %w", tenant, err)
	}
	return nil
}
