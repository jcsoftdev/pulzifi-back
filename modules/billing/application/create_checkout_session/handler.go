package createcheckoutsession

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
)

// Sentinel errors returned by the use case.
var (
	ErrInvalidBillingCycle = errors.New("billing: cycle must be 'monthly' or 'yearly'")
	ErrMissingPriceID      = errors.New("billing: plan has no Stripe price configured for this billing cycle")
)

// Handler orchestrates checkout session creation.
type Handler struct {
	gateway      services.StripeGateway
	customerRepo repositories.CustomerRepository
}

// NewHandler returns a Handler with its dependencies injected.
func NewHandler(gateway services.StripeGateway, customerRepo repositories.CustomerRepository) *Handler {
	return &Handler{gateway: gateway, customerRepo: customerRepo}
}

// Handle runs the create_checkout_session use case.
//
//  1. Validate billing cycle.
//  2. Resolve the Stripe price ID from the caller-supplied plan fields.
//  3. Lazily ensure a Stripe customer exists for the org.
//  4. Persist the customer mapping if newly created.
//  5. Create the Stripe checkout session and return the URL.
func (h *Handler) Handle(ctx context.Context, req Request) (*Response, error) {
	// 1. Validate billing cycle
	if req.BillingCycle != "monthly" && req.BillingCycle != "yearly" {
		return nil, ErrInvalidBillingCycle
	}

	// 2. Resolve price ID
	var priceID string
	switch req.BillingCycle {
	case "monthly":
		priceID = req.StripePriceIDMonthly
	case "yearly":
		priceID = req.StripePriceIDYearly
	}
	if priceID == "" {
		return nil, ErrMissingPriceID
	}

	// 3. Lazy customer creation
	customerID, err := h.gateway.EnsureCustomer(ctx, req.OrgID, req.OrgEmail, req.OrgName)
	if err != nil {
		return nil, err
	}

	// 4. Persist customer mapping if not already stored
	orgID, err := uuid.Parse(req.OrgID)
	if err == nil { // if OrgID is a valid UUID, persist
		existing, _ := h.customerRepo.FindByOrgID(ctx, orgID)
		if existing == nil {
			_ = h.customerRepo.Save(ctx, &entities.Customer{
				OrgID:            orgID,
				StripeCustomerID: customerID,
			})
		}
	}

	// 5. Create checkout session
	url, err := h.gateway.CreateCheckoutSession(ctx, services.CheckoutInput{
		CustomerID: customerID,
		PriceID:    priceID,
		SuccessURL: req.SuccessURL,
		CancelURL:  req.CancelURL,
	})
	if err != nil {
		return nil, err
	}

	return &Response{CheckoutURL: url}, nil
}
