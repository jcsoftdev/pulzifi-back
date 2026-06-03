package application_test

import (
	"context"
	"encoding/base64"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	snapp "github.com/jcsoftdev/pulzifi-back/modules/snapshot/application"
	snapEntities "github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/repositories"
	snapServices "github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/services/mocks"
	sharedHTML "github.com/jcsoftdev/pulzifi-back/shared/html"
)

// TestE2E_MonitoringPipeline_ContentChange_FiresAlertAndInsight is the end-to-end
// (P1) test for the monitoring core: it drives a single check through the real
// SnapshotWorker.ExecuteCheck pipeline — extraction → storage → content-diff
// change detection → alert creation → async insight dispatch → check finalization —
// using in-memory fakes for every external boundary (no Chromium, no LLM, no MinIO).
//
// The trigger is a structural content change (Stage 2 content-block diff) with no
// previous screenshot, which is the most deterministic detection path: it does not
// depend on PNG pixel comparison internals.
func TestE2E_MonitoringPipeline_ContentChange_FiresAlertAndInsight(t *testing.T) {
	const (
		schemaName = "tenant_e2e"
		targetURL  = "https://example.com/pricing"
		prevHTML   = `<html><body><h1>Pricing</h1><p>Price is 100 dollars</p></body></html>`
		newHTML    = `<html><body><h1>Pricing</h1><p>Price is 200 dollars</p></body></html>`
		newText    = "Pricing Price is 200 dollars"
		prevHTMLURL = "tenant_e2e/prev.html"
	)

	pageID := uuid.New()

	// ── Object storage fake: in-memory blob map keyed by object name. ──────────
	storage := newFakeObjectStorage()
	// Seed the previous check's HTML so the content-diff stage can fetch it.
	storage.seed(prevHTMLURL, []byte(prevHTML))

	// ── Previous successful check: differs structurally from the new content. ──
	prevCheck := &snapEntities.Check{
		ID:               uuid.New(),
		PageID:           pageID,
		Status:           "success",
		HTMLSnapshotURL:  prevHTMLURL,
		ContentBlockHash: sharedHTML.HashContentBlocks(sharedHTML.ExtractContentBlocks(prevHTML)),
		// ScreenshotURL intentionally empty: skips pixel cross-validation so the
		// content diff alone decides the verdict (deterministic).
	}

	// ── Pending current check the worker will load and finalize. ───────────────
	currentCheck := &snapEntities.Check{
		ID:        uuid.New(),
		PageID:    pageID,
		Status:    "pending",
		CheckedAt: time.Now(),
	}

	checkRepo := newFakeCheckRepo()
	checkRepo.put(currentCheck)
	checkRepo.prev = prevCheck

	extractor := &mocks.MockExtractorClient{
		ExtractResult: &snapServices.ExtractorResult{
			Title:            "Pricing",
			HTML:             newHTML,
			Text:             newText,
			ScreenshotBase64: base64.StdEncoding.EncodeToString([]byte("current-screenshot-bytes")),
		},
	}

	alertCreator := &mocks.MockAlertCreator{}
	insight := newFakeInsightDispatcher()

	worker := snapp.NewSnapshotWorker(snapp.SnapshotWorkerDeps{
		ObjectStorage:    storage,
		ExtractorClient:  extractor,
		MonitoringReader: &fakeMonitoringReader{},
		CheckRepoFactory: func(string) snapServices.CheckRepository { return checkRepo },
		AlertCreator:     alertCreator,
		// NotificationEmailer left nil — worker guards nil and skips email.
		InsightDispatcher: insight,
		FrontendURL:       "https://app.test",
	})

	// ── Act: run the full check pipeline. ──────────────────────────────────────
	if err := worker.ExecuteCheck(context.Background(), currentCheck.ID, targetURL, schemaName); err != nil {
		t.Fatalf("ExecuteCheck returned error: %v", err)
	}

	// ── Assert: check finalized as a successful, change-detected check. ────────
	got, err := checkRepo.GetByID(context.Background(), currentCheck.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Status != "success" {
		t.Errorf("check status = %q, want %q", got.Status, "success")
	}
	if !got.ChangeDetected {
		t.Error("ChangeDetected = false, want true")
	}
	if got.ChangeType != "content" {
		t.Errorf("ChangeType = %q, want %q", got.ChangeType, "content")
	}
	if got.ContentDiffJSON == "" {
		t.Error("ContentDiffJSON is empty, want a serialized content diff")
	}

	// ── Assert: an alert was created with the content-change contract. ─────────
	if alertCreator.CreateCalls != 1 {
		t.Fatalf("AlertCreator.Create calls = %d, want 1", alertCreator.CreateCalls)
	}

	// ── Assert: insight generation was dispatched (async goroutine). ───────────
	select {
	case req := <-insight.dispatched:
		if req.CheckID != currentCheck.ID {
			t.Errorf("insight CheckID = %v, want %v", req.CheckID, currentCheck.ID)
		}
		if req.NewText != newText {
			t.Errorf("insight NewText = %q, want %q", req.NewText, newText)
		}
		if req.PageURL != targetURL {
			t.Errorf("insight PageURL = %q, want %q", req.PageURL, targetURL)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for insight dispatch")
	}
}

// TestE2E_MonitoringPipeline_NoPreviousCheck_NoChangeNoAlert verifies the
// first-ever check (no baseline to compare against) stores a snapshot but does
// not detect a change, create an alert, or dispatch insights.
func TestE2E_MonitoringPipeline_NoPreviousCheck_NoChangeNoAlert(t *testing.T) {
	pageID := uuid.New()

	storage := newFakeObjectStorage()
	currentCheck := &snapEntities.Check{ID: uuid.New(), PageID: pageID, Status: "pending", CheckedAt: time.Now()}

	checkRepo := newFakeCheckRepo()
	checkRepo.put(currentCheck)
	// prev stays nil — no baseline.

	extractor := &mocks.MockExtractorClient{
		ExtractResult: &snapServices.ExtractorResult{
			Title:            "Pricing",
			HTML:             `<html><body><p>First capture</p></body></html>`,
			Text:             "First capture",
			ScreenshotBase64: base64.StdEncoding.EncodeToString([]byte("first-screenshot")),
		},
	}
	alertCreator := &mocks.MockAlertCreator{}
	insight := newFakeInsightDispatcher()

	worker := snapp.NewSnapshotWorker(snapp.SnapshotWorkerDeps{
		ObjectStorage:     storage,
		ExtractorClient:   extractor,
		MonitoringReader:  &fakeMonitoringReader{},
		CheckRepoFactory:  func(string) snapServices.CheckRepository { return checkRepo },
		AlertCreator:      alertCreator,
		InsightDispatcher: insight,
		FrontendURL:       "https://app.test",
	})

	if err := worker.ExecuteCheck(context.Background(), currentCheck.ID, "https://example.com", "tenant_e2e"); err != nil {
		t.Fatalf("ExecuteCheck returned error: %v", err)
	}

	got, _ := checkRepo.GetByID(context.Background(), currentCheck.ID)
	if got.Status != "success" {
		t.Errorf("check status = %q, want %q", got.Status, "success")
	}
	if got.ChangeDetected {
		t.Error("ChangeDetected = true on first check, want false")
	}
	if alertCreator.CreateCalls != 0 {
		t.Errorf("AlertCreator.Create calls = %d, want 0 on first check", alertCreator.CreateCalls)
	}
	select {
	case <-insight.dispatched:
		t.Error("insight dispatched on first check, want none")
	case <-time.After(200 * time.Millisecond):
		// expected: no dispatch
	}
}

// ─── Fakes ───────────────────────────────────────────────────────────────────

// fakeObjectStorage is an in-memory ObjectStorage. Upload returns the object name
// as its URL; Download resolves that URL back to the stored bytes.
type fakeObjectStorage struct {
	mu   sync.Mutex
	blob map[string][]byte
}

func newFakeObjectStorage() *fakeObjectStorage {
	return &fakeObjectStorage{blob: make(map[string][]byte)}
}

func (f *fakeObjectStorage) seed(url string, data []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.blob[url] = data
}

func (f *fakeObjectStorage) Upload(_ context.Context, objectName string, reader io.Reader, _ int64, _ string) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.blob[objectName] = data
	return objectName, nil
}

func (f *fakeObjectStorage) Download(_ context.Context, objectURL string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.blob[objectURL], nil
}

func (f *fakeObjectStorage) Delete(_ context.Context, objectName string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.blob, objectName)
	return nil
}

