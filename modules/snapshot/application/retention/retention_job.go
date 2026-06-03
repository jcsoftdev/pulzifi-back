// Package retention implements the snapshot storage retention job.
// It iterates every known tenant schema, finds checks whose checked_at
// is older than the tenant's storage_period_days, deletes the associated
// objects from object storage, and clears the storage URL columns in the DB.
package retention

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// RetentionJobDeps holds the dependencies for RetentionJob.
type RetentionJobDeps struct {
	// RepoFactory returns a RetentionRepository scoped to the given tenant schema.
	RepoFactory func(schemaName string) repositories.RetentionRepository
	// Storage is the shared object storage client used to delete objects.
	Storage repositories.ObjectStorage
	// Tenants is the list of tenant schema names to process on each run.
	// The list is refreshed by the scheduler before each RunOnce call.
	Tenants []string
	// DefaultRetentionDays is used when a tenant's storage_period_days is zero or unset.
	DefaultRetentionDays int
}

// RetentionJob deletes expired snapshot objects across all tenant schemas.
type RetentionJob struct {
	deps RetentionJobDeps
}

// NewRetentionJob creates a RetentionJob from the given deps.
func NewRetentionJob(deps RetentionJobDeps) *RetentionJob {
	if deps.DefaultRetentionDays <= 0 {
		deps.DefaultRetentionDays = 7
	}
	return &RetentionJob{deps: deps}
}

// SetTenants replaces the tenant list used on the next RunOnce call.
// This allows the scheduler to refresh the list before each sweep without
// reconstructing the job.
func (j *RetentionJob) SetTenants(tenants []string) {
	j.deps.Tenants = tenants
}

// RunOnce performs one retention sweep across all configured tenant schemas.
// Each tenant is processed sequentially to avoid overwhelming the storage backend.
// Do not call concurrently with SetTenants; the scheduler's single ticker goroutine
// guarantees this in normal operation.
func (j *RetentionJob) RunOnce(ctx context.Context) error {
	for _, tenant := range j.deps.Tenants {
		if err := j.processTenant(ctx, tenant); err != nil {
			// Log and continue — a single bad tenant must not block others.
			logger.Warn("retention: tenant sweep failed",
				zap.String("tenant", tenant),
				zap.Error(err))
		}
	}
	return nil
}

// processTenant runs the retention sweep for one tenant schema.
func (j *RetentionJob) processTenant(ctx context.Context, tenant string) error {
	repo := j.deps.RepoFactory(tenant)

	days, err := repo.GetStoragePeriodDays(ctx)
	if err != nil {
		return err
	}
	if days <= 0 {
		days = j.deps.DefaultRetentionDays
	}

	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	checks, err := repo.ListExpiredChecks(ctx, cutoff)
	if err != nil {
		return err
	}
	if len(checks) == 0 {
		return nil
	}

	logger.Info("retention: processing expired checks",
		zap.String("tenant", tenant),
		zap.Int("count", len(checks)),
		zap.Int("storage_period_days", days))

	var succeeded []uuid.UUID
	for _, check := range checks {
		if j.deleteCheckObjects(ctx, check) {
			succeeded = append(succeeded, check.ID)
		}
	}

	if len(succeeded) == 0 {
		return nil
	}

	if err := repo.NullifyStorageURLs(ctx, succeeded); err != nil {
		logger.Error("retention: failed to nullify storage URLs",
			zap.String("tenant", tenant),
			zap.Int("count", len(succeeded)),
			zap.Error(err))
		return err
	}

	logger.Info("retention: nullified storage URLs",
		zap.String("tenant", tenant),
		zap.Int("count", len(succeeded)))

	return nil
}

// deleteCheckObjects deletes all storage objects for a single check.
// Returns true if all deletions succeeded (or there were no URLs to delete),
// false if any deletion failed (in which case the DB reference is preserved).
func (j *RetentionJob) deleteCheckObjects(ctx context.Context, check repositories.ExpiredCheck) bool {
	urls := []string{check.ScreenshotURL, check.HTMLSnapshotURL, check.DiffImageURL}
	for _, u := range urls {
		if u == "" {
			continue
		}
		// Delegate URL-to-object-name translation to the storage backend.
		// Each implementation (minio, cloudinary) knows its own URL shape.
		// An empty result means the URL is malformed or does not belong to this
		// backend; skip Delete so we never issue a bogus delete.
		name := j.deps.Storage.ObjectNameFromURL(u)
		if name == "" {
			continue
		}
		if err := j.deps.Storage.Delete(ctx, name); err != nil {
			logger.Warn("retention: failed to delete object",
				zap.String("url", u),
				zap.String("check_id", check.ID.String()),
				zap.Error(err))
			return false
		}
	}
	return true
}
