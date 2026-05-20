package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
)

// ErrPlanNotFound is returned by PlanAssigner.Assign when the Stripe price ID
// cannot be resolved to any row in public.plans. Callers (e.g. handle_webhook)
// MUST treat this as a no-op warning rather than a fatal error, to prevent
// Stripe retry storms when an unknown price_id arrives.
var ErrPlanNotFound = errors.New("billing: no plan found for stripe price ID")

// AssignInput carries the data needed to assign (or update) a plan for an org.
type AssignInput struct {
	OrgID                uuid.UUID
	StripeSubscriptionID string
	StripePriceID        string
	StripeCustomerID     string
	BillingStatus        entities.BillingStatus
	CurrentPeriodEnd     time.Time
	PaymentStatus        string
}

// PlanAssigner maps a Stripe price ID to the correct internal plan and persists
// the subscription state. The concrete implementation lives in cmd/wiring/billing/.
type PlanAssigner interface {
	// Assign resolves the plan from the price ID and upserts the organisation_plans row.
	// It also persists stripe_customer_id on public.organizations if not already set.
	// Returns ErrPlanNotFound if the price ID is not in public.plans.
	Assign(ctx context.Context, in AssignInput) error
}
