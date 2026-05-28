package updatesubscription

// Response carries the result of an in-place subscription change.
//
// Usage-based proration breakdown (Model B):
//   - UsageCreditCents:    refund for the unused portion of the current plan
//     based on remaining checks (NOT remaining time)
//   - AccountCreditCents:  the customer's Stripe balance (negative = credit)
//     applied automatically to upcoming invoices
//   - NewPlanChargeCents:  full price of the new plan starting today
//   - AmountDueCents:      max(0, NewPlan - UsageCredit - AccountCredit)
//
// On Preview=true only the breakdown is populated. On non-preview, the
// subscription has already been mutated.
type Response struct {
	SubscriptionID      string `json:"subscription_id"`
	NewPriceID          string `json:"new_price_id"`
	BillingCycle        string `json:"billing_cycle"`

	UsageCreditCents     int64  `json:"usage_credit_cents"`
	AccountCreditCents   int64  `json:"account_credit_cents"`
	NewPlanChargeCents   int64  `json:"new_plan_charge_cents"`
	AmountDueCents       int64  `json:"amount_due_cents"`
	Currency             string `json:"currency"`

	ChecksRemaining int `json:"checks_remaining"`
	ChecksAllowed   int `json:"checks_allowed"`

	Preview bool `json:"preview"`
}
