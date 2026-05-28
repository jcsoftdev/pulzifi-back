package services

import "context"

// CheckoutInput carries the parameters required to create a Stripe Checkout session.
type CheckoutInput struct {
	CustomerID string // Stripe customer ID (already ensured before this call)
	PriceID    string // Stripe price ID from plans.stripe_price_id_{monthly,yearly}
	SuccessURL string
	CancelURL  string

	// PromotionCodeID, when non-empty, pre-applies a Stripe promotion code as a
	// Checkout discount (the "Try for $1" apply-link path). Mutually exclusive
	// with the hosted promo-code field: when set, AllowPromotionCodes is NOT
	// enabled. When empty, the gateway enables AllowPromotionCodes so customers
	// can still type a code manually.
	PromotionCodeID string
}

// StripeEvent is a minimal representation of a parsed Stripe webhook event.
// The full raw payload is preserved so handlers can unmarshal event-specific data.
type StripeEvent struct {
	ID      string
	Type    string
	RawData []byte // raw JSON of event.Data.Raw
}

// StripeSubscription is the billing state as returned by Stripe for a subscription.
type StripeSubscription struct {
	ID               string
	Status           string // Stripe's status string (active, past_due, canceled, etc.)
	CurrentPeriodEnd int64  // Unix timestamp
	CustomerID       string
	PriceID          string // first item's price ID

	// Active discount (coupon) on the subscription, if any. Populated by
	// RetrieveSubscription. Drives the "free month gifted" banner: a gift is a
	// discount whose coupon metadata purpose == "gift_one_month".
	DiscountAmountOffCents int64
	DiscountPercentOff     int64
	DiscountPurpose        string // coupon metadata "purpose" (e.g. gift_one_month, first_month_1usd)

	// Scheduled cancellation state. CancelAtPeriodEnd is true when the user has
	// requested cancellation but the subscription stays active until the paid
	// period ends. CancelAt is the Unix timestamp Stripe will end it (0 when
	// not scheduled).
	CancelAtPeriodEnd bool
	CancelAt          int64

	// Plan-gift state (derived from subscription metadata, set by a gift
	// schedule). GiftPlanActive is true while the customer is in the free
	// gifted-plan phase; GiftPlanCode is the plan they're using for free and
	// GiftRevertPlanCode is the plan they return to when the gift ends (at
	// CurrentPeriodEnd / TrialEnd). Empty/false when no plan gift is running.
	GiftPlanActive     bool
	GiftPlanCode       string
	GiftRevertPlanCode string
}

