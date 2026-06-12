package orgwiring

import (
	"context"

	"github.com/google/uuid"
	orgservices "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/featureflags"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// minioObjectStore is the minimal MinIO operations we need for sweeping.
// Matches the subset of snapshot's minio.Client used here.
type minioObjectStore interface {
	// ListAndDeleteByPrefix removes all objects under the given prefix.
	// Returns nil when the prefix is empty or no objects exist.
	DeleteByPrefix(ctx context.Context, prefix string) error
}

// storageSweepAdapter implements orgservices.StorageSweeper.
// It is best-effort: it never returns an error that aborts the cascade
// (the use case ignores sweep errors anyway, but we keep the contract clean).
type storageSweepAdapter struct {
	minio    minioObjectStore
	ffReader *featureflags.Reader
}

// compile-time assertion
var _ orgservices.StorageSweeper = (*storageSweepAdapter)(nil)

// NewStorageSweepAdapter builds a StorageSweeper backed by a real MinIO client
// and the shared feature-flags reader.
func NewStorageSweepAdapter(minio minioObjectStore, ffReader *featureflags.Reader) orgservices.StorageSweeper {
	return &storageSweepAdapter{minio: minio, ffReader: ffReader}
}

// SweepTenant removes all objects under {schema}/ when the
// snapshot.private_storage flag is on for the org. No-op otherwise.
func (a *storageSweepAdapter) SweepTenant(ctx context.Context, orgID uuid.UUID, schema string) error {
	on, err := a.ffReader.IsOn(ctx, orgID, "snapshot.private_storage")
	if err != nil || !on {
		logger.Debug("storage_sweep_adapter: private_storage off or flag error, skipping",
			zap.String("org_id", orgID.String()),
			zap.String("schema", schema),
		)
		return nil
	}

	prefix := schema + "/"
	if delErr := a.minio.DeleteByPrefix(ctx, prefix); delErr != nil {
		// Return the error; the use case is responsible for logging WARN and continuing.
		return delErr
	}
	return nil
}

// ── nopStorageSweeper ─────────────────────────────────────────────────────────

// nopStorageSweeper is injected when object storage is not configured.
// Returns nil immediately without any storage calls.
type nopStorageSweeper struct{}

var _ orgservices.StorageSweeper = (*nopStorageSweeper)(nil)

// NopStorageSweeper returns a no-op StorageSweeper for use when MinIO is
// unavailable or the provider is Cloudinary (not prefix-sweepable at MVP).
func NopStorageSweeper() orgservices.StorageSweeper {
	return &nopStorageSweeper{}
}

func (n *nopStorageSweeper) SweepTenant(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}
