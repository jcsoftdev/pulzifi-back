package persistence

import (
	"context"
	"database/sql"

	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/repositories"
)

// PlanPostgresRepository implements repositories.PlanRepository against the public schema.
type PlanPostgresRepository struct {
	db *sql.DB
}

// NewPlanPostgresRepository creates a new plan repository.
func NewPlanPostgresRepository(db *sql.DB) repositories.PlanRepository {
	return &PlanPostgresRepository{db: db}
}

func (r *PlanPostgresRepository) ListActive(ctx context.Context) ([]*entities.Plan, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, code, name, description, checks_allowed_monthly, is_active, storage_period_days
		FROM public.plans
		WHERE is_active = TRUE
		ORDER BY checks_allowed_monthly ASC
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var plans []*entities.Plan
	for rows.Next() {
		p := &entities.Plan{}
		if err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.Description, &p.ChecksAllowedMonthly, &p.IsActive, &p.StoragePeriodDays); err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}
	return plans, rows.Err()
}

func (r *PlanPostgresRepository) GetByCode(ctx context.Context, code string) (*entities.Plan, error) {
	p := &entities.Plan{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, code, name, description, checks_allowed_monthly, is_active, storage_period_days
		FROM public.plans
		WHERE code = $1 AND is_active = TRUE
		LIMIT 1
	`, code).Scan(&p.ID, &p.Code, &p.Name, &p.Description, &p.ChecksAllowedMonthly, &p.IsActive, &p.StoragePeriodDays)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}
