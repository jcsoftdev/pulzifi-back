package orchestrator

import (
	"context"
	"errors"

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
}

type PageRepository interface {
	UpdateLastChecked(ctx context.Context, pageID uuid.UUID) error
}

type RepositoryFactory interface {
	GetCheckRepository(tenant string) CheckRepository
	GetPageRepository(tenant string) PageRepository
	GetUsageRepository(tenant string) UsageRepository
}

type Orchestrator struct {
	repoFactory RepositoryFactory
	dispatcher  JobDispatcher
}

func NewOrchestrator(repoFactory RepositoryFactory, dispatcher JobDispatcher) *Orchestrator {
	return &Orchestrator{
		repoFactory: repoFactory,
		dispatcher:  dispatcher,
	}
}

type CheckJob struct {
	PageID     uuid.UUID
	URL        string
	SchemaName string
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
		return errors.New("quota exceeded")
	}

	// 2. Create Check (Pending)
	check := entities.NewCheck(job.PageID, "pending", false)
	if err := checkRepo.Create(ctx, check); err != nil {
		logger.Error("Failed to create check", zap.Error(err))
		return err
	}

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
	return o.dispatcher.Dispatch(ctx, check.ID, job.URL, job.SchemaName)
}
