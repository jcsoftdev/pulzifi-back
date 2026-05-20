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
	BillingStatus        BillingStatus
	CurrentPeriodEnd     *time.Time
	PaymentStatus        string
	UpdatedAt            time.Time
}
