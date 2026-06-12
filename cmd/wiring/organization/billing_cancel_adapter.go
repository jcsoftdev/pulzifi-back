// Package orgwiring provides cross-module adapters that implement organization
// domain ports. Implementations live here so the organization module never
// imports billing or snapshot packages directly (hexagonal boundary rule).
package orgwiring

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	billingservices "github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
	orgservices "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
)

// planRow is the minimal data we read from public.organization_plans.
type planRow struct {
	subscriptionID string
	billingStatus  string
	amount         int64 // monthly price in cents; 0 = trial/free
}

// planReader is an abstraction over the DB so tests can inject an in-memory fake.
type planReader interface {
	readActivePlan(ctx context.Context, orgID uuid.UUID) (*planRow, error)
}

// ── billingCancelAdapter ──────────────────────────────────────────────────────

// billingCancelAdapter implements orgservices.BillingCanceller by looking up
// the active Stripe subscription from organization_plans and delegating
// cancellation to the billing StripeGateway.
type billingCancelAdapter struct {
	reader  planReader
	gateway billingservices.StripeGateway
}

// compile-time assertion
var _ orgservices.BillingCanceller = (*billingCancelAdapter)(nil)

// NewBillingCancelAdapter builds a production adapter backed by a real *sql.DB.
func NewBillingCancelAdapter(db *sql.DB, gateway billingservices.StripeGateway) orgservices.BillingCanceller {
	return &billingCancelAdapter{
		reader:  &sqlPlanReader{db: db},
		gateway: gateway,
	}
}

// NewBillingCancelAdapterForTest builds an adapter with an injected planReader.
// Used by tests to inject an in-memory fake instead of a real database.
func NewBillingCancelAdapterForTest(reader planReader, gateway billingservices.StripeGateway) orgservices.BillingCanceller {
	return &billingCancelAdapter{reader: reader, gateway: gateway}
}

// CancelForOrg resolves the active Stripe subscription for the org and
// cancels it immediately. Returns ErrBillingActive only when the subscription
// is on a paid active/past_due plan AND the cancel call fails — in that case
// the caller MUST abort the deletion cascade and return 409.
func (a *billingCancelAdapter) CancelForOrg(ctx context.Context, orgID uuid.UUID) error {
	plan, err := a.reader.readActivePlan(ctx, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No active plan row — nothing to cancel.
			return nil
		}
		return fmt.Errorf("billing_cancel_adapter: read plan: %w", err)
	}
	if plan == nil || plan.subscriptionID == "" {
		return nil
	}

	isPaidActive := (plan.billingStatus == "active" || plan.billingStatus == "past_due") && plan.amount > 0

	_, cancelErr := a.gateway.CancelSubscriptionNow(ctx, plan.subscriptionID)
	if cancelErr != nil {
		if isPaidActive {
			// Paid + active/past_due + cancel failed → abort (DAO-3 abort rule).
			return orgservices.ErrBillingActive
		}
		// Trial / free / non-paid: cancel failure is ignored (best-effort).
		return nil
	}
	return nil
}

// ── sqlPlanReader ─────────────────────────────────────────────────────────────

// sqlPlanReader reads from public.organization_plans using raw SQL (same
// pattern as cmd/wiring/billing/plan_assigner.go; no billing module import).
type sqlPlanReader struct{ db *sql.DB }

func (r *sqlPlanReader) readActivePlan(ctx context.Context, orgID uuid.UUID) (*planRow, error) {
	var row planRow
	// amount is read from plans.amount_cents_monthly (added in migration 000026).
	// Falls back to 0 when no plan is joined (free/trial).
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(op.stripe_subscription_id, ''),
		       COALESCE(op.billing_status, ''),
		       COALESCE(p.amount_cents_monthly, 0)
		  FROM public.organization_plans op
		  LEFT JOIN public.plans p ON p.id = op.plan_id
		 WHERE op.organization_id = $1
		   AND op.status          = 'active'
		   AND op.deleted_at      IS NULL
		 LIMIT 1
	`, orgID).Scan(&row.subscriptionID, &row.billingStatus, &row.amount)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// ── nopBillingCanceller ───────────────────────────────────────────────────────

// nopBillingCanceller is injected when BILLING_ENABLED=false. Returns nil
// immediately without any Stripe calls (OQ-5).
type nopBillingCanceller struct{}

var _ orgservices.BillingCanceller = (*nopBillingCanceller)(nil)

// NopBillingCanceller returns a no-op BillingCanceller for use when
// BILLING_ENABLED is false.
func NopBillingCanceller() orgservices.BillingCanceller {
	return &nopBillingCanceller{}
}

func (n *nopBillingCanceller) CancelForOrg(_ context.Context, _ uuid.UUID) error {
	return nil
}

