package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
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
	generateinsights "github.com/jcsoftdev/pulzifi-back/modules/insight/application/generate_insights"
	insightservices "github.com/jcsoftdev/pulzifi-back/modules/insight/domain/services"
	integrationPersistence "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/webhook"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	monPersistence "github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/repositories"
	imagecompare "github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/extractor"
	sharedHTML "github.com/jcsoftdev/pulzifi-back/shared/html"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"go.uber.org/zap"
)

type SnapshotWorker struct {
	objectStorage      repositories.ObjectStorage
	extractorClient    *extractor.HTTPClient
	db                 *sql.DB
	insightHandler     *generateinsights.GenerateInsightsHandler
	emailProvider      emailservices.EmailProvider
	frontendURL        string
	visionAnalyzer     insightservices.VisionAnalyzer
	pixelDiffThreshold float64
	onCheckDone        func(pageID uuid.UUID, checkJSON []byte)
}

// SetOnCheckDone registers a callback invoked after every check completes
// (success or error) with the serialized CheckResponse payload.
func (s *SnapshotWorker) SetOnCheckDone(fn func(pageID uuid.UUID, checkJSON []byte)) {
	s.onCheckDone = fn
}

// SetVisionAnalyzer sets the vision AI analyzer for screenshot comparison.
func (s *SnapshotWorker) SetVisionAnalyzer(analyzer insightservices.VisionAnalyzer) {
	s.visionAnalyzer = analyzer
}

// SetPixelDiffThreshold sets the threshold for pixel comparison (default 0.001).
func (s *SnapshotWorker) SetPixelDiffThreshold(threshold float64) {
	s.pixelDiffThreshold = threshold
}

// notifyCheckDone serializes a check into the same DTO format the frontend
// expects and invokes the onCheckDone callback if set.
func (s *SnapshotWorker) notifyCheckDone(check *entities.Check) {
	if s.onCheckDone == nil {
		return
	}
	type checkResponse struct {
		ID              uuid.UUID `json:"id"`
		PageID          uuid.UUID `json:"page_id"`
		Status          string    `json:"status"`
		ScreenshotURL   string    `json:"screenshot_url"`
		HTMLSnapshotURL string    `json:"html_snapshot_url"`
		ChangeDetected  bool      `json:"change_detected"`
		ChangeType      string    `json:"change_type"`
		ErrorMessage    string    `json:"error_message,omitempty"`
		CheckedAt       string    `json:"checked_at"`
	}
	payload, err := json.Marshal(checkResponse{
		ID:              check.ID,
		PageID:          check.PageID,
		Status:          check.Status,
		ScreenshotURL:   check.ScreenshotURL,
		HTMLSnapshotURL: check.HTMLSnapshotURL,
		ChangeDetected:  check.ChangeDetected,
		ChangeType:      check.ChangeType,
		ErrorMessage:    check.ErrorMessage,
		CheckedAt:       check.CheckedAt.Format("2006-01-02T15:04:05.999999Z"),
	})
	if err != nil {
		logger.Error("Failed to serialize check for SSE notification", zap.Error(err))
		return
	}
	s.onCheckDone(check.PageID, payload)
}

