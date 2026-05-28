// Package managecoupons implements the SUPER_ADMIN coupon CRUD use cases:
// create a first-month-$1 promo (coupon + promotion code), list all promotion
// codes, and revoke (deactivate) one. Coupon state lives entirely in Stripe —
// no local mirror table.
package managecoupons

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
)

// Sentinels.
var (
	ErrInvalidBillingCycle = errors.New("billing: cycle must be 'monthly' or 'yearly'")
	ErrPlanNotFound        = errors.New("billing: plan not found")
	ErrMissingPriceID      = errors.New("billing: plan has no Stripe price for this cycle")
)

// targetAmountCents is the post-discount price of the first cycle. $1.00.
const targetAmountCents = 100

// Handler bundles the coupon CRUD operations behind one struct so the HTTP
// layer wires a single dependency.
type Handler struct {
	gateway     services.StripeGateway
	planRepo    repositories.PlanRepository
	frontendURL string
}

// NewHandler constructs the coupon management handler.
func NewHandler(gateway services.StripeGateway, planRepo repositories.PlanRepository, frontendURL string) *Handler {
	return &Handler{gateway: gateway, planRepo: planRepo, frontendURL: strings.TrimRight(frontendURL, "/")}
}

// Create resolves the plan price, computes amount_off so the first cycle costs
// $1, creates the Stripe coupon + promotion code, and returns the redeemable
// code plus an apply link.
func (h *Handler) Create(ctx context.Context, req CreateRequest) (*CreateResponse, error) {
	if req.BillingCycle != "monthly" && req.BillingCycle != "yearly" {
		return nil, ErrInvalidBillingCycle
	}

	priceID, err := h.planRepo.FindStripePriceID(ctx, req.PlanCode, req.BillingCycle)
	if err != nil {
		if errors.Is(err, repositories.ErrPlanNotFound) {
			return nil, ErrPlanNotFound
		}
		return nil, err
	}
	if priceID == "" {
		return nil, ErrMissingPriceID
	}

	priceCents, currency, err := h.gateway.RetrievePriceAmount(ctx, priceID)
	if err != nil {
		return nil, fmt.Errorf("retrieve price amount: %w", err)
	}

	amountOff := priceCents - targetAmountCents
	if amountOff < 0 {
		amountOff = 0
	}

	couponID, err := h.gateway.CreateCoupon(ctx, amountOff, currency, services.CouponMetadata{
		PlanCode:     req.PlanCode,
		BillingCycle: req.BillingCycle,
		Purpose:      "first_month_1usd",
	})
	if err != nil {
		return nil, fmt.Errorf("create coupon: %w", err)
	}

	pc, err := h.gateway.CreatePromotionCode(ctx, couponID, strings.ToUpper(strings.TrimSpace(req.Code)), req.MaxRedemptions, req.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("create promotion code: %w", err)
	}

	return &CreateResponse{
		CouponID:        couponID,
		PromotionCodeID: pc.ID,
		Code:            pc.Code,
		AmountOffCents:  pc.AmountOffCents,
		Currency:        pc.Currency,
		ApplyURL:        fmt.Sprintf("%s/settings/billing?promo=%s", h.frontendURL, pc.Code),
	}, nil
}

// List returns all promotion codes for the admin table.
func (h *Handler) List(ctx context.Context) ([]PromotionCodeDTO, error) {
	codes, err := h.gateway.ListPromotionCodes(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]PromotionCodeDTO, 0, len(codes))
	for _, pc := range codes {
		out = append(out, PromotionCodeDTO{
			ID:             pc.ID,
			Code:           pc.Code,
			AmountOffCents: pc.AmountOffCents,
			Currency:       pc.Currency,
			Active:         pc.Active,
			MaxRedemptions: pc.MaxRedemptions,
			TimesRedeemed:  pc.TimesRedeemed,
			ExpiresAt:      pc.ExpiresAt,
			PlanCode:       pc.PlanCode,
			BillingCycle:   pc.BillingCycle,
			ApplyURL:       fmt.Sprintf("%s/settings/billing?promo=%s", h.frontendURL, pc.Code),
		})
	}
	return out, nil
}

// Revoke deactivates a promotion code (Stripe codes cannot be deleted).
func (h *Handler) Revoke(ctx context.Context, promotionCodeID string) error {
	if promotionCodeID == "" {
		return errors.New("billing: promotion code id required")
	}
	return h.gateway.DeactivatePromotionCode(ctx, promotionCodeID)
}
