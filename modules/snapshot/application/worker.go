package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	snapEntities "github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/repositories"
	snapServices "github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	sharedHTML "github.com/jcsoftdev/pulzifi-back/shared/html"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type SnapshotWorker struct {
	objectStorage       repositories.ObjectStorage
	extractorClient     snapServices.ExtractorClient
	monitoringReader    snapServices.MonitoringReader
	checkRepoFactory    func(schemaName string) snapServices.CheckRepository
	alertCreator        snapServices.AlertCreator
	notificationEmailer snapServices.NotificationEmailer
	insightDispatcher   snapServices.InsightDispatcher
	visionAnalyzer      snapServices.VisionAnalyzer
	bus                 eventbus.MessageBus
	storageFlagReader   repositories.StorageFlagReader
	sectionNamer        sharedHTML.SectionNamer
	frontendURL         string
	pixelDiffThreshold  float64
	onCheckDone         func(pageID uuid.UUID, checkJSON []byte)
}

// SetOnCheckDone registers a callback invoked after every check completes
// (success or error) with the serialized CheckResponse payload.
func (s *SnapshotWorker) SetOnCheckDone(fn func(pageID uuid.UUID, checkJSON []byte)) {
	s.onCheckDone = fn
}

// SetVisionAnalyzer sets the vision AI analyzer for screenshot comparison.
func (s *SnapshotWorker) SetVisionAnalyzer(analyzer snapServices.VisionAnalyzer) {
	s.visionAnalyzer = analyzer
}

// SetPixelDiffThreshold sets the threshold for pixel comparison (default 0.001).
func (s *SnapshotWorker) SetPixelDiffThreshold(threshold float64) {
	s.pixelDiffThreshold = threshold
}

// SetMessageBus enables domain-event publishing for cross-module integration dispatch.
// If not set, the worker silently skips publishing (e.g. in tests / standalone runs).
func (s *SnapshotWorker) SetMessageBus(bus eventbus.MessageBus) {
	s.bus = bus
}

