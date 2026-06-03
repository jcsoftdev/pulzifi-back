package intwiring

import (
	"context"

	"github.com/google/uuid"
)

// ChannelEntitlementAdapter implements integration/domain/services.ChannelEntitlement.
// It uses OrgPlanLookup to fetch the org's active plan code and checks
// membership against the configured paid-plans list.
//
// An org is considered "paid" (all channels allowed) when its active plan code
// is in paidPlans.  An org with no active plan (empty code) or an unrecognised
// plan code is treated as "free" (email-only).
type ChannelEntitlementAdapter struct {
	plans     *OrgPlanLookup
	paidPlans []string
}

// NewChannelEntitlementAdapter constructs the adapter.
// paidPlans should come from config.IntegrationPaidPlans.
func NewChannelEntitlementAdapter(plans *OrgPlanLookup, paidPlans []string) *ChannelEntitlementAdapter {
	return &ChannelEntitlementAdapter{plans: plans, paidPlans: paidPlans}
}

// IsPaid returns true when the org has an active plan whose code is in the
// configured paid-plans list.
func (a *ChannelEntitlementAdapter) IsPaid(ctx context.Context, orgID uuid.UUID) (bool, error) {
	code, err := a.plans.PlanCode(ctx, orgID)
	if err != nil {
		return false, err
	}
	for _, p := range a.paidPlans {
		if p == code {
			return true, nil
		}
	}
	return false, nil
}
