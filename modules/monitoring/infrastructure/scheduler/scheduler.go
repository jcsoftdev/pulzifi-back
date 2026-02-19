package scheduler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/orchestrator"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type Scheduler struct {
	db           *sql.DB
	orchestrator *orchestrator.Orchestrator
	wakeUp       chan struct{}
}

func NewScheduler(db *sql.DB, orchestrator *orchestrator.Orchestrator) *Scheduler {
	return &Scheduler{
		db:           db,
		orchestrator: orchestrator,
		wakeUp:       make(chan struct{}, 1),
	}
}

// WakeUp signals the scheduler to check for tasks immediately
// It is non-blocking and coalesces multiple signals into one
func (s *Scheduler) WakeUp() {
	select {
	case s.wakeUp <- struct{}{}:
	default:
		// Channel is full, meaning a signal is already pending.
		// We don't need to queue another one.
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	go func() {
		for {
			nextRun := s.getNextRunTime(ctx)
			now := time.Now()

			var waitDuration time.Duration
			if nextRun.IsZero() {
				// No tasks pending. In split API/worker mode wake-up signals are in-process only,
				// so we keep a short polling interval to discover newly enabled configs.
				waitDuration = 15 * time.Second
			} else {
				waitDuration = nextRun.Sub(now)
				if waitDuration < 0 {
					waitDuration = 0
				}
			}

			logger.Debug("Scheduler sleeping", zap.Duration("duration", waitDuration))

			timer := time.NewTimer(waitDuration)

			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-s.wakeUp:
				// Woken up by signal (e.g. new task or config change)
				timer.Stop()
				logger.Info("Scheduler woken up by signal")
				// Loop again to re-calculate next run time immediately
				continue
			case <-timer.C:
				// Timer expired, run check
				s.runCheck(ctx)
			}
		}
	}()
	logger.Info("Monitoring Scheduler started (Wake-up Channel Mode)")
}

func (s *Scheduler) getNextRunTime(ctx context.Context) time.Time {
	// This is tricky in a multi-tenant DB without a unified view.
	// We would need to query all schemas or have a central table.
	// For MVP, polling is safer. But to strictly follow instructions:
	// We can iterate all tenants and find the minimum next_run_at.
	// Optimization: This might be heavy if many tenants.
	// Let's stick to the current implementation but rename/adjust if needed.
	// Actually, the user PROMPT says "Wake-up exact".
	// Let's try to implement a global query if possible, or iterate efficiently.

	// Since we don't have a global index of next_run_at across schemas,
	// we will iterate all schemas to find the earliest time.

	rows, err := s.db.QueryContext(ctx, "SELECT schema_name FROM organizations WHERE deleted_at IS NULL")
	if err != nil {
		return time.Now().Add(1 * time.Minute)
	}
	defer rows.Close()

	minNextRun := time.Time{}

	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			continue
		}

		// Calculate min next_run based on check_frequency and last_checked_at
		// Using a default past date (2000-01-01) for null last_checked_at to ensure it runs immediately
		q := fmt.Sprintf(`
			SELECT MIN(
				CASE 
					WHEN mc.check_frequency = '30m' THEN COALESCE(p.last_checked_at, '2000-01-01'::timestamp) + INTERVAL '30 minutes'
					WHEN mc.check_frequency = '1h' THEN COALESCE(p.last_checked_at, '2000-01-01'::timestamp) + INTERVAL '1 hour'
					WHEN mc.check_frequency = '1 hr' THEN COALESCE(p.last_checked_at, '2000-01-01'::timestamp) + INTERVAL '1 hour'
					WHEN mc.check_frequency = '2h' THEN COALESCE(p.last_checked_at, '2000-01-01'::timestamp) + INTERVAL '2 hours'
					WHEN mc.check_frequency = '2 hr' THEN COALESCE(p.last_checked_at, '2000-01-01'::timestamp) + INTERVAL '2 hours'
					WHEN mc.check_frequency = '8h' THEN COALESCE(p.last_checked_at, '2000-01-01'::timestamp) + INTERVAL '8 hours'
					WHEN mc.check_frequency = '8 hr' THEN COALESCE(p.last_checked_at, '2000-01-01'::timestamp) + INTERVAL '8 hours'
					WHEN mc.check_frequency = '24h' THEN COALESCE(p.last_checked_at, '2000-01-01'::timestamp) + INTERVAL '1 day'
					WHEN mc.check_frequency = '1d' THEN COALESCE(p.last_checked_at, '2000-01-01'::timestamp) + INTERVAL '1 day'
					WHEN mc.check_frequency = '48h' THEN COALESCE(p.last_checked_at, '2000-01-01'::timestamp) + INTERVAL '48 hours'
					WHEN mc.check_frequency = '2d' THEN COALESCE(p.last_checked_at, '2000-01-01'::timestamp) + INTERVAL '48 hours'
					WHEN mc.check_frequency = 'Every 30 minutes' THEN COALESCE(p.last_checked_at, '2000-01-01'::timestamp) + INTERVAL '30 minutes'
					WHEN mc.check_frequency = 'Every 1 hour' THEN COALESCE(p.last_checked_at, '2000-01-01'::timestamp) + INTERVAL '1 hour'
					WHEN mc.check_frequency = 'Every 2 hours' THEN COALESCE(p.last_checked_at, '2000-01-01'::timestamp) + INTERVAL '2 hours'
					WHEN mc.check_frequency = 'Every 8 hours' THEN COALESCE(p.last_checked_at, '2000-01-01'::timestamp) + INTERVAL '8 hours'
					WHEN mc.check_frequency = 'Every day' THEN COALESCE(p.last_checked_at, '2000-01-01'::timestamp) + INTERVAL '1 day'
					WHEN mc.check_frequency = 'Every 48 hours' THEN COALESCE(p.last_checked_at, '2000-01-01'::timestamp) + INTERVAL '48 hours'
					ELSE '2100-01-01'::timestamp
				END
			)
			FROM %s.monitoring_configs mc
			JOIN %s.pages p ON mc.page_id = p.id
			WHERE mc.deleted_at IS NULL AND p.deleted_at IS NULL AND mc.check_frequency != 'Off'
		`, schema, schema)

		var nextRun sql.NullTime
		if err := s.db.QueryRowContext(ctx, q).Scan(&nextRun); err == nil && nextRun.Valid {
			if minNextRun.IsZero() || nextRun.Time.Before(minNextRun) {
				minNextRun = nextRun.Time
			}
		}
	}

	return minNextRun
}

