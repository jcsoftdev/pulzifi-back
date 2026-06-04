package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
	"github.com/lib/pq"
)

type MonitoringConfigPostgresRepository struct {
	db     *sql.DB
	tenant string
}

func NewMonitoringConfigPostgresRepository(db *sql.DB, tenant string) *MonitoringConfigPostgresRepository {
	return &MonitoringConfigPostgresRepository{db: db, tenant: tenant}
}

func marshalStringSlice(s []string) []byte {
	if s == nil {
		s = []string{}
	}
	b, _ := json.Marshal(s)
	return b
}

func marshalSelectorOffsets(o *entities.SelectorOffsets) []byte {
	if o == nil {
		o = &entities.SelectorOffsets{}
	}
	b, _ := json.Marshal(o)
	return b
}

func (r *MonitoringConfigPostgresRepository) Create(ctx context.Context, config *entities.MonitoringConfig) error {
	insightTypesJSON := marshalStringSlice(config.EnabledInsightTypes)
	alertConditionsJSON := marshalStringSlice(config.EnabledAlertConditions)
	selectorOffsetsJSON := marshalSelectorOffsets(config.SelectorOffsets)
	ignoreSelectorsJSON := marshalStringSlice(config.IgnoreSelectors)
	q := `INSERT INTO monitoring_configs
		(id, page_id, check_frequency, schedule_type, timezone, block_ads_cookies,
		 enabled_insight_types, enabled_alert_conditions, custom_alert_condition,
		 selector_type, css_selector, xpath_selector, selector_offsets, ignore_selectors,
		 created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`
	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q,
			config.ID, config.PageID, config.CheckFrequency, config.ScheduleType,
			config.Timezone, config.BlockAdsCookies,
			string(insightTypesJSON), string(alertConditionsJSON), config.CustomAlertCondition,
			config.SelectorType, config.CSSSelector, config.XPathSelector, string(selectorOffsetsJSON),
			string(ignoreSelectorsJSON),
			config.CreatedAt, config.UpdatedAt,
		)
		return err
	})
}

func (r *MonitoringConfigPostgresRepository) GetByPageID(ctx context.Context, pageID uuid.UUID) (*entities.MonitoringConfig, error) {
	var c entities.MonitoringConfig
	var insightTypesRaw, alertConditionsRaw, selectorOffsetsRaw, ignoreSelectorsRaw []byte
	q := `SELECT id, page_id, check_frequency, schedule_type, timezone, block_ads_cookies,
		         enabled_insight_types, enabled_alert_conditions, custom_alert_condition,
		         COALESCE(selector_type, 'full_page'), COALESCE(css_selector, ''), COALESCE(xpath_selector, ''),
		         COALESCE(selector_offsets, '{"top":0,"right":0,"bottom":0,"left":0}')::text,
		         COALESCE(ignore_selectors, '[]')::text,
		         created_at, updated_at
		  FROM monitoring_configs WHERE page_id = $1 AND deleted_at IS NULL`

	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, q, pageID).Scan(
			&c.ID, &c.PageID, &c.CheckFrequency, &c.ScheduleType, &c.Timezone, &c.BlockAdsCookies,
			&insightTypesRaw, &alertConditionsRaw, &c.CustomAlertCondition,
			&c.SelectorType, &c.CSSSelector, &c.XPathSelector, &selectorOffsetsRaw,
			&ignoreSelectorsRaw,
			&c.CreatedAt, &c.UpdatedAt,
		)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(insightTypesRaw) > 0 {
		_ = json.Unmarshal(insightTypesRaw, &c.EnabledInsightTypes)
	}
	if len(alertConditionsRaw) > 0 {
		_ = json.Unmarshal(alertConditionsRaw, &c.EnabledAlertConditions)
	}
	if len(selectorOffsetsRaw) > 0 {
		var offsets entities.SelectorOffsets
		if json.Unmarshal(selectorOffsetsRaw, &offsets) == nil {
			c.SelectorOffsets = &offsets
		}
	}
	if len(ignoreSelectorsRaw) > 0 {
		_ = json.Unmarshal(ignoreSelectorsRaw, &c.IgnoreSelectors)
	}
	return &c, nil
}

