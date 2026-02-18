package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	monPersistence "github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/extractor"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"go.uber.org/zap"
)

type SnapshotWorker struct {
	objectStorage   repositories.ObjectStorage
	extractorClient *extractor.HTTPClient
	db              *sql.DB
}

func NewSnapshotWorker(objectStorage repositories.ObjectStorage, extractorClient *extractor.HTTPClient, db *sql.DB) *SnapshotWorker {
	return &SnapshotWorker{
		objectStorage:   objectStorage,
		extractorClient: extractorClient,
		db:              db,
	}
}

func (s *SnapshotWorker) ExecuteCheck(ctx context.Context, checkID uuid.UUID, targetURL string, schemaName string) error {
	logger.Info("SnapshotWorker executing check", zap.String("check_id", checkID.String()), zap.String("url", targetURL))

	checkRepo := monPersistence.NewCheckPostgresRepository(s.db, schemaName)
	check, err := checkRepo.GetByID(ctx, checkID)
	if err != nil {
		return err
	}
	if check == nil {
		return fmt.Errorf("check not found: %s", checkID)
	}

	// Find previous successful check for comparison
	// Ideally we need a better query, but for now getting latest by page
	// (which might be the current one or previous one)
	// We will skip comparison if we can't reliably find the previous one
	// or implement a simple check
	// prevCheck, _ := checkRepo.GetLatestByPage(ctx, check.PageID)

	startTime := time.Now()
	res, err := s.extractorClient.Extract(ctx, targetURL)
	duration := int(time.Since(startTime).Milliseconds())

	if err != nil {
		check.Status = "error"
		check.ErrorMessage = err.Error()
		check.DurationMs = duration
		return checkRepo.Update(ctx, check)
	}

	// Process Results
	imgBytes, err := base64.StdEncoding.DecodeString(res.ScreenshotBase64)
	if err != nil {
		return fmt.Errorf("failed to decode screenshot: %w", err)
	}

	ts := time.Now().Unix()
	imgName := fmt.Sprintf("%s/%d.png", check.PageID, ts)
	htmlName := fmt.Sprintf("%s/%d.html", check.PageID, ts)

	// Upload
	imgURL, err := s.objectStorage.Upload(ctx, imgName, bytes.NewReader(imgBytes), int64(len(imgBytes)), "image/png")
	if err != nil {
		logger.Error("Failed to upload screenshot", zap.Error(err))
		// Continue even if upload fails? Or fail?
		// Let's continue but mark error?
		// For now fail.
		return err
	}
	htmlURL, err := s.objectStorage.Upload(ctx, htmlName, strings.NewReader(res.HTML), int64(len(res.HTML)), "text/html")
	if err != nil {
		return err
	}

	// Hashes
	contentHash := sha256.Sum256([]byte(res.Text))
	contentHashStr := hex.EncodeToString(contentHash[:])

	// Update Check
	check.Status = "success"
	check.DurationMs = duration
	check.ScreenshotURL = imgURL
	check.HTMLSnapshotURL = htmlURL
	check.ContentHash = contentHashStr
	check.ChangeDetected = false
	check.ChangeType = ""

	// Compare logic (simplified)
	// Fetch the actual previous successful check
	prevCheck := s.getPreviousSuccessfulCheck(ctx, checkRepo, check.PageID, check.ID)

	if prevCheck != nil && prevCheck.ContentHash != "" {
		if prevCheck.ContentHash != contentHashStr {
			check.ChangeDetected = true
			check.ChangeType = "content"
			s.createAlert(ctx, schemaName, check)
		}
	}

	if err := checkRepo.Update(ctx, check); err != nil {
		return err
	}

	if err := s.updatePageSnapshotMetadata(ctx, schemaName, check.PageID, imgURL, check.ChangeDetected); err != nil {
		logger.Error("Failed to update page snapshot metadata", zap.Error(err), zap.String("page_id", check.PageID.String()))
	}

	return nil
}

func (s *SnapshotWorker) getPreviousSuccessfulCheck(ctx context.Context, repo *monPersistence.CheckPostgresRepository, pageID, currentCheckID uuid.UUID) *entities.Check {
	// We need to list checks and find the first one that is success and not current ID
	// This is inefficient but works for now without modifying repo interface too much
	checks, err := repo.ListByPage(ctx, pageID)
	if err != nil {
		return nil
	}
	for _, c := range checks {
		if c.ID != currentCheckID && c.Status == "success" {
			return c
		}
	}
	return nil
}

func (s *SnapshotWorker) createAlert(ctx context.Context, schemaName string, check *entities.Check) {
	if _, err := s.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(schemaName)); err != nil {
		logger.Error("Failed to set search path for alert", zap.Error(err))
		return
	}

	// Insert Alert
	q := `INSERT INTO alerts (workspace_id, page_id, check_id, type, title, description) 
		  SELECT workspace_id, $1, $2, 'content_change', 'Content Changed', 'The page content has changed.' 
		  FROM pages WHERE id = $1`

	if _, err := s.db.ExecContext(ctx, q, check.PageID, check.ID); err != nil {
		logger.Error("Failed to create alert", zap.Error(err))
	} else {
		logger.Info("Alert created", zap.String("check_id", check.ID.String()))
	}
}

func (s *SnapshotWorker) updatePageSnapshotMetadata(ctx context.Context, schemaName string, pageID uuid.UUID, thumbnailURL string, changeDetected bool) error {
	if _, err := s.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(schemaName)); err != nil {
		return err
	}

	q := `UPDATE pages
		SET thumbnail_url = $1,
			last_change_detected_at = CASE WHEN $2 THEN NOW() ELSE last_change_detected_at END
		WHERE id = $3`

	_, err := s.db.ExecContext(ctx, q, thumbnailURL, changeDetected, pageID)
	return err
}
