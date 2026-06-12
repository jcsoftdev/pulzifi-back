package orgwiring

import (
	"context"
	"database/sql"

	orgservices "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
)

// schemaDropperAdapter wraps shared/database.DeprovisionTenantSchema behind
// the org domain's SchemaDropper port. The use case already wraps this call
// in context.WithTimeout before invoking it; the adapter itself is stateless.
type schemaDropperAdapter struct {
	db *sql.DB
}

// compile-time assertion
var _ orgservices.SchemaDropper = (*schemaDropperAdapter)(nil)

// NewSchemaDropperAdapter builds a SchemaDropper backed by the shared DB pool.
func NewSchemaDropperAdapter(db *sql.DB) orgservices.SchemaDropper {
	return &schemaDropperAdapter{db: db}
}

// DropTenantSchema validates the schema name and executes DROP SCHEMA CASCADE.
//
// NOTE: the ctx parameter is intentionally ignored (OQ-3). DeprovisionTenantSchema
// uses context.Background() internally, so the drop is NOT bounded by the 30-second
// budget the handler establishes via context.WithTimeout. The operation is bounded
// only by the PostgreSQL statement timeout (lock_timeout / statement_timeout) and the
// server-level idle-in-transaction timeout. Callers should be aware that a hung DROP
// will outlive the request context and must be terminated via PostgreSQL tooling.
func (a *schemaDropperAdapter) DropTenantSchema(_ context.Context, schema string) error {
	return database.DeprovisionTenantSchema(a.db, schema)
}
