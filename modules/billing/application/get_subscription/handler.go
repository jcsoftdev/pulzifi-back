package getsubscription

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/repositories"
)

// ErrSubscriptionNotFound is returned when the org has no subscription record.
var ErrSubscriptionNotFound = errors.New("billing: subscription not found for org")

// Handler retrieves the current subscription state for an organisation.
type Handler struct {
	subRepo repositories.SubscriptionRepository
}

// NewHandler returns a Handler with its dependencies injected.
func NewHandler(subRepo repositories.SubscriptionRepository) *Handler {
	return &Handler{subRepo: subRepo}
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

	return &Response{
		OrgID:                sub.OrgID.String(),
		PlanID:               sub.PlanID.String(),
		BillingStatus:        string(sub.BillingStatus),
		PaymentStatus:        sub.PaymentStatus,
		StripeSubscriptionID: sub.StripeSubscriptionID,
		CurrentPeriodEnd:     sub.CurrentPeriodEnd,
	}, nil
}
