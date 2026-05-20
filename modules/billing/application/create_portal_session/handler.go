package createportalsession

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
)

// ErrNoStripeCustomer is returned when the org has not completed checkout yet
// and therefore has no Stripe customer ID — the portal cannot be opened.
var ErrNoStripeCustomer = errors.New("billing: org has no Stripe customer; complete checkout first")

// Handler orchestrates portal session creation.
type Handler struct {
	gateway      services.StripeGateway
	customerRepo repositories.CustomerRepository
}

// NewHandler returns a Handler with its dependencies injected.
func NewHandler(gateway services.StripeGateway, customerRepo repositories.CustomerRepository) *Handler {
	return &Handler{gateway: gateway, customerRepo: customerRepo}
}

// Handle runs the create_portal_session use case.
//
//  1. Look up the Stripe customer ID for the org.
//  2. Return ErrNoStripeCustomer if none found.
//  3. Call gateway to create a portal session and return the URL.
func (h *Handler) Handle(ctx context.Context, req Request) (*Response, error) {
	orgID, err := uuid.Parse(req.OrgID)
	if err != nil {
		return nil, errors.New("billing: invalid org ID")
	}

	customer, err := h.customerRepo.FindByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, ErrNoStripeCustomer
	}

	url, err := h.gateway.CreatePortalSession(ctx, customer.StripeCustomerID, req.ReturnURL)
	if err != nil {
		return nil, err
	}

	return &Response{PortalURL: url}, nil
}
