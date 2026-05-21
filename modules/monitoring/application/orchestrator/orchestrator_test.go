package orchestrator

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
)

// ── Test doubles ──────────────────────────────────────────────────────────────

type mockUsageRepo struct {
	hasQuotaResult bool
	hasQuotaErr    error
	logUsageErr    error

	hasQuotaCalls int
	logUsageCalls int
}

func (m *mockUsageRepo) HasQuota(_ context.Context) (bool, error) {
	m.hasQuotaCalls++
	return m.hasQuotaResult, m.hasQuotaErr
}

func (m *mockUsageRepo) LogUsage(_ context.Context, _, _ uuid.UUID) error {
	m.logUsageCalls++
	return m.logUsageErr
}

type mockCheckRepo struct {
	createErr error
	updateErr error

	createCalls int
	updateCalls int
	lastStatus  string
}

func (m *mockCheckRepo) Create(_ context.Context, check *entities.Check) error {
	m.createCalls++
	// Simulate assigning an ID so the orchestrator can use it.
	if check.ID == uuid.Nil {
		check.ID = uuid.New()
	}
	return m.createErr
}

func (m *mockCheckRepo) GetByID(_ context.Context, _ uuid.UUID) (*entities.Check, error) {
	return nil, nil
}

func (m *mockCheckRepo) Update(_ context.Context, check *entities.Check) error {
	m.updateCalls++
	m.lastStatus = check.Status
	return m.updateErr
}

type mockPageRepo struct {
	updateLastCheckedErr error
	updateLastCheckedCalls int
}

func (m *mockPageRepo) UpdateLastChecked(_ context.Context, _ uuid.UUID) error {
	m.updateLastCheckedCalls++
	return m.updateLastCheckedErr
}

type mockJobDispatcher struct {
	dispatchErr error
	dispatchCalls int
}

func (m *mockJobDispatcher) Dispatch(_ context.Context, _ uuid.UUID, _ string, _ string) error {
	m.dispatchCalls++
	return m.dispatchErr
}

// mockRepoFactory returns the configured mocks for the same tenant on every call.
type mockRepoFactory struct {
	usageRepo *mockUsageRepo
	checkRepo *mockCheckRepo
	pageRepo  *mockPageRepo
}

func (f *mockRepoFactory) GetUsageRepository(_ string) UsageRepository {
	return f.usageRepo
}

func (f *mockRepoFactory) GetCheckRepository(_ string) CheckRepository {
	return f.checkRepo
}