func (r *MonitoringConfigPostgresRepository) Update(ctx context.Context, config *entities.MonitoringConfig) error {
	config.UpdatedAt = time.Now()
	insightTypesJSON, err := json.Marshal(config.EnabledInsightTypes)
	if err != nil {
		return err
	}
	alertConditionsJSON, err := json.Marshal(config.EnabledAlertConditions)
	if err != nil {
		return err
	}
	selectorOffsetsJSON := marshalSelectorOffsets(config.SelectorOffsets)
	ignoreSelectorsJSON := marshalStringSlice(config.IgnoreSelectors)
	q := `UPDATE monitoring_configs
		  SET check_frequency = $1, schedule_type = $2, timezone = $3, block_ads_cookies = $4,
		      enabled_insight_types = $5, enabled_alert_conditions = $6, custom_alert_condition = $7,
		      selector_type = $8, css_selector = $9, xpath_selector = $10, selector_offsets = $11,
		      ignore_selectors = $12, updated_at = $13
		  WHERE id = $14 AND deleted_at IS NULL`
	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q,
			config.CheckFrequency, config.ScheduleType, config.Timezone, config.BlockAdsCookies,
			string(insightTypesJSON), string(alertConditionsJSON), config.CustomAlertCondition,
			config.SelectorType, config.CSSSelector, config.XPathSelector, string(selectorOffsetsJSON),
			string(ignoreSelectorsJSON),
			config.UpdatedAt, config.ID,
		)
		return err
	})
}

func (r *MonitoringConfigPostgresRepository) BulkUpdateFrequency(ctx context.Context, pageIDs []uuid.UUID, frequency string) error {
	if len(pageIDs) == 0 {
		return nil
	}
	now := time.Now()
	placeholders := make([]string, len(pageIDs))
	args := make([]interface{}, len(pageIDs)+2)
	args[0] = frequency
	args[1] = now
	for i, id := range pageIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+3)
		args[i+2] = id
	}
	q := `UPDATE monitoring_configs SET check_frequency = $1, updated_at = $2 WHERE page_id IN (` + strings.Join(placeholders, ", ") + `) AND deleted_at IS NULL`
	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q, args...)
		return err
	})
}

// buildNextRunAtCases generates the CASE WHEN fragment that computes
// NOW() + interval for a given check_frequency.  Used by queue-mode
// queries to advance next_run_at atomically at claim time.
func buildNextRunAtCases() string {
	allKeys := entities.AllFrequencyKeys()
	var cases string
	for canonical, keys := range allKeys {
		pgInterval := entities.FrequencyToPostgresInterval[canonical]
		for _, k := range keys {
			cases += fmt.Sprintf(
				"\n\t\t\t\tWHEN mc.check_frequency = '%s' THEN NOW() + INTERVAL '%s'",
				k, pgInterval)
		}
	}
	return cases
}

func buildDueConditions() string {
	allKeys := entities.AllFrequencyKeys()
	var conditions []string
	for canonical, keys := range allKeys {
		pgInterval := entities.FrequencyToPostgresInterval[canonical]
		for _, k := range keys {
			conditions = append(conditions,
				fmt.Sprintf("(mc.check_frequency = '%s' AND (p.last_checked_at IS NULL OR p.last_checked_at < NOW() - INTERVAL '%s'))", k, pgInterval))
		}
	}
	return strings.Join(conditions, " OR\n\t\t\t")
}

func (r *MonitoringConfigPostgresRepository) GetDueSnapshotTasks(ctx context.Context) ([]entities.SnapshotTask, error) {
	// Atomically claim due tasks: SELECT with FOR UPDATE SKIP LOCKED then UPDATE in one
	// round-trip. This prevents concurrent scheduler instances from picking up the same
	// page and creating duplicate check records.
	qTenant := pq.QuoteIdentifier(r.tenant)
	// Advance next_run_at alongside last_checked_at even in poll mode. Poll mode
	// itself never reads next_run_at (it claims on last_checked_at), but keeping
	// the column current means switching to queue mode — or any external reader —
	// never sees a stale next_run_at that points into the past.
	q := fmt.Sprintf(`
		WITH candidates AS (
			SELECT p.id, mc.check_frequency
			FROM %[1]s.pages p
			JOIN %[1]s.monitoring_configs mc ON p.id = mc.page_id
			WHERE p.deleted_at IS NULL AND mc.deleted_at IS NULL
			AND mc.check_frequency != 'Off'
			AND (
				%[2]s
			)
			LIMIT 50
			FOR UPDATE OF p SKIP LOCKED
		)
		UPDATE %[1]s.pages p
		SET last_checked_at = NOW(),
		    next_run_at = CASE %[3]s
		                    ELSE NOW() + INTERVAL '1 hour'
		                  END
		FROM candidates mc
		WHERE p.id = mc.id
		RETURNING p.id, p.url
	`, qTenant, buildDueConditions(), buildNextRunAtCases())

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var tasks []entities.SnapshotTask
	for rows.Next() {
		var t entities.SnapshotTask
		if err := rows.Scan(&t.PageID, &t.URL); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (r *MonitoringConfigPostgresRepository) GetPageURL(ctx context.Context, pageID uuid.UUID) (string, error) {
	var url string
	q := `SELECT url FROM pages WHERE id = $1 AND deleted_at IS NULL`
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, q, pageID).Scan(&url)
	})
	if err != nil {
		return "", err
	}
	return url, nil
}

