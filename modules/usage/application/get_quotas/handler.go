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

	// Find the existing period first.
	ut, err := h.usageRepo.FindCurrent(ctx, now)
	if err != nil {
		return nil, fmt.Errorf("find current period: %w", err)
	}

	// Always fetch the active plan so we can expose MaxPages/MaxWorkspaces
	// regardless of whether a usage_tracking period already exists.
	const unlimitedSentinel = 2147483647
	maxPages := unlimitedSentinel
	maxWorkspaces := unlimitedSentinel

	planInfo, planErr := h.orgPlans.GetActivePlanForTenant(ctx, req.Tenant)
	if planErr != nil {
		// If we already have a period, we can still return quota data without
		// the plan limits — default to unlimited sentinel rather than failing.
		if ut == nil {
			return nil, fmt.Errorf("get active plan: %w", planErr)
		}
		// ut != nil: degrade gracefully, limits stay at sentinel.
	} else if planInfo != nil {
		maxPages = planInfo.MaxPages
		maxWorkspaces = planInfo.MaxWorkspaces
	}

	// If no active period, create one from the plan.
	if ut == nil {
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
		MaxPages:            maxPages,
		MaxWorkspaces:       maxWorkspaces,
		Message:             "get usage quotas",
	}, nil
}