func (f *mockRepoFactory) GetPageRepository(_ string) PageRepository {
	return f.pageRepo
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func makeJob() CheckJob {
	return CheckJob{
		PageID:     uuid.New(),
		URL:        "https://example.com",
		SchemaName: "test_tenant",
	}
}

func TestOrchestrator_ScheduleCheck_HappyPath(t *testing.T) {
	usageRepo := &mockUsageRepo{hasQuotaResult: true}
	checkRepo := &mockCheckRepo{}
	pageRepo := &mockPageRepo{}
	dispatcher := &mockJobDispatcher{}

	factory := &mockRepoFactory{usageRepo: usageRepo, checkRepo: checkRepo, pageRepo: pageRepo}
	o := NewOrchestrator(factory, dispatcher)

	err := o.ScheduleCheck(context.Background(), makeJob())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if usageRepo.hasQuotaCalls != 1 {
		t.Errorf("HasQuota calls: want 1, got %d", usageRepo.hasQuotaCalls)
	}
	if usageRepo.logUsageCalls != 1 {
		t.Errorf("LogUsage calls: want 1, got %d", usageRepo.logUsageCalls)
	}
	if checkRepo.createCalls != 1 {
		t.Errorf("Check Create calls: want 1, got %d", checkRepo.createCalls)
	}
	if pageRepo.updateLastCheckedCalls != 1 {
		t.Errorf("UpdateLastChecked calls: want 1, got %d", pageRepo.updateLastCheckedCalls)
	}
	if dispatcher.dispatchCalls != 1 {
		t.Errorf("Dispatch calls: want 1, got %d", dispatcher.dispatchCalls)
	}
}

func TestOrchestrator_ScheduleCheck_QuotaExceeded(t *testing.T) {
	usageRepo := &mockUsageRepo{hasQuotaResult: false}
	checkRepo := &mockCheckRepo{}
	pageRepo := &mockPageRepo{}
	dispatcher := &mockJobDispatcher{}

	factory := &mockRepoFactory{usageRepo: usageRepo, checkRepo: checkRepo, pageRepo: pageRepo}
	o := NewOrchestrator(factory, dispatcher)

	err := o.ScheduleCheck(context.Background(), makeJob())
	if !errors.Is(err, ErrQuotaExceeded) {
		t.Fatalf("expected ErrQuotaExceeded, got: %v", err)
	}

	// Check must NOT be created
	if checkRepo.createCalls != 0 {
		t.Errorf("expected 0 Check.Create calls when quota exceeded, got %d", checkRepo.createCalls)
	}
	// Dispatch must NOT be called
	if dispatcher.dispatchCalls != 0 {
		t.Errorf("expected 0 Dispatch calls when quota exceeded, got %d", dispatcher.dispatchCalls)
	}
	// Page timestamp MUST be updated even on quota exceeded
	if pageRepo.updateLastCheckedCalls != 1 {
		t.Errorf("expected UpdateLastChecked to be called even on quota exceeded, got %d", pageRepo.updateLastCheckedCalls)
	}
}

func TestOrchestrator_ScheduleCheck_DispatchFailure_MarksCheckAsError(t *testing.T) {
	usageRepo := &mockUsageRepo{hasQuotaResult: true}
	checkRepo := &mockCheckRepo{}
	pageRepo := &mockPageRepo{}
	dispatchErr := errors.New("extractor unreachable")
	dispatcher := &mockJobDispatcher{dispatchErr: dispatchErr}

	factory := &mockRepoFactory{usageRepo: usageRepo, checkRepo: checkRepo, pageRepo: pageRepo}
	o := NewOrchestrator(factory, dispatcher)

	err := o.ScheduleCheck(context.Background(), makeJob())
	if err == nil {
		t.Fatal("expected error from dispatch failure, got nil")
	}
	if !errors.Is(err, dispatchErr) {
		t.Errorf("expected dispatch error %v, got %v", dispatchErr, err)
	}

	// Check.Update must have been called with status "error"
	if checkRepo.updateCalls != 1 {
		t.Errorf("expected 1 Check.Update call on dispatch failure, got %d", checkRepo.updateCalls)
	}
	if checkRepo.lastStatus != "error" {
		t.Errorf("expected check status 'error', got %q", checkRepo.lastStatus)
	}
}

func TestOrchestrator_ScheduleCheck_HasQuotaRepoError_ReturnsImmediately(t *testing.T) {
	dbErr := errors.New("database connection lost")
	usageRepo := &mockUsageRepo{hasQuotaErr: dbErr}
	checkRepo := &mockCheckRepo{}
	pageRepo := &mockPageRepo{}
	dispatcher := &mockJobDispatcher{}

	factory := &mockRepoFactory{usageRepo: usageRepo, checkRepo: checkRepo, pageRepo: pageRepo}
	o := NewOrchestrator(factory, dispatcher)

	err := o.ScheduleCheck(context.Background(), makeJob())
	if err == nil {
		t.Fatal("expected error from HasQuota failure, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Errorf("expected db error %v, got %v", dbErr, err)
	}

	// Nothing should be created or dispatched
	if checkRepo.createCalls != 0 {
		t.Errorf("expected 0 Check.Create calls, got %d", checkRepo.createCalls)
	}
	if dispatcher.dispatchCalls != 0 {
		t.Errorf("expected 0 Dispatch calls, got %d", dispatcher.dispatchCalls)
	}
}
