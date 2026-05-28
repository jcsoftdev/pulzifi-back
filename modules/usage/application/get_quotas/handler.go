package getquotas

import (
	"context"
	"fmt"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/services"
)

// Handler handles the get_quotas use case.
type Handler struct {
	usageRepo repositories.UsageTrackingRepository
	orgPlans  repositories.OrganizationPlanRepository
	billing   services.BillingPeriodCalculator
}

// NewHandler creates a new get quotas handler.
func NewHandler(
	usageRepo repositories.UsageTrackingRepository,
	orgPlans repositories.OrganizationPlanRepository,
	billing services.BillingPeriodCalculator,
) *Handler {
	return &Handler{usageRepo: usageRepo, orgPlans: orgPlans, billing: billing}
}

// Handle returns the current billing period quotas for the tenant.
func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	now := time.Now()

	// Find the existing period first
	ut, err := h.usageRepo.FindCurrent(ctx, now)
	if err != nil {
		return nil, fmt.Errorf("find current period: %w", err)
	}

	// If no active period, create one from the org's plan
	if ut == nil {
		planInfo, err := h.orgPlans.GetActivePlanForTenant(ctx, req.Tenant)
		if err != nil {
			return nil, fmt.Errorf("get active plan: %w", err)
		}
		if planInfo == nil {
			return nil, fmt.Errorf("no active plan found for tenant %s", req.Tenant)
		}

		startedAt, ok := planInfo.StartedAt.(time.Time)
		if !ok {
			return nil, fmt.Errorf("invalid plan start date")
		}

		periodStart, periodEnd := h.billing.For(now, startedAt.Day())
		ut = &entities.UsageTracking{
			PeriodStart:       periodStart,
			PeriodEnd:         periodEnd,
			ChecksAllowed:     planInfo.ChecksAllowed,
			StoragePeriodDays: planInfo.StoragePeriodDays,
			AIInsightsAllowed: planInfo.AIInsightsAllowed,
		}

		if insertErr := h.usageRepo.Insert(ctx, ut); insertErr != nil {
			return nil, fmt.Errorf("create usage period: %w", insertErr)
		}
	}

	var refill interface{}
	if ut.NextRefillAt != nil {
		refill = ut.NextRefillAt.UTC().Format(time.RFC3339)
	}

	const unlimitedSentinel = 2147483647
	aiUnlimited := ut.AIInsightsAllowed >= unlimitedSentinel
	aiRemaining := ut.AIInsightsAllowed - ut.AIInsightsUsed
	if aiRemaining < 0 {
		aiRemaining = 0
	}
	if aiUnlimited {
		// Hide the gigantic sentinel from the API surface.
		aiRemaining = -1
	}

	return &Response{
		ChecksUsed:          ut.ChecksUsed,
		ChecksAllowed:       ut.ChecksAllowed,
		NextRefillAt:        refill,
		StoragePeriodDays:   ut.StoragePeriodDays,
		AIInsightsUsed:      ut.AIInsightsUsed,
		AIInsightsAllowed:   ut.AIInsightsAllowed,
		AIInsightsRemaining: aiRemaining,
		AIInsightsUnlimited: aiUnlimited,
		Message:             "get usage quotas",
	}, nil
}
