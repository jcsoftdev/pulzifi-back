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
}
