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