func (s *Scheduler) runCheck(ctx context.Context) {
	// Query all schema names
	rows, err := s.db.QueryContext(ctx, "SELECT schema_name FROM organizations WHERE deleted_at IS NULL")
	if err != nil {
		logger.Error("Scheduler failed to fetch organizations", zap.Error(err))
		return
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			continue
		}
		schemas = append(schemas, schema)
	}

	for _, schema := range schemas {
		s.processTenant(ctx, schema)
	}
}

// TriggerPageCheck schedules one immediate check for a specific page within a tenant schema.
// This is used when a user updates monitoring frequency and expects an immediate run.
func (s *Scheduler) TriggerPageCheck(ctx context.Context, schema string, pageID uuid.UUID) error {
	q := fmt.Sprintf(`
		SELECT p.url
		FROM %s.pages p
		JOIN %s.monitoring_configs mc ON mc.page_id = p.id
		WHERE p.id = $1
		  AND p.deleted_at IS NULL
		  AND mc.deleted_at IS NULL
		  AND mc.check_frequency != 'Off'
		LIMIT 1
	`, schema, schema)

	var url string
	if err := s.db.QueryRowContext(ctx, q, pageID).Scan(&url); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	// Mark the page as checked now so the scheduler loop doesn't also pick it up
	// before the async goroutine below has a chance to call UpdateLastChecked.
	repo := persistence.NewMonitoringConfigPostgresRepository(s.db, schema)
	if err := repo.UpdateLastCheckedAt(ctx, pageID); err != nil {
		logger.Error("TriggerPageCheck: failed to pre-update last_checked_at", zap.String("page_id", pageID.String()), zap.Error(err))
	}

	job := orchestrator.CheckJob{
		PageID:     pageID,
		URL:        url,
		SchemaName: schema,
	}

	go func(j orchestrator.CheckJob) {
		if err := s.orchestrator.ScheduleCheck(context.Background(), j); err != nil {
			if errors.Is(err, orchestrator.ErrQuotaExceeded) {
				logger.Warn("Immediate check skipped due to quota", zap.String("page_id", j.PageID.String()), zap.String("schema", j.SchemaName))
				return
			}
			logger.Error("Failed to schedule immediate check", zap.String("page_id", j.PageID.String()), zap.Error(err))
		}
	}(job)

	return nil
}

func (s *Scheduler) processTenant(ctx context.Context, schema string) {
	// Scheduler logic: Calculate next_run_at (Find due tasks)
	// We use the existing repository logic to find due tasks
	repo := persistence.NewMonitoringConfigPostgresRepository(s.db, schema)
	tasks, err := repo.GetDueSnapshotTasks(ctx)
	if err != nil {
		logger.Error("Failed to get due tasks", zap.String("tenant", schema), zap.Error(err))
		return
	}

	for _, task := range tasks {
		// Pre-update last_checked_at synchronously so that the next scheduler
		// iteration doesn't see this page as due again before the goroutine commits.
		if err := repo.UpdateLastCheckedAt(ctx, task.PageID); err != nil {
			logger.Error("Failed to pre-update last_checked_at", zap.String("page_id", task.PageID.String()), zap.Error(err))
			continue
		}

		// Create Job (In-Memory)
		job := orchestrator.CheckJob{
			PageID:     task.PageID,
			URL:        task.URL,
			SchemaName: schema,
		}

		// Send to Orchestrator
		// Using goroutine to avoid blocking the scheduler loop
		go func(j orchestrator.CheckJob) {
			// Create a detached context for the job execution
			if err := s.orchestrator.ScheduleCheck(context.Background(), j); err != nil {
				if errors.Is(err, orchestrator.ErrQuotaExceeded) {
					logger.Warn("Check not scheduled due to quota", zap.String("page_id", j.PageID.String()), zap.String("schema", j.SchemaName))
					return
				}
				logger.Error("Failed to schedule check", zap.String("page_id", j.PageID.String()), zap.Error(err))
			}
		}(job)
	}
}
