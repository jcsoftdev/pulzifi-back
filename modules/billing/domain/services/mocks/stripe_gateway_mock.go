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

	// Cancel / resume
	CancelSubscriptionResult services.StripeSubscription
	CancelSubscriptionErr    error
	CancelSubscriptionFn     func(ctx context.Context, subID string) (services.StripeSubscription, error)
	ResumeSubscriptionResult services.StripeSubscription
	ResumeSubscriptionErr    error
	ResumeSubscriptionFn     func(ctx context.Context, subID string) (services.StripeSubscription, error)

	// CreditCustomerBalance
	CreditCustomerBalanceErr error
	CreditCustomerBalanceFn  func(ctx context.Context, customerID string, amountCents int64, currency, description string) error
	CreditCustomerBalanceCalls int
	LastCreditCustomerID       string
	LastCreditAmountCents      int64
	LastCreditCurrency         string

	// RemoveSubscriptionDiscount
	RemoveSubscriptionDiscountErr error
	RemoveSubscriptionDiscountFn  func(ctx context.Context, subID string) error
	RemoveDiscountCalls           int
	LastRemoveDiscountSubID       string

	// GiftPlanSchedule
	GiftPlanScheduleRevertAt    int64
	GiftPlanScheduleErr         error
	GiftPlanScheduleFn          func(ctx context.Context, subID, giftPriceID, currentPriceID, giftCode, currentCode string) (int64, error)
	GiftPlanScheduleCalls       int
	LastGiftScheduleSubID       string
	LastGiftScheduleGiftPriceID string
	LastGiftScheduleGiftCode    string

	// CreateGiftSubscription
	CreateGiftSubscriptionResult services.StripeSubscription
	CreateGiftSubscriptionErr    error
	CreateGiftSubscriptionFn     func(ctx context.Context, customerID, giftPriceID, giftCode string) (services.StripeSubscription, error)
	CreateGiftSubscriptionCalls  int
	LastGiftSubCustomerID        string
	LastGiftSubPriceID           string
	LastGiftSubCode              string

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

	// CreateCoupon
	CreateCouponResult string
	CreateCouponErr    error
	CreateCouponFn     func(ctx context.Context, amountOffCents int64, currency string, meta services.CouponMetadata) (string, error)
	LastCouponAmount   int64

	// ApplyCouponToSubscription
	ApplyCouponToSubscriptionErr error
	ApplyCouponToSubscriptionFn  func(ctx context.Context, subID, couponID string) error
	LastApplySubID               string
	LastApplyCouponID            string

	// CreatePromotionCode
	CreatePromotionCodeResult services.PromotionCode
	CreatePromotionCodeErr    error
	CreatePromotionCodeFn     func(ctx context.Context, couponID, code string, maxRedemptions, expiresAt int64) (services.PromotionCode, error)

	// ListPromotionCodes
	ListPromotionCodesResult []services.PromotionCode
	ListPromotionCodesErr    error
	ListPromotionCodesFn     func(ctx context.Context) ([]services.PromotionCode, error)

	// DeactivatePromotionCode
	DeactivatePromotionCodeErr error
	DeactivatePromotionCodeFn  func(ctx context.Context, promotionCodeID string) error
	LastDeactivatedPromoID     string

	// FindPromotionCodeByCode
	FindPromotionCodeByCodeResult services.PromotionCode
	FindPromotionCodeByCodeErr    error
	FindPromotionCodeByCodeFn     func(ctx context.Context, code string) (services.PromotionCode, error)

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
	CreateCouponCalls                    int
	CreatePromotionCodeCalls             int
	ListPromotionCodesCalls              int
	DeactivatePromotionCodeCalls         int
	FindPromotionCodeByCodeCalls         int
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

func (m *MockStripeGateway) GiftPlanSchedule(ctx context.Context, subID, giftPriceID, currentPriceID, giftCode, currentCode string) (int64, error) {
	m.GiftPlanScheduleCalls++
	m.LastGiftScheduleSubID = subID
	m.LastGiftScheduleGiftPriceID = giftPriceID
	m.LastGiftScheduleGiftCode = giftCode
	if m.GiftPlanScheduleFn != nil {
		return m.GiftPlanScheduleFn(ctx, subID, giftPriceID, currentPriceID, giftCode, currentCode)
	}
	return m.GiftPlanScheduleRevertAt, m.GiftPlanScheduleErr
}

