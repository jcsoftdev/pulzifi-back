package giftmonth

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

var schemaNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// Handler handles the gift_month use case.
type Handler struct {
	orgPlans  repositories.OrganizationPlanRepository
	usageRepo repositories.UsageTrackingRepository
	billing   services.BillingPeriodCalculator
}

// NewHandler creates a new gift month handler.
func NewHandler(
	orgPlans repositories.OrganizationPlanRepository,
	usageRepo repositories.UsageTrackingRepository,
	billing services.BillingPeriodCalculator,
) *Handler {
	return &Handler{orgPlans: orgPlans, usageRepo: usageRepo, billing: billing}
}

// Handle grants an extra billing period to the organization.
func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	planInfo, err := h.orgPlans.GetActivePlanForOrg(ctx, req.OrgID)
	if err != nil {
		return nil, fmt.Errorf("get active plan: %w", err)
	}
	if planInfo == nil {
		return nil, fmt.Errorf("organization has no active plan")
	}

	if !schemaNameRegex.MatchString(planInfo.SchemaName) {
		return nil, fmt.Errorf("invalid organization schema")
	}

	startedAt, ok := planInfo.StartedAt.(time.Time)
	if !ok {
		return nil, fmt.Errorf("invalid plan start date")
	}

	anchorDay := startedAt.Day()
	now := time.Now()

	// Ensure current period exists
	currentStart, currentEnd := h.billing.For(now, anchorDay)
	if insertErr := h.usageRepo.Insert(ctx, &entities.UsageTracking{
		PeriodStart:   currentStart,
		PeriodEnd:     currentEnd,
		ChecksAllowed: planInfo.ChecksAllowed,
	}); insertErr != nil {
		logger.Warn("gift_month: failed to ensure current period (non-fatal)", zap.Error(insertErr))
	}

	// Find latest period end to gift the next period
	latestEnd, err := h.usageRepo.LatestPeriodEnd(ctx)
	if err != nil {
		return nil, fmt.Errorf("find latest period: %w", err)
	}

	nextPeriodAnchor := latestEnd.AddDate(0, 0, 1)
	giftStart, giftEnd := h.billing.For(nextPeriodAnchor, anchorDay)
	giftNextRefill := giftEnd.AddDate(0, 0, 1)

	if err := h.usageRepo.Insert(ctx, &entities.UsageTracking{
		PeriodStart:   giftStart,
		PeriodEnd:     giftEnd,
		ChecksAllowed: planInfo.ChecksAllowed,
		NextRefillAt:  &giftNextRefill,
	}); err != nil {
		return nil, fmt.Errorf("insert gifted period: %w", err)
	}

	return &Response{
		OrganizationID: req.OrgID,
		GiftedPeriod: GiftedPeriod{
			PeriodStart:   giftStart.Format("2006-01-02"),
			PeriodEnd:     giftEnd.Format("2006-01-02"),
			ChecksAllowed: planInfo.ChecksAllowed,
			NextRefillAt:  giftNextRefill.Format(time.RFC3339),
		},
		Message: "free month gifted successfully",
	}, nil
}
