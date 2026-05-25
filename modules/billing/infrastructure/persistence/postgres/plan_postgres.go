package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/repositories"
)

var _ repositories.PlanRepository = (*PlanPostgresRepository)(nil)

// PlanPostgresRepository reads plan rows from public.plans.
type PlanPostgresRepository struct {
	db *sql.DB
}

func NewPlanPostgresRepository(db *sql.DB) *PlanPostgresRepository {
	return &PlanPostgresRepository{db: db}
}

// FindStripePriceID resolves a plan code + billing cycle to its Stripe price ID.
// Returns repositories.ErrPlanNotFound when no active plan with that code exists.
func (r *PlanPostgresRepository) FindStripePriceID(ctx context.Context, planCode, billingCycle string) (string, error) {
	const query = `
		SELECT COALESCE(stripe_price_id_monthly, ''), COALESCE(stripe_price_id_yearly, '')
		FROM public.plans
		WHERE code = $1 AND is_active = TRUE
		LIMIT 1
	`
	var monthly, yearly string
	err := r.db.QueryRowContext(ctx, query, planCode).Scan(&monthly, &yearly)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", repositories.ErrPlanNotFound
		}
		return "", err
	}

	switch billingCycle {
	case "monthly":
		return monthly, nil
	case "yearly":
		return yearly, nil
	default:
		return "", nil
	}
}
