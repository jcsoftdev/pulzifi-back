package getquotastatus

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/services"
)

// Handler handles the get-quota-status use case.
// This is a read-only operation — it MUST NOT consume or modify quota.
type Handler struct {
	planLimits services.PlanLimits
}

// NewHandler creates a new Handler.
func NewHandler(planLimits services.PlanLimits) *Handler {
	return &Handler{planLimits: planLimits}
}

// Handle returns the current quota status for the given workspace's organisation.
// Results are scoped to the authenticated tenant (enforced at the HTTP layer).
func (h *Handler) Handle(ctx context.Context, workspaceID uuid.UUID) (*Response, error) {
	// Read limit from plan (REQ-QUOTA-STATUS-01, REQ-QUOTA-PLAN-01)
	checksPerDay, err := h.planLimits.GetChecksPerDay(ctx, workspaceID.String())
	if err != nil {
		return nil, err
	}

	// Read current usage (REQ-QUOTA-STATUS-05 — read-only)
	checksUsed, err := h.planLimits.GetChecksUsedToday(ctx)
	if err != nil {
		return nil, err
	}

	// Build response
	resp := &Response{
		ChecksUsed: checksUsed,
		ResetsAt:   nextUTCMidnight(),
	}

	// 0 = unlimited → nil in response (REQ-QUOTA-PLAN-02, REQ-QUOTA-STATUS-02)
	if checksPerDay > 0 {
		limit := checksPerDay
		resp.ChecksLimit = &limit
	}

	return resp, nil
}

// nextUTCMidnight returns the start of the next calendar day in UTC.
func nextUTCMidnight() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
}
