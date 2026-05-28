// Package reconcilesubscription implements the defense-by-sync flow that pulls
// Stripe's view of a customer's subscriptions and materialises them into
// public.organization_plans. It is invoked when a Stripe customer is linked to
// a local org for the first time, closing the race window where webhooks may
// have arrived before the org existed (orphan-customer scenario).
package reconcilesubscription

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Input carries the identifiers needed to reconcile billing state for an org.
type Input struct {
	OrgID            uuid.UUID
	StripeCustomerID string
}

// Result reports what the reconcile pass produced. All fields are best-effort
// — none of them should drive control flow on the caller side.
type Result struct {
	SubscriptionsFound      int
	SubscriptionsApplied    int
	DeferredEventsProcessed int
}

// Handler is the application service that orchestrates Stripe → DB
// reconciliation for a single (org, customer) pair.
type Handler struct {
	gateway      services.StripeGateway
	planAssigner services.PlanAssigner
	webhookRepo  repositories.WebhookEventRepository
}

// NewHandler constructs a reconcile handler with its dependencies.
func NewHandler(
	gateway services.StripeGateway,
	planAssigner services.PlanAssigner,
	webhookRepo repositories.WebhookEventRepository,
) *Handler {
	return &Handler{
		gateway:      gateway,
		planAssigner: planAssigner,
		webhookRepo:  webhookRepo,
	}
}

// Reconcile is the narrow signature consumed by the checkout flow's optional
// Reconciler dependency. It delegates to Handle and discards the diagnostic
// Result, returning only the terminal error (or nil for partial success).
func (h *Handler) Reconcile(ctx context.Context, orgID uuid.UUID, stripeCustomerID string) error {
	_, err := h.Handle(ctx, Input{OrgID: orgID, StripeCustomerID: stripeCustomerID})
	return err
}

// Handle reconciles Stripe's view of the customer's subscriptions with the
// local org_plans table. It also marks any deferred webhook events for the
// customer as processed (they have been superseded by the live Stripe data).
//
// The implementation is best-effort: per-step errors are logged but do not
// halt the overall pass, because reconciliation may be invoked from a code
// path that already has a successful Stripe Save behind it and we must not
// block the user's flow on Stripe API hiccups.
func (h *Handler) Handle(ctx context.Context, in Input) (*Result, error) {
	if in.StripeCustomerID == "" {
		return nil, errors.New("reconcile: stripe customer id is required")
	}

	res := &Result{}

	subs, err := h.gateway.ListSubscriptions(ctx, in.StripeCustomerID)
	if err != nil {
		return res, err
	}
	res.SubscriptionsFound = len(subs)

	for _, sub := range subs {
		if !isReconcileableStatus(sub.Status) {
			continue
		}
		billingStatus, _ := entities.BillingStatusFromString(sub.Status)
		assignErr := h.planAssigner.Assign(ctx, services.AssignInput{
			OrgID:                in.OrgID,
			StripeSubscriptionID: sub.ID,
			StripePriceID:        sub.PriceID,
			StripeCustomerID:     in.StripeCustomerID,
			BillingStatus:        billingStatus,
			CurrentPeriodEnd:     time.Unix(sub.CurrentPeriodEnd, 0),
			PaymentStatus:        paymentStatusFor(sub.Status),
		})
		if assignErr != nil {
			logger.Warn("reconcile: plan assign failed",
				zap.String("subscription_id", sub.ID),
				zap.String("price_id", sub.PriceID),
				zap.Error(assignErr),
			)
			continue
		}
		res.SubscriptionsApplied++
	}

	// Mark deferred events processed — they have been superseded by the
	// authoritative Stripe state we just applied. Best-effort.
	deferred, err := h.webhookRepo.FindDeferredByCustomer(ctx, in.StripeCustomerID)
	if err != nil {
		logger.Warn("reconcile: list deferred events failed", zap.Error(err))
		return res, nil //nolint:nilerr — partial success is acceptable
	}
	for _, ev := range deferred {
		if markErr := h.webhookRepo.MarkProcessed(ctx, ev.EventID, entities.WebhookEventProcessed); markErr != nil {
			logger.Warn("reconcile: mark deferred event processed failed",
				zap.String("event_id", ev.EventID),
				zap.Error(markErr),
			)
			continue
		}
		res.DeferredEventsProcessed++
	}

	return res, nil
}

// isReconcileableStatus returns true for Stripe statuses that imply the
// customer is on a paid (or trialing-to-paid) plan we should materialise.
// canceled / incomplete_expired / unpaid are skipped — they do not represent
// an active plan to assign.
func isReconcileableStatus(status string) bool {
	switch strings.ToLower(status) {
	case "active", "trialing", "past_due":
		return true
	}
	return false
}

// paymentStatusFor maps a Stripe subscription status to our payment_status column.
func paymentStatusFor(status string) string {
	if strings.EqualFold(status, "past_due") {
		return "grace_period"
	}
	return "ok"
}
