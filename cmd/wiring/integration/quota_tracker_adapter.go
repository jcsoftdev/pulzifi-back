package intwiring

import (
	"context"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/shared/integrationusage"
)

// AllowedFor binds (ctx, orgID) -> (allowed, error) for a specific service.
type AllowedFor func(ctx context.Context, orgID uuid.UUID) (int, error)

// TwilioQuotaAdapter satisfies twilioprovider.QuotaTracker by currying the
// AllowedFor callback into integrationusage.Tracker.CheckAndIncrement.
type TwilioQuotaAdapter struct {
	tracker    *integrationusage.Tracker
	allowedFor AllowedFor
}

func NewTwilioQuotaAdapter(tracker *integrationusage.Tracker, allowedFor AllowedFor) *TwilioQuotaAdapter {
	return &TwilioQuotaAdapter{tracker: tracker, allowedFor: allowedFor}
}

func (a *TwilioQuotaAdapter) CheckAndIncrement(ctx context.Context, orgID uuid.UUID, serviceType string) error {
	return a.tracker.CheckAndIncrement(ctx, orgID, serviceType, integrationusage.AllowedFunc(a.allowedFor))
}
