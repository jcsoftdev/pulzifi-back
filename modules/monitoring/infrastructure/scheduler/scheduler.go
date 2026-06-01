package scheduler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/orchestrator"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

// listExistingTenantSchemas returns active tenant schemas from public.organizations
// that actually exist in information_schema. Skips orphan org rows whose schema
// was never provisioned (or was dropped) — querying them produces noisy errors.
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
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// ResolveFrequency delegates to the shared domain definition.
func ResolveFrequency(freq string) (time.Duration, bool) {
	return entities.ResolveFrequency(freq)
}

// buildFrequencySQLCases generates the CASE WHEN SQL fragment dynamically
// from the shared frequency definitions in the domain layer.
func buildFrequencySQLCases() string {
	allKeys := entities.AllFrequencyKeys()

	var cases string
	for canonical, keys := range allKeys {
		pgInterval := entities.FrequencyToPostgresInterval[canonical]
		for _, k := range keys {
			cases += fmt.Sprintf(
				"\n\t\t\t\t\tWHEN mc.check_frequency = '%s' THEN COALESCE(p.last_checked_at, '2000-01-01'::timestamp) + INTERVAL '%s'",
				k, pgInterval)
		}
	}
	return cases
}

// buildTriggerNextRunAtCases generates the CASE fragment that computes
// NOW() + the page's configured frequency interval. Used by TriggerPageCheck
// so a manual "Run Now" advances next_run_at by the REAL frequency (a daily
// page must not re-fire in 5 minutes once queue mode is active).
func buildTriggerNextRunAtCases() string {
	allKeys := entities.AllFrequencyKeys()
	var cases string
	for canonical, keys := range allKeys {
		pgInterval := entities.FrequencyToPostgresInterval[canonical]
		for _, k := range keys {
			cases += fmt.Sprintf(
				"\n\t\t\t\t\t\tWHEN mc.check_frequency = '%s' THEN NOW() + INTERVAL '%s'",
				k, pgInterval)
		}
	}
	return cases
}

type Scheduler struct {
	db            *sql.DB
	orchestrator  *orchestrator.Orchestrator
	wakeUp        chan struct{}
	schedulerMode string // "poll" (default) | "queue"
}

func NewScheduler(db *sql.DB, orchestrator *orchestrator.Orchestrator) *Scheduler {
	return &Scheduler{
		db:            db,
		orchestrator:  orchestrator,
		wakeUp:        make(chan struct{}, 1),
		schedulerMode: "poll",
	}
}

// NewSchedulerWithMode creates a Scheduler with an explicit mode.
// mode must be "poll" or "queue"; any other value defaults to "poll".
func NewSchedulerWithMode(db *sql.DB, o *orchestrator.Orchestrator, mode string) *Scheduler {
	if mode != "queue" {
		mode = "poll"
	}
	return &Scheduler{
		db:            db,
		orchestrator:  o,
		wakeUp:        make(chan struct{}, 1),
		schedulerMode: mode,
	}
}

