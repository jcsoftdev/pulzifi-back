package socialwiring

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
)

// orgLookup is a shared helper that resolves an organisation UUID from a tenant
// schema name. It is used by both the AlertCreator and PlanLimits adapters.
type orgLookup struct{ db *sql.DB }

// newOrgLookup creates an orgLookup backed by the given DB pool.
func newOrgLookup(db *sql.DB) *orgLookup { return &orgLookup{db: db} }

// NewOrgLookup is the exported constructor for use by cmd/server/modules.go.
// Allows building a context-aware AlertCreator for the HTTP module path.
func NewOrgLookup(db *sql.DB) *orgLookup { return newOrgLookup(db) }

// lookupOrgID resolves the org UUID from the tenant schema name.
func (l *orgLookup) lookupOrgID(ctx context.Context, tenant string) (uuid.UUID, error) {
	var idStr string
	err := l.db.QueryRowContext(ctx,
		`SELECT id::text FROM public.organizations WHERE schema_name = $1 AND deleted_at IS NULL LIMIT 1`,
		tenant,
	).Scan(&idStr)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, fmt.Errorf("social wiring: org not found for tenant %q", tenant)
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("social wiring: lookup org: %w", err)
	}
	return uuid.Parse(idStr)
}

// ── PlanLimits adapter ────────────────────────────────────────────────────────

// planLimitsAdapter implements services.PlanLimits via raw SQL on
// public.plans + public.organization_plans (REQ-WIRING-03).
//
// Plan fields are resolved using the tenant schema name extracted from the
// request context (set by the tenant middleware). The workspaceID parameter
// is accepted by the interface but not used for the schema join — the tenant
// identifies the organisation uniquely at runtime.
type planLimitsAdapter struct {
	db     *sql.DB
	lookup *orgLookup
}

// Compile-time interface check.
var _ services.PlanLimits = (*planLimitsAdapter)(nil)

// NewPlanLimits constructs a PlanLimits adapter backed by the shared DB pool.
func NewPlanLimits(db *sql.DB) services.PlanLimits {
	return &planLimitsAdapter{db: db, lookup: newOrgLookup(db)}
}

// GetProfilesLimit returns the maximum number of social profiles allowed for the
// workspace's organisation. Returns 0 for unlimited (NULL in DB), -1 when the
// feature is disabled (social_profiles_limit=0 in DB).
// The tenant is resolved from the request context via the tenant middleware.
func (a *planLimitsAdapter) GetProfilesLimit(ctx context.Context, _ string) (int, error) {
	val, err := a.queryPlanFieldByTenant(ctx, "social_profiles_limit")
	if err != nil {
		return 0, err
	}
	return normalisePlanValue(val), nil
}

// GetChecksPerDay returns the maximum number of social checks per day for the
// workspace's organisation. Returns 0 for unlimited (NULL in DB), -1 for disabled.
// The tenant is resolved from the request context via the tenant middleware.
func (a *planLimitsAdapter) GetChecksPerDay(ctx context.Context, _ string) (int, error) {
	val, err := a.queryPlanFieldByTenant(ctx, "social_checks_per_day")
	if err != nil {
		return 0, err
	}
	return normalisePlanValue(val), nil
}

// GetChecksUsedToday returns the number of social checks consumed today.
// The tenant is resolved from request context; this is a read-only query
// against the tenant schema's social_check_usage table, used by get_quota_status.
func (a *planLimitsAdapter) GetChecksUsedToday(ctx context.Context) (int, error) {
	tenant := middleware.GetTenantFromContext(ctx)
	if tenant == "" {
		return 0, nil // no tenant in context (e.g. scheduler path); caller uses quota repo directly
	}

	var count int
	err := a.db.QueryRowContext(ctx,
		fmt.Sprintf(`SELECT COALESCE(checks_used, 0) FROM %q.social_check_usage WHERE usage_date = CURRENT_DATE`, tenant),
	).Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("social plan limits: get checks used today: %w", err)
	}
	return count, nil
}

// queryPlanFieldByTenant reads a nullable integer plan field for the current
// tenant's active organisation plan. The tenant is extracted from the context.
// Returns nil when no active plan exists (treated as unlimited by the caller).
func (a *planLimitsAdapter) queryPlanFieldByTenant(ctx context.Context, column string) (*int, error) {
	if column != "social_profiles_limit" && column != "social_checks_per_day" {
		return nil, fmt.Errorf("planLimitsAdapter: unknown plan column %q", column)
	}

	tenant := middleware.GetTenantFromContext(ctx)
	if tenant == "" {
		// No tenant context (e.g. unit test path) → unlimited.
		return nil, nil
	}

	orgID, err := a.lookup.lookupOrgID(ctx, tenant)
	if err != nil {
		return nil, err
	}

	// Hard-coded column names (validated above) are safe to interpolate.
	q := fmt.Sprintf(`
		SELECT p.%s
		  FROM public.plans p
		  JOIN public.organization_plans op ON op.plan_id = p.id
		 WHERE op.organization_id = $1
		   AND op.status          = 'active'
		   AND op.deleted_at      IS NULL
		   AND p.is_active        = TRUE
		 ORDER BY op.started_at DESC
		 LIMIT 1
	`, column)

	var val sql.NullInt64
	err = a.db.QueryRowContext(ctx, q, orgID).Scan(&val)
	if errors.Is(err, sql.ErrNoRows) {
		// No active plan → treat as unlimited.
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("planLimitsAdapter: query %s for org %s: %w", column, orgID, err)
	}
	if !val.Valid {
		// NULL in DB → unlimited.
		return nil, nil
	}
	v := int(val.Int64)
	return &v, nil
}

// normalisePlanValue converts a nullable DB value to the domain convention:
//   - nil (NULL) → 0 (unlimited)
//   - 0           → -1 (feature disabled)
//   - >0          → the limit value
func normalisePlanValue(v *int) int {
	if v == nil {
		return 0 // NULL = unlimited
	}
	if *v == 0 {
		return -1 // 0 in DB = feature disabled
	}
	return *v
}

// GetChecksPerDayByTenant is a helper used by the tenant factory to pre-load the
// daily limit when constructing a run_check.Handler for the scheduler. It reads
// the plan limit for the given tenant schema without a request context.
func GetChecksPerDayByTenant(ctx context.Context, db *sql.DB, tenant string) (int, error) {
	lookup := newOrgLookup(db)
	orgID, err := lookup.lookupOrgID(ctx, tenant)
	if err != nil {
		// Unknown tenant or no org → unlimited (0).
		return 0, nil
	}

	var val sql.NullInt64
	err = db.QueryRowContext(ctx, `
		SELECT p.social_checks_per_day
		  FROM public.plans p
		  JOIN public.organization_plans op ON op.plan_id = p.id
		 WHERE op.organization_id = $1
		   AND op.status          = 'active'
		   AND op.deleted_at      IS NULL
		   AND p.is_active        = TRUE
		 ORDER BY op.started_at DESC
		 LIMIT 1
	`, orgID).Scan(&val)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, nil // best-effort; scheduler treats unknown as unlimited
	}
	if !val.Valid {
		return 0, nil
	}
	return normalisePlanValue(func() *int { v := int(val.Int64); return &v }()), nil
}
