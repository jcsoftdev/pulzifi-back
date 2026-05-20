// Package stripe provides the concrete implementation of the billing StripeGateway interface.
// It wraps github.com/stripe/stripe-go/v79 and maps Stripe SDK types to domain types.
package stripe

import (
	"context"
	"fmt"

	stripe "github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/checkout/session"
	"github.com/stripe/stripe-go/v79/customer"
	"github.com/stripe/stripe-go/v79/subscription"
	"github.com/stripe/stripe-go/v79/webhook"

	billingservices "github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"

	billingportalclient "github.com/stripe/stripe-go/v79/billingportal/session"
)

// Compile-time interface check.
var _ billingservices.StripeGateway = (*Gateway)(nil)

// Gateway is the concrete Stripe implementation of billing.StripeGateway.
// It is constructed with the Stripe API key and webhook secret from config.
// The stripe-go SDK is initialised in the constructor via stripe.Key global.
type Gateway struct {
	webhookSecret string
}

// NewGateway creates a Gateway and sets the Stripe API key on the global SDK client.
// apiKey must be a live or test secret key (sk_live_* or sk_test_*).
// webhookSecret must be the signing secret from the Stripe webhook dashboard (whsec_*).
func NewGateway(apiKey, webhookSecret string) *Gateway {
	stripe.Key = apiKey
	return &Gateway{webhookSecret: webhookSecret}
}

// EnsureCustomer returns the existing Stripe customer ID for the org, or creates one.
// It searches by email first to avoid duplicate Stripe customers across retries.
func (g *Gateway) EnsureCustomer(_ context.Context, orgID, email, name string) (string, error) {
	// Look up by email — Stripe allows filtering customers by exact email.
	params := &stripe.CustomerListParams{}
	params.Email = stripe.String(email)
	params.Limit = stripe.Int64(1)

	iter := customer.List(params)
	for iter.Next() {
		return iter.Customer().ID, nil
	}
	if err := iter.Err(); err != nil {
		return "", fmt.Errorf("billing: stripe customer lookup: %w", err)
	}

	// No existing customer — create one.
	created, err := customer.New(&stripe.CustomerParams{
		Email:    stripe.String(email),
		Name:     stripe.String(name),
		Metadata: map[string]string{"org_id": orgID},
	})
	if err != nil {
		return "", fmt.Errorf("billing: stripe customer create: %w", err)
	}

	return created.ID, nil
}

// CreateCheckoutSession creates a hosted Stripe Checkout session for a subscription.
func (g *Gateway) CreateCheckoutSession(_ context.Context, in billingservices.CheckoutInput) (string, error) {
	qty := stripe.Int64(1)
	mode := string(stripe.CheckoutSessionModeSubscription)

	params := &stripe.CheckoutSessionParams{
		Customer:   stripe.String(in.CustomerID),
		SuccessURL: stripe.String(in.SuccessURL),
		CancelURL:  stripe.String(in.CancelURL),
		Mode:       stripe.String(mode),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(in.PriceID),
				Quantity: qty,
			},
		},
	}

	s, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("billing: stripe checkout session: %w", err)
	}

	return s.URL, nil
}

// CreatePortalSession creates a Stripe Customer Portal session and returns its URL.
func (g *Gateway) CreatePortalSession(_ context.Context, customerID, returnURL string) (string, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	}

	s, err := billingportalclient.New(params)
	if err != nil {
		return "", fmt.Errorf("billing: stripe portal session: %w", err)
	}

	return s.URL, nil
}

// ConstructEvent validates the Stripe-Signature header and parses the raw webhook payload.
// Returns ErrInvalidSignature (via the caller) on any validation failure.
func (g *Gateway) ConstructEvent(payload []byte, signature, secret string) (billingservices.StripeEvent, error) {
	event, err := webhook.ConstructEvent(payload, signature, secret)
	if err != nil {
		return billingservices.StripeEvent{}, err
	}

	raw := []byte{}
	if event.Data != nil {
		raw = []byte(event.Data.Raw)
	}

	return billingservices.StripeEvent{
		ID:      event.ID,
		Type:    string(event.Type),
		RawData: raw,
	}, nil
}

// RetrieveSubscription fetches the current subscription state from Stripe.
func (g *Gateway) RetrieveSubscription(_ context.Context, subID string) (billingservices.StripeSubscription, error) {
	sub, err := subscription.Get(subID, nil)
	if err != nil {
		return billingservices.StripeSubscription{}, fmt.Errorf("billing: stripe retrieve subscription: %w", err)
	}

	var priceID string
	if sub.Items != nil && len(sub.Items.Data) > 0 && sub.Items.Data[0].Price != nil {
		priceID = sub.Items.Data[0].Price.ID
	}

	var customerID string
	if sub.Customer != nil {
		customerID = sub.Customer.ID
	}

	return billingservices.StripeSubscription{
		ID:               sub.ID,
		Status:           string(sub.Status),
		CurrentPeriodEnd: sub.CurrentPeriodEnd,
		CustomerID:       customerID,
		PriceID:          priceID,
	}, nil
}
