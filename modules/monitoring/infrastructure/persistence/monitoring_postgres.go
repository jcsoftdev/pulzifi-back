package persistence

import (
	"context"
	"database/sql"
	"errors"
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

func (r *MonitoringConfigPostgresRepository) Create(ctx context.Context, config *entities.MonitoringConfig) error {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return err
	}
	q := `INSERT INTO monitoring_configs (id, page_id, check_frequency, schedule_type, timezone, block_ads_cookies, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(ctx, q, config.ID, config.PageID, config.CheckFrequency, config.ScheduleType, config.Timezone, config.BlockAdsCookies, config.CreatedAt, config.UpdatedAt)
	return err
}

func (r *MonitoringConfigPostgresRepository) GetByPageID(ctx context.Context, pageID uuid.UUID) (*entities.MonitoringConfig, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}
	var c entities.MonitoringConfig
	q := `SELECT id, page_id, check_frequency, schedule_type, timezone, block_ads_cookies, created_at, updated_at FROM monitoring_configs WHERE page_id = $1 AND deleted_at IS NULL`
	err := r.db.QueryRowContext(ctx, q, pageID).Scan(&c.ID, &c.PageID, &c.CheckFrequency, &c.ScheduleType, &c.Timezone, &c.BlockAdsCookies, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *MonitoringConfigPostgresRepository) Update(ctx context.Context, config *entities.MonitoringConfig) error {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return err
	}
	config.UpdatedAt = time.Now()
	q := `UPDATE monitoring_configs SET check_frequency = $1, schedule_type = $2, timezone = $3, block_ads_cookies = $4, updated_at = $5 WHERE id = $6 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, q, config.CheckFrequency, config.ScheduleType, config.Timezone, config.BlockAdsCookies, config.UpdatedAt, config.ID)
	return err
}

func (r *MonitoringConfigPostgresRepository) GetDueSnapshotTasks(ctx context.Context) ([]entities.SnapshotTask, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}

	q := `
		SELECT p.id, p.url
		FROM pages p
		JOIN monitoring_configs mc ON p.id = mc.page_id
		WHERE p.deleted_at IS NULL AND mc.deleted_at IS NULL
		AND mc.check_frequency != 'Off'
		AND (
			(mc.check_frequency = '30m' AND (p.last_checked_at IS NULL OR p.last_checked_at < NOW() - INTERVAL '30 minutes')) OR
			(mc.check_frequency = '1h' AND (p.last_checked_at IS NULL OR p.last_checked_at < NOW() - INTERVAL '1 hour')) OR
			(mc.check_frequency = '1 hr' AND (p.last_checked_at IS NULL OR p.last_checked_at < NOW() - INTERVAL '1 hour')) OR
			(mc.check_frequency = '2h' AND (p.last_checked_at IS NULL OR p.last_checked_at < NOW() - INTERVAL '2 hours')) OR
			(mc.check_frequency = '2 hr' AND (p.last_checked_at IS NULL OR p.last_checked_at < NOW() - INTERVAL '2 hours')) OR
			(mc.check_frequency = '8h' AND (p.last_checked_at IS NULL OR p.last_checked_at < NOW() - INTERVAL '8 hours')) OR
			(mc.check_frequency = '8 hr' AND (p.last_checked_at IS NULL OR p.last_checked_at < NOW() - INTERVAL '8 hours')) OR
			(mc.check_frequency = '24h' AND (p.last_checked_at IS NULL OR p.last_checked_at < NOW() - INTERVAL '1 day')) OR
			(mc.check_frequency = '1d' AND (p.last_checked_at IS NULL OR p.last_checked_at < NOW() - INTERVAL '1 day')) OR
			(mc.check_frequency = '48h' AND (p.last_checked_at IS NULL OR p.last_checked_at < NOW() - INTERVAL '48 hours')) OR
			(mc.check_frequency = '2d' AND (p.last_checked_at IS NULL OR p.last_checked_at < NOW() - INTERVAL '48 hours')) OR
			(mc.check_frequency = 'Every 30 minutes' AND (p.last_checked_at IS NULL OR p.last_checked_at < NOW() - INTERVAL '30 minutes')) OR
			(mc.check_frequency = 'Every 1 hour' AND (p.last_checked_at IS NULL OR p.last_checked_at < NOW() - INTERVAL '1 hour')) OR
			(mc.check_frequency = 'Every 2 hours' AND (p.last_checked_at IS NULL OR p.last_checked_at < NOW() - INTERVAL '2 hours')) OR
			(mc.check_frequency = 'Every 8 hours' AND (p.last_checked_at IS NULL OR p.last_checked_at < NOW() - INTERVAL '8 hours')) OR
			(mc.check_frequency = 'Every day' AND (p.last_checked_at IS NULL OR p.last_checked_at < NOW() - INTERVAL '1 day')) OR
			(mc.check_frequency = 'Every 48 hours' AND (p.last_checked_at IS NULL OR p.last_checked_at < NOW() - INTERVAL '48 hours'))
		)
		LIMIT 50
	`

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
