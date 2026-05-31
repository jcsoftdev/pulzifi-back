package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ExpiredCheck holds the minimal data the retention job needs to clean up
// one check's object storage artifacts and nullify their DB references.
type ExpiredCheck struct {
	ID              uuid.UUID
	ScreenshotURL   string
	HTMLSnapshotURL string
	DiffImageURL    string
}

// RetentionRepository is the port for querying and clearing snapshot storage
// references that have exceeded the tenant's retention period.
type RetentionRepository interface {
	// GetStoragePeriodDays returns the configured retention period for the given
	// tenant schema. Returns DefaultRetentionDays when the row is absent or zero.
	GetStoragePeriodDays(ctx context.Context) (int, error)

	// ListExpiredChecks returns all checks for the tenant whose checked_at is
	// older than cutoff and that still have at least one storage URL set.
	ListExpiredChecks(ctx context.Context, cutoff time.Time) ([]ExpiredCheck, error)

	// NullifyStorageURLs clears the screenshot_url, html_snapshot_url, and
	// diff_image_url columns for the given check IDs in a single update.
	NullifyStorageURLs(ctx context.Context, ids []uuid.UUID) error
}
