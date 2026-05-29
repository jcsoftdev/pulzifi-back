// Package updatesubscription implements in-place subscription plan changes
// using a USAGE-BASED proration model (Model B):
//
//   1. UsageCredit = priceOld × (checks_remaining / checks_allowed)
//   2. AccountCredit = customer.balance (Stripe holds existing credit here)
//   3. NewPlanCharge = priceNew (full price, new billing cycle starts today)
//   4. AmountDue = max(0, NewPlanCharge - UsageCredit - AccountCredit)
//
// Apply path (when Preview=false):
//   a. Persist UsageCredit as a negative InvoiceItem on the Stripe customer
//   b. Call subscription.Update with billing_cycle_anchor=now and
//      proration_behavior=none so Stripe issues a fresh invoice for the
//      new plan at full price; the negative InvoiceItem reduces the total
//      and Stripe auto-applies customer.balance.
//
// Unlike create_checkout_session (which creates NEW Stripe checkout flows)
// this use case mutates an existing Stripe Subscription via the Subscriptions
// API. No Stripe-hosted UI is involved.
package updatesubscription

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
)

// Sentinel errors.
var (
	ErrInvalidBillingCycle  = errors.New("billing: cycle must be 'monthly' or 'yearly'")
	ErrMissingPriceID       = errors.New("billing: plan has no Stripe price configured for this billing cycle")
	ErrPlanNotFound         = errors.New("billing: plan not found")
	ErrNoActiveSubscription = errors.New("billing: organization has no active stripe subscription")
)

// Handler orchestrates the in-place usage-based plan change.
type Handler struct {
	gateway      services.StripeGateway
	planRepo     repositories.PlanRepository
	subRepo      repositories.SubscriptionRepository
	usageReader  services.UsageReader
	planAssigner services.PlanAssigner
}

// NewHandler wires the dependencies. PlanAssigner is invoked synchronously
// after Stripe accepts the update so the client's subsequent subscription
// refetch reads the new plan immediately, without waiting for the
// customer.subscription.updated webhook to arrive (which can take seconds).
// The webhook is still the authoritative writer and re-applies idempotently.
func NewHandler(
	gateway services.StripeGateway,
	planRepo repositories.PlanRepository,
	subRepo repositories.SubscriptionRepository,
	usageReader services.UsageReader,
	planAssigner services.PlanAssigner,
) *Handler {
	return &Handler{
		gateway:      gateway,
		planRepo:     planRepo,
		subRepo:      subRepo,
		usageReader:  usageReader,
		planAssigner: planAssigner,
	}
}

// Handle runs the use case. See package docstring for sequence.
func (h *Handler) Handle(ctx context.Context, req Request) (*Response, error) {
	if req.BillingCycle != "monthly" && req.BillingCycle != "yearly" {
		return nil, ErrInvalidBillingCycle
	}

	orgID, err := uuid.Parse(req.OrgID)
	if err != nil {
		return nil, ErrNoActiveSubscription
	}

	newPriceID, err := h.resolveNewPriceID(ctx, req)
	if err != nil {
		return nil, err
	}

	sub, err := h.subRepo.FindByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if sub == nil || sub.StripeSubscriptionID == "" {
		return nil, ErrNoActiveSubscription
	}

	resp, snap, err := h.computeProration(ctx, req, orgID, sub, newPriceID)
	if err != nil {
		return nil, err
	}

	if req.Preview {
		return resp, nil
	}

	if err := h.applyPlanChange(ctx, orgID, sub, newPriceID, resp, snap); err != nil {
		return nil, err
	}

	resp.Preview = false
	return resp, nil
}

// resolveNewPriceID resolves the target Stripe price id for the requested plan
// and billing cycle, mapping repository errors to the use case sentinels.
func (h *Handler) resolveNewPriceID(ctx context.Context, req Request) (string, error) {
	newPriceID, err := h.planRepo.FindStripePriceID(ctx, req.PlanID, req.BillingCycle)
	if err != nil {
		if errors.Is(err, repositories.ErrPlanNotFound) {
			return "", ErrPlanNotFound
		}
		return "", err
	}
	if newPriceID == "" {
		return "", ErrMissingPriceID
	}
	return newPriceID, nil
}

