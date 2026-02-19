package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	alertentities "github.com/jcsoftdev/pulzifi-back/modules/alert/domain/entities"
	alertPersistence "github.com/jcsoftdev/pulzifi-back/modules/alert/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	monPersistence "github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/persistence"
	generateinsights "github.com/jcsoftdev/pulzifi-back/modules/insight/application/generate_insights"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/extractor"
	sharedHTML "github.com/jcsoftdev/pulzifi-back/shared/html"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"go.uber.org/zap"
)

type SnapshotWorker struct {
	objectStorage   repositories.ObjectStorage
	extractorClient *extractor.HTTPClient
	db              *sql.DB
	insightHandler  *generateinsights.GenerateInsightsHandler
}

func NewSnapshotWorker(objectStorage repositories.ObjectStorage, extractorClient *extractor.HTTPClient, db *sql.DB, insightHandler *generateinsights.GenerateInsightsHandler) *SnapshotWorker {
	return &SnapshotWorker{
		objectStorage:   objectStorage,
		extractorClient: extractorClient,
		db:              db,
		insightHandler:  insightHandler,
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

			// Load insight preferences for this page
			configRepo := monPersistence.NewMonitoringConfigPostgresRepository(s.db, schemaName)
			pageConfig, _ := configRepo.GetByPageID(ctx, check.PageID)

			enabledInsightTypes := []string{"marketing", "market_analysis"}
			enabledAlertConditions := []string{"any_changes"}
			if pageConfig != nil {
				if len(pageConfig.EnabledInsightTypes) > 0 {
					enabledInsightTypes = pageConfig.EnabledInsightTypes
				}
				if len(pageConfig.EnabledAlertConditions) > 0 {
					enabledAlertConditions = pageConfig.EnabledAlertConditions
				}
			}

			// Only alert if "any_changes" is an enabled alert condition
			if sliceContains(enabledAlertConditions, "any_changes") {
				s.createAlert(ctx, schemaName, check)
			}

			// Generate insights for enabled types
			if s.insightHandler != nil && len(enabledInsightTypes) > 0 {
				prevText := s.fetchTextFromURL(prevCheck.HTMLSnapshotURL)
				go s.generateInsightsAsync(check, targetURL, prevText, res.Text, schemaName, enabledInsightTypes)
			}
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
	check, err := repo.GetPreviousSuccessfulByPage(ctx, pageID, currentCheckID)
	if err != nil {
		return nil
	}
	return check
}

func (s *SnapshotWorker) createAlert(ctx context.Context, schemaName string, check *entities.Check) {
	if _, err := s.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(schemaName)); err != nil {
		logger.Error("Failed to set search path for alert", zap.Error(err))
		return
	}

	var workspaceID uuid.UUID
	if err := s.db.QueryRowContext(ctx, `SELECT workspace_id FROM pages WHERE id = $1`, check.PageID).Scan(&workspaceID); err != nil {
		logger.Error("Failed to get workspace_id for alert", zap.Error(err), zap.String("page_id", check.PageID.String()))
		return
	}

	alert := alertentities.NewAlert(workspaceID, check.PageID, check.ID, "content_change", "Content Changed", "The page content has changed.")
	alertRepo := alertPersistence.NewAlertPostgresRepository(s.db, schemaName)
	if err := alertRepo.Create(ctx, alert); err != nil {
		logger.Error("Failed to create alert", zap.Error(err))
	} else {
		logger.Info("Alert created", zap.String("check_id", check.ID.String()))
	}
}

// generateInsightsAsync calls the insight handler in the background after a change is detected.
func (s *SnapshotWorker) generateInsightsAsync(check *entities.Check, pageURL, prevText, newText, schemaName string, enabledTypes []string) {
	genCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	err := s.insightHandler.Handle(genCtx, &generateinsights.Request{
		PageID:              check.PageID,
		CheckID:             check.ID,
		PageURL:             pageURL,
		PrevText:            prevText,
		NewText:             newText,
		SchemaName:          schemaName,
		EnabledInsightTypes: enabledTypes,
	})
	if err != nil {
		logger.Error("Failed to generate insights", zap.String("check_id", check.ID.String()), zap.Error(err))
	} else {
		logger.Info("Insights generated successfully", zap.String("check_id", check.ID.String()))
	}
}

// sliceContains reports whether s contains target.
func sliceContains(s []string, target string) bool {
	for _, v := range s {
		if v == target {
			return true
		}
	}
	return false
}

// fetchTextFromURL downloads HTML from the given URL and extracts plain text.
// It retries up to 3 times with exponential backoff (1s, 2s, 4s) and uses a
// 30-second timeout per attempt.
func (s *SnapshotWorker) fetchTextFromURL(rawURL string) string {
	if rawURL == "" {
		return ""
	}

	const maxRetries = 3
	backoff := 1 * time.Second

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			cancel()
			logger.Error("Failed to create request for text extraction",
				zap.String("url", rawURL), zap.Error(err))
			return ""
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			cancel()
			lastErr = err
			logger.Warn("Fetch attempt failed, will retry",
				zap.String("url", rawURL),
				zap.Int("attempt", attempt),
				zap.Int("max_retries", maxRetries),
				zap.Duration("backoff", backoff),
				zap.Error(err))
			if attempt < maxRetries {
				time.Sleep(backoff)
				backoff *= 2
			}
			continue
		}

		content, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		cancel()
		if err != nil {
			lastErr = err
			logger.Warn("Failed to read HTML body, will retry",
				zap.String("url", rawURL),
				zap.Int("attempt", attempt),
				zap.Int("max_retries", maxRetries),
				zap.Duration("backoff", backoff),
				zap.Error(err))
			if attempt < maxRetries {
				time.Sleep(backoff)
				backoff *= 2
			}
			continue
		}

		return sharedHTML.ExtractText(string(content))
	}

	logger.Error("All fetch attempts exhausted for text extraction",
		zap.String("url", rawURL),
		zap.Int("attempts", maxRetries),
		zap.Error(lastErr))
	return ""
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
