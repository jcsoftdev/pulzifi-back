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

	// ListSubscriptions
	ListSubscriptionsResult []services.StripeSubscription
	ListSubscriptionsErr    error
	ListSubscriptionsFn     func(ctx context.Context, customerID string) ([]services.StripeSubscription, error)

	// UpdateSubscriptionItem
	UpdateSubscriptionItemResult services.StripeSubscription
	UpdateSubscriptionItemErr    error
	UpdateSubscriptionItemFn     func(ctx context.Context, subID, newPriceID, prorationBehavior string) (services.StripeSubscription, error)

	// PreviewProration
	PreviewProrationAmount   int64
	PreviewProrationCurrency string
	PreviewProrationErr      error
	PreviewProrationFn       func(ctx context.Context, subID, newPriceID string) (int64, string, error)

	// CreateRefundCreditInvoiceItem
	CreateRefundCreditInvoiceItemErr error
	CreateRefundCreditInvoiceItemFn  func(ctx context.Context, customerID string, amountCents int64, currency, description string) error
	LastRefundCreditAmount           int64
	LastRefundCreditCurrency         string
	LastRefundCreditDescription      string

	// UpdateSubscriptionAnchorNow
	UpdateSubscriptionAnchorNowResult services.StripeSubscription
	UpdateSubscriptionAnchorNowErr    error
	UpdateSubscriptionAnchorNowFn     func(ctx context.Context, subID, newPriceID string) (services.StripeSubscription, error)

	// RetrievePriceAmount — key by priceID
	RetrievePriceAmountByID map[string]int64
	RetrievePriceCurrency   string
	RetrievePriceAmountErr  error
	RetrievePriceAmountFn   func(ctx context.Context, priceID string) (int64, string, error)

	// Call counters
	EnsureCustomerCalls            int
	CreateCheckoutSessionCalls     int
	CreatePortalSessionCalls       int
	ConstructEventCalls            int
	RetrieveSubscriptionCalls      int
	RetrieveCustomerBalanceCalls   int
	ListSubscriptionsCalls               int
	UpdateSubscriptionItemCalls          int
	PreviewProrationCalls                int
	CreateRefundCreditInvoiceItemCalls   int
	UpdateSubscriptionAnchorNowCalls     int
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

func (m *MockStripeGateway) ListSubscriptions(ctx context.Context, customerID string) ([]services.StripeSubscription, error) {
	m.ListSubscriptionsCalls++
	if m.ListSubscriptionsFn != nil {
		return m.ListSubscriptionsFn(ctx, customerID)
	}
	return m.ListSubscriptionsResult, m.ListSubscriptionsErr
}

func (m *MockStripeGateway) UpdateSubscriptionItem(ctx context.Context, subID, newPriceID, prorationBehavior string) (services.StripeSubscription, error) {
	m.UpdateSubscriptionItemCalls++
	if m.UpdateSubscriptionItemFn != nil {
		return m.UpdateSubscriptionItemFn(ctx, subID, newPriceID, prorationBehavior)
	}
	return m.UpdateSubscriptionItemResult, m.UpdateSubscriptionItemErr
}

func (m *MockStripeGateway) PreviewProration(ctx context.Context, subID, newPriceID string) (int64, string, error) {
	m.PreviewProrationCalls++
	if m.PreviewProrationFn != nil {
		return m.PreviewProrationFn(ctx, subID, newPriceID)
	}
	return m.PreviewProrationAmount, m.PreviewProrationCurrency, m.PreviewProrationErr
}

func (m *MockStripeGateway) CreateRefundCreditInvoiceItem(ctx context.Context, customerID string, amountCents int64, currency, description string) error {
	m.CreateRefundCreditInvoiceItemCalls++
	m.LastRefundCreditAmount = amountCents
	m.LastRefundCreditCurrency = currency
	m.LastRefundCreditDescription = description
	if m.CreateRefundCreditInvoiceItemFn != nil {
		return m.CreateRefundCreditInvoiceItemFn(ctx, customerID, amountCents, currency, description)
	}
	return m.CreateRefundCreditInvoiceItemErr
}

func (m *MockStripeGateway) UpdateSubscriptionAnchorNow(ctx context.Context, subID, newPriceID string) (services.StripeSubscription, error) {
	m.UpdateSubscriptionAnchorNowCalls++
	if m.UpdateSubscriptionAnchorNowFn != nil {
		return m.UpdateSubscriptionAnchorNowFn(ctx, subID, newPriceID)
	}
	return m.UpdateSubscriptionAnchorNowResult, m.UpdateSubscriptionAnchorNowErr
}

func (m *MockStripeGateway) RetrievePriceAmount(ctx context.Context, priceID string) (int64, string, error) {
	if m.RetrievePriceAmountFn != nil {
		return m.RetrievePriceAmountFn(ctx, priceID)
	}
	if m.RetrievePriceAmountByID != nil {
		if amt, ok := m.RetrievePriceAmountByID[priceID]; ok {
			currency := m.RetrievePriceCurrency
			if currency == "" {
				currency = "usd"
			}
			return amt, currency, m.RetrievePriceAmountErr
		}
	}
	return 0, m.RetrievePriceCurrency, m.RetrievePriceAmountErr
}