// computeProration gathers Stripe prices, usage, and account credit to build the
// preview Response (usage-based Model B). It also returns the usage snapshot so
// the apply path can reuse the remaining-checks figures without re-reading.
func (h *Handler) computeProration(ctx context.Context, req Request, orgID uuid.UUID, sub *entities.Subscription, newPriceID string) (*Response, services.UsageSnapshot, error) {
	// Resolve current price id from Stripe (the Subscription entity does not
	// carry it). Use this to look up the current unit amount for the credit
	// calculation.
	currentSub, err := h.gateway.RetrieveSubscription(ctx, sub.StripeSubscriptionID)
	if err != nil {
		return nil, services.UsageSnapshot{}, fmt.Errorf("retrieve current sub: %w", err)
	}
	oldPriceCents, currency, err := h.gateway.RetrievePriceAmount(ctx, currentSub.PriceID)
	if err != nil {
		return nil, services.UsageSnapshot{}, fmt.Errorf("retrieve old price: %w", err)
	}
	newPriceCents, newCurrency, err := h.gateway.RetrievePriceAmount(ctx, newPriceID)
	if err != nil {
		return nil, services.UsageSnapshot{}, fmt.Errorf("retrieve new price: %w", err)
	}
	if newCurrency != "" {
		currency = newCurrency
	}

	// Usage credit — based on remaining checks of the current plan period.
	snap, err := h.usageReader.ReadActiveUsage(ctx, orgID)
	if err != nil {
		return nil, services.UsageSnapshot{}, fmt.Errorf("read usage: %w", err)
	}
	usageCredit := int64(float64(oldPriceCents) * snap.RemainingRatio())
	if usageCredit < 0 {
		usageCredit = 0
	}
	checksRemaining := snap.ChecksAllowed - snap.ChecksUsed
	if checksRemaining < 0 {
		checksRemaining = 0
	}

	// Account credit — Stripe stores account credit as a NEGATIVE balance.
	// We expose it here as a positive cents value (the "savings" the customer
	// will see applied to this invoice).
	var accountCredit int64
	if sub.StripeCustomerID != "" {
		bal, balCurrency, balErr := h.gateway.RetrieveCustomerBalance(ctx, sub.StripeCustomerID)
		if balErr == nil && bal < 0 {
			accountCredit = -bal
			if balCurrency != "" {
				currency = balCurrency
			}
		}
	}

	amountDue := newPriceCents - usageCredit - accountCredit
	if amountDue < 0 {
		amountDue = 0
	}

	resp := &Response{
		SubscriptionID:     sub.StripeSubscriptionID,
		NewPriceID:         newPriceID,
		BillingCycle:       req.BillingCycle,
		UsageCreditCents:   usageCredit,
		AccountCreditCents: accountCredit,
		NewPlanChargeCents: newPriceCents,
		AmountDueCents:     amountDue,
		Currency:           currency,
		ChecksRemaining:    checksRemaining,
		ChecksAllowed:      snap.ChecksAllowed,
		Preview:            req.Preview,
	}
	return resp, snap, nil
}

// applyPlanChange performs the live Stripe mutation: post the usage credit,
// update the subscription with a fresh billing-cycle anchor, then sync the plan
// locally so the next read reflects the new plan immediately.
func (h *Handler) applyPlanChange(ctx context.Context, orgID uuid.UUID, sub *entities.Subscription, newPriceID string, resp *Response, snap services.UsageSnapshot) error {
	// Apply: create the refund invoice item BEFORE updating the subscription
	// so the negative line is on the next invoice Stripe finalises.
	if resp.UsageCreditCents > 0 {
		description := fmt.Sprintf(
			"Usage-based credit on plan change (%d / %d checks remaining)",
			resp.ChecksRemaining, snap.ChecksAllowed,
		)
		if err := h.gateway.CreateRefundCreditInvoiceItem(ctx, sub.StripeCustomerID, resp.UsageCreditCents, resp.Currency, description); err != nil {
			return fmt.Errorf("apply usage credit: %w", err)
		}
	}

	updated, err := h.gateway.UpdateSubscriptionAnchorNow(ctx, sub.StripeSubscriptionID, newPriceID)
	if err != nil {
		return fmt.Errorf("update subscription: %w", err)
	}

	// Persist locally so the client's next read of /billing/subscription
	// shows the new plan immediately. The webhook will re-apply the same
	// state — PlanAssigner.Assign is idempotent.
	billingStatus, _ := entities.BillingStatusFromString(updated.Status)
	if assignErr := h.planAssigner.Assign(ctx, services.AssignInput{
		OrgID:                orgID,
		StripeSubscriptionID: updated.ID,
		StripePriceID:        updated.PriceID,
		StripeCustomerID:     updated.CustomerID,
		BillingStatus:        billingStatus,
		CurrentPeriodEnd:     time.Unix(updated.CurrentPeriodEnd, 0),
		PaymentStatus:        "ok",
	}); assignErr != nil {
		// Don't fail the request — Stripe state is already mutated. The
		// webhook will re-apply within seconds.
		fmt.Printf("update_subscription: sync plan assign failed (org=%s): %v\n", orgID, assignErr)
	}

	return nil
}
