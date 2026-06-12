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
// Context is passed but DeprovisionTenantSchema uses context.Background() internally
// (OQ-3 design decision); the use case wraps the call with context.WithTimeout.
func (a *schemaDropperAdapter) DropTenantSchema(_ context.Context, schema string) error {
	return database.DeprovisionTenantSchema(a.db, schema)
}