// StripeGateway abstracts all Stripe API interactions behind a testable interface.
// The concrete implementation lives in infrastructure/stripe/gateway.go.
type StripeGateway interface {
	// EnsureCustomer returns an existing Stripe customer ID for the org or creates one.
	EnsureCustomer(ctx context.Context, orgID, email, name string) (customerID string, err error)

	// CreateCheckoutSession creates a hosted Checkout session and returns the URL.
	CreateCheckoutSession(ctx context.Context, in CheckoutInput) (url string, err error)

	// CreatePortalSession creates a Billing Portal session and returns the URL.
	CreatePortalSession(ctx context.Context, customerID, returnURL string) (url string, err error)

	// ConstructEvent parses and validates a Stripe webhook payload.
	// Returns an error when the signature is invalid or the payload is malformed.
	ConstructEvent(payload []byte, signature, secret string) (StripeEvent, error)

	// RetrieveSubscription fetches current subscription state from Stripe.
	RetrieveSubscription(ctx context.Context, subID string) (StripeSubscription, error)

	// ListSubscriptions returns all non-deleted subscriptions for a customer.
	// Used by ReconcileFromStripe to discover paid plans that webhooks missed.
	// Status filtering (active, past_due, trialing) is the caller's responsibility.
	ListSubscriptions(ctx context.Context, customerID string) ([]StripeSubscription, error)

	// UpdateSubscriptionItem swaps the (single) line item on an existing
	// subscription to a new price, with proration handled by Stripe.
	// prorationBehavior values: "create_prorations" (default for upgrades —
	// generates a prorated invoice on the next cycle) or "always_invoice"
	// (charges the prorated amount immediately).
	UpdateSubscriptionItem(ctx context.Context, subID, newPriceID, prorationBehavior string) (StripeSubscription, error)

	// PreviewProration returns the prorated invoice total (in cents) that
	// would be billed if the subscription's price changed to newPriceID NOW.
	// Used by the UI to show "Te cobramos $X prorrateado hoy" before the user
	// confirms the upgrade. Returns currency in lowercase ISO ("usd").
	PreviewProration(ctx context.Context, subID, newPriceID string) (amountCents int64, currency string, err error)

	// CreateRefundCreditInvoiceItem attaches a negative-amount InvoiceItem to
	// the customer so the next invoice subtracts the supplied credit. Used by
	// the usage-based upgrade path to refund the unused portion of the prior
	// plan based on remaining checks (not time). amountCents MUST be positive;
	// the gateway negates it before sending to Stripe.
	CreateRefundCreditInvoiceItem(ctx context.Context, customerID string, amountCents int64, currency, description string) error

	// UpdateSubscriptionAnchorNow swaps the subscription's price to newPriceID
	// and resets the billing cycle to start today (billing_cycle_anchor=now)
	// with proration_behavior=none. This is the usage-based companion to
	// UpdateSubscriptionItem — the caller must have already applied any
	// custom credit (e.g. via CreateRefundCreditInvoiceItem) before invoking
	// this method so the resulting invoice nets out correctly.
	UpdateSubscriptionAnchorNow(ctx context.Context, subID, newPriceID string) (StripeSubscription, error)

	// RetrievePriceAmount returns the unit_amount + currency of a Stripe Price.
	// Used by the usage-based proration path to compute the refund credit and
	// the new-plan charge in our own units (without trusting any local
	// hardcoded price catalog).
	RetrievePriceAmount(ctx context.Context, priceID string) (amountCents int64, currency string, err error)

	// ── Coupons / promotion codes (first-month-$1 promos) ──────────────────

	// CreateCoupon creates a once-applied amount_off coupon and returns its ID.
	CreateCoupon(ctx context.Context, amountOffCents int64, currency string, meta CouponMetadata) (couponID string, err error)

	// ApplyCouponToSubscription attaches a coupon to an EXISTING subscription so
	// the discount lands on the next invoice. Used by gift (and any
	// apply-discount-to-existing-sub flow). proration is not triggered.
	ApplyCouponToSubscription(ctx context.Context, subID, couponID string) error

	// CreatePromotionCode wraps a coupon in a customer-facing code. code may be
	// empty to let Stripe auto-generate. maxRedemptions / expiresAt are passed
	// through when > 0.
	CreatePromotionCode(ctx context.Context, couponID, code string, maxRedemptions, expiresAt int64) (PromotionCode, error)

	// ListPromotionCodes returns all promotion codes (active + inactive) for
	// the admin coupon list.
	ListPromotionCodes(ctx context.Context) ([]PromotionCode, error)

	// DeactivatePromotionCode flips a promotion code to inactive (Stripe codes
	// cannot be deleted). Existing redemptions are unaffected.
	DeactivatePromotionCode(ctx context.Context, promotionCodeID string) error

	// FindPromotionCodeByCode resolves a human code to its promotion code for
	// the auto-apply checkout path. Returns an error when not found / inactive.
	FindPromotionCodeByCode(ctx context.Context, code string) (PromotionCode, error)

	// RetrieveCustomerBalance returns the customer's account balance in cents.
	// Negative values represent credit available to the customer (auto-applied
	// to upcoming invoices). Positive values represent pending amounts owed.
	RetrieveCustomerBalance(ctx context.Context, customerID string) (cents int64, currency string, err error)

	// GiftPlanSchedule converts the subscription into a 2-phase schedule: phase
	// 1 runs giftPriceID FREE (as a trial) for one billing cycle so the
	// customer immediately uses the gifted plan's features; phase 2 reverts to
	// currentPriceID at the normal price, ongoing. Used when the gifted plan is
	// a HIGHER tier than the customer's current plan. giftCode/currentCode are
	// stamped into subscription metadata so the UI can explain the gift.
	// Returns the Unix timestamp the gift ends (revert date).
	GiftPlanSchedule(ctx context.Context, subID, giftPriceID, currentPriceID, giftCode, currentCode string) (revertAt int64, err error)

	// CreateGiftSubscription creates a brand-new subscription on giftPriceID for
	// a customer that has no active subscription (e.g. a trial / newly invited
	// org). The subscription runs as a FREE trial for one month and, because no
	// payment method is attached, auto-cancels at trial end (missing_payment_
	// method=cancel) — giving the org a free month of the gifted plan that then
	// simply ends (no active plan). giftCode is stamped into metadata for the
	// UI. Returns the created subscription (CurrentPeriodEnd = trial end).
	CreateGiftSubscription(ctx context.Context, customerID, giftPriceID, giftCode string) (StripeSubscription, error)

	// CreditCustomerBalance grants the customer a credit on their Stripe
	// account balance. amountCents MUST be positive — the gateway records it as
	// a NEGATIVE balance transaction (Stripe's credit convention) so Stripe
	// auto-applies it to upcoming invoices and carries any leftover forward.
	// Used by gifts and duplicate-payment refunds: unlike an amount_off coupon,
	// balance credit spans multiple invoices and never discards leftover value.
	CreditCustomerBalance(ctx context.Context, customerID string, amountCents int64, currency, description string) error

	// ── Cancellation ──────────────────────────────────────────────────────

	// CancelSubscriptionAtPeriodEnd schedules the subscription to end when the
	// current paid period closes (cancel_at_period_end=true). The user keeps
	// access until then; Stripe emits customer.subscription.deleted at period
	// end, which deactivates the local plan.
	CancelSubscriptionAtPeriodEnd(ctx context.Context, subID string) (StripeSubscription, error)

	// ResumeSubscription clears a pending period-end cancellation
	// (cancel_at_period_end=false), restoring normal renewal.
	ResumeSubscription(ctx context.Context, subID string) (StripeSubscription, error)

	// RemoveSubscriptionDiscount deletes the active discount (coupon) from a
	// subscription. Used by the one-off migration that converts legacy gift
	// coupons into balance credit.
	RemoveSubscriptionDiscount(ctx context.Context, subID string) error
}
