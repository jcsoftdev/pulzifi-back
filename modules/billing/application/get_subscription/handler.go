package getsubscription

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// ErrSubscriptionNotFound is returned when the org has no subscription record.
var ErrSubscriptionNotFound = errors.New("billing: subscription not found for org")

// Handler retrieves the current subscription state for an organisation.
type Handler struct {
	subRepo repositories.SubscriptionRepository
	gateway services.StripeGateway // optional; nil → credit balance is not populated
}

// NewHandler returns a Handler with its dependencies injected.
func NewHandler(subRepo repositories.SubscriptionRepository) *Handler {
	return &Handler{subRepo: subRepo}
}

// WithStripeGateway enables live credit-balance lookups when a customer
// ID is present on the loaded subscription. Returns the handler for chaining.
func (h *Handler) WithStripeGateway(g services.StripeGateway) *Handler {
	h.gateway = g
	return h
}

// Handle runs the get_subscription use case.
//
//  1. Parse orgID.
//  2. Load subscription from repo.
//  3. Return 404-equivalent error if none found.
func (h *Handler) Handle(ctx context.Context, orgIDStr string) (*Response, error) {
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return nil, ErrSubscriptionNotFound
	}

	sub, err := h.subRepo.FindByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, ErrSubscriptionNotFound
	}

	resp := &Response{
		OrgID:                sub.OrgID.String(),
		PlanID:               sub.PlanID.String(),
		PlanCode:             sub.PlanCode,
		PlanName:             sub.PlanName,
		BillingCycle:         sub.BillingCycle,
		BillingStatus:        string(sub.BillingStatus),
		PaymentStatus:        sub.PaymentStatus,
		StripeCustomerID:     sub.StripeCustomerID,
		StripeSubscriptionID: sub.StripeSubscriptionID,
		CurrentPeriodEnd:     sub.CurrentPeriodEnd,
	}

	if h.gateway != nil && sub.StripeCustomerID != "" {
		h.populateCreditBalance(ctx, sub.StripeCustomerID, resp)
	}

	if h.gateway != nil && sub.StripeSubscriptionID != "" {
		h.populateLiveSubscription(ctx, sub.StripeSubscriptionID, resp)
	}

	return resp, nil
}

// populateCreditBalance fetches the live credit balance from Stripe when
// possible. Stripe stores balance as a SIGNED integer: negative = credit
// available to the customer (auto-applied to upcoming invoices). Surface the
// absolute value so the UI can show "$X credit available — applied to next
// invoice" instead of the user thinking they lost their money on a duplicate
// purchase. Non-fatal: failures are logged and the credit fields stay empty.
func (h *Handler) populateCreditBalance(ctx context.Context, customerID string, resp *Response) {
	balance, currency, err := h.gateway.RetrieveCustomerBalance(ctx, customerID)
	if err != nil {
		logger.Warn("billing: failed to fetch customer balance",
			zap.String("customer_id", customerID),
			zap.Error(err),
		)
		return
	}
	if balance < 0 {
		resp.CreditBalanceCents = -balance
		resp.CreditBalanceCurrency = currency
	}
}

// populateLiveSubscription surfaces gift discounts, scheduled cancellation, and
// active plan gifts from the live Stripe subscription. Non-fatal: failures are
// logged and the related fields stay empty.
func (h *Handler) populateLiveSubscription(ctx context.Context, subscriptionID string, resp *Response) {
	live, err := h.gateway.RetrieveSubscription(ctx, subscriptionID)
	if err != nil {
		logger.Warn("billing: failed to fetch live subscription for gift info",
			zap.String("subscription_id", subscriptionID),
			zap.Error(err),
		)
		return
	}

	// Surface an active gift discount so the UI can show a clear "free month"
	// banner instead of a generic credit line. A gift is a subscription
	// discount whose coupon metadata purpose == "gift_one_month".
	if live.DiscountPurpose == "gift_one_month" && live.DiscountAmountOffCents > 0 {
		resp.GiftActive = true
		resp.GiftAmountOffCents = live.DiscountAmountOffCents
	}
	// Surface scheduled cancellation so the UI can show "Cancels on X"
	// + a Resume action.
	resp.CancelAtPeriodEnd = live.CancelAtPeriodEnd
	if live.CancelAt > 0 {
		t := time.Unix(live.CancelAt, 0).UTC()
		resp.CancelAt = &t
	}

	// Surface an active plan gift (free higher-tier plan for a cycle).
	// The gift ends at the current period end, when the schedule
	// reverts to the original plan.
	if live.GiftPlanActive {
		resp.PlanGiftActive = true
		resp.GiftPlanCode = live.GiftPlanCode
		resp.GiftRevertPlanCode = live.GiftRevertPlanCode
		if live.CurrentPeriodEnd > 0 {
			t := time.Unix(live.CurrentPeriodEnd, 0).UTC()
			resp.GiftEndsAt = &t
		}
	}
}
