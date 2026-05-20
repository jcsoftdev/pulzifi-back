// Package postgres provides PostgreSQL-backed repository implementations for the billing module.
// All queries target the PUBLIC schema — Stripe data is organisation-level, not tenant-scoped.
// DO NOT prefix queries with SET search_path.
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/repositories"
)

// Compile-time interface check.
var _ repositories.SubscriptionRepository = (*SubscriptionPostgresRepository)(nil)

// SubscriptionPostgresRepository is the Postgres implementation of SubscriptionRepository.
// It queries public.organization_plans joined with public.organizations to materialise a
// Subscription entity without touching any tenant schema.
type SubscriptionPostgresRepository struct {
	db *sql.DB
}

// NewSubscriptionPostgresRepository returns a new SubscriptionPostgresRepository.
func NewSubscriptionPostgresRepository(db *sql.DB) *SubscriptionPostgresRepository {
	return &SubscriptionPostgresRepository{db: db}
}

// findByQuery is a helper that executes the shared SELECT and scans one row.
func (r *SubscriptionPostgresRepository) findByQuery(ctx context.Context, query string, arg any) (*entities.Subscription, error) {
	var (
		sub              entities.Subscription
		stripeSubID      sql.NullString
		stripeCustID     sql.NullString
		billingStatus    sql.NullString
		currentPeriodEnd sql.NullTime
		paymentStatus    sql.NullString
	)

	err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&sub.OrgID,
		&stripeSubID,
		&stripeCustID,
		&sub.PlanID,
		&billingStatus,
		&currentPeriodEnd,
		&paymentStatus,
		&sub.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	sub.StripeSubscriptionID = stripeSubID.String
	sub.StripeCustomerID = stripeCustID.String
	sub.BillingStatus = entities.BillingStatus(billingStatus.String)
	sub.PaymentStatus = paymentStatus.String
	if currentPeriodEnd.Valid {
		t := currentPeriodEnd.Time
		sub.CurrentPeriodEnd = &t
	}

	return &sub, nil
}

// FindByOrgID returns the subscription for the given organisation, or nil if none exists.
func (r *SubscriptionPostgresRepository) FindByOrgID(ctx context.Context, orgID uuid.UUID) (*entities.Subscription, error) {
	const query = `
		SELECT
			op.organization_id,
			op.stripe_subscription_id,
			o.stripe_customer_id,
			op.plan_id,
			op.billing_status,
			op.current_period_end,
			op.payment_status,
			op.updated_at
		FROM public.organization_plans op
		JOIN public.organizations o ON o.id = op.organization_id
		WHERE op.organization_id = $1
		LIMIT 1
	`
	return r.findByQuery(ctx, query, orgID)
}

// FindByStripeSubscriptionID looks up by the Stripe subscription identifier.
func (r *SubscriptionPostgresRepository) FindByStripeSubscriptionID(ctx context.Context, subID string) (*entities.Subscription, error) {
	const query = `
		SELECT
			op.organization_id,
			op.stripe_subscription_id,
			o.stripe_customer_id,
			op.plan_id,
			op.billing_status,
			op.current_period_end,
			op.payment_status,
			op.updated_at
		FROM public.organization_plans op
		JOIN public.organizations o ON o.id = op.organization_id
		WHERE op.stripe_subscription_id = $1
		LIMIT 1
	`
	return r.findByQuery(ctx, query, subID)
}

// Save inserts a new subscription record into organization_plans.
// It does NOT upsert — use Update for existing rows.
func (r *SubscriptionPostgresRepository) Save(ctx context.Context, sub *entities.Subscription) error {
	const query = `
		INSERT INTO public.organization_plans
			(organization_id, plan_id, stripe_subscription_id, stripe_price_id, billing_status, current_period_end, payment_status, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		sub.OrgID,
		sub.PlanID,
		nullString(sub.StripeSubscriptionID),
		nullString(sub.StripeCustomerID), // stripe_price_id — placeholder; wiring provides actual price id
		string(sub.BillingStatus),
		nullTime(sub.CurrentPeriodEnd),
		nullString(sub.PaymentStatus),
		time.Now(),
	)
	return err
}

// Update overwrites the mutable billing columns for an existing organization_plans row.
func (r *SubscriptionPostgresRepository) Update(ctx context.Context, sub *entities.Subscription) error {
	const query = `
		UPDATE public.organization_plans
		SET
			stripe_subscription_id = $2,
			billing_status         = $3,
			current_period_end     = $4,
			payment_status         = $5,
			updated_at             = $6
		WHERE organization_id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		sub.OrgID,
		nullString(sub.StripeSubscriptionID),
		string(sub.BillingStatus),
		nullTime(sub.CurrentPeriodEnd),
		nullString(sub.PaymentStatus),
		time.Now(),
	)
	return err
}

// ── helpers ───────────────────────────────────────────────────────────────────

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
