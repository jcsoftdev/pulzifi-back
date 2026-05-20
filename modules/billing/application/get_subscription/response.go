package getsubscription

import "time"

// Response is the output DTO for the GetSubscription use case.
type Response struct {
	OrgID                string     `json:"org_id"`
	PlanID               string     `json:"plan_id"`
	BillingStatus        string     `json:"billing_status"`        // Stripe billing status; empty if no Stripe sub
	PaymentStatus        string     `json:"payment_status"`        // "ok" | "past_due" | "grace_period"
	StripeSubscriptionID string     `json:"stripe_subscription_id"` // empty if no Stripe sub
	CurrentPeriodEnd     *time.Time `json:"current_period_end"`    // nil if no Stripe sub
}
