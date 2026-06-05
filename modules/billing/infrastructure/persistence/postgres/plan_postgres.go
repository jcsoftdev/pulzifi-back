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

// UpsertStripePricing writes the Stripe price ID + cached amount for one cycle
// of an existing plan. Only the targeted cycle's columns are updated; the other
// cycle is preserved. Returns ErrPlanNotFound when the code does not exist.
func (r *PlanPostgresRepository) UpsertStripePricing(ctx context.Context, planCode, cycle, priceID string, amountCents int64, currency string) error {
	var column string
	var amountColumn string
	switch cycle {
	case "monthly":
		column, amountColumn = "stripe_price_id_monthly", "amount_cents_monthly"
	case "yearly":
		column, amountColumn = "stripe_price_id_yearly", "amount_cents_yearly"
	default:
		return errors.New("billing: invalid billing cycle " + cycle)
	}

	// #nosec G201 -- column names are from a fixed allowlist above, never user input.
	query := `
		UPDATE public.plans
		SET ` + column + ` = $2, ` + amountColumn + ` = $3, currency = $4, updated_at = NOW()
		WHERE code = $1`
	res, err := r.db.ExecContext(ctx, query, planCode, priceID, amountCents, currency)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return repositories.ErrPlanNotFound
	}
	return nil
}
