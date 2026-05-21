package trialstatus

import (
	"context"
	"math"
	"time"
)

// PlanSnapshot is the minimal view this use case needs about an organisation's
// currently active plan row. Implementations live in the usage module's
// infrastructure layer; tests use an inmem fake.
type PlanSnapshot struct {
	PlanCode    string
	TrialEndsAt *time.Time
	ConvertedAt *time.Time
}

// OrgPlanReader is the port the use case depends on. It resolves the active
// organization_plans row for a given tenant (subdomain).
type OrgPlanReader interface {
	ActivePlanBySubdomain(ctx context.Context, subdomain string) (*PlanSnapshot, error)
}

// Request carries the resolved tenant for which to compute the trial status.
type Request struct {
	Subdomain string
}

// Handler computes the public trial status for a tenant.
type Handler struct {
	reader OrgPlanReader
	nowFn  func() time.Time
}

// NewHandler constructs the trial-status handler.
func NewHandler(reader OrgPlanReader) *Handler {
	return &Handler{reader: reader, nowFn: time.Now}
}

// Handle reads the active organization_plans row and derives the trial state.
//
// Semantics:
//   - No active plan, or plan code != "trial": IsTrial=false, no other flags set.
//   - converted_at != NULL: IsTrial=false, Converted=true (legacy trial that paid).
//   - trial_ends_at < now() AND converted_at IS NULL: IsExpired=true,
//     NeedsUpgrade=true, DaysRemaining=0.
//   - otherwise: IsTrial=true with positive DaysRemaining.
func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	snap, err := h.reader.ActivePlanBySubdomain(ctx, req.Subdomain)
	if err != nil {
		return nil, err
	}

	resp := &Response{}
	if snap == nil || snap.PlanCode != "trial" {
		// Not on the trial plan — nothing else to report.
		return resp, nil
	}

	if snap.ConvertedAt != nil {
		resp.Converted = true
		return resp, nil
	}

	resp.IsTrial = true
	if snap.TrialEndsAt != nil {
		resp.TrialEndsAt = snap.TrialEndsAt
		remaining := snap.TrialEndsAt.Sub(h.nowFn())
		if remaining <= 0 {
			resp.IsExpired = true
			resp.NeedsUpgrade = true
			resp.DaysRemaining = 0
		} else {
			// Ceil so that a few-hour remainder still surfaces as "1 day left".
			resp.DaysRemaining = int(math.Ceil(remaining.Hours() / 24))
		}
	}
	return resp, nil
}
