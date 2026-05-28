// Package cancelsubscription implements the user-initiated "cancel my plan"
// operation. Cancellation is scheduled for the END of the current paid period
// (cancel_at_period_end=true): the org keeps full access until then. When the
// period closes Stripe emits customer.subscription.deleted, which deactivates
// the local plan (no replacement → no active plan / read-only). Resume clears a
// pending cancellation before it takes effect.
package cancelsubscription

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
)

// ErrNoActiveSubscription is returned when the org has no Stripe subscription to
// cancel or resume.
var ErrNoActiveSubscription = errors.New("billing: organization has no active stripe subscription")

// Handler schedules or clears a period-end cancellation.
type Handler struct {
	gateway services.StripeGateway
	subRepo repositories.SubscriptionRepository
}

// NewHandler constructs the cancel handler.
func NewHandler(gateway services.StripeGateway, subRepo repositories.SubscriptionRepository) *Handler {
	return &Handler{gateway: gateway, subRepo: subRepo}
}

// Response reports the resulting cancellation state.
type Response struct {
	SubscriptionID    string     `json:"subscription_id"`
	CancelAtPeriodEnd bool       `json:"cancel_at_period_end"`
	CancelAt          *time.Time `json:"cancel_at"`
	Message           string     `json:"message"`
}

// Cancel schedules the org's subscription to end at the current period close.
func (h *Handler) Cancel(ctx context.Context, orgID uuid.UUID) (*Response, error) {
	sub, err := h.resolveSub(ctx, orgID)
	if err != nil {
		return nil, err
	}

	updated, err := h.gateway.CancelSubscriptionAtPeriodEnd(ctx, sub.StripeSubscriptionID)
	if err != nil {
		return nil, fmt.Errorf("cancel subscription: %w", err)
	}

	return toResponse(updated, "plan will be canceled at the end of the current billing period"), nil
}

// Resume clears a pending period-end cancellation, restoring normal renewal.
func (h *Handler) Resume(ctx context.Context, orgID uuid.UUID) (*Response, error) {
	sub, err := h.resolveSub(ctx, orgID)
	if err != nil {
		return nil, err
	}

	updated, err := h.gateway.ResumeSubscription(ctx, sub.StripeSubscriptionID)
	if err != nil {
		return nil, fmt.Errorf("resume subscription: %w", err)
	}

	return toResponse(updated, "plan cancellation reverted — your subscription will renew normally"), nil
}

func (h *Handler) resolveSub(ctx context.Context, orgID uuid.UUID) (subInfo, error) {
	sub, err := h.subRepo.FindByOrgID(ctx, orgID)
	if err != nil {
		return subInfo{}, fmt.Errorf("lookup subscription: %w", err)
	}
	if sub == nil || sub.StripeSubscriptionID == "" {
		return subInfo{}, ErrNoActiveSubscription
	}
	return subInfo{StripeSubscriptionID: sub.StripeSubscriptionID}, nil
}

type subInfo struct {
	StripeSubscriptionID string
}

func toResponse(s services.StripeSubscription, msg string) *Response {
	resp := &Response{
		SubscriptionID:    s.ID,
		CancelAtPeriodEnd: s.CancelAtPeriodEnd,
		Message:           msg,
	}
	if s.CancelAt > 0 {
		t := time.Unix(s.CancelAt, 0).UTC()
		resp.CancelAt = &t
	}
	return resp
}