func NewSnapshotWorker(objectStorage repositories.ObjectStorage, extractorClient *extractor.HTTPClient, db *sql.DB, insightHandler *generateinsights.GenerateInsightsHandler, emailProvider emailservices.EmailProvider, frontendURL string) *SnapshotWorker {
	return &SnapshotWorker{
		objectStorage:      objectStorage,
		extractorClient:    extractorClient,
		db:                 db,
		insightHandler:     insightHandler,
		emailProvider:      emailProvider,
		frontendURL:        frontendURL,
		pixelDiffThreshold: 0.001, // default
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

	// markError is a helper that ensures the check reaches "error" status.
	markError := func(msg string, duration int) error {
		check.Status = "error"
		check.ErrorMessage = msg
		check.DurationMs = duration
		if updateErr := checkRepo.Update(ctx, check); updateErr != nil {
			logger.Error("Failed to mark check as error", zap.Error(updateErr), zap.String("check_id", checkID.String()))
			return updateErr
		}
		s.notifyCheckDone(check)
		return fmt.Errorf("%s", msg)
	}

	// Fetch monitoring config before extraction to get block_ads_cookies + selector settings
	configRepo := monPersistence.NewMonitoringConfigPostgresRepository(s.db, schemaName)
	pageConfig, configErr := configRepo.GetByPageID(ctx, check.PageID)
	if configErr != nil {
		logger.Warn("Failed to load monitoring config, using defaults",
			zap.String("page_id", check.PageID.String()), zap.Error(configErr))
	}

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

	extractOpts := extractor.ExtractOptions{}
	if pageConfig != nil {
		extractOpts.BlockAdsCookies = pageConfig.BlockAdsCookies
		extractOpts.IgnoreSelectors = pageConfig.IgnoreSelectors

		switch pageConfig.SelectorType {
		case "element":
			if pageConfig.CSSSelector != "" {
				extractOpts.Selector = pageConfig.CSSSelector
				extractOpts.SelectorXPath = pageConfig.XPathSelector
				if pageConfig.SelectorOffsets != nil {
					extractOpts.SelectorOffsets = &extractor.SelectorOffsets{
						Top:    pageConfig.SelectorOffsets.Top,
						Right:  pageConfig.SelectorOffsets.Right,
						Bottom: pageConfig.SelectorOffsets.Bottom,
						Left:   pageConfig.SelectorOffsets.Left,
					}
				}
			}
		case "sections":
			// Query sections and pass them to the extractor for a single page load.
			sectionRepo := monPersistence.NewMonitoredSectionPostgresRepository(s.db, schemaName)
			pageSections, err := sectionRepo.ListByPageID(ctx, check.PageID)
			if err != nil {
				logger.Error("Failed to load monitored sections, falling back to full-page",
					zap.String("page_id", check.PageID.String()), zap.Error(err))
				break
			}
			if len(pageSections) == 0 {
				logger.Warn("No monitored sections found for sections-mode page, falling back to full-page",
					zap.String("page_id", check.PageID.String()))
				break
			}

			logger.Info("Processing sections mode",
				zap.String("page_id", check.PageID.String()),
				zap.Int("section_count", len(pageSections)))

			sectionsByID := make(map[uuid.UUID]*entities.MonitoredSection, len(pageSections))
			for _, sec := range pageSections {
				sectionsByID[sec.ID] = sec
				opt := extractor.SectionExtractOption{ID: sec.ID.String(), Selector: sec.CSSSelector, SelectorXPath: sec.XPathSelector}
				if sec.SelectorOffsets != nil {
					opt.Offsets = &extractor.SelectorOffsets{Top: sec.SelectorOffsets.Top, Right: sec.SelectorOffsets.Right, Bottom: sec.SelectorOffsets.Bottom, Left: sec.SelectorOffsets.Left}
				}
				extractOpts.Sections = append(extractOpts.Sections, opt)
			}

			startTime := time.Now()
			res, err := s.extractorClient.Extract(ctx, targetURL, extractOpts)
			duration := int(time.Since(startTime).Milliseconds())
			if err != nil {
				return markError(err.Error(), duration)
			}

			logger.Info("Extractor returned sections result",
				zap.String("page_id", check.PageID.String()),
				zap.Int("section_results", len(res.Sections)),
				zap.Bool("has_full_page_screenshot", res.ScreenshotBase64 != ""))

			// Store full-page screenshot on the parent check if available.
			if res.ScreenshotBase64 != "" {
				if imgBytes, decErr := base64.StdEncoding.DecodeString(res.ScreenshotBase64); decErr == nil && len(imgBytes) > 0 {
					ts := time.Now().Unix()
					imgName := fmt.Sprintf("%s/%d.png", check.PageID, ts)
					if imgURL, upErr := s.objectStorage.Upload(ctx, imgName, bytes.NewReader(imgBytes), int64(len(imgBytes)), "image/png"); upErr == nil {
						check.ScreenshotURL = imgURL
					}
				}
			}

			// Mark parent check as complete.
			check.Status = "success"
			check.DurationMs = duration
			if err := checkRepo.Update(ctx, check); err != nil {
				return err
			}
			s.notifyCheckDone(check)

			anyChanged := s.processSectionsFromExtractor(ctx, checkRepo, schemaName, check.ID, check.PageID, sectionsByID, res.Sections, targetURL, enabledAlertConditions)
			if anyChanged {
				check.ChangeDetected = true
				check.ChangeType = "content"
				if err := checkRepo.Update(ctx, check); err != nil {
					logger.Error("Failed to update parent check with section changes",
						zap.Error(err), zap.String("check_id", check.ID.String()))
				}
				// Re-notify SSE with updated change status
				s.notifyCheckDone(check)
			}
			return nil
		}
	}

	startTime := time.Now()
	res, err := s.extractorClient.Extract(ctx, targetURL, extractOpts)
	duration := int(time.Since(startTime).Milliseconds())

	if err != nil {
		return markError(err.Error(), duration)
	}

	// Process Results
	imgBytes, err := base64.StdEncoding.DecodeString(res.ScreenshotBase64)
	if err != nil {
		return markError(fmt.Sprintf("failed to decode screenshot: %v", err), duration)
	}

	ts := time.Now().Unix()
	imgName := fmt.Sprintf("%s/%d.png", check.PageID, ts)
	htmlName := fmt.Sprintf("%s/%d.html", check.PageID, ts)

	// Upload
	if s.objectStorage == nil {
		return markError("object storage client is not configured", duration)
	}
	imgURL, err := s.objectStorage.Upload(ctx, imgName, bytes.NewReader(imgBytes), int64(len(imgBytes)), "image/png")
	if err != nil {
		return markError(fmt.Sprintf("failed to upload screenshot: %v", err), duration)
	}
	htmlURL, err := s.objectStorage.Upload(ctx, htmlName, strings.NewReader(res.HTML), int64(len(res.HTML)), "text/html")
	if err != nil {
		return markError(fmt.Sprintf("failed to upload html snapshot: %v", err), duration)
	}

	// Content hash — extract text from HTML (deterministic) instead of
	// Playwright's innerText (rendering-dependent, varies across runs).
	contentHash := sha256.Sum256([]byte(sharedHTML.ExtractText(res.HTML)))
	contentHashStr := hex.EncodeToString(contentHash[:])

	// Content block hash — structural content representation for content-first detection.
	contentBlocks := sharedHTML.ExtractContentBlocks(res.HTML)
	contentBlockHash := sharedHTML.HashContentBlocks(contentBlocks)

	// Screenshot hash (pixel-based)
	screenshotHash := imagecompare.HashScreenshot(imgBytes)

	// Update Check
	check.Status = "success"
	check.DurationMs = duration
	check.ScreenshotURL = imgURL
	check.HTMLSnapshotURL = htmlURL
	check.ContentHash = contentHashStr
	check.ContentBlockHash = contentBlockHash
	check.ScreenshotHash = screenshotHash
	check.ChangeDetected = false
	check.ChangeType = ""

	// Fetch previous successful check for comparison
	prevCheck := s.getPreviousSuccessfulCheck(ctx, checkRepo, check.PageID, check.ID)

	if prevCheck != nil {
		changeDetected, changeSummary, contentDiff := s.detectChange(ctx, prevCheck, check, imgBytes, res.ScreenshotBase64, targetURL, res.HTML)

		if changeDetected {
			check.ChangeDetected = true
			check.ChangeType = "content"
			check.VisionChangeSummary = changeSummary

			// Store pre-computed content diff for the frontend
			if contentDiff != nil && contentDiff.HasChanges {
				if diffJSON, err := json.Marshal(contentDiff); err == nil {
					check.ContentDiffJSON = string(diffJSON)
				}
			}

			// Only alert if "any_changes" is an enabled alert condition
			if sliceContains(enabledAlertConditions, "any_changes") {
				s.createAlert(ctx, schemaName, check, targetURL, changeSummary)
			}

			// Generate insights for enabled types
			if s.insightHandler != nil && len(enabledInsightTypes) > 0 {
				var diffText string
				if contentDiff != nil && contentDiff.HasChanges {
					diffText = sharedHTML.FormatDiffForAI(contentDiff)
				}
				prevText := s.fetchTextFromURL(prevCheck.HTMLSnapshotURL)
				go s.generateInsightsAsync(check, targetURL, prevText, res.Text, schemaName, enabledInsightTypes, diffText)
			}
		}
	}

	if err := checkRepo.Update(ctx, check); err != nil {
		return err
	}

	s.notifyCheckDone(check)

	if err := s.updatePageSnapshotMetadata(ctx, schemaName, check.PageID, imgURL, check.ChangeDetected); err != nil {
		logger.Error("Failed to update page snapshot metadata", zap.Error(err), zap.String("page_id", check.PageID.String()))
	}

	return nil
}

// detectChange runs a multi-stage content-first change detection pipeline:
//
//	Stage 1: Content block hash (fast structural identity)
//	Stage 2: Content block diff (structural comparison for diff text)
//	Stage 3: Pixel comparison (visual-only changes when content is unchanged)
//	Stage 4: Vision AI semantic analysis (optional)
//	Stage 5: Normalized text hash fallback (legacy compatibility)
//
// Returns (changeDetected, changeSummary, contentDiff)
func (s *SnapshotWorker) detectChange(ctx context.Context, prevCheck, currCheck *entities.Check, currImgBytes []byte, currBase64 string, pageURL string, currHTML string) (bool, string, *sharedHTML.ContentDiff) {
	pageID := currCheck.PageID.String()

	// ── Stage 1: Content block hash comparison ───────────────────────────
	if prevCheck.ContentBlockHash != "" {
		if prevCheck.ContentBlockHash == currCheck.ContentBlockHash {
			logger.Info("Content block hash identical — checking visual changes",
				zap.String("page_id", pageID))

			// Content is structurally identical — check for visual-only changes.
			// ── Stage 3: Pixel comparison (secondary) ────────────────────
			if prevCheck.ScreenshotHash != "" && prevCheck.ScreenshotHash != currCheck.ScreenshotHash {
				if prevCheck.ScreenshotURL != "" {
					prevImgBytes := s.downloadScreenshot(prevCheck.ScreenshotURL)
					if len(prevImgBytes) > 0 {
						result, err := imagecompare.CompareScreenshots(prevImgBytes, currImgBytes, s.pixelDiffThreshold)
						if err == nil && !result.Identical && result.DiffRatio >= s.pixelDiffThreshold {
							// ── Stage 4: Vision AI (optional) ────────────
							if s.visionAnalyzer != nil {
								prevB64 := base64.StdEncoding.EncodeToString(prevImgBytes)
								visionResult, vErr := s.visionAnalyzer.AnalyzeChange(ctx, prevB64, currBase64, pageURL)
								if vErr != nil {
									logger.Error("Vision AI failed, reporting visual change",
										zap.Error(vErr), zap.String("page_id", pageID))
									return true, "", nil
								}
								if !visionResult.HasMeaningfulChange {
									logger.Info("Vision AI says no meaningful visual change",
										zap.String("page_id", pageID))
									return false, "", nil
								}
								return true, visionResult.ChangeSummary, nil
							}
							logger.Info("Visual-only change detected via pixel diff",
								zap.String("page_id", pageID),
								zap.Float64("diff_ratio", result.DiffRatio))
							return true, "", nil
						}
					}
				}
			}
			return false, "", nil
		}

		// ── Stage 2: Content block diff ──────────────────────────────────
		// Content block hashes differ — compute structural diff.
		logger.Info("Content block hash differs, computing structural diff",
			zap.String("page_id", pageID))

		prevHTML := s.fetchHTMLFromURL(prevCheck.HTMLSnapshotURL)
		var contentDiff *sharedHTML.ContentDiff
		if prevHTML != "" {
			prevBlocks := sharedHTML.ExtractContentBlocks(prevHTML)
			currBlocks := sharedHTML.ExtractContentBlocks(currHTML)
			contentDiff = sharedHTML.DiffContentBlocks(prevBlocks, currBlocks)
		}

		if contentDiff != nil && contentDiff.HasChanges {
			logger.Info("Content change detected via structural diff",
				zap.String("page_id", pageID),
				zap.Int("total_changes", contentDiff.TotalChanges))
			return true, "", contentDiff
		}

		// Diff computation failed or showed no changes despite hash difference.
		// Fall through to pixel comparison.
		logger.Info("Content block hash differed but diff empty, falling through to pixel",
			zap.String("page_id", pageID))
	}

	// ── Stage 3: Pixel comparison ────────────────────────────────────────
	if prevCheck.ScreenshotHash != "" {
		if prevCheck.ScreenshotHash == currCheck.ScreenshotHash {
			logger.Info("Screenshot hash identical — no change", zap.String("page_id", pageID))
			return false, "", nil
		}

		if prevCheck.ScreenshotURL != "" {
			prevImgBytes := s.downloadScreenshot(prevCheck.ScreenshotURL)
			if len(prevImgBytes) > 0 {
				result, err := imagecompare.CompareScreenshots(prevImgBytes, currImgBytes, s.pixelDiffThreshold)
				if err == nil {
					if result.Identical || result.DiffRatio < s.pixelDiffThreshold {
						logger.Info("Pixel diff below threshold — no meaningful change",
							zap.String("page_id", pageID),
							zap.Float64("diff_ratio", result.DiffRatio))
						return false, "", nil
					}

					// ── Stage 4: Vision AI analysis (optional) ───────────
					if s.visionAnalyzer != nil {
						prevB64 := base64.StdEncoding.EncodeToString(prevImgBytes)
						visionResult, vErr := s.visionAnalyzer.AnalyzeChange(ctx, prevB64, currBase64, pageURL)
						if vErr != nil {
							logger.Error("Vision AI failed, reporting change based on pixel diff",
								zap.Error(vErr), zap.String("page_id", pageID))
							return true, "", nil
						}
						if !visionResult.HasMeaningfulChange {
							return false, "", nil
						}
						return true, visionResult.ChangeSummary, nil
					}

					logger.Info("Pixel diff above threshold, reporting change",
						zap.String("page_id", pageID),
						zap.Float64("diff_ratio", result.DiffRatio))
					return true, "", nil
				}
				logger.Error("Pixel comparison failed", zap.Error(err), zap.String("page_id", pageID))
			}
		}
	}

	// ── Stage 5: Normalized text hash fallback ───────────────────────────
	// Used when previous check has no content block hash (legacy checks).
	if prevCheck.ContentHash != "" && prevCheck.ContentHash != currCheck.ContentHash {
		logger.Info("Change detected via normalized text hash",
			zap.String("page_id", pageID))
		return true, "", nil
	}

	return false, "", nil
}

// downloadScreenshot fetches a screenshot using the object storage client.
func (s *SnapshotWorker) downloadScreenshot(url string) []byte {
	if url == "" {
		return nil
	}

	if s.objectStorage == nil {
		logger.Error("Object storage not configured, cannot download screenshot", zap.String("url", url))
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	data, err := s.objectStorage.Download(ctx, url)
	if err != nil {
		logger.Error("Failed to download screenshot via object storage client", zap.String("url", url), zap.Error(err))
		return nil
	}

	return data
}

func (s *SnapshotWorker) getPreviousSuccessfulCheck(ctx context.Context, repo *monPersistence.CheckPostgresRepository, pageID, currentCheckID uuid.UUID) *entities.Check {
	check, err := repo.GetPreviousSuccessfulByPage(ctx, pageID, currentCheckID)
	if err != nil {
		return nil
	}
	return check
}


func (s *SnapshotWorker) createAlert(ctx context.Context, schemaName string, check *entities.Check, pageURL string, changeSummary string) {
	if _, err := s.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(schemaName)); err != nil {
		logger.Error("Failed to set search path for alert", zap.Error(err))
		return
	}

	var workspaceID uuid.UUID
	if err := s.db.QueryRowContext(ctx, `SELECT workspace_id FROM pages WHERE id = $1`, check.PageID).Scan(&workspaceID); err != nil {
		logger.Error("Failed to get workspace_id for alert", zap.Error(err), zap.String("page_id", check.PageID.String()))
		return
	}

	// Use Vision AI summary if available, otherwise generic message
	alertTitle := "Content Changed"
	alertDescription := "The page content has changed."
	if changeSummary != "" {
		alertTitle = changeSummary
		alertDescription = changeSummary
	}

	alert := alertentities.NewAlert(workspaceID, check.PageID, check.ID, "content_change", alertTitle, alertDescription)
	alert.ChangeSummary = changeSummary

	alertRepo := alertPersistence.NewAlertPostgresRepository(s.db, schemaName)
	if err := alertRepo.Create(ctx, alert); err != nil {
		logger.Error("Failed to create alert", zap.Error(err))
	} else {
		logger.Info("Alert created", zap.String("check_id", check.ID.String()))
	}

	// Send email notifications asynchronously
	go s.sendAlertEmails(schemaName, check, pageURL, changeSummary)

	// Dispatch webhooks (Slack, Discord, Teams) asynchronously
	go s.dispatchWebhooks(schemaName, check, pageURL, changeSummary)
}

// sendAlertEmails queries notification preferences for the page and sends email alerts.
func (s *SnapshotWorker) sendAlertEmails(schemaName string, check *entities.Check, pageURL string, changeSummary string) {
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

	// Filter preferences by change_types: empty means all types, otherwise must include "page_change"
	var filteredPrefs []*entities.NotificationPreference
	for _, pref := range prefs {
		if len(pref.ChangeTypes) == 0 || sliceContains(pref.ChangeTypes, "page_change") {
			filteredPrefs = append(filteredPrefs, pref)
		}
	}
	if len(filteredPrefs) == 0 {
		return
	}

	dashboardURL := fmt.Sprintf("%s/workspaces", s.frontendURL)
	changeType := check.ChangeType
	if changeType == "" {
		changeType = "content"
	}
	// Use change summary as the change type description if available
	if changeSummary != "" {
		changeType = changeSummary
	}
	subject, html := templates.AlertNotification(pageURL, changeType, dashboardURL)

	for _, pref := range filteredPrefs {
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
func (s *SnapshotWorker) dispatchWebhooks(schemaName string, check *entities.Check, pageURL string, changeSummary string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	integrationRepo := integrationPersistence.NewIntegrationPostgresRepository(s.db, schemaName)
	sender := webhook.NewSender()

	changeType := check.ChangeType
	if changeType == "" {
		changeType = "content"
	}
	if changeSummary != "" {
		changeType = changeSummary
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
func (s *SnapshotWorker) generateInsightsAsync(check *entities.Check, pageURL, prevText, newText, schemaName string, enabledTypes []string, diffText string) {
	genCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	err := s.insightHandler.Handle(genCtx, &generateinsights.Request{
		PageID:              check.PageID,
		CheckID:             check.ID,
		PageURL:             pageURL,
		PrevText:            prevText,
		NewText:             newText,
		DiffText:            diffText,
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

// fetchHTMLFromURL downloads raw HTML from the given URL. Uses the object storage
// client. Returns empty string on failure.
func (s *SnapshotWorker) fetchHTMLFromURL(rawURL string) string {
	if rawURL == "" {
		return ""
	}

	if s.objectStorage == nil {
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	data, err := s.objectStorage.Download(ctx, rawURL)
	if err != nil {
		logger.Error("Failed to download HTML snapshot", zap.String("url", rawURL), zap.Error(err))
		return ""
	}

	return string(data)
}

// processSectionsFromExtractor creates one check per monitored section using the
// section screenshots returned by the extractor (single page load). Returns true
// if any section detected a change. Alerts are aggregated into a single notification
// per page instead of one per section.
func (s *SnapshotWorker) processSectionsFromExtractor(
	ctx context.Context,
	checkRepo *monPersistence.CheckPostgresRepository,
	schemaName string,
	parentCheckID uuid.UUID,
	pageID uuid.UUID,
	sectionsByID map[uuid.UUID]*entities.MonitoredSection,
	sectionResults []extractor.SectionExtractResult,
	targetURL string,
	enabledAlertConditions []string,
) bool {
	anyChanged := false
	firstScreenshotURL := ""
	var changeSummaries []string

	for i := range sectionResults {
		sec := &sectionResults[i]
		if sec.ScreenshotBase64 == "" {
			logger.Warn("Section has no screenshot (selector may not have matched)",
				zap.String("section_id", sec.ID),
				zap.Bool("selector_matched", sec.SelectorMatched))
			continue
		}

		sectionID, err := uuid.Parse(sec.ID)
		if err != nil {
			logger.Warn("Invalid section ID in extractor result", zap.String("id", sec.ID))
			continue
		}
		section, ok := sectionsByID[sectionID]
		if !ok {
			continue
		}

		imgBytes, err := base64.StdEncoding.DecodeString(sec.ScreenshotBase64)
		if err != nil || len(imgBytes) == 0 {
			logger.Warn("Failed to decode section screenshot", zap.String("section_id", sec.ID), zap.Error(err))
			continue
		}

		ts := time.Now().Unix()
		imgName := fmt.Sprintf("%s/sections/%s/%d.png", pageID, sectionID, ts)
		imgURL, err := s.objectStorage.Upload(ctx, imgName, bytes.NewReader(imgBytes), int64(len(imgBytes)), "image/png")
		if err != nil {
			logger.Error("Failed to upload section screenshot", zap.String("section_id", sec.ID), zap.Error(err))
			continue
		}

		htmlURL := ""
		if sec.HTML != "" {
			htmlName := fmt.Sprintf("%s/sections/%s/%d.html", pageID, sectionID, ts)
			htmlURL, _ = s.objectStorage.Upload(ctx, htmlName, strings.NewReader(sec.HTML), int64(len(sec.HTML)), "text/html")
		}

		contentHash := sha256.Sum256([]byte(sharedHTML.ExtractText(sec.HTML)))
		contentHashStr := hex.EncodeToString(contentHash[:])

		sectionContentBlocks := sharedHTML.ExtractContentBlocks(sec.HTML)
		sectionContentBlockHash := sharedHTML.HashContentBlocks(sectionContentBlocks)

		sectionCheck := entities.NewCheck(pageID, "success", false)
		sectionCheck.SectionID = &section.ID
		sectionCheck.ParentCheckID = &parentCheckID
		sectionCheck.ScreenshotURL = imgURL
		sectionCheck.HTMLSnapshotURL = htmlURL
		sectionCheck.ContentHash = contentHashStr
		sectionCheck.ContentBlockHash = sectionContentBlockHash
		sectionCheck.ScreenshotHash = imagecompare.HashScreenshot(imgBytes)

		prevSectionCheck, _ := checkRepo.GetPreviousSuccessfulBySection(ctx, pageID, &section.ID, uuid.Nil)
		if prevSectionCheck != nil {
			changeDetected, changeSummary, contentDiff := s.detectChange(ctx, prevSectionCheck, sectionCheck, imgBytes, sec.ScreenshotBase64, targetURL, sec.HTML)
			if changeDetected {
				sectionCheck.ChangeDetected = true
				sectionCheck.ChangeType = "content"
				sectionCheck.VisionChangeSummary = changeSummary
				if contentDiff != nil && contentDiff.HasChanges {
					if diffJSON, err := json.Marshal(contentDiff); err == nil {
						sectionCheck.ContentDiffJSON = string(diffJSON)
					}
				}
				anyChanged = true
				if changeSummary != "" {
					changeSummaries = append(changeSummaries, changeSummary)
				}
			}
		}

		if err := checkRepo.Create(ctx, sectionCheck); err != nil {
			logger.Error("Failed to create section check", zap.String("section_id", sec.ID), zap.Error(err))
			continue
		}
		s.notifyCheckDone(sectionCheck)

		if firstScreenshotURL == "" {
			firstScreenshotURL = imgURL
		}
	}

	// Create a single aggregated alert for all section changes
	if anyChanged && sliceContains(enabledAlertConditions, "any_changes") {
		aggregatedSummary := strings.Join(changeSummaries, "; ")
		// Use the parent check for the alert to avoid FK issues with section checks
		parentCheck, err := checkRepo.GetByID(ctx, parentCheckID)
		if err != nil {
			logger.Error("Failed to retrieve parent check for aggregated alert",
				zap.Error(err), zap.String("parent_check_id", parentCheckID.String()))
		}
		if parentCheck != nil {
			s.createAlert(ctx, schemaName, parentCheck, targetURL, aggregatedSummary)
		}
	}

	if firstScreenshotURL != "" {
		if err := s.updatePageSnapshotMetadata(ctx, schemaName, pageID, firstScreenshotURL, anyChanged); err != nil {
			logger.Error("Failed to update page snapshot metadata from sections", zap.Error(err))
		}
	}

	return anyChanged
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
