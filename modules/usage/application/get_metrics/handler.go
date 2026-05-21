package getmetrics

import (
	"context"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Handler handles the get_metrics use case.
type Handler struct {
	usageRepo repositories.UsageTrackingRepository
	orgPlans  repositories.OrganizationPlanRepository
	billing   services.BillingPeriodCalculator
}

// NewHandler creates a new get metrics handler.
func NewHandler(
	usageRepo repositories.UsageTrackingRepository,
	orgPlans repositories.OrganizationPlanRepository,
	billing services.BillingPeriodCalculator,
) *Handler {
	return &Handler{usageRepo: usageRepo, orgPlans: orgPlans, billing: billing}
}

// Handle fetches all usage metrics for the given tenant.
func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	total, success, failed, _ := h.usageRepo.CountChecks(ctx)
	pages, _ := h.usageRepo.CountPages(ctx)
	workspaces, _ := h.usageRepo.CountWorkspaces(ctx)
	alerts, _ := h.usageRepo.CountAlerts(ctx)

	var checksUsed, checksAllowed int
	if ut, err := h.usageRepo.FindCurrent(ctx, time.Now()); err == nil && ut != nil {
		checksUsed = ut.ChecksUsed
		checksAllowed = ut.ChecksAllowed
	} else if err != nil {
		logger.Warn("get_metrics: failed to find current usage period", zap.Error(err), zap.String("tenant", req.Tenant))
	}

	return &Response{
		Checks:        CheckMetrics{Total: total, Success: success, Failed: failed},
		Pages:         pages,
		Workspaces:    workspaces,
		Alerts:        alerts,
		ChecksUsed:    checksUsed,
		ChecksAllowed: checksAllowed,
	}, nil
}
