package billingwiring

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
)

// UsageReader implements services.UsageReader against the tenant's
// usage_tracking table. The tenant schema is resolved via public.organizations.
// The query MUST stay on the tenant schema-qualified table name — never use
// SET search_path here because billing handlers may be invoked while another
// tenant's path is active.
type UsageReader struct {
	db *sql.DB
}

// NewUsageReader constructs the reader backed by the application's *sql.DB.
func NewUsageReader(db *sql.DB) *UsageReader {
	return &UsageReader{db: db}
}

// Compile-time interface check.
var _ services.UsageReader = (*UsageReader)(nil)

// ReadActiveUsage returns the checks_used / checks_allowed pair for the
// usage_tracking row covering today. Empty snapshot (no error) when no active
// row exists — callers must NOT treat that as a failure; it just means there
// is no usage credit to apply.
func (r *UsageReader) ReadActiveUsage(ctx context.Context, orgID uuid.UUID) (services.UsageSnapshot, error) {
	var schemaName string
	err := r.db.QueryRowContext(ctx, `
		SELECT schema_name
		  FROM public.organizations
		 WHERE id = $1
		   AND deleted_at IS NULL
	`, orgID).Scan(&schemaName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.UsageSnapshot{}, nil
		}
		return services.UsageSnapshot{}, fmt.Errorf("usage reader: lookup schema: %w", err)
	}
	if schemaName == "" {
		return services.UsageSnapshot{}, nil
	}

	// Defense-in-depth — schema names are admin-controlled identifiers but we
	// validate here too since we will interpolate it into the query.
	for _, c := range schemaName {
		if !(c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return services.UsageSnapshot{}, fmt.Errorf("usage reader: unsafe schema_name %q", schemaName)
		}
	}

	now := time.Now()
	query := fmt.Sprintf(`
		SELECT COALESCE(checks_used, 0), COALESCE(checks_allowed, 0)
		  FROM %q.usage_tracking
		 WHERE period_start <= $1
		   AND period_end   >= $1
		 ORDER BY period_start DESC
		 LIMIT 1
	`, schemaName)

	var snap services.UsageSnapshot
	if err := r.db.QueryRowContext(ctx, query, now).Scan(&snap.ChecksUsed, &snap.ChecksAllowed); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.UsageSnapshot{}, nil
		}
		return services.UsageSnapshot{}, fmt.Errorf("usage reader: query usage: %w", err)
	}
	return snap, nil
}
