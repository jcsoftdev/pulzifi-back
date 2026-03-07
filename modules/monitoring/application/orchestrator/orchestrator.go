package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type JobDispatcher interface {
	Dispatch(ctx context.Context, checkID uuid.UUID, url string, schemaName string) error
}

type UsageRepository interface {
	HasQuota(ctx context.Context) (bool, error)
	LogUsage(ctx context.Context, pageID, checkID uuid.UUID) error
}

type CheckRepository interface {
	Create(ctx context.Context, check *entities.Check) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Check, error)
	Update(ctx context.Context, check *entities.Check) error
}

type PageRepository interface {
	UpdateLastChecked(ctx context.Context, pageID uuid.UUID) error
}

type RepositoryFactory interface {
	GetCheckRepository(tenant string) CheckRepository
	GetPageRepository(tenant string) PageRepository
	GetUsageRepository(tenant string) UsageRepository
}

// checkEvent is the DTO published via SSE when a check is created or fails at dispatch.
// It mirrors listchecks.CheckResponse to keep the SSE payload consistent.
type checkEvent struct {
	ID              uuid.UUID `json:"id"`
	PageID          uuid.UUID `json:"page_id"`
	Status          string    `json:"status"`
	ScreenshotURL   string    `json:"screenshot_url"`
	HTMLSnapshotURL string    `json:"html_snapshot_url"`
	ChangeDetected  bool      `json:"change_detected"`
	ChangeType      string    `json:"change_type"`
	ErrorMessage    string    `json:"error_message,omitempty"`
	CheckedAt       time.Time `json:"checked_at"`
}

type Orchestrator struct {
	repoFactory    RepositoryFactory
	dispatcher     JobDispatcher
	onCheckCreated func(pageID uuid.UUID, checkJSON []byte)
}

var ErrQuotaExceeded = errors.New("quota exceeded")

func NewOrchestrator(repoFactory RepositoryFactory, dispatcher JobDispatcher) *Orchestrator {
	return &Orchestrator{
		repoFactory: repoFactory,
		dispatcher:  dispatcher,
	}
}

// SetOnCheckCreated registers a callback invoked when a check record is created
// (status=pending) or when dispatch fails (status=error). This is used to push
// real-time SSE events via CheckBroker.
func (o *Orchestrator) SetOnCheckCreated(fn func(pageID uuid.UUID, checkJSON []byte)) {
	o.onCheckCreated = fn
}

type CheckJob struct {
	PageID     uuid.UUID
	URL        string
	SchemaName string
	SectionID  *uuid.UUID // nil = full-page check; non-nil = section-specific check
}

// HasQuota checks whether the given tenant has remaining quota for the current billing period.
func (o *Orchestrator) HasQuota(ctx context.Context, tenant string) (bool, error) {
	return o.repoFactory.GetUsageRepository(tenant).HasQuota(ctx)
}

func (o *Orchestrator) ScheduleCheck(ctx context.Context, job CheckJob) error {
	logger.Info("Scheduling check", zap.String("page_id", job.PageID.String()), zap.String("schema", job.SchemaName))

	// Get Repositories for this tenant
	usageRepo := o.repoFactory.GetUsageRepository(job.SchemaName)
	checkRepo := o.repoFactory.GetCheckRepository(job.SchemaName)
	pageRepo := o.repoFactory.GetPageRepository(job.SchemaName)

	// 1. Validate Usage
	hasQuota, err := usageRepo.HasQuota(ctx)
	if err != nil {
		logger.Error("Failed to check quota", zap.Error(err))
		return err
	}
	if !hasQuota {
		logger.Warn("Quota exceeded for tenant", zap.String("schema", job.SchemaName))
		if err := pageRepo.UpdateLastChecked(ctx, job.PageID); err != nil {
			logger.Error("Failed to update last_checked_at", zap.Error(err))
		}
		return ErrQuotaExceeded
	}

	// 2. Create Check (Pending)
	check := entities.NewCheck(job.PageID, "pending", false)
	check.SectionID = job.SectionID
	if err := checkRepo.Create(ctx, check); err != nil {
		logger.Error("Failed to create check", zap.Error(err))
		return err
	}

	// Notify SSE subscribers about the pending check.
	o.publishCheckEvent(check)

	// 3. Log Usage
	if err := usageRepo.LogUsage(ctx, job.PageID, check.ID); err != nil {
		// Log error but proceed
		logger.Error("Failed to log usage", zap.Error(err))
	}

	// 4. Update Page
	if err := pageRepo.UpdateLastChecked(ctx, job.PageID); err != nil {
		logger.Error("Failed to update last_checked_at", zap.Error(err))
	}

	// 5. Dispatch Job
	if err := o.dispatcher.Dispatch(ctx, check.ID, job.URL, job.SchemaName); err != nil {
		check.Status = "error"
		check.ErrorMessage = fmt.Sprintf("dispatch failed: %v", err)
		if updateErr := checkRepo.Update(ctx, check); updateErr != nil {
			logger.Error("Failed to mark check as error after dispatch failure", zap.Error(updateErr), zap.String("check_id", check.ID.String()))
		}
		// Notify SSE subscribers about the dispatch failure.
		o.publishCheckEvent(check)
		return err
	}
	return nil
}

// publishCheckEvent serializes the check entity and invokes the onCheckCreated callback.
func (o *Orchestrator) publishCheckEvent(check *entities.Check) {
	if o.onCheckCreated == nil {
		return
	}
	payload, err := json.Marshal(checkEvent{
		ID:              check.ID,
		PageID:          check.PageID,
		Status:          check.Status,
		ScreenshotURL:   check.ScreenshotURL,
		HTMLSnapshotURL: check.HTMLSnapshotURL,
		ChangeDetected:  check.ChangeDetected,
		ChangeType:      check.ChangeType,
		ErrorMessage:    check.ErrorMessage,
		CheckedAt:       check.CheckedAt,
	})
	if err != nil {
		logger.Error("Failed to marshal check event", zap.Error(err))
		return
	}
	o.onCheckCreated(check.PageID, payload)
}
