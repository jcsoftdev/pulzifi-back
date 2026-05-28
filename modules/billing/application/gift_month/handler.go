// Package giftmonth implements the SUPER_ADMIN "gift one month" operation.
//
// The gift always uses the plan with the HIGHER access of {current, gifted}:
//
//   - Gifted plan tier > current tier (e.g. Starter user, gift Pro):
//     the customer IMMEDIATELY uses the gifted plan FREE for one cycle via a
//     Stripe subscription schedule (phase 1 = gifted plan trial, phase 2 =
//     auto-revert to the current plan). Their normal plan is effectively
//     paused for the gift month, then resumes.
//
//   - Gifted plan tier <= current tier (e.g. Pro user, gift Starter):
//     we do NOT downgrade. The gift is BANKED as a Stripe customer balance
//     credit worth one month of the gifted plan. The customer keeps their
//     higher plan; the credit auto-applies to upcoming invoices, so their
//     next renewal is prorated and they only pay the remainder.
//
// Balance credit (vs an amount_off coupon) is deliberate: Stripe auto-applies
// it across invoices and carries leftover forward.
package giftmonth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Sentinels.
var (
	ErrNoActiveSubscription = errors.New("billing: organization has no active stripe subscription to gift")
	ErrPlanNotFound         = errors.New("billing: gift plan not found")
	ErrMissingPriceID       = errors.New("billing: gift plan has no monthly Stripe price")
)

// planTier ranks plans by access level. Higher = more access. Unknown → -1.
func planTier(code string) int {
	switch code {
	case "trial":
		return 0
	case "starter":
		return 1
	case "pro":
		return 2
	case "enterprise":
		return 3
	default:
		return -1
	}
}

// Gift modes returned to the caller.
const (
	ModePlanGift      = "plan_gift"      // immediate free use of the gifted (higher) plan, then revert
	ModeBalanceCredit = "balance_credit" // banked credit; customer keeps their (higher) plan
)

// Handler applies a one-month gift, choosing schedule vs balance by tier.
type Handler struct {
	gateway      services.StripeGateway
	subRepo      repositories.SubscriptionRepository
	planRepo     repositories.PlanRepository
	planAssigner services.PlanAssigner
}

// NewHandler constructs the gift handler. planAssigner lets plan-changing gifts
// reflect IMMEDIATELY (synchronous local write) instead of waiting on the
// Stripe webhook — the webhook re-applies idempotently afterwards.
func NewHandler(gateway services.StripeGateway, subRepo repositories.SubscriptionRepository, planRepo repositories.PlanRepository, planAssigner services.PlanAssigner) *Handler {
	return &Handler{gateway: gateway, subRepo: subRepo, planRepo: planRepo, planAssigner: planAssigner}
}

// syncPlan writes the gifted plan locally so the recipient sees it without
// waiting for the customer.subscription.* webhook. Non-fatal: Stripe state is
// already mutated and the webhook re-applies idempotently.
func (h *Handler) syncPlan(ctx context.Context, orgID uuid.UUID, subID, priceID, customerID, status string, periodEnd int64) {
	if h.planAssigner == nil {
		return
	}
	billingStatus, _ := entities.BillingStatusFromString(status)
	if assignErr := h.planAssigner.Assign(ctx, services.AssignInput{
		OrgID:                orgID,
		StripeSubscriptionID: subID,
		StripePriceID:        priceID,
		StripeCustomerID:     customerID,
		BillingStatus:        billingStatus,
		CurrentPeriodEnd:     time.Unix(periodEnd, 0),
		PaymentStatus:        "ok",
	}); assignErr != nil {
		logger.Warn("gift: synchronous plan assign failed (webhook will reconcile)",
			zap.String("org_id", orgID.String()), zap.Error(assignErr))
	}
}

// Response reports the applied gift.
type Response struct {
	OrgID        string `json:"org_id"`
	CustomerID   string `json:"customer_id"`
	GiftPlanCode string `json:"gift_plan_code"`
	// Mode is ModePlanGift (used the gifted plan now) or ModeBalanceCredit
	// (banked as credit). Drives the admin success message + the user banner.
	Mode string `json:"mode"`
	// AmountCents is the gifted plan's monthly value (balance credit amount, or
	// the value of the free month for plan gifts).
	AmountCents int64 `json:"amount_cents"`
	Currency    string `json:"currency"`
	// RevertAt is the Unix timestamp the plan gift ends (0 for balance gifts).
	RevertAt int64  `json:"revert_at"`
	Message  string `json:"message"`
}

