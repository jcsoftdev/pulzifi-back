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

// frequencyIntervals maps compact frequency keys to their durations.
// These match the canonical values from migration 000006_normalize_frequencies.
var frequencyIntervals = map[string]time.Duration{
	"5m":   5 * time.Minute,
	"10m":  10 * time.Minute,
	"15m":  15 * time.Minute,
	"30m":  30 * time.Minute,
	"1h":   1 * time.Hour,
	"2h":   2 * time.Hour,
	"4h":   4 * time.Hour,
	"6h":   6 * time.Hour,
	"12h":  12 * time.Hour,
	"24h":  24 * time.Hour,
	"168h": 168 * time.Hour,
}

// frequencyAliases maps old verbose format strings to canonical short keys.
var frequencyAliases = map[string]string{
	"Every 5 minutes":  "5m",
	"Every 10 minutes": "10m",
	"Every 15 minutes": "15m",
	"Every 30 minutes": "30m",
	"Every hour":       "1h",
	"Every 1 hour":     "1h",
	"1 hr":             "1h",
	"Every 2 hours":    "2h",
	"2 hr":             "2h",
	"Every 4 hours":    "4h",
	"4 hr":             "4h",
	"Every 6 hours":    "6h",
	"6 hr":             "6h",
	"Every 12 hours":   "12h",
	"12 hr":            "12h",
	"Every day":        "24h",
	"1d":               "24h",
	"Every week":       "168h",
	"7d":               "168h",
}

// ResolveFrequency returns the duration for a given frequency string,
// handling both canonical short keys and verbose aliases.
func ResolveFrequency(freq string) (time.Duration, bool) {
	if d, ok := frequencyIntervals[freq]; ok {
		return d, true
	}
	if canonical, ok := frequencyAliases[freq]; ok {
		return frequencyIntervals[canonical], true
	}
	return 0, false
}

// buildFrequencySQLCases generates the CASE WHEN SQL fragment dynamically
// from the frequencyIntervals and frequencyAliases maps.
func buildFrequencySQLCases() string {
	// Collect all keys that resolve to each duration
	type entry struct {
		keys     []string
		interval string
	}
	// Map from canonical key to postgres interval string
	intervalSQL := map[string]string{
		"5m":   "5 minutes",
		"10m":  "10 minutes",
		"15m":  "15 minutes",
		"30m":  "30 minutes",
		"1h":   "1 hour",
		"2h":   "2 hours",
		"4h":   "4 hours",
		"6h":   "6 hours",
		"12h":  "12 hours",
		"24h":  "1 day",
		"168h": "7 days",
	}

	// Build all keys per canonical key
	allKeys := make(map[string][]string)
	for canonical := range frequencyIntervals {
		allKeys[canonical] = append(allKeys[canonical], canonical)
	}
	for alias, canonical := range frequencyAliases {
		allKeys[canonical] = append(allKeys[canonical], alias)
	}

	var cases string
	for canonical, keys := range allKeys {
		pgInterval := intervalSQL[canonical]
		for _, k := range keys {
			cases += fmt.Sprintf(
				"\n\t\t\t\t\tWHEN mc.check_frequency = '%s' THEN COALESCE(p.last_checked_at, '2000-01-01'::timestamp) + INTERVAL '%s'",
				k, pgInterval)
		}
	}
	return cases
}

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
		cases := buildFrequencySQLCases()
		q := fmt.Sprintf(`
			SELECT MIN(
				CASE %s
					ELSE '2100-01-01'::timestamp
				END
			)
			FROM %s.monitoring_configs mc
			JOIN %s.pages p ON mc.page_id = p.id
			WHERE mc.deleted_at IS NULL AND p.deleted_at IS NULL AND mc.check_frequency != 'Off'
		`, cases, schema, schema)

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
