package services

import "context"

// CheckoutInput carries the parameters required to create a Stripe Checkout session.
type CheckoutInput struct {
	CustomerID string // Stripe customer ID (already ensured before this call)
	PriceID    string // Stripe price ID from plans.stripe_price_id_{monthly,yearly}
	SuccessURL string
	CancelURL  string
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

	// RetrieveCustomerBalance returns the customer's account balance in cents.
	// Negative values represent credit available to the customer (auto-applied
	// to upcoming invoices). Positive values represent pending amounts owed.
	RetrieveCustomerBalance(ctx context.Context, customerID string) (cents int64, currency string, err error)
}
