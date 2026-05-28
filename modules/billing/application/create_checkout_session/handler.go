package createcheckoutsession

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Reconciler is the optional dependency that materialises Stripe state into
// public.organization_plans whenever a Stripe customer is freshly linked to an
// org. Implemented by application/reconcile_subscription.Handler. Kept narrow
// here to avoid an import cycle between application packages.
type Reconciler interface {
	Reconcile(ctx context.Context, orgID uuid.UUID, stripeCustomerID string) error
}

// Sentinel errors returned by the use case.
var (
	ErrInvalidBillingCycle = errors.New("billing: cycle must be 'monthly' or 'yearly'")
	ErrMissingPriceID      = errors.New("billing: plan has no Stripe price configured for this billing cycle")
	ErrPlanNotFound        = errors.New("billing: plan not found")
)

// Handler orchestrates checkout session creation.
type Handler struct {
	gateway      services.StripeGateway
	customerRepo repositories.CustomerRepository
	planRepo     repositories.PlanRepository
	reconciler   Reconciler // optional; nil → reconcile-on-link is skipped
}

// NewHandler returns a Handler with its dependencies injected.
func NewHandler(gateway services.StripeGateway, customerRepo repositories.CustomerRepository, planRepo repositories.PlanRepository) *Handler {
	return &Handler{gateway: gateway, customerRepo: customerRepo, planRepo: planRepo}
}

// WithReconciler wires the reconcile-on-link hook and returns the handler for
// chaining. When set, the Reconciler runs once per checkout invocation after
// CustomerRepository.Save persists a fresh customer linkage.
func (h *Handler) WithReconciler(r Reconciler) *Handler {
	h.reconciler = r
	return h
}

// Handle runs the create_checkout_session use case.
//
//  1. Validate billing cycle.
//  2. Resolve the Stripe price ID from the plan code via PlanRepository.
//  3. Lazily ensure a Stripe customer exists for the org.
//  4. Persist the customer mapping if newly created.
//  5. Create the Stripe checkout session and return the URL.
func (h *Handler) Handle(ctx context.Context, req Request) (*Response, error) {
	if req.BillingCycle != "monthly" && req.BillingCycle != "yearly" {
		return nil, ErrInvalidBillingCycle
	}

	priceID, err := h.planRepo.FindStripePriceID(ctx, req.PlanID, req.BillingCycle)
	if err != nil {
		if errors.Is(err, repositories.ErrPlanNotFound) {
			return nil, ErrPlanNotFound
		}
		return nil, err
	}
	if priceID == "" {
		return nil, ErrMissingPriceID
	}

	customerID, err := h.gateway.EnsureCustomer(ctx, req.OrgID, req.OrgEmail, req.OrgName)
	if err != nil {
		return nil, err
	}

	orgID, err := uuid.Parse(req.OrgID)
	if err == nil {
		existing, _ := h.customerRepo.FindByOrgID(ctx, orgID)
		if existing == nil {
			if saveErr := h.customerRepo.Save(ctx, &entities.Customer{
				OrgID:            orgID,
				StripeCustomerID: customerID,
			}); saveErr != nil {
				logger.Warn("create_checkout_session: persist customer linkage failed",
					zap.String("org_id", req.OrgID),
					zap.String("customer_id", customerID),
					zap.Error(saveErr),
				)
			} else if h.reconciler != nil {
				// Fresh link → reconcile in case Stripe has paid subs we missed
				// (orphan webhooks, guest payments, etc.). Best-effort: log + continue.
				if recErr := h.reconciler.Reconcile(ctx, orgID, customerID); recErr != nil {
					logger.Warn("create_checkout_session: reconcile-on-link failed",
						zap.String("org_id", req.OrgID),
						zap.String("customer_id", customerID),
						zap.Error(recErr),
					)
				}
			}
		}
	}

	// Resolve an optional promo code (apply-link path) to its Stripe
	// promotion code id. Invalid / expired / unknown codes are ignored — the
	// hosted promo-code field stays available so checkout never breaks on a
	// bad code.
	var promotionCodeID string
	if req.PromotionCode != "" {
		pc, pcErr := h.gateway.FindPromotionCodeByCode(ctx, req.PromotionCode)
		if pcErr != nil || !pc.Active {
			logger.Warn("create_checkout_session: promo code ignored",
				zap.String("code", req.PromotionCode),
				zap.Error(pcErr),
			)
		} else {
			promotionCodeID = pc.ID
		}
	}

	url, err := h.gateway.CreateCheckoutSession(ctx, services.CheckoutInput{
		CustomerID:      customerID,
		PriceID:         priceID,
		SuccessURL:      req.SuccessURL,
		CancelURL:       req.CancelURL,
		PromotionCodeID: promotionCodeID,
	})
	if err != nil {
		return nil, err
	}

	return &Response{CheckoutURL: url}, nil
}