func (f *fakeObjectStorage) EnsureBucket(context.Context) error { return nil }

// ObjectNameFromURL returns the storedURL as-is because this fake stores blobs
// keyed by the objectName, and Upload returns that same objectName as the URL.
func (f *fakeObjectStorage) ObjectNameFromURL(storedURL string) string {
	return storedURL
}

func (f *fakeObjectStorage) PresignedGetURL(_ context.Context, storedURL string, _ time.Duration) (string, error) {
	return storedURL, nil
}

var _ repositories.ObjectStorage = (*fakeObjectStorage)(nil)

// fakeCheckRepo is an in-memory CheckRepository. `prev` is returned as the
// previous successful check for any page.
type fakeCheckRepo struct {
	mu     sync.Mutex
	checks map[uuid.UUID]*snapEntities.Check
	prev   *snapEntities.Check
}

func newFakeCheckRepo() *fakeCheckRepo {
	return &fakeCheckRepo{checks: make(map[uuid.UUID]*snapEntities.Check)}
}

func (r *fakeCheckRepo) put(c *snapEntities.Check) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checks[c.ID] = c
}

func (r *fakeCheckRepo) Create(_ context.Context, c *snapEntities.Check) error {
	r.put(c)
	return nil
}

func (r *fakeCheckRepo) GetByID(_ context.Context, id uuid.UUID) (*snapEntities.Check, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.checks[id], nil
}

