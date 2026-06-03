package application_test

import (
	"context"
	"io"
	"testing"
	"time"

	snapp "github.com/jcsoftdev/pulzifi-back/modules/snapshot/application"
	snapentities "github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/entities"
	snaprepos "github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/repositories"
	snapservices "github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/services"
	"github.com/google/uuid"
)

// ── stubs ─────────────────────────────────────────────────────────────────────

// stubObjectStorage satisfies repositories.ObjectStorage.
type stubObjectStorage struct{}

func (stubObjectStorage) Upload(_ context.Context, _ string, _ io.Reader, _ int64, _ string) (string, error) {
	return "https://example.com/snap.png", nil
}
func (stubObjectStorage) Download(_ context.Context, _ string) ([]byte, error) { return nil, nil }
func (stubObjectStorage) Delete(_ context.Context, _ string) error              { return nil }
func (stubObjectStorage) EnsureBucket(_ context.Context) error                 { return nil }
func (stubObjectStorage) ObjectNameFromURL(storedURL string) string             { return storedURL }
func (stubObjectStorage) PresignedGetURL(_ context.Context, storedURL string, _ time.Duration) (string, error) {
	return storedURL, nil
}

var _ snaprepos.ObjectStorage = stubObjectStorage{}

// stubExtractorClient satisfies services.ExtractorClient.
type stubExtractorClient struct{}

func (stubExtractorClient) Extract(_ context.Context, _ string, _ snapservices.ExtractOptions) (*snapservices.ExtractorResult, error) {
	return &snapservices.ExtractorResult{ScreenshotBase64: "ZmFrZS1wbmc="}, nil
}

var _ snapservices.ExtractorClient = stubExtractorClient{}

// stubCheckRepo satisfies services.CheckRepository.
type stubCheckRepo struct{}

func (stubCheckRepo) Create(_ context.Context, _ *snapentities.Check) error { return nil }
func (stubCheckRepo) GetByID(_ context.Context, _ uuid.UUID) (*snapentities.Check, error) {
	return nil, nil
}
func (stubCheckRepo) Update(_ context.Context, _ *snapentities.Check) error { return nil }
func (stubCheckRepo) GetPreviousSuccessfulByPage(_ context.Context, _, _ uuid.UUID) (*snapentities.Check, error) {
	return nil, nil
}
func (stubCheckRepo) GetPreviousSuccessfulBySection(_ context.Context, _ uuid.UUID, _ *uuid.UUID, _ uuid.UUID) (*snapentities.Check, error) {
	return nil, nil
}

var _ snapservices.CheckRepository = stubCheckRepo{}

// TestSnapshotServiceCanBeConstructed verifies that NewSnapshotWorker and
// NewSnapshotService accept their dependencies and do not panic on construction.
// (Snapshot has no HTTP layer — this is the structural smoke test instead.)
func TestSnapshotServiceCanBeConstructed(t *testing.T) {
	storage := stubObjectStorage{}
	extractor := stubExtractorClient{}
	checkRepoFactory := func(_ string) snapservices.CheckRepository {
		return stubCheckRepo{}
	}

	worker := snapp.NewSnapshotWorker(snapp.SnapshotWorkerDeps{
		ObjectStorage:    storage,
		ExtractorClient:  extractor,
		CheckRepoFactory: checkRepoFactory,
	})
	if worker == nil {
		t.Fatal("expected non-nil SnapshotWorker")
	}

	svc := snapp.NewSnapshotService(storage, worker, checkRepoFactory)
	if svc == nil {
		t.Fatal("expected non-nil SnapshotService")
	}
}
