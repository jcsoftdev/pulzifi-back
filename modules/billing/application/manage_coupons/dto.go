package managecoupons

// CreateRequest is the admin input to mint a first-month-$1 promo.
type CreateRequest struct {
	PlanCode       string // "starter" | "pro" | ...
	BillingCycle   string // "monthly" | "yearly"
	Code           string // optional human code; empty → Stripe auto-generates
	MaxRedemptions int64  // 0 = unlimited
	ExpiresAt      int64  // unix seconds; 0 = no expiry
}

// CreateResponse is returned after a promo is created.
type CreateResponse struct {
	CouponID        string `json:"coupon_id"`
	PromotionCodeID string `json:"promotion_code_id"`
	Code            string `json:"code"`
	AmountOffCents  int64  `json:"amount_off_cents"`
	Currency        string `json:"currency"`
	ApplyURL        string `json:"apply_url"`
}

// PromotionCodeDTO is one row in the admin coupon list.
type PromotionCodeDTO struct {
	ID             string `json:"id"`
	Code           string `json:"code"`
	AmountOffCents int64  `json:"amount_off_cents"`
	Currency       string `json:"currency"`
	Active         bool   `json:"active"`
	MaxRedemptions int64  `json:"max_redemptions"`
	TimesRedeemed  int64  `json:"times_redeemed"`
	ExpiresAt      int64  `json:"expires_at"`
	PlanCode       string `json:"plan_code"`
	BillingCycle   string `json:"billing_cycle"`
	ApplyURL       string `json:"apply_url"`
}
