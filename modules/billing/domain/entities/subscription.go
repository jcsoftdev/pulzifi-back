package entities

import (
	"time"

	"github.com/google/uuid"
)

// Subscription represents the Stripe-backed subscription state for an organisation.
type Subscription struct {
	OrgID                uuid.UUID
	StripeSubscriptionID string
	StripeCustomerID     string
	PlanID               uuid.UUID
	PlanCode             string // public.plans.code — frontend-facing identifier ("starter"|"pro"|...)
	PlanName             string // public.plans.name — display label
	BillingCycle         string // "monthly" | "yearly" | "" — derived from the active stripe_price_id
	BillingStatus        BillingStatus
	CurrentPeriodEnd     *time.Time
	PaymentStatus        string
	UpdatedAt            time.Time
}
