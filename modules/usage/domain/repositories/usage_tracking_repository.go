package repositories

import (
	"context"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/entities"
)

// UsageTrackingRepository manages usage_tracking rows in the tenant schema.
// Implementations MUST call middleware.GetSetSearchPathSQL(tenant) before queries.
type UsageTrackingRepository interface {
	// FindCurrent returns the active billing period covering today, or nil if none exists.
	FindCurrent(ctx context.Context, today time.Time) (*entities.UsageTracking, error)

	// Insert creates a new usage_tracking row (ON CONFLICT DO NOTHING).
	Insert(ctx context.Context, ut *entities.UsageTracking) error

	// LatestPeriodEnd returns the latest period_end date across all rows.
	LatestPeriodEnd(ctx context.Context) (time.Time, error)

	// SyncChecksAllowed updates checks_allowed for any active period row.
	SyncChecksAllowed(ctx context.Context, checksAllowed int) error

	// CountChecks returns (total, success, failed) from the checks table.
	CountChecks(ctx context.Context) (total, success, failed int, err error)

	// CountPages returns non-deleted page count.
	CountPages(ctx context.Context) (int, error)

	// CountWorkspaces returns non-deleted workspace count.
	CountWorkspaces(ctx context.Context) (int, error)

	// CountAlerts returns alert count.
	CountAlerts(ctx context.Context) (int, error)
}
