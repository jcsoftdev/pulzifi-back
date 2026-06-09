// Package scheduler provides the social profile check scheduler.
// It polls every 30 seconds for due profiles across all tenant schemas and
// dispatches run_check jobs to a bounded worker pool.
// Gated by SOCIAL_ENABLED: Start() is a no-op when the flag is false.
package scheduler

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/google/uuid"
	runcheck "github.com/jcsoftdev/pulzifi-back/modules/social/application/run_check"
	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/social/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/social/infrastructure/persistence/postgres"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

const (
	// pollInterval is the fixed tick rate of the scheduler loop (REQ-SCHED-01).
	pollInterval = 30 * time.Second

	// claimBatchSize is the maximum number of profiles claimed per tenant per tick.
	claimBatchSize = 50

	// workerConcurrency is the bounded worker pool size.
	workerConcurrency = 10
)

// TenantRepoFactory creates per-tenant repository and quota instances.
// The concrete implementation lives in cmd/wiring/social/.
type TenantRepoFactory interface {
	ProfileRepo(tenant string) repositories.ProfileRepository
	SnapshotRepo(tenant string) repositories.SnapshotRepository
	ChangeRepo(tenant string) repositories.ChangeRepository
	CheckQuota(tenant string) services.CheckQuota
}

// CheckHandlerFactory creates a run_check.Handler for a given tenant, wiring
// all dependencies (fetcher, media store, alert creator, plan limits) that are
// resolved by the wiring layer.
type CheckHandlerFactory interface {
	NewHandler(tenant string) *runcheck.Handler
}

// Scheduler polls tenant schemas for due social profiles and dispatches checks.
type Scheduler struct {
	db             *sql.DB
	handlerFactory CheckHandlerFactory
	enabled        bool

	jobs chan schedulerJob
	wg   sync.WaitGroup
}

type schedulerJob struct {
	profileID uuid.UUID
	tenant    string
	handler   *runcheck.Handler
}

// NewScheduler creates a social scheduler.
//
//   - db             — shared database connection (used to list tenants)
//   - handlerFactory — per-tenant run_check.Handler factory (wired in cmd/wiring/social)
//   - enabled        — mirrors SOCIAL_ENABLED; when false Start() is a no-op
func NewScheduler(db *sql.DB, handlerFactory CheckHandlerFactory, enabled bool) *Scheduler {
	return &Scheduler{
		db:             db,
		handlerFactory: handlerFactory,
		enabled:        enabled,
		jobs:           make(chan schedulerJob, workerConcurrency*2),
	}
}

// Start launches the scheduler goroutine and the worker pool.
// It is a no-op when SOCIAL_ENABLED=false.
func (s *Scheduler) Start(ctx context.Context) {
	if !s.enabled {
		logger.Info("Social scheduler disabled (SOCIAL_ENABLED=false)")
		return
	}

	// Launch bounded worker pool.
	for i := 0; i < workerConcurrency; i++ {
		s.wg.Add(1)
		go s.worker(ctx)
	}

	// Launch polling loop.
	go s.loop(ctx)

	logger.Info("Social scheduler started", zap.Int("workers", workerConcurrency))
}

// loop ticks every pollInterval and processes due profiles for all tenants.
func (s *Scheduler) loop(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			close(s.jobs)
			s.wg.Wait()
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

// tick processes all tenants once.
func (s *Scheduler) tick(ctx context.Context) {
	tenants, err := listExistingTenantSchemas(ctx, s.db)
	if err != nil {
		logger.Error("Social scheduler: failed to list tenant schemas", zap.Error(err))
		return
	}

	for _, tenant := range tenants {
		s.processTenant(ctx, tenant)
	}
}

// processTenant claims due profiles for one tenant and enqueues run_check jobs.
func (s *Scheduler) processTenant(ctx context.Context, tenant string) {
	profileRepo := postgres.NewProfilePostgresRepository(s.db, tenant)

	profiles, err := profileRepo.ListDue(ctx, time.Now().UTC(), claimBatchSize)
	if err != nil {
		logger.Error("Social scheduler: list due profiles failed",
			zap.String("tenant", tenant), zap.Error(err))
		return
	}

	if len(profiles) == 0 {
		return
	}

	handler := s.handlerFactory.NewHandler(tenant)

	for _, p := range profiles {
		select {
		case <-ctx.Done():
			return
		case s.jobs <- schedulerJob{profileID: p.ID, tenant: tenant, handler: handler}:
		}
	}
}

// worker consumes jobs from the channel and runs the check.
func (s *Scheduler) worker(ctx context.Context) {
	defer s.wg.Done()

	for j := range s.jobs {
		s.runJob(ctx, j)
	}
}

// runJob executes a single run_check for one profile.
func (s *Scheduler) runJob(ctx context.Context, j schedulerJob) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Social scheduler: panic in run_check",
				zap.String("profile_id", j.profileID.String()),
				zap.String("tenant", j.tenant),
				zap.Any("panic", r),
				zap.Stack("stack"),
			)
		}
	}()

	// run_check may return ErrQuotaExceeded — push next_check_at to next midnight.
	_, runErr := j.handler.Handle(ctx, j.profileID)
	if runErr != nil {
		if isQuotaExceeded(runErr) {
			// REQ-SCHED-06: quota exhausted → reschedule to next UTC midnight.
			s.pushToNextMidnight(ctx, j)
			return
		}
		// ErrFetchFailed is expected (handled by run_check internally) — log at debug.
		logger.Debug("Social scheduler: run_check completed with error",
			zap.String("profile_id", j.profileID.String()),
			zap.String("tenant", j.tenant),
			zap.Error(runErr),
		)
	}
}

// pushToNextMidnight sets next_check_at to the start of the next UTC day.
// This prevents the scheduler from hot-looping on quota-exhausted profiles.
func (s *Scheduler) pushToNextMidnight(ctx context.Context, j schedulerJob) {
	repo := postgres.NewProfilePostgresRepository(s.db, j.tenant)

	profile, err := repo.GetByID(ctx, j.profileID)
	if err != nil {
		return
	}

	now := time.Now().UTC()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	profile.NextCheckAt = &nextMidnight
	profile.UpdatedAt = now

	if err := repo.Update(ctx, profile); err != nil {
		logger.Warn("Social scheduler: failed to push next_check_at to midnight",
			zap.String("profile_id", j.profileID.String()),
			zap.Error(err),
		)
	}
}

// listExistingTenantSchemas returns active tenant schemas that actually exist in
// information_schema. Mirrors the pattern from monitoring scheduler.
func listExistingTenantSchemas(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT o.schema_name
		FROM public.organizations o
		JOIN information_schema.schemata s ON s.schema_name = o.schema_name
		WHERE o.schema_name IS NOT NULL AND o.deleted_at IS NULL`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return nil, err
		}
		out = append(out, schema)
	}
	return out, rows.Err()
}

// isQuotaExceeded reports whether err is domainerrors.ErrQuotaExceeded.
func isQuotaExceeded(err error) bool {
	return err != nil && err.Error() == domainerrors.ErrQuotaExceeded.Error()
}
