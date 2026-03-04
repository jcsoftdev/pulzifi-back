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
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
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
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return err
	}
	insightTypesJSON := marshalStringSlice(config.EnabledInsightTypes)
	alertConditionsJSON := marshalStringSlice(config.EnabledAlertConditions)
	selectorOffsetsJSON := marshalSelectorOffsets(config.SelectorOffsets)
	q := `INSERT INTO monitoring_configs
		(id, page_id, check_frequency, schedule_type, timezone, block_ads_cookies,
		 enabled_insight_types, enabled_alert_conditions, custom_alert_condition,
		 selector_type, css_selector, xpath_selector, selector_offsets,
		 created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`
	_, err := r.db.ExecContext(ctx, q,
		config.ID, config.PageID, config.CheckFrequency, config.ScheduleType,
		config.Timezone, config.BlockAdsCookies,
		string(insightTypesJSON), string(alertConditionsJSON), config.CustomAlertCondition,
		config.SelectorType, config.CSSSelector, config.XPathSelector, string(selectorOffsetsJSON),
		config.CreatedAt, config.UpdatedAt,
	)
	return err
}

func (r *MonitoringConfigPostgresRepository) GetByPageID(ctx context.Context, pageID uuid.UUID) (*entities.MonitoringConfig, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}
	var c entities.MonitoringConfig
	var insightTypesRaw, alertConditionsRaw, selectorOffsetsRaw []byte
	q := `SELECT id, page_id, check_frequency, schedule_type, timezone, block_ads_cookies,
		         enabled_insight_types, enabled_alert_conditions, custom_alert_condition,
		         COALESCE(selector_type, 'full_page'), COALESCE(css_selector, ''), COALESCE(xpath_selector, ''),
		         COALESCE(selector_offsets, '{"top":0,"right":0,"bottom":0,"left":0}')::text,
		         created_at, updated_at
		  FROM monitoring_configs WHERE page_id = $1 AND deleted_at IS NULL`
	err := r.db.QueryRowContext(ctx, q, pageID).Scan(
		&c.ID, &c.PageID, &c.CheckFrequency, &c.ScheduleType, &c.Timezone, &c.BlockAdsCookies,
		&insightTypesRaw, &alertConditionsRaw, &c.CustomAlertCondition,
		&c.SelectorType, &c.CSSSelector, &c.XPathSelector, &selectorOffsetsRaw,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
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
	return &c, nil
}

func (r *MonitoringConfigPostgresRepository) Update(ctx context.Context, config *entities.MonitoringConfig) error {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return err
	}
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
	q := `UPDATE monitoring_configs
		  SET check_frequency = $1, schedule_type = $2, timezone = $3, block_ads_cookies = $4,
		      enabled_insight_types = $5, enabled_alert_conditions = $6, custom_alert_condition = $7,
		      selector_type = $8, css_selector = $9, xpath_selector = $10, selector_offsets = $11,
		      updated_at = $12
		  WHERE id = $13 AND deleted_at IS NULL`
	_, err = r.db.ExecContext(ctx, q,
		config.CheckFrequency, config.ScheduleType, config.Timezone, config.BlockAdsCookies,
		string(insightTypesJSON), string(alertConditionsJSON), config.CustomAlertCondition,
		config.SelectorType, config.CSSSelector, config.XPathSelector, string(selectorOffsetsJSON),
		config.UpdatedAt, config.ID,
	)
	return err
}

func (r *MonitoringConfigPostgresRepository) BulkUpdateFrequency(ctx context.Context, pageIDs []uuid.UUID, frequency string) error {
	if len(pageIDs) == 0 {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return err
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
	_, err := r.db.ExecContext(ctx, q, args...)
	return err
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
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}

	q := fmt.Sprintf(`
		SELECT p.id, p.url
		FROM pages p
		JOIN monitoring_configs mc ON p.id = mc.page_id
		WHERE p.deleted_at IS NULL AND mc.deleted_at IS NULL
		AND mc.check_frequency != 'Off'
		AND (
			%s
		)
		LIMIT 50
	`, buildDueConditions())

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return "", err
	}
	var url string
	q := `SELECT url FROM pages WHERE id = $1 AND deleted_at IS NULL`
	err := r.db.QueryRowContext(ctx, q, pageID).Scan(&url)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (r *MonitoringConfigPostgresRepository) UpdateLastCheckedAt(ctx context.Context, pageID uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return err
	}
	q := `UPDATE pages SET last_checked_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, q, pageID)
	return err
}

func (r *MonitoringConfigPostgresRepository) MarkPageDueNow(ctx context.Context, pageID uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return err
	}
	q := `UPDATE pages SET last_checked_at = NULL WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, q, pageID)
	return err
}
