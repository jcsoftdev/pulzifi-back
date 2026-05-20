package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	snapEntities "github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/repositories"
	snapServices "github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Port type aliases for external callers constructing SnapshotWorkerDeps.
type NotificationEmailerPort = snapServices.NotificationEmailer

// SnapshotService orchestrates a one-shot snapshot capture from an event-bus request.
// It creates a Check record, delegates execution to SnapshotWorker, then reads back
// the result to produce a SnapshotResult.
type SnapshotService struct {
	worker           *SnapshotWorker
	checkRepoFactory func(schemaName string) snapServices.CheckRepository
}

// NewSnapshotService creates a SnapshotService backed by the given SnapshotWorker.
func NewSnapshotService(objectStorage repositories.ObjectStorage, worker *SnapshotWorker, checkRepoFactory func(schemaName string) snapServices.CheckRepository) *SnapshotService {
	return &SnapshotService{
		worker:           worker,
		checkRepoFactory: checkRepoFactory,
	}
}

func (s *SnapshotService) CaptureAndUpload(ctx context.Context, req snapEntities.SnapshotRequest) (*snapEntities.SnapshotResult, error) {
	logger.Info("SnapshotService: CaptureAndUpload", zap.String("page_id", req.PageID), zap.String("url", req.URL))

	pageID, err := uuid.Parse(req.PageID)
	if err != nil {
		return nil, fmt.Errorf("invalid page_id: %w", err)
	}

	// Create a new Check entity
	check := snapEntities.NewCheck(pageID, "running", false)
	check.CheckedAt = time.Now()

	checkRepo := s.checkRepoFactory(req.SchemaName)
	if err := checkRepo.Create(ctx, check); err != nil {
		return nil, fmt.Errorf("failed to create check: %w", err)
	}

	// Execute the check using the worker
	if err := s.worker.ExecuteCheck(ctx, check.ID, req.URL, req.SchemaName); err != nil {
		// Worker updates the check status to error
		logger.Error("SnapshotService: Check execution failed", zap.Error(err))
		return nil, err
	}

	// Fetch the updated check to get results
	updatedCheck, err := checkRepo.GetByID(ctx, check.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated check: %w", err)
	}
	if updatedCheck == nil {
		return nil, fmt.Errorf("check disappeared")
	}

	result := &snapEntities.SnapshotResult{
		PageID:      req.PageID,
		URL:         req.URL,
		SchemaName:  req.SchemaName,
		ImageURL:    updatedCheck.ScreenshotURL,
		HTMLURL:     updatedCheck.HTMLSnapshotURL,
		ContentHash: updatedCheck.ContentHash,
		Status:      updatedCheck.Status,
		CreatedAt:   updatedCheck.CheckedAt,
	}

	return result, nil
}