// minNextRunAtQueue iterates all tenant schemas and returns the global
// MIN(next_run_at) across all non-deleted, scheduled pages.  Used by queue
// mode instead of the O(n×freq-CASE) getNextRunTime scan.
func (s *Scheduler) minNextRunAtQueue(ctx context.Context) time.Time {
	schemas, err := listExistingTenantSchemas(ctx, s.db)
	if err != nil {
		return time.Now().Add(1 * time.Minute)
	}

	var minT time.Time
	for _, schema := range schemas {
		qSchema := pq.QuoteIdentifier(schema)
		q := fmt.Sprintf(`
			SELECT MIN(next_run_at)
			FROM %s.pages
			WHERE deleted_at IS NULL AND next_run_at IS NOT NULL
		`, qSchema)

		var t sql.NullTime
		if err := s.db.QueryRowContext(ctx, q).Scan(&t); err == nil && t.Valid {
			if minT.IsZero() || t.Time.Before(minT) {
				minT = t.Time
			}
		}
	}
	return minT
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
			var nextRun time.Time
			if s.schedulerMode == "queue" {
				nextRun = s.minNextRunAtQueue(ctx)
			} else {
				nextRun = s.getNextRunTime(ctx)
			}
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

	schemas, err := listExistingTenantSchemas(ctx, s.db)
	if err != nil {
		return time.Now().Add(1 * time.Minute)
	}

	minNextRun := time.Time{}
	cases := buildFrequencySQLCases()

	for _, schema := range schemas {
		qSchema := pq.QuoteIdentifier(schema)
		q := fmt.Sprintf(`
			SELECT MIN(
				CASE %s
					ELSE '2100-01-01'::timestamp
				END
			)
			FROM %s.monitoring_configs mc
			JOIN %s.pages p ON mc.page_id = p.id
			WHERE mc.deleted_at IS NULL AND p.deleted_at IS NULL AND mc.check_frequency != 'Off'
		`, cases, qSchema, qSchema)

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
	schemas, err := listExistingTenantSchemas(ctx, s.db)
	if err != nil {
		logger.Error("Scheduler failed to fetch organizations", zap.Error(err))
		return
	}

	for _, schema := range schemas {
		s.processTenant(ctx, schema)
	}
}

// TriggerPageCheck schedules one immediate check for a specific page within a tenant schema.
// This is used for manual "Run Now" actions — no monitoring config is required.
func (s *Scheduler) TriggerPageCheck(ctx context.Context, schema string, pageID uuid.UUID) error {
	qSchema := pq.QuoteIdentifier(schema)
	q := fmt.Sprintf(`
		SELECT p.url
		FROM %s.pages p
		WHERE p.id = $1
		  AND p.deleted_at IS NULL
		LIMIT 1
	`, qSchema)

	var url string
	if err := s.db.QueryRowContext(ctx, q, pageID).Scan(&url); err != nil {
		if err == sql.ErrNoRows {
			return orchestrator.ErrPageNotFound
		}
		return err
	}

	// Synchronous quota pre-check so the caller (HTTP handler) can detect quota exceeded.
	hasQuota, err := s.orchestrator.HasQuota(ctx, schema)
	if err != nil {
		return fmt.Errorf("quota check: %w", err)
	}
	if !hasQuota {
		return orchestrator.ErrQuotaExceeded
	}

	// Atomically claim the page with a debounce guard: only succeed if the page
	// was not already claimed within the last 5 seconds.  This makes concurrent
	// "Run Now" requests idempotent — the second caller gets 0 rows affected and
	// returns without scheduling a duplicate check.
	//
	// next_run_at is advanced at claim time so queue mode stays consistent even
	// for manual triggers.  It is set from the page's REAL configured frequency
	// (via a correlated subquery) — NOT a fixed interval — so queue mode does not
	// re-fire a daily page minutes after a manual "Run Now".  In poll mode the
	// column is present but unused.
	claimQ := fmt.Sprintf(`
		UPDATE %[1]s.pages p
		SET last_checked_at = NOW(),
		    next_run_at = (
		        SELECT CASE %[2]s
		            ELSE NULL
		        END
		        FROM %[1]s.monitoring_configs mc
		        WHERE mc.page_id = p.id AND mc.deleted_at IS NULL AND mc.check_frequency <> 'Off'
		        LIMIT 1
		    )
		WHERE p.id = $1
		  AND p.deleted_at IS NULL
		  AND (p.last_checked_at IS NULL OR p.last_checked_at < NOW() - INTERVAL '5 seconds')
		RETURNING p.id
	`, qSchema, buildTriggerNextRunAtCases())
	var claimedID uuid.UUID
	claimErr := s.db.QueryRowContext(ctx, claimQ, pageID).Scan(&claimedID)
	if claimErr != nil {
		if errors.Is(claimErr, sql.ErrNoRows) {
			// Another instance (or a concurrent Run Now) already claimed this page
			// within the 5-second debounce window — treat as idempotent success.
			logger.Info("TriggerPageCheck: page already claimed within debounce window, skipping",
				zap.String("page_id", pageID.String()))
			return nil
		}
		logger.Error("TriggerPageCheck: failed to claim page", zap.String("page_id", pageID.String()), zap.Error(claimErr))
		// Non-fatal: proceed to schedule even if the claim UPDATE failed.
	}

	job := orchestrator.CheckJob{
		PageID:     pageID,
		URL:        url,
		SchemaName: schema,
	}

	go func() {
		if err := s.orchestrator.ScheduleCheck(context.Background(), job); err != nil {
			if errors.Is(err, orchestrator.ErrQuotaExceeded) {
				logger.Warn("Immediate check skipped due to quota", zap.String("page_id", job.PageID.String()), zap.String("schema", job.SchemaName))
				return
			}
			logger.Error("Failed to schedule immediate check", zap.String("page_id", job.PageID.String()), zap.Error(err))
		}
	}()

	return nil
}

func (s *Scheduler) processTenant(ctx context.Context, schema string) {
	repo := persistence.NewMonitoringConfigPostgresRepository(s.db, schema)

	var tasks []entities.SnapshotTask
	var err error
	if s.schedulerMode == "queue" {
		tasks, err = repo.GetDueSnapshotTasksQueue(ctx)
	} else {
		tasks, err = repo.GetDueSnapshotTasks(ctx)
	}
	if err != nil {
		logger.Error("Failed to get due tasks", zap.String("tenant", schema), zap.Error(err))
		return
	}

	for _, task := range tasks {
		// last_checked_at is already set to NOW() by the atomic GetDueSnapshotTasks query,
		// so no separate pre-update is needed.
		job := orchestrator.CheckJob{
			PageID:     task.PageID,
			URL:        task.URL,
			SchemaName: schema,
		}
		go func() {
			if err := s.orchestrator.ScheduleCheck(context.Background(), job); err != nil {
				if errors.Is(err, orchestrator.ErrQuotaExceeded) {
					logger.Warn("Check not scheduled due to quota", zap.String("page_id", job.PageID.String()), zap.String("schema", job.SchemaName))
					return
				}
				logger.Error("Failed to schedule check", zap.String("page_id", job.PageID.String()), zap.Error(err))
			}
		}()
	}
}