func (r *fakeCheckRepo) Update(_ context.Context, c *snapEntities.Check) error {
	r.put(c)
	return nil
}

func (r *fakeCheckRepo) GetPreviousSuccessfulByPage(_ context.Context, _, _ uuid.UUID) (*snapEntities.Check, error) {
	return r.prev, nil
}

func (r *fakeCheckRepo) GetPreviousSuccessfulBySection(_ context.Context, _ uuid.UUID, _ *uuid.UUID, _ uuid.UUID) (*snapEntities.Check, error) {
	return nil, nil
}

var _ snapServices.CheckRepository = (*fakeCheckRepo)(nil)

// fakeInsightDispatcher records each dispatch on a buffered channel so the test
// can assert the async insight generation fired.
type fakeInsightDispatcher struct {
	dispatched chan snapServices.InsightRequest
}

func newFakeInsightDispatcher() *fakeInsightDispatcher {
	return &fakeInsightDispatcher{dispatched: make(chan snapServices.InsightRequest, 4)}
}

func (d *fakeInsightDispatcher) Dispatch(_ context.Context, req snapServices.InsightRequest) error {
	d.dispatched <- req
	return nil
}

var _ snapServices.InsightDispatcher = (*fakeInsightDispatcher)(nil)

// fakeMonitoringReader returns benign defaults for every read so the pipeline
// runs without a database.
type fakeMonitoringReader struct{}

func (fakeMonitoringReader) GetConfigByPageID(context.Context, string, uuid.UUID) (*snapEntities.MonitoringConfig, error) {
	return nil, nil
}

func (fakeMonitoringReader) ListSectionsByPageID(context.Context, string, uuid.UUID) ([]*snapEntities.MonitoredSection, error) {
	return nil, nil
}

func (fakeMonitoringReader) GetEmailEnabledByPage(context.Context, string, uuid.UUID) ([]*snapEntities.NotificationPreference, error) {
	return nil, nil
}

func (fakeMonitoringReader) GetUserEmail(context.Context, uuid.UUID) (string, error) {
	return "", nil
}

func (fakeMonitoringReader) GetWorkspaceIDForPage(context.Context, string, uuid.UUID) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (fakeMonitoringReader) GetPageTitle(context.Context, string, uuid.UUID) (string, error) {
	return "Test Page", nil
}

func (fakeMonitoringReader) GetOrgIDBySchemaName(context.Context, string) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (fakeMonitoringReader) UpdatePageSnapshotMetadata(context.Context, string, uuid.UUID, string, bool) error {
	return nil
}

var _ snapServices.MonitoringReader = (*fakeMonitoringReader)(nil)
