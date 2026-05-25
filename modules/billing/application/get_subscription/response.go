package getsubscription

import "time"

// Response is the output DTO for the GetSubscription use case.
type Response struct {
	OrgID                string     `json:"org_id"`
	PlanID               string     `json:"plan_id"`
	PlanCode             string     `json:"plan_code"`              // public.plans.code — stable frontend key
	PlanName             string     `json:"plan_name"`              // public.plans.name — display label
	BillingCycle         string     `json:"billing_cycle"`          // "monthly" | "yearly" | "" — derived from active stripe_price_id
	BillingStatus        string     `json:"billing_status"`         // Stripe billing status; empty if no Stripe sub
	PaymentStatus        string     `json:"payment_status"`         // "ok" | "past_due" | "grace_period"
	StripeCustomerID     string     `json:"stripe_customer_id"`     // empty if org never went through Checkout
	StripeSubscriptionID string     `json:"stripe_subscription_id"` // empty if no Stripe sub
	CurrentPeriodEnd     *time.Time `json:"current_period_end"`     // nil if no Stripe sub
	// CreditBalanceCents is the absolute value of the customer's negative
	// Stripe balance — money the customer has available for upcoming invoices.
	// Zero when the customer has no credit (or a positive balance owed).
	CreditBalanceCents int64  `json:"credit_balance_cents"`
	CreditBalanceCurrency string `json:"credit_balance_currency"` // ISO currency, e.g. "usd"
}