// Handle gifts ONE month of planCode, always honoring the higher-access plan.
// orgEmail/orgName are used only when the org has no Stripe customer yet (new /
// trial org) so a customer + free trial subscription can be created.
func (h *Handler) Handle(ctx context.Context, orgID uuid.UUID, planCode, orgEmail, orgName string) (*Response, error) {
	sub, err := h.subRepo.FindByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("lookup subscription: %w", err)
	}

	giftPriceID, err := h.planRepo.FindStripePriceID(ctx, planCode, "monthly")
	if err != nil {
		if errors.Is(err, repositories.ErrPlanNotFound) {
			return nil, ErrPlanNotFound
		}
		return nil, err
	}
	if giftPriceID == "" {
		return nil, ErrMissingPriceID
	}

	monthlyCents, currency, err := h.gateway.RetrievePriceAmount(ctx, giftPriceID)
	if err != nil {
		return nil, fmt.Errorf("retrieve gift plan price: %w", err)
	}
	if monthlyCents <= 0 {
		return nil, fmt.Errorf("billing: computed gift amount is zero")
	}

	hasActiveSub := sub != nil && sub.StripeSubscriptionID != ""

	// Upgrade-gift: existing paid sub AND gifted plan is a higher tier → use the
	// gifted plan free for a cycle (schedule), then revert to the current plan.
	if hasActiveSub && planTier(planCode) > planTier(sub.PlanCode) {
		currentSub, err := h.gateway.RetrieveSubscription(ctx, sub.StripeSubscriptionID)
		if err != nil {
			return nil, fmt.Errorf("retrieve current sub: %w", err)
		}
		if currentSub.PriceID == "" {
			return nil, fmt.Errorf("billing: current subscription has no price to revert to")
		}
		revertAt, err := h.gateway.GiftPlanSchedule(ctx, sub.StripeSubscriptionID, giftPriceID, currentSub.PriceID, planCode, sub.PlanCode)
		if err != nil {
			return nil, fmt.Errorf("schedule plan gift: %w", err)
		}
		// Reflect the gifted plan immediately (trialing during the free phase).
		h.syncPlan(ctx, orgID, sub.StripeSubscriptionID, giftPriceID, sub.StripeCustomerID, "trialing", revertAt)
		return &Response{
			OrgID:        orgID.String(),
			CustomerID:   sub.StripeCustomerID,
			GiftPlanCode: planCode,
			Mode:         ModePlanGift,
			AmountCents:  monthlyCents,
			Currency:     currency,
			RevertAt:     revertAt,
			Message:      fmt.Sprintf("gifted one free month of %s — active now, reverts to %s when it ends", planCode, sub.PlanCode),
		}, nil
	}

	// Bank-gift: existing paid sub, gifted plan same/lower tier → keep the
	// higher plan, credit the balance so the next invoice is prorated.
	if hasActiveSub {
		description := fmt.Sprintf("Gift: one month of %s plan (banked credit)", planCode)
		if err := h.gateway.CreditCustomerBalance(ctx, sub.StripeCustomerID, monthlyCents, currency, description); err != nil {
			return nil, fmt.Errorf("credit gift balance: %w", err)
		}
		return &Response{
			OrgID:        orgID.String(),
			CustomerID:   sub.StripeCustomerID,
			GiftPlanCode: planCode,
			Mode:         ModeBalanceCredit,
			AmountCents:  monthlyCents,
			Currency:     currency,
			Message:      fmt.Sprintf("gifted one month of %s — banked as credit, applied to upcoming invoices", planCode),
		}, nil
	}

	// New-user gift: no active subscription (trial / freshly invited org). Make
	// sure the org has a Stripe customer, then start a FREE one-month trial of
	// the gifted plan that auto-cancels at the end (no payment method).
	customerID, err := h.gateway.EnsureCustomer(ctx, orgID.String(), orgEmail, orgName)
	if err != nil {
		return nil, fmt.Errorf("ensure customer for gift: %w", err)
	}
	giftSub, err := h.gateway.CreateGiftSubscription(ctx, customerID, giftPriceID, planCode)
	if err != nil {
		return nil, fmt.Errorf("create gift subscription: %w", err)
	}
	// Reflect the gifted plan immediately so the recipient sees it now.
	status := giftSub.Status
	if status == "" {
		status = "trialing"
	}
	h.syncPlan(ctx, orgID, giftSub.ID, giftPriceID, customerID, status, giftSub.CurrentPeriodEnd)
	return &Response{
		OrgID:        orgID.String(),
		CustomerID:   customerID,
		GiftPlanCode: planCode,
		Mode:         ModePlanGift,
		AmountCents:  monthlyCents,
		Currency:     currency,
		RevertAt:     giftSub.CurrentPeriodEnd,
		Message:      fmt.Sprintf("gifted one free month of %s — active now, ends automatically (no charge)", planCode),
	}, nil
}
