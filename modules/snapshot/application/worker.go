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
	emailservices "github.com/jcsoftdev/pulzifi-back/modules/email/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/email/infrastructure/templates"
	integrationPersistence "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/webhook"
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
	emailProvider   emailservices.EmailProvider
	frontendURL     string
}

func NewSnapshotWorker(objectStorage repositories.ObjectStorage, extractorClient *extractor.HTTPClient, db *sql.DB, insightHandler *generateinsights.GenerateInsightsHandler, emailProvider emailservices.EmailProvider, frontendURL string) *SnapshotWorker {
	return &SnapshotWorker{
		objectStorage:   objectStorage,
		extractorClient: extractorClient,
		db:              db,
		insightHandler:  insightHandler,
		emailProvider:   emailProvider,
		frontendURL:     frontendURL,
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
				s.createAlert(ctx, schemaName, check, targetURL)
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

func (s *SnapshotWorker) createAlert(ctx context.Context, schemaName string, check *entities.Check, pageURL string) {
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

	// Send email notifications asynchronously
	go s.sendAlertEmails(schemaName, check, pageURL)

	// Dispatch webhooks (Slack, Discord, Teams) asynchronously
	go s.dispatchWebhooks(schemaName, check, pageURL)
}

// sendAlertEmails queries notification preferences for the page and sends email alerts.
func (s *SnapshotWorker) sendAlertEmails(schemaName string, check *entities.Check, pageURL string) {
	if s.emailProvider == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	notifRepo := monPersistence.NewNotificationPreferencePostgresRepository(s.db, schemaName)
	prefs, err := notifRepo.GetEmailEnabledByPage(ctx, check.PageID)
	if err != nil {
		logger.Error("Failed to get email-enabled preferences", zap.Error(err))
		return
	}

	if len(prefs) == 0 {
		return
	}

	dashboardURL := fmt.Sprintf("%s/workspaces", s.frontendURL)
	changeType := check.ChangeType
	if changeType == "" {
		changeType = "content"
	}
	subject, html := templates.AlertNotification(pageURL, changeType, dashboardURL)

	for _, pref := range prefs {
		// Look up user email
		var email string
		if err := s.db.QueryRowContext(ctx, `SELECT email FROM public.users WHERE id = $1`, pref.UserID).Scan(&email); err != nil {
			logger.Error("Failed to get user email for alert notification", zap.Error(err), zap.String("user_id", pref.UserID.String()))
			continue
		}
		if err := s.emailProvider.Send(ctx, email, subject, html); err != nil {
			logger.Error("Failed to send alert email", zap.Error(err), zap.String("email", email))
		}
	}
}

// dispatchWebhooks sends webhook notifications to enabled Slack/Discord/Teams integrations.
func (s *SnapshotWorker) dispatchWebhooks(schemaName string, check *entities.Check, pageURL string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	integrationRepo := integrationPersistence.NewIntegrationPostgresRepository(s.db, schemaName)
	sender := webhook.NewSender()

	changeType := check.ChangeType
	if changeType == "" {
		changeType = "content"
	}

	for _, serviceType := range []string{"slack", "discord", "teams"} {
		integrations, err := integrationRepo.ListByServiceType(ctx, serviceType)
		if err != nil {
			logger.Error("Failed to list integrations", zap.Error(err), zap.String("service_type", serviceType))
			continue
		}
		for _, integration := range integrations {
			if !integration.Enabled {
				continue
			}
			if err := sender.Dispatch(ctx, integration, pageURL, changeType); err != nil {
				logger.Error("Failed to dispatch webhook",
					zap.Error(err),
					zap.String("service_type", serviceType),
					zap.String("integration_id", integration.ID.String()))
			}
		}
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
