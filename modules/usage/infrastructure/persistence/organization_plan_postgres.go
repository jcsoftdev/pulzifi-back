package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/repositories"
)

// orgPlanDB is the concrete executor interface shared by *sql.DB and *sql.Tx.
type orgPlanDB interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// OrganizationPlanPostgresRepository implements repositories.OrganizationPlanRepository.
type OrganizationPlanPostgresRepository struct {
	db   *sql.DB
	exec orgPlanDB
}

// NewOrganizationPlanPostgresRepository creates a new organization plan repository.
func NewOrganizationPlanPostgresRepository(db *sql.DB) repositories.OrganizationPlanRepository {
	return &OrganizationPlanPostgresRepository{db: db, exec: db}
}

// WithTx returns a transaction-scoped copy. tx must be a *sql.Tx.
func (r *OrganizationPlanPostgresRepository) WithTx(tx interface{}) repositories.OrganizationPlanRepository {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok || sqlTx == nil {
		return r
	}
	return &OrganizationPlanPostgresRepository{db: r.db, exec: sqlTx}
}

func (r *OrganizationPlanPostgresRepository) ListOrgsWithPlans(ctx context.Context) ([]*repositories.OrgWithPlan, error) {
	rows, err := r.exec.QueryContext(ctx, `
		SELECT
			o.id,
			o.name,
			o.subdomain,
			o.schema_name,
			COALESCE(p.code, ''),
			COALESCE(p.name, ''),
			COALESCE(p.checks_allowed_monthly, 0),
			COALESCE(p.storage_period_days, 7)
		FROM public.organizations o
		LEFT JOIN LATERAL (
			SELECT op.plan_id
			FROM public.organization_plans op
			WHERE op.organization_id = o.id
				AND op.deleted_at IS NULL
				AND op.status = 'active'
			ORDER BY op.started_at DESC
			LIMIT 1
		) active_plan ON TRUE
		LEFT JOIN public.plans p ON p.id = active_plan.plan_id
		WHERE o.deleted_at IS NULL
		ORDER BY o.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []*repositories.OrgWithPlan
	for rows.Next() {
		o := &repositories.OrgWithPlan{}
		if err := rows.Scan(&o.ID, &o.Name, &o.Subdomain, &o.SchemaName, &o.PlanCode, &o.PlanName, &o.ChecksAllowed, &o.StoragePeriodDays); err != nil {
			return nil, err
		}
		result = append(result, o)
	}
	return result, rows.Err()
}

func (r *OrganizationPlanPostgresRepository) DeactivateActive(ctx context.Context, orgID uuid.UUID) error {
	_, err := r.exec.ExecContext(ctx, `
		UPDATE public.organization_plans
		SET status = 'inactive', ended_at = NOW(), updated_at = NOW()
		WHERE organization_id = $1
		  AND status = 'active'
		  AND deleted_at IS NULL
	`, orgID)
	return err
}

func (r *OrganizationPlanPostgresRepository) Insert(ctx context.Context, plan *entities.OrganizationPlan) error {
	var createdBy interface{}
	if plan.CreatedBy != nil {
		createdBy = plan.CreatedBy.String()
	}
	_, err := r.exec.ExecContext(ctx, `
		INSERT INTO public.organization_plans (organization_id, plan_id, status, started_at, created_by, created_at, updated_at)
		VALUES ($1, $2, 'active', NOW(), NULLIF($3, '')::uuid, NOW(), NOW())
	`, plan.OrganizationID, plan.PlanID, createdBy)
	return err
}

func (r *OrganizationPlanPostgresRepository) GetSchemaName(ctx context.Context, orgID uuid.UUID) (string, error) {
	var schemaName string
	err := r.exec.QueryRowContext(ctx, `SELECT schema_name FROM public.organizations WHERE id = $1`, orgID).Scan(&schemaName)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return schemaName, err
}

func (r *OrganizationPlanPostgresRepository) GetActivePlanForOrg(ctx context.Context, orgID uuid.UUID) (*repositories.OrgPlanInfo, error) {
	var info repositories.OrgPlanInfo
	var startedAt time.Time
	var aiAllowed sql.NullInt64
	var maxPages, maxWorkspaces sql.NullInt64
	err := r.exec.QueryRowContext(ctx, `
		SELECT o.schema_name, COALESCE(p.checks_allowed_monthly, 2147483647), COALESCE(p.storage_period_days, 7),
		       p.ai_insights_allowed_monthly, op.started_at,
		       p.max_pages, p.max_workspaces
		FROM public.organizations o
		JOIN public.organization_plans op ON op.organization_id = o.id
			AND op.status = 'active' AND op.deleted_at IS NULL
		JOIN public.plans p ON p.id = op.plan_id
		WHERE o.id = $1
		ORDER BY op.started_at DESC
		LIMIT 1
	`, orgID).Scan(&info.SchemaName, &info.ChecksAllowed, &info.StoragePeriodDays, &aiAllowed, &startedAt, &maxPages, &maxWorkspaces)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	info.AIInsightsAllowed = resolveNullInt(aiAllowed)
	info.MaxPages = resolveNullInt(maxPages)
	info.MaxWorkspaces = resolveNullInt(maxWorkspaces)
	info.StartedAt = startedAt
	return &info, nil
}

func (r *OrganizationPlanPostgresRepository) GetActivePlanForTenant(ctx context.Context, tenant string) (*repositories.OrgPlanInfo, error) {
	var info repositories.OrgPlanInfo
	var startedAt time.Time
	var aiAllowed sql.NullInt64
	var maxPages, maxWorkspaces sql.NullInt64
	err := r.exec.QueryRowContext(ctx, `
		SELECT COALESCE(p.checks_allowed_monthly, 2147483647), COALESCE(p.storage_period_days, 7),
		       p.ai_insights_allowed_monthly, op.started_at,
		       p.max_pages, p.max_workspaces
		FROM public.organizations o
		JOIN public.organization_plans op ON op.organization_id = o.id
			AND op.status = 'active' AND op.deleted_at IS NULL
		JOIN public.plans p ON p.id = op.plan_id
		WHERE o.schema_name = $1
		ORDER BY op.started_at DESC
		LIMIT 1
	`, tenant).Scan(&info.ChecksAllowed, &info.StoragePeriodDays, &aiAllowed, &startedAt, &maxPages, &maxWorkspaces)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	info.AIInsightsAllowed = resolveNullInt(aiAllowed)
	info.MaxPages = resolveNullInt(maxPages)
	info.MaxWorkspaces = resolveNullInt(maxWorkspaces)
	info.StartedAt = startedAt
	return &info, nil
}

// resolveNullInt converts a nullable plan column into an application-level
// integer. NULL (Enterprise/trial = unlimited) → INT max sentinel (2147483647)
// so "used < allowed" comparisons work without nullable branching.
func resolveNullInt(v sql.NullInt64) int {
	if !v.Valid {
		return 2147483647
	}
	return int(v.Int64)
}
