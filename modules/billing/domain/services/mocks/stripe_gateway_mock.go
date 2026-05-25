package mocks

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
)

// compile-time interface check
var _ services.StripeGateway = (*MockStripeGateway)(nil)

// MockStripeGateway is a hand-rolled test double for services.StripeGateway.
// Tests set the *Result / *Err fields or provide function hooks for fine-grained control.
type MockStripeGateway struct {
	// EnsureCustomer
	EnsureCustomerResult string
	EnsureCustomerErr    error
	EnsureCustomerFn     func(ctx context.Context, orgID, email, name string) (string, error)

	// CreateCheckoutSession
	CreateCheckoutSessionResult string
	CreateCheckoutSessionErr    error
	CreateCheckoutSessionFn     func(ctx context.Context, in services.CheckoutInput) (string, error)

	// CreatePortalSession
	CreatePortalSessionResult string
	CreatePortalSessionErr    error
	CreatePortalSessionFn     func(ctx context.Context, customerID, returnURL string) (string, error)

	// ConstructEvent
	ConstructEventResult services.StripeEvent
	ConstructEventErr    error
	ConstructEventFn     func(payload []byte, signature, secret string) (services.StripeEvent, error)

	// RetrieveSubscription
	RetrieveSubscriptionResult services.StripeSubscription
	RetrieveSubscriptionErr    error
	RetrieveSubscriptionFn     func(ctx context.Context, subID string) (services.StripeSubscription, error)

	// RetrieveCustomerBalance
	RetrieveCustomerBalanceCents    int64
	RetrieveCustomerBalanceCurrency string
	RetrieveCustomerBalanceErr      error
	RetrieveCustomerBalanceFn       func(ctx context.Context, customerID string) (int64, string, error)

	// Call counters
	EnsureCustomerCalls            int
	CreateCheckoutSessionCalls     int
	CreatePortalSessionCalls       int
	ConstructEventCalls            int
	RetrieveSubscriptionCalls      int
	RetrieveCustomerBalanceCalls   int
}

func (m *MockStripeGateway) EnsureCustomer(ctx context.Context, orgID, email, name string) (string, error) {
	m.EnsureCustomerCalls++
	if m.EnsureCustomerFn != nil {
		return m.EnsureCustomerFn(ctx, orgID, email, name)
	}
	return m.EnsureCustomerResult, m.EnsureCustomerErr
}

func (m *MockStripeGateway) CreateCheckoutSession(ctx context.Context, in services.CheckoutInput) (string, error) {
	m.CreateCheckoutSessionCalls++
	if m.CreateCheckoutSessionFn != nil {
		return m.CreateCheckoutSessionFn(ctx, in)
	}
	return m.CreateCheckoutSessionResult, m.CreateCheckoutSessionErr
}

func (m *MockStripeGateway) CreatePortalSession(ctx context.Context, customerID, returnURL string) (string, error) {
	m.CreatePortalSessionCalls++
	if m.CreatePortalSessionFn != nil {
		return m.CreatePortalSessionFn(ctx, customerID, returnURL)
	}
	return m.CreatePortalSessionResult, m.CreatePortalSessionErr
}

func (m *MockStripeGateway) ConstructEvent(payload []byte, signature, secret string) (services.StripeEvent, error) {
	m.ConstructEventCalls++
	if m.ConstructEventFn != nil {
		return m.ConstructEventFn(payload, signature, secret)
	}
	return m.ConstructEventResult, m.ConstructEventErr
}

func (m *MockStripeGateway) RetrieveSubscription(ctx context.Context, subID string) (services.StripeSubscription, error) {
	m.RetrieveSubscriptionCalls++
	if m.RetrieveSubscriptionFn != nil {
		return m.RetrieveSubscriptionFn(ctx, subID)
	}
	return m.RetrieveSubscriptionResult, m.RetrieveSubscriptionErr
}

func (m *MockStripeGateway) RetrieveCustomerBalance(ctx context.Context, customerID string) (int64, string, error) {
	m.RetrieveCustomerBalanceCalls++
	if m.RetrieveCustomerBalanceFn != nil {
		return m.RetrieveCustomerBalanceFn(ctx, customerID)
	}
	return m.RetrieveCustomerBalanceCents, m.RetrieveCustomerBalanceCurrency, m.RetrieveCustomerBalanceErr
}