func (m *MockStripeGateway) CreateGiftSubscription(ctx context.Context, customerID, giftPriceID, giftCode string) (services.StripeSubscription, error) {
	m.CreateGiftSubscriptionCalls++
	m.LastGiftSubCustomerID = customerID
	m.LastGiftSubPriceID = giftPriceID
	m.LastGiftSubCode = giftCode
	if m.CreateGiftSubscriptionFn != nil {
		return m.CreateGiftSubscriptionFn(ctx, customerID, giftPriceID, giftCode)
	}
	return m.CreateGiftSubscriptionResult, m.CreateGiftSubscriptionErr
}

func (m *MockStripeGateway) CreditCustomerBalance(ctx context.Context, customerID string, amountCents int64, currency, description string) error {
	m.CreditCustomerBalanceCalls++
	m.LastCreditCustomerID = customerID
	m.LastCreditAmountCents = amountCents
	m.LastCreditCurrency = currency
	if m.CreditCustomerBalanceFn != nil {
		return m.CreditCustomerBalanceFn(ctx, customerID, amountCents, currency, description)
	}
	return m.CreditCustomerBalanceErr
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

func (m *MockStripeGateway) CreateCoupon(ctx context.Context, amountOffCents int64, currency string, meta services.CouponMetadata) (string, error) {
	m.CreateCouponCalls++
	m.LastCouponAmount = amountOffCents
	if m.CreateCouponFn != nil {
		return m.CreateCouponFn(ctx, amountOffCents, currency, meta)
	}
	return m.CreateCouponResult, m.CreateCouponErr
}

func (m *MockStripeGateway) CreatePromotionCode(ctx context.Context, couponID, code string, maxRedemptions, expiresAt int64) (services.PromotionCode, error) {
	m.CreatePromotionCodeCalls++
	if m.CreatePromotionCodeFn != nil {
		return m.CreatePromotionCodeFn(ctx, couponID, code, maxRedemptions, expiresAt)
	}
	return m.CreatePromotionCodeResult, m.CreatePromotionCodeErr
}

func (m *MockStripeGateway) ListPromotionCodes(ctx context.Context) ([]services.PromotionCode, error) {
	m.ListPromotionCodesCalls++
	if m.ListPromotionCodesFn != nil {
		return m.ListPromotionCodesFn(ctx)
	}
	return m.ListPromotionCodesResult, m.ListPromotionCodesErr
}

func (m *MockStripeGateway) DeactivatePromotionCode(ctx context.Context, promotionCodeID string) error {
	m.DeactivatePromotionCodeCalls++
	m.LastDeactivatedPromoID = promotionCodeID
	if m.DeactivatePromotionCodeFn != nil {
		return m.DeactivatePromotionCodeFn(ctx, promotionCodeID)
	}
	return m.DeactivatePromotionCodeErr
}

func (m *MockStripeGateway) FindPromotionCodeByCode(ctx context.Context, code string) (services.PromotionCode, error) {
	m.FindPromotionCodeByCodeCalls++
	if m.FindPromotionCodeByCodeFn != nil {
		return m.FindPromotionCodeByCodeFn(ctx, code)
	}
	return m.FindPromotionCodeByCodeResult, m.FindPromotionCodeByCodeErr
}

func (m *MockStripeGateway) ApplyCouponToSubscription(ctx context.Context, subID, couponID string) error {
	m.LastApplySubID = subID
	m.LastApplyCouponID = couponID
	if m.ApplyCouponToSubscriptionFn != nil {
		return m.ApplyCouponToSubscriptionFn(ctx, subID, couponID)
	}
	return m.ApplyCouponToSubscriptionErr
}

func (m *MockStripeGateway) CancelSubscriptionAtPeriodEnd(ctx context.Context, subID string) (services.StripeSubscription, error) {
	if m.CancelSubscriptionFn != nil {
		return m.CancelSubscriptionFn(ctx, subID)
	}
	return m.CancelSubscriptionResult, m.CancelSubscriptionErr
}

func (m *MockStripeGateway) ResumeSubscription(ctx context.Context, subID string) (services.StripeSubscription, error) {
	if m.ResumeSubscriptionFn != nil {
		return m.ResumeSubscriptionFn(ctx, subID)
	}
	return m.ResumeSubscriptionResult, m.ResumeSubscriptionErr
}

func (m *MockStripeGateway) RemoveSubscriptionDiscount(ctx context.Context, subID string) error {
	m.RemoveDiscountCalls++
	m.LastRemoveDiscountSubID = subID
	if m.RemoveSubscriptionDiscountFn != nil {
		return m.RemoveSubscriptionDiscountFn(ctx, subID)
	}
	return m.RemoveSubscriptionDiscountErr
}