func (r *MonitoringConfigPostgresRepository) UpdateLastCheckedAt(ctx context.Context, pageID uuid.UUID) error {
	q := fmt.Sprintf(`UPDATE %s.pages SET last_checked_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, pq.QuoteIdentifier(r.tenant))
	_, err := r.db.ExecContext(ctx, q, pageID)
	return err
}

func (r *MonitoringConfigPostgresRepository) MarkPageDueNow(ctx context.Context, pageID uuid.UUID) error {
	q := fmt.Sprintf(`UPDATE %s.pages SET last_checked_at = NULL WHERE id = $1 AND deleted_at IS NULL`, pq.QuoteIdentifier(r.tenant))
	_, err := r.db.ExecContext(ctx, q, pageID)
	return err
}

func (r *MonitoringConfigPostgresRepository) GetLastCheckedAt(ctx context.Context, pageID uuid.UUID) (*time.Time, error) {
	q := fmt.Sprintf(`SELECT last_checked_at FROM %s.pages WHERE id = $1 AND deleted_at IS NULL`, pq.QuoteIdentifier(r.tenant))
	var lastCheckedAt *time.Time
	err := r.db.QueryRowContext(ctx, q, pageID).Scan(&lastCheckedAt)
	if err != nil {
		return nil, err
	}
	return lastCheckedAt, nil
}

// GetDueSnapshotTasksQueue is the queue-mode replacement for GetDueSnapshotTasks.
// It claims pages WHERE next_run_at <= NOW() (FOR UPDATE SKIP LOCKED) and in the
// same atomic UPDATE advances next_run_at = NOW() + interval so concurrent
// scheduler instances never double-claim the same page.
//
// Only active when SCHEDULER_MODE=queue; poll mode uses GetDueSnapshotTasks.
func (r *MonitoringConfigPostgresRepository) GetDueSnapshotTasksQueue(ctx context.Context) ([]entities.SnapshotTask, error) {
	qTenant := pq.QuoteIdentifier(r.tenant)
	nextRunCases := buildNextRunAtCases()

	q := fmt.Sprintf(`
		WITH candidates AS (
			SELECT p.id, mc.check_frequency
			FROM %[1]s.pages p
			JOIN %[1]s.monitoring_configs mc ON p.id = mc.page_id
			WHERE p.deleted_at IS NULL AND mc.deleted_at IS NULL
			AND mc.check_frequency != 'Off'
			AND p.next_run_at IS NOT NULL
			AND p.next_run_at <= NOW()
			LIMIT 50
			FOR UPDATE OF p SKIP LOCKED
		)
		UPDATE %[1]s.pages p
		SET last_checked_at = NOW(),
		    next_run_at = CASE %s
		                    ELSE NOW() + INTERVAL '1 hour'
		                  END
		FROM candidates mc
		WHERE p.id = mc.id
		RETURNING p.id, p.url
	`, qTenant, nextRunCases)

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var tasks []entities.SnapshotTask
	for rows.Next() {
		var t entities.SnapshotTask
		if err := rows.Scan(&t.PageID, &t.URL); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// UpdateNextRunAt recomputes next_run_at for a page based on its current
// check_frequency.  Called after a check completes in queue mode so the page
// stays scheduled even when the check completion path runs outside the
// atomic claim window (e.g. orchestrator.PageRepository.UpdateLastChecked).
//
// In poll mode this is a no-op (the column is unused) — the UPDATE is
// harmless because it only touches rows that already have a non-NULL
// check_frequency config.
func (r *MonitoringConfigPostgresRepository) UpdateNextRunAt(ctx context.Context, pageID uuid.UUID) error {
	qTenant := pq.QuoteIdentifier(r.tenant)
	nextRunCases := buildNextRunAtCases()

	q := fmt.Sprintf(`
		UPDATE %[1]s.pages p
		SET next_run_at = CASE %s
		                    ELSE NOW() + INTERVAL '1 hour'
		                  END
		FROM %[1]s.monitoring_configs mc
		WHERE p.id = $1
		  AND p.deleted_at IS NULL
		  AND mc.page_id = p.id
		  AND mc.deleted_at IS NULL
		  AND mc.check_frequency != 'Off'
	`, qTenant, nextRunCases)
	_, err := r.db.ExecContext(ctx, q, pageID)
	return err
}
