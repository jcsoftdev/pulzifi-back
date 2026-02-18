package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	generateinsights "github.com/jcsoftdev/pulzifi-back/modules/insight/application/generate_insights"
	monEntities "github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	monPersistence "github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/persistence"
	snapEntities "github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/extractor"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type SnapshotService struct {
	worker *SnapshotWorker
	db     *sql.DB
}

func NewSnapshotService(objectStorage repositories.ObjectStorage, extractorClient *extractor.HTTPClient, db *sql.DB, insightHandler *generateinsights.GenerateInsightsHandler) *SnapshotService {
	return &SnapshotService{
		worker: NewSnapshotWorker(objectStorage, extractorClient, db, insightHandler),
		db:     db,
	}
}

func (s *SnapshotService) CaptureAndUpload(ctx context.Context, req snapEntities.SnapshotRequest) (*snapEntities.SnapshotResult, error) {
	logger.Info("SnapshotService: CaptureAndUpload", zap.String("page_id", req.PageID), zap.String("url", req.URL))

	pageID, err := uuid.Parse(req.PageID)
	if err != nil {
		return nil, fmt.Errorf("invalid page_id: %w", err)
	}

	// Create a new Check entity
	check := monEntities.Check{
		ID:        uuid.New(),
		PageID:    pageID,
		Status:    "running",
		CheckedAt: time.Now(),
	}

	checkRepo := monPersistence.NewCheckPostgresRepository(s.db, req.SchemaName)
	if err := checkRepo.Create(ctx, &check); err != nil {
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
