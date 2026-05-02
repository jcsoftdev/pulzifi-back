package intwiring

import (
	"context"

	"github.com/google/uuid"
)

// TwilioAllowedFor returns the AllowedFunc for Twilio quota tracking.
// Looks up the org's active plan; returns paidQuota if the plan code is in
// PaidPlans, else 0 (free tier — calls would be rejected upstream by tier
// check but defensively here too).
func TwilioAllowedFor(plans *OrgPlanLookup, paidPlans []string, paidQuota int) func(ctx context.Context, orgID uuid.UUID) (int, error) {
	return func(ctx context.Context, orgID uuid.UUID) (int, error) {
		code, err := plans.PlanCode(ctx, orgID)
		if err != nil {
			return 0, err
		}
		for _, p := range paidPlans {
			if p == code {
				return paidQuota, nil
			}
		}
		return 0, nil
	}
}