// notifyCheckDone serializes a check into the same DTO format the frontend
// expects and invokes the onCheckDone callback if set.
func (s *SnapshotWorker) notifyCheckDone(check *snapEntities.Check) {
	if s.onCheckDone == nil {
		return
	}
	type checkResponse struct {
		ID              uuid.UUID  `json:"id"`
		PageID          uuid.UUID  `json:"page_id"`
		SectionID       *uuid.UUID `json:"section_id,omitempty"`
		Status          string     `json:"status"`
		ScreenshotURL   string     `json:"screenshot_url"`
		HTMLSnapshotURL string     `json:"html_snapshot_url"`
		ChangeDetected  bool       `json:"change_detected"`
		ChangeType      string     `json:"change_type"`
		ErrorMessage    string     `json:"error_message,omitempty"`
		CheckedAt       string     `json:"checked_at"`
	}
	// SectionID must travel with the event: the checks-history UI keeps only
	// full-page checks and filters section checks by this field — without it a
	// section completion renders as a duplicate history entry.
	payload, err := json.Marshal(checkResponse{
		ID:              check.ID,
		PageID:          check.PageID,
		SectionID:       check.SectionID,
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

// SnapshotWorkerDeps groups all dependencies for the SnapshotWorker.
type SnapshotWorkerDeps struct {
	ObjectStorage       repositories.ObjectStorage
	ExtractorClient     snapServices.ExtractorClient
	MonitoringReader    snapServices.MonitoringReader
	CheckRepoFactory    func(schemaName string) snapServices.CheckRepository
	AlertCreator        snapServices.AlertCreator
	NotificationEmailer snapServices.NotificationEmailer
	InsightDispatcher   snapServices.InsightDispatcher
	StorageFlagReader   repositories.StorageFlagReader
	SectionNamer        sharedHTML.SectionNamer
	FrontendURL         string
}

func NewSnapshotWorker(deps SnapshotWorkerDeps) *SnapshotWorker {
	return &SnapshotWorker{
		objectStorage:       deps.ObjectStorage,
		extractorClient:     deps.ExtractorClient,
		monitoringReader:    deps.MonitoringReader,
		checkRepoFactory:    deps.CheckRepoFactory,
		alertCreator:        deps.AlertCreator,
		notificationEmailer: deps.NotificationEmailer,
		insightDispatcher:   deps.InsightDispatcher,
		storageFlagReader:   deps.StorageFlagReader,
		sectionNamer:        deps.SectionNamer,
		frontendURL:         deps.FrontendURL,
		pixelDiffThreshold:  0.001, // default
	}
}

// objectKey returns the storage object key for the given suffix, optionally
// prefixed with {schemaName}/ when the per-org "snapshot.private_storage" flag
// is enabled. Default-deny: any error or nil reader means no prefix.
func (s *SnapshotWorker) objectKey(ctx context.Context, schemaName, suffix string) string {
	if s.storageFlagReader == nil || schemaName == "" {
		return suffix
	}
	enabled, err := s.storageFlagReader.IsPrivateStorageEnabled(ctx, schemaName)
	if err != nil || !enabled {
		return suffix
	}
	return schemaName + "/" + suffix
}

func (s *SnapshotWorker) ExecuteCheck(ctx context.Context, checkID uuid.UUID, targetURL string, schemaName string) error {
	logger.Info("SnapshotWorker executing check", zap.String("check_id", checkID.String()), zap.String("url", targetURL))

	checkRepo := s.checkRepoFactory(schemaName)
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
	pageConfig, configErr := s.monitoringReader.GetConfigByPageID(ctx, schemaName, check.PageID)
	if configErr != nil {
		logger.Warn("Failed to load monitoring config, using defaults",
			zap.String("page_id", check.PageID.String()), zap.Error(configErr))
	}

	enabledInsightTypes, enabledAlertConditions := resolveEnabledTypes(pageConfig)

	extractOpts := buildElementExtractOptions(pageConfig)

	// Sections mode fully handles the check (extraction, per-section processing,
	// parent-check update) and short-circuits the standard full-page path.
	if isSectionsMode(pageConfig) {
		handled, err := s.executeSectionsCheck(ctx, checkRepo, schemaName, check, targetURL, extractOpts, enabledAlertConditions, markError)
		if handled {
			return err
		}
		// Sections unavailable — fall through to full-page processing.
	}

	startTime := time.Now()
	res, err := s.extractorClient.Extract(ctx, targetURL, extractOpts)
	duration := int(time.Since(startTime).Milliseconds())

	if err != nil {
		return markError(err.Error(), duration)
	}

	ts := time.Now().Unix()
	imgBytes, imgURL, storeErr := s.storeFullPageSnapshot(ctx, schemaName, check, res, ts, duration, markError)
	if storeErr != nil {
		return storeErr
	}

	// Fetch previous successful check for comparison
	prevCheck := s.getPreviousSuccessfulCheck(ctx, checkRepo, check.PageID, check.ID)

	if prevCheck != nil {
		s.applyChangeDetection(ctx, schemaName, check, prevCheck, res, imgBytes, ts, targetURL, enabledAlertConditions, enabledInsightTypes)
	}

	if err := checkRepo.Update(ctx, check); err != nil {
		return err
	}

	s.notifyCheckDone(check)

	if err := s.monitoringReader.UpdatePageSnapshotMetadata(ctx, schemaName, check.PageID, imgURL, check.ChangeDetected); err != nil {
		logger.Error("Failed to update page snapshot metadata", zap.Error(err), zap.String("page_id", check.PageID.String()))
	}

	return nil
}

// resolveEnabledTypes returns the enabled insight types and alert conditions for
// a page, falling back to defaults when the config is absent or empty.
func resolveEnabledTypes(pageConfig *snapEntities.MonitoringConfig) (enabledInsightTypes, enabledAlertConditions []string) {
	enabledInsightTypes = []string{"marketing", "market_analysis"}
	enabledAlertConditions = []string{"any_changes"}
	if pageConfig != nil {
		if len(pageConfig.EnabledInsightTypes) > 0 {
			enabledInsightTypes = pageConfig.EnabledInsightTypes
		}
		if len(pageConfig.EnabledAlertConditions) > 0 {
			enabledAlertConditions = pageConfig.EnabledAlertConditions
		}
	}
	return enabledInsightTypes, enabledAlertConditions
}

// storeFullPageSnapshot decodes the screenshot, uploads the screenshot and HTML,
// computes the content/screenshot hashes, and populates the check's snapshot
// fields. On any failure it invokes markError and returns the resulting error so
// the caller can return immediately.
func (s *SnapshotWorker) storeFullPageSnapshot(
	ctx context.Context,
	schemaName string,
	check *snapEntities.Check,
	res *snapServices.ExtractorResult,
	ts int64,
	duration int,
	markError func(msg string, duration int) error,
) (imgBytes []byte, imgURL string, err error) {
	// Process Results
	imgBytes, err = base64.StdEncoding.DecodeString(res.ScreenshotBase64)
	if err != nil {
		return nil, "", markError(fmt.Sprintf("failed to decode screenshot: %v", err), duration)
	}

	imgName := s.objectKey(ctx, schemaName, fmt.Sprintf("%s/%d.png", check.PageID, ts))
	htmlName := s.objectKey(ctx, schemaName, fmt.Sprintf("%s/%d.html", check.PageID, ts))

	// Upload
	if s.objectStorage == nil {
		return nil, "", markError("object storage client is not configured", duration)
	}
	imgURL, err = s.objectStorage.Upload(ctx, imgName, bytes.NewReader(imgBytes), int64(len(imgBytes)), "image/png")
	if err != nil {
		return nil, "", markError(fmt.Sprintf("failed to upload screenshot: %v", err), duration)
	}
	htmlURL, err := s.objectStorage.Upload(ctx, htmlName, strings.NewReader(res.HTML), int64(len(res.HTML)), "text/html")
	if err != nil {
		return nil, "", markError(fmt.Sprintf("failed to upload html snapshot: %v", err), duration)
	}

	// Content hash — extract text from HTML (deterministic) instead of
	// Playwright's innerText (rendering-dependent, varies across runs).
	contentHash := sha256.Sum256([]byte(sharedHTML.ExtractText(res.HTML)))
	contentHashStr := hex.EncodeToString(contentHash[:])

	// Content block hash — structural content representation for content-first detection.
	contentBlocks := sharedHTML.ExtractContentBlocks(res.HTML)
	contentBlockHash := sharedHTML.HashContentBlocks(contentBlocks)

	// Screenshot hash (pixel-based)
	screenshotHash := snapServices.HashScreenshot(imgBytes)

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

	return imgBytes, imgURL, nil
}

// isSectionsMode reports whether the page is configured for section-based monitoring.
func isSectionsMode(pageConfig *snapEntities.MonitoringConfig) bool {
	return pageConfig != nil && pageConfig.SelectorType == "sections"
}

// buildElementExtractOptions builds the base extract options from a page config,
// including element-selector settings. Section options are added separately.
func buildElementExtractOptions(pageConfig *snapEntities.MonitoringConfig) snapServices.ExtractOptions {
	extractOpts := snapServices.ExtractOptions{}
	if pageConfig == nil {
		return extractOpts
	}

	extractOpts.BlockAdsCookies = pageConfig.BlockAdsCookies
	extractOpts.IgnoreSelectors = pageConfig.IgnoreSelectors

	if pageConfig.SelectorType == "element" && pageConfig.CSSSelector != "" {
		extractOpts.Selector = pageConfig.CSSSelector
		extractOpts.SelectorXPath = pageConfig.XPathSelector
		if pageConfig.SelectorOffsets != nil {
			extractOpts.SelectorOffsets = &snapServices.SelectorOffsets{
				Top:    pageConfig.SelectorOffsets.Top,
				Right:  pageConfig.SelectorOffsets.Right,
				Bottom: pageConfig.SelectorOffsets.Bottom,
				Left:   pageConfig.SelectorOffsets.Left,
			}
		}
	}

	return extractOpts
}

// executeSectionsCheck runs the sections-mode pipeline. It returns handled=false
// (with no side effects) when sections are unavailable so the caller can fall
// back to full-page processing. When handled=true, err carries the check outcome.
func (s *SnapshotWorker) executeSectionsCheck(
	ctx context.Context,
	checkRepo snapServices.CheckRepository,
	schemaName string,
	check *snapEntities.Check,
	targetURL string,
	extractOpts snapServices.ExtractOptions,
	enabledAlertConditions []string,
	markError func(msg string, duration int) error,
) (handled bool, err error) {
	// Query sections and pass them to the extractor for a single page load.
	pageSections, err := s.monitoringReader.ListSectionsByPageID(ctx, schemaName, check.PageID)
	if err != nil {
		logger.Error("Failed to load monitored sections, falling back to full-page",
			zap.String("page_id", check.PageID.String()), zap.Error(err))
		return false, nil
	}
	if len(pageSections) == 0 {
		logger.Warn("No monitored sections found for sections-mode page, falling back to full-page",
			zap.String("page_id", check.PageID.String()))
		return false, nil
	}

	logger.Info("Processing sections mode",
		zap.String("page_id", check.PageID.String()),
		zap.Int("section_count", len(pageSections)))

	sectionsByID := make(map[uuid.UUID]*snapEntities.MonitoredSection, len(pageSections))
	for _, sec := range pageSections {
		sectionsByID[sec.ID] = sec
		opt := snapServices.SectionExtractOption{ID: sec.ID.String(), Selector: sec.CSSSelector, SelectorXPath: sec.XPathSelector}
		if sec.SelectorOffsets != nil {
			opt.Offsets = &snapServices.SelectorOffsets{Top: sec.SelectorOffsets.Top, Right: sec.SelectorOffsets.Right, Bottom: sec.SelectorOffsets.Bottom, Left: sec.SelectorOffsets.Left}
		}
		extractOpts.Sections = append(extractOpts.Sections, opt)
	}

	startTime := time.Now()
	res, err := s.extractorClient.Extract(ctx, targetURL, extractOpts)
	duration := int(time.Since(startTime).Milliseconds())
	if err != nil {
		return true, markError(err.Error(), duration)
	}

	logger.Info("Extractor returned sections result",
		zap.String("page_id", check.PageID.String()),
		zap.Int("section_results", len(res.Sections)),
		zap.Bool("has_full_page_screenshot", res.ScreenshotBase64 != ""))

	// Store full-page screenshot on the parent check if available.
	if res.ScreenshotBase64 != "" {
		if imgBytes, decErr := base64.StdEncoding.DecodeString(res.ScreenshotBase64); decErr == nil && len(imgBytes) > 0 {
			ts := time.Now().Unix()
			imgName := s.objectKey(ctx, schemaName, fmt.Sprintf("%s/%d.png", check.PageID, ts))
			if imgURL, upErr := s.objectStorage.Upload(ctx, imgName, bytes.NewReader(imgBytes), int64(len(imgBytes)), "image/png"); upErr == nil {
				check.ScreenshotURL = imgURL
			}
		}
	}

	// Mark parent check as complete.
	check.Status = "success"
	check.DurationMs = duration
	if err := checkRepo.Update(ctx, check); err != nil {
		return true, err
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
	return true, nil
}

// applyChangeDetection runs change detection against the previous check and,
// when a change is found, populates the diff fields on check, uploads the diff
// image, fires alerts, and kicks off async insight generation.
func (s *SnapshotWorker) applyChangeDetection(
	ctx context.Context,
	schemaName string,
	check, prevCheck *snapEntities.Check,
	res *snapServices.ExtractorResult,
	imgBytes []byte,
	ts int64,
	targetURL string,
	enabledAlertConditions, enabledInsightTypes []string,
) {
	changeDetected, changeSummary, contentDiff, diffImgBytes := s.detectChange(ctx, prevCheck, check, imgBytes, res.ScreenshotBase64, targetURL, res.HTML)
	if !changeDetected {
		return
	}

	check.ChangeDetected = true
	check.ChangeType = "content"
	check.VisionChangeSummary = changeSummary

	// Upload diff image if generated
	if len(diffImgBytes) > 0 {
		diffName := s.objectKey(ctx, schemaName, fmt.Sprintf("%s/diffs/%d.png", check.PageID, ts))
		if diffURL, upErr := s.objectStorage.Upload(ctx, diffName, bytes.NewReader(diffImgBytes), int64(len(diffImgBytes)), "image/png"); upErr == nil {
			check.DiffImageURL = diffURL
		} else {
			logger.Error("Failed to upload diff image", zap.Error(upErr), zap.String("page_id", check.PageID.String()))
		}
	}

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
	if s.insightDispatcher != nil && len(enabledInsightTypes) > 0 {
		var diffText string
		if contentDiff != nil && contentDiff.HasChanges {
			diffText = sharedHTML.FormatDiffForAI(contentDiff)
		}
		prevText := s.fetchTextFromURL(prevCheck.HTMLSnapshotURL)
		go s.generateInsightsAsync(check, targetURL, prevText, res.Text, schemaName, enabledInsightTypes, diffText)
	}
}

// detectChange runs a multi-stage content-first change detection pipeline:
//
//	Stage 1: Content block hash (fast structural identity)
//	Stage 2: Content block diff (structural comparison for diff text)
//	Stage 3: Pixel comparison (visual-only changes when content is unchanged)
//	Stage 4: Vision AI semantic analysis (optional)
//	Stage 5: Normalized text hash fallback (legacy compatibility)
//
// Returns (changeDetected, changeSummary, contentDiff, diffImageBytes)
func (s *SnapshotWorker) detectChange(ctx context.Context, prevCheck, currCheck *snapEntities.Check, currImgBytes []byte, currBase64 string, pageURL string, currHTML string) (bool, string, *sharedHTML.ContentDiff, []byte) {
	pageID := currCheck.PageID.String()

	// ── Stage 1: Content block hash comparison ───────────────────────────
	if prevCheck.ContentBlockHash != "" {
		if prevCheck.ContentBlockHash == currCheck.ContentBlockHash {
			logger.Info("Content block hash identical — checking visual changes",
				zap.String("page_id", pageID))

			// Content is structurally identical — check for visual-only changes.
			res := s.detectVisualOnlyChange(ctx, prevCheck, currCheck, currImgBytes, currBase64, pageURL, pageID)
			return res.detected, res.summary, res.diff, res.img
		}

		// ── Stage 2: Content block diff ──────────────────────────────────
		if res, decided := s.detectContentBlockDiff(prevCheck, currCheck, currImgBytes, currHTML, pageID); decided {
			return res.detected, res.summary, res.diff, res.img
		}
	}

	// ── Stage 3: Pixel comparison ────────────────────────────────────────
	if res, decided := s.detectPixelChange(ctx, prevCheck, currCheck, currImgBytes, currBase64, pageURL, pageID); decided {
		return res.detected, res.summary, res.diff, res.img
	}

	// ── Stage 5: Normalized text hash fallback ───────────────────────────
	if res, decided := s.detectTextHashFallback(prevCheck, currCheck, currImgBytes, pageID); decided {
		return res.detected, res.summary, res.diff, res.img
	}

	return false, "", nil, nil
}

// changeDecision carries the outcome of a single detection stage.
type changeDecision struct {
	detected bool
	summary  string
	diff     *sharedHTML.ContentDiff
	img      []byte
}

// detectVisualOnlyChange handles the case where content block hashes are
// identical: it looks for visual-only changes via pixel comparison and optional
// Vision AI. This stage always reaches a verdict (no fall-through).
func (s *SnapshotWorker) detectVisualOnlyChange(ctx context.Context, prevCheck, currCheck *snapEntities.Check, currImgBytes []byte, currBase64, pageURL, pageID string) changeDecision {
	// ── Stage 3: Pixel comparison (secondary) ────────────────────
	if prevCheck.ScreenshotHash == "" || prevCheck.ScreenshotHash == currCheck.ScreenshotHash || prevCheck.ScreenshotURL == "" {
		return changeDecision{}
	}

	prevImgBytes := s.downloadScreenshot(prevCheck.ScreenshotURL)
	if len(prevImgBytes) == 0 {
		return changeDecision{}
	}

	result, err := snapServices.CompareScreenshots(prevImgBytes, currImgBytes, s.pixelDiffThreshold)
	if err != nil {
		logger.Error("Pixel comparison failed (image decode error)",
			zap.Error(err), zap.String("page_id", pageID))
		return changeDecision{}
	}
	if result.Identical || result.DiffRatio < s.pixelDiffThreshold {
		return changeDecision{}
	}

	// ── Stage 4: Vision AI (optional) ────────────
	diffImg := snapServices.GenerateDiffImage(prevImgBytes, currImgBytes)
	if s.visionAnalyzer != nil {
		prevB64 := base64.StdEncoding.EncodeToString(prevImgBytes)
		visionResult, vErr := s.visionAnalyzer.AnalyzeChange(ctx, prevB64, currBase64, pageURL)
		if vErr != nil {
			logger.Error("Vision AI failed, reporting visual change",
				zap.Error(vErr), zap.String("page_id", pageID))
			return changeDecision{detected: true, img: diffImg}
		}
		if !visionResult.HasMeaningfulChange {
			logger.Info("Vision AI says no meaningful visual change",
				zap.String("page_id", pageID))
			return changeDecision{}
		}
		return changeDecision{detected: true, summary: visionResult.ChangeSummary, img: diffImg}
	}
	logger.Info("Visual-only change detected via pixel diff",
		zap.String("page_id", pageID),
		zap.Float64("diff_ratio", result.DiffRatio),
		zap.Int("diff_count", result.DiffCount),
		zap.Int("total_pixels", result.TotalPixels))
	return changeDecision{detected: true, img: diffImg}
}

// detectContentBlockDiff handles differing content block hashes by computing a
// structural diff and cross-validating with pixel comparison. decided=false
// means no structural change was found and detection should fall through.
func (s *SnapshotWorker) detectContentBlockDiff(prevCheck, currCheck *snapEntities.Check, currImgBytes []byte, currHTML, pageID string) (changeDecision, bool) {
	// Content block hashes differ — compute structural diff.
	logger.Info("Content block hash differs, computing structural diff",
		zap.String("page_id", pageID),
		zap.String("prev_hash", prevCheck.ContentBlockHash),
		zap.String("curr_hash", currCheck.ContentBlockHash))

	prevHTML := s.fetchHTMLFromURL(prevCheck.HTMLSnapshotURL)
	var contentDiff *sharedHTML.ContentDiff
	if prevHTML != "" {
		prevBlocks := sharedHTML.ExtractContentBlocks(prevHTML)
		currBlocks := sharedHTML.ExtractContentBlocks(currHTML)
		contentDiff = sharedHTML.DiffContentBlocks(prevBlocks, currBlocks)
		sharedHTML.EnrichSections(contentDiff, s.sectionNamer)
	}

	if contentDiff == nil || !contentDiff.HasChanges {
		// Diff computation failed or showed no changes despite hash difference.
		// Fall through to pixel comparison.
		logger.Info("Content block hash differed but diff empty, falling through to pixel",
			zap.String("page_id", pageID))
		return changeDecision{}, false
	}

	logContentDiffs(contentDiff, pageID)
	logger.Info("Content change detected via structural diff",
		zap.String("page_id", pageID),
		zap.Int("total_changes", contentDiff.TotalChanges))

	// Cross-validate: if the visual appearance hasn't changed meaningfully,
	// treat the content diff as noise (e.g., typewriter animations, counters,
	// live clocks, ad rotation). This prevents false positives from dynamic
	// text that doesn't affect the user-visible page.
	if prevCheck.ScreenshotURL != "" {
		prevImgBytes := s.downloadScreenshot(prevCheck.ScreenshotURL)
		if len(prevImgBytes) > 0 {
			pixResult, pixErr := snapServices.CompareScreenshots(prevImgBytes, currImgBytes, s.pixelDiffThreshold)
			if pixErr == nil {
				if pixResult.Identical || pixResult.DiffRatio < s.pixelDiffThreshold {
					logger.Info("Content diff suppressed — pixel diff below threshold (likely animation noise)",
						zap.String("page_id", pageID),
						zap.Int("content_changes", contentDiff.TotalChanges),
						zap.Float64("pixel_diff_ratio", pixResult.DiffRatio),
						zap.Float64("threshold", s.pixelDiffThreshold))
					return changeDecision{}, true
				}
				// Visual change confirmed — report as real content change
				logger.Info("Content diff confirmed by pixel comparison",
					zap.String("page_id", pageID),
					zap.Float64("pixel_diff_ratio", pixResult.DiffRatio))
				diffImg := snapServices.GenerateDiffImage(prevImgBytes, currImgBytes)
				return changeDecision{detected: true, diff: contentDiff, img: diffImg}, true
			}
			logger.Error("Cross-validation pixel comparison failed, reporting content change",
				zap.Error(pixErr), zap.String("page_id", pageID))
			diffImg := snapServices.GenerateDiffImage(prevImgBytes, currImgBytes)
			return changeDecision{detected: true, diff: contentDiff, img: diffImg}, true
		}
	}
	// No previous screenshot available for cross-validation — report change
	return changeDecision{detected: true, diff: contentDiff}, true
}

// logContentDiffs logs the first few structural diffs for diagnosis.
func logContentDiffs(contentDiff *sharedHTML.ContentDiff, pageID string) {
	for i, d := range contentDiff.Diffs {
		if i >= 5 {
			logger.Info("... and more content diffs",
				zap.String("page_id", pageID),
				zap.Int("remaining", contentDiff.TotalChanges-5))
			break
		}
		switch d.Op {
		case sharedHTML.DiffAdded:
			logger.Info("Content diff: ADDED",
				zap.String("page_id", pageID),
				zap.String("type", string(d.Block.Type)),
				zap.String("text", truncateForLog(d.Block.Text, 120)))
		case sharedHTML.DiffRemoved:
			logger.Info("Content diff: REMOVED",
				zap.String("page_id", pageID),
				zap.String("type", string(d.Block.Type)),
				zap.String("text", truncateForLog(d.Block.Text, 120)))
		case sharedHTML.DiffChanged:
			oldText := ""
			if d.OldBlock != nil {
				oldText = truncateForLog(d.OldBlock.Text, 80)
			}
			logger.Info("Content diff: CHANGED",
				zap.String("page_id", pageID),
				zap.String("type", string(d.Block.Type)),
				zap.String("old_text", oldText),
				zap.String("new_text", truncateForLog(d.Block.Text, 80)))
		}
	}
}

// detectPixelChange is the primary pixel-comparison stage. decided=false means
// no screenshot data was available to reach a verdict.
func (s *SnapshotWorker) detectPixelChange(ctx context.Context, prevCheck, currCheck *snapEntities.Check, currImgBytes []byte, currBase64, pageURL, pageID string) (changeDecision, bool) {
	if prevCheck.ScreenshotHash == "" {
		return changeDecision{}, false
	}

	if prevCheck.ScreenshotHash == currCheck.ScreenshotHash {
		logger.Info("Screenshot hash identical — no change", zap.String("page_id", pageID))
		return changeDecision{}, true
	}

	if prevCheck.ScreenshotURL == "" {
		return changeDecision{}, false
	}

	prevImgBytes := s.downloadScreenshot(prevCheck.ScreenshotURL)
	if len(prevImgBytes) == 0 {
		return changeDecision{}, false
	}

	result, err := snapServices.CompareScreenshots(prevImgBytes, currImgBytes, s.pixelDiffThreshold)
	switch {
	case err != nil:
		logger.Error("Pixel comparison failed (image decode error)",
			zap.Error(err), zap.String("page_id", pageID))
		return changeDecision{}, false
	case result.Identical || result.DiffRatio < s.pixelDiffThreshold:
		logger.Info("Pixel diff below threshold — no meaningful change",
			zap.String("page_id", pageID),
			zap.Float64("diff_ratio", result.DiffRatio))
		return changeDecision{}, true
	default:
		diffImg := snapServices.GenerateDiffImage(prevImgBytes, currImgBytes)

		// ── Stage 4: Vision AI analysis (optional) ───────────
		if s.visionAnalyzer != nil {
			prevB64 := base64.StdEncoding.EncodeToString(prevImgBytes)
			visionResult, vErr := s.visionAnalyzer.AnalyzeChange(ctx, prevB64, currBase64, pageURL)
			if vErr != nil {
				logger.Error("Vision AI failed, reporting change based on pixel diff",
					zap.Error(vErr), zap.String("page_id", pageID))
				return changeDecision{detected: true, img: diffImg}, true
			}
			if !visionResult.HasMeaningfulChange {
				return changeDecision{}, true
			}
			return changeDecision{detected: true, summary: visionResult.ChangeSummary, img: diffImg}, true
		}

		logger.Info("Pixel diff above threshold, reporting change",
			zap.String("page_id", pageID),
			zap.Float64("diff_ratio", result.DiffRatio),
			zap.Int("diff_count", result.DiffCount),
			zap.Int("total_pixels", result.TotalPixels),
			zap.Float64("threshold", s.pixelDiffThreshold))
		return changeDecision{detected: true, img: diffImg}, true
	}
}

// detectTextHashFallback is the legacy normalized-text-hash stage used when the
// previous check predates content block hashing.
func (s *SnapshotWorker) detectTextHashFallback(prevCheck, currCheck *snapEntities.Check, currImgBytes []byte, pageID string) (changeDecision, bool) {
	if prevCheck.ContentHash == "" || prevCheck.ContentHash == currCheck.ContentHash {
		return changeDecision{}, false
	}

	logger.Info("Change detected via normalized text hash (Stage 5 fallback)",
		zap.String("page_id", pageID),
		zap.String("prev_content_hash", prevCheck.ContentHash[:16]+"..."),
		zap.String("curr_content_hash", currCheck.ContentHash[:16]+"..."))
	// Generate diff image if previous screenshot is available
	var diffImg []byte
	if prevCheck.ScreenshotURL != "" {
		prevImgBytes := s.downloadScreenshot(prevCheck.ScreenshotURL)
		if len(prevImgBytes) > 0 {
			diffImg = snapServices.GenerateDiffImage(prevImgBytes, currImgBytes)
		}
	}
	return changeDecision{detected: true, img: diffImg}, true
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

func (s *SnapshotWorker) getPreviousSuccessfulCheck(ctx context.Context, repo snapServices.CheckRepository, pageID, currentCheckID uuid.UUID) *snapEntities.Check {
	check, err := repo.GetPreviousSuccessfulByPage(ctx, pageID, currentCheckID)
	if err != nil {
		return nil
	}
	return check
}

func (s *SnapshotWorker) createAlert(ctx context.Context, schemaName string, check *snapEntities.Check, pageURL string, changeSummary string) {
	if s.alertCreator == nil {
		return
	}

	workspaceID, err := s.monitoringReader.GetWorkspaceIDForPage(ctx, schemaName, check.PageID)
	if err != nil {
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

	input := snapServices.AlertInput{
		SchemaName:    schemaName,
		WorkspaceID:   workspaceID,
		PageID:        check.PageID,
		CheckID:       check.ID,
		AlertType:     "content_change",
		Title:         alertTitle,
		Description:   alertDescription,
		ChangeSummary: changeSummary,
	}

	if err := s.alertCreator.Create(ctx, input); err != nil {
		logger.Error("Failed to create alert", zap.Error(err))
	} else {
		logger.Info("Alert created", zap.String("check_id", check.ID.String()))
	}

	// Change-alert emails are delivered through the integration dispatch path
	// (publishChangeDetected → integration dispatch_event → email/Slack/… destinations),
	// which also covers settings-configured recipients, retry, and delivery history.
	// Publish change.detected domain event so the integration dispatcher can fan-out.
	go s.publishChangeDetected(ctx, schemaName, check, pageURL, changeSummary)
}

// publishChangeDetected publishes a change.detected DomainEvent to the EventBus so that
// the integration dispatch_event use case can fan-out to all enabled destinations.
func (s *SnapshotWorker) publishChangeDetected(parentCtx context.Context, tenant string, check *snapEntities.Check, pageURL, changeSummary string) {
	if s.bus == nil {
		return // dispatch disabled — bus not wired (tests, standalone runs)
	}

	// Use a fresh context with timeout — parent may be near-cancellation by the time we publish.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = parentCtx // reserved for future correlation-id propagation; unused now

	orgID, err := s.monitoringReader.GetOrgIDBySchemaName(ctx, tenant)
	if err != nil {
		logger.Error("publishChangeDetected: org lookup failed", zap.Error(err), zap.String("tenant", tenant))
		return
	}

	workspaceID, err := s.monitoringReader.GetWorkspaceIDForPage(ctx, tenant, check.PageID)
	if err != nil {
		logger.Error("publishChangeDetected: page/workspace lookup failed", zap.Error(err), zap.String("page_id", check.PageID.String()))
		return
	}

	pageTitle, err := s.monitoringReader.GetPageTitle(ctx, tenant, check.PageID)
	if err != nil {
		logger.Error("publishChangeDetected: page title lookup failed", zap.Error(err), zap.String("page_id", check.PageID.String()))
		return
	}

	// The dashboard deep link is built from the tenant SUBDOMAIN, which differs
	// from the schema name (hyphens vs underscores). Best-effort: an empty
	// subdomain just drops the deep link (email falls back to the page URL).
	subdomain, err := s.monitoringReader.GetSubdomainBySchemaName(ctx, tenant)
	if err != nil {
		logger.Warn("publishChangeDetected: subdomain lookup failed — deep link will fall back to page URL",
			zap.Error(err), zap.String("tenant", tenant))
	}

	changeType := check.ChangeType
	if changeType == "" {
		changeType = "content"
	}

	payload := map[string]any{
		"page_url":       pageURL,
		"page_title":     pageTitle,
		"change_type":    changeType,
		"diff_summary":   changeSummary,
		"check_id":       check.ID.String(),
		"page_id":        check.PageID.String(),
		"workspace_id":   workspaceID.String(),
		"tenant":         tenant,
		"subdomain":      subdomain,
		"diff_image_url": check.DiffImageURL,
		"changed_at":     check.CheckedAt.Format(time.RFC3339),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		logger.Error("publishChangeDetected: marshal payload", zap.Error(err))
		return
	}

	pid := check.PageID
	wsid := workspaceID
	ev := eventbus.DomainEvent{
		Type:        eventbus.TopicChangeDetected,
		OrgID:       orgID,
		WorkspaceID: &wsid,
		PageID:      &pid,
		Tenant:      tenant,
		Data:        data,
	}
	if err := eventbus.PublishDomainEvent(s.bus, ev); err != nil {
		logger.Error("publishChangeDetected: publish", zap.Error(err))
	}
}

// generateInsightsAsync calls the insight dispatcher in the background after a change is detected.
func (s *SnapshotWorker) generateInsightsAsync(check *snapEntities.Check, pageURL, prevText, newText, schemaName string, enabledTypes []string, diffText string) {
	genCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	err := s.insightDispatcher.Dispatch(genCtx, snapServices.InsightRequest{
		PageID:              check.PageID,
		CheckID:             check.ID,
		PageURL:             pageURL,
		PrevText:            prevText,
		NewText:             newText,
		DiffText:            diffText,
		SchemaName:          schemaName,
		EnabledInsightTypes: enabledTypes,
	})
	switch {
	case err == nil:
		logger.Info("Insights generated successfully", zap.String("check_id", check.ID.String()))
	case errors.Is(err, snapServices.ErrInsightQuotaExhausted):
		// Normal business condition — tenant hit their monthly cap. Not a failure.
		logger.Info("Insights skipped — monthly AI quota reached", zap.String("check_id", check.ID.String()))
	default:
		logger.Error("Failed to generate insights", zap.String("check_id", check.ID.String()), zap.Error(err))
	}
}

// truncateForLog truncates a string for log output.
func truncateForLog(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
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

// fetchTextFromURL downloads HTML from the given URL via the object storage client
// and extracts plain text. Falls back to empty string on any error.
func (s *SnapshotWorker) fetchTextFromURL(rawURL string) string {
	html := s.fetchHTMLFromURL(rawURL)
	if html == "" {
		return ""
	}
	return sharedHTML.ExtractText(html)
}

// fetchHTMLFromURL downloads raw HTML from the given URL using the object storage client.
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
	checkRepo snapServices.CheckRepository,
	schemaName string,
	parentCheckID uuid.UUID,
	pageID uuid.UUID,
	sectionsByID map[uuid.UUID]*snapEntities.MonitoredSection,
	sectionResults []snapServices.SectionExtractResult,
	targetURL string,
	enabledAlertConditions []string,
) bool {
	anyChanged := false
	firstScreenshotURL := ""
	var changeSummaries []string

	for i := range sectionResults {
		outcome, ok := s.processSingleSection(ctx, checkRepo, schemaName, parentCheckID, pageID, sectionsByID, &sectionResults[i], targetURL)
		if !ok {
			continue
		}

		if outcome.changed {
			anyChanged = true
			if outcome.changeSummary != "" {
				changeSummaries = append(changeSummaries, outcome.changeSummary)
			}
		}

		if firstScreenshotURL == "" {
			firstScreenshotURL = outcome.imgURL
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
		if err := s.monitoringReader.UpdatePageSnapshotMetadata(ctx, schemaName, pageID, firstScreenshotURL, anyChanged); err != nil {
			logger.Error("Failed to update page snapshot metadata from sections", zap.Error(err))
		}
	}

	return anyChanged
}

// sectionOutcome reports the result of processing a single monitored section.
type sectionOutcome struct {
	imgURL        string
	changeSummary string
	changed       bool
}

// processSingleSection uploads a section's screenshot/HTML, builds and persists
// its check, and runs change detection against the previous section check. It
// returns ok=false when the section was skipped (no screenshot, invalid ID,
// decode/upload/create failure).
func (s *SnapshotWorker) processSingleSection(
	ctx context.Context,
	checkRepo snapServices.CheckRepository,
	schemaName string,
	parentCheckID uuid.UUID,
	pageID uuid.UUID,
	sectionsByID map[uuid.UUID]*snapEntities.MonitoredSection,
	sec *snapServices.SectionExtractResult,
	targetURL string,
) (sectionOutcome, bool) {
	if sec.ScreenshotBase64 == "" {
		logger.Warn("Section has no screenshot (selector may not have matched)",
			zap.String("section_id", sec.ID),
			zap.Bool("selector_matched", sec.SelectorMatched))
		return sectionOutcome{}, false
	}

	sectionID, err := uuid.Parse(sec.ID)
	if err != nil {
		logger.Warn("Invalid section ID in extractor result", zap.String("id", sec.ID))
		return sectionOutcome{}, false
	}
	section, ok := sectionsByID[sectionID]
	if !ok {
		return sectionOutcome{}, false
	}

	imgBytes, err := base64.StdEncoding.DecodeString(sec.ScreenshotBase64)
	if err != nil || len(imgBytes) == 0 {
		logger.Warn("Failed to decode section screenshot", zap.String("section_id", sec.ID), zap.Error(err))
		return sectionOutcome{}, false
	}

	ts := time.Now().Unix()
	imgName := s.objectKey(ctx, schemaName, fmt.Sprintf("%s/sections/%s/%d.png", pageID, sectionID, ts))
	imgURL, err := s.objectStorage.Upload(ctx, imgName, bytes.NewReader(imgBytes), int64(len(imgBytes)), "image/png")
	if err != nil {
		logger.Error("Failed to upload section screenshot", zap.String("section_id", sec.ID), zap.Error(err))
		return sectionOutcome{}, false
	}

	htmlURL := ""
	if sec.HTML != "" {
		htmlName := s.objectKey(ctx, schemaName, fmt.Sprintf("%s/sections/%s/%d.html", pageID, sectionID, ts))
		htmlURL, _ = s.objectStorage.Upload(ctx, htmlName, strings.NewReader(sec.HTML), int64(len(sec.HTML)), "text/html")
	}

	contentHash := sha256.Sum256([]byte(sharedHTML.ExtractText(sec.HTML)))
	contentHashStr := hex.EncodeToString(contentHash[:])

	sectionContentBlocks := sharedHTML.ExtractContentBlocks(sec.HTML)
	sectionContentBlockHash := sharedHTML.HashContentBlocks(sectionContentBlocks)

	sectionCheck := snapEntities.NewCheck(pageID, "success", false)
	sectionCheck.SectionID = &section.ID
	sectionCheck.ParentCheckID = &parentCheckID
	sectionCheck.ScreenshotURL = imgURL
	sectionCheck.HTMLSnapshotURL = htmlURL
	sectionCheck.ContentHash = contentHashStr
	sectionCheck.ContentBlockHash = sectionContentBlockHash
	sectionCheck.ScreenshotHash = snapServices.HashScreenshot(imgBytes)

	changed, changeSummary := s.detectSectionChange(ctx, checkRepo, schemaName, sectionCheck, section, pageID, sectionID, ts, imgBytes, sec, targetURL)

	if err := checkRepo.Create(ctx, sectionCheck); err != nil {
		logger.Error("Failed to create section check", zap.String("section_id", sec.ID), zap.Error(err))
		return sectionOutcome{}, false
	}
	s.notifyCheckDone(sectionCheck)

	return sectionOutcome{imgURL: imgURL, changeSummary: changeSummary, changed: changed}, true
}

// detectSectionChange runs change detection for a section against its previous
// successful check and, when a change is found, populates the section check's
// diff fields and uploads the diff image. Returns whether a change was detected
// and its summary.
func (s *SnapshotWorker) detectSectionChange(
	ctx context.Context,
	checkRepo snapServices.CheckRepository,
	schemaName string,
	sectionCheck *snapEntities.Check,
	section *snapEntities.MonitoredSection,
	pageID, sectionID uuid.UUID,
	ts int64,
	imgBytes []byte,
	sec *snapServices.SectionExtractResult,
	targetURL string,
) (changed bool, changeSummary string) {
	prevSectionCheck, _ := checkRepo.GetPreviousSuccessfulBySection(ctx, pageID, &section.ID, uuid.Nil)
	if prevSectionCheck == nil {
		return false, ""
	}

	changeDetected, summary, contentDiff, diffImgBytes := s.detectChange(ctx, prevSectionCheck, sectionCheck, imgBytes, sec.ScreenshotBase64, targetURL, sec.HTML)
	if !changeDetected {
		return false, ""
	}

	sectionCheck.ChangeDetected = true
	sectionCheck.ChangeType = "content"
	sectionCheck.VisionChangeSummary = summary
	// Upload diff image for section check
	if len(diffImgBytes) > 0 {
		diffName := s.objectKey(ctx, schemaName, fmt.Sprintf("%s/sections/%s/%d_diff.png", pageID, sectionID, ts))
		if diffURL, upErr := s.objectStorage.Upload(ctx, diffName, bytes.NewReader(diffImgBytes), int64(len(diffImgBytes)), "image/png"); upErr == nil {
			sectionCheck.DiffImageURL = diffURL
		}
	}
	if contentDiff != nil && contentDiff.HasChanges {
		if diffJSON, err := json.Marshal(contentDiff); err == nil {
			sectionCheck.ContentDiffJSON = string(diffJSON)
		}
	}
	return true, summary
}
