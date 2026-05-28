package services

// PromotionCode is the domain view of a Stripe promotion code wrapping a
// first-month discount coupon. Customer-facing token + redemption metadata.
type PromotionCode struct {
	ID             string
	Code           string
	CouponID       string
	AmountOffCents int64
	Currency       string
	Active         bool
	MaxRedemptions int64 // 0 = unlimited
	TimesRedeemed  int64
	ExpiresAt      int64  // unix seconds; 0 = no expiry
	PlanCode       string // from coupon metadata
	BillingCycle   string // from coupon metadata
}

// CouponMetadata is attached to the Stripe coupon so the list view can show
// which plan/cycle a promo targets without a local mirror table.
type CouponMetadata struct {
	PlanCode     string
	BillingCycle string
	Purpose      string // e.g. "first_month_1usd"
}
