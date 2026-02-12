package persistence

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
)

type MonitoringPagePostgresRepository struct {
	db     *sql.DB
	tenant string
}

func NewMonitoringPagePostgresRepository(db *sql.DB, tenant string) *MonitoringPagePostgresRepository {
	return &MonitoringPagePostgresRepository{
		db:     db,
		tenant: tenant,
	}
}

func (r *MonitoringPagePostgresRepository) UpdateLastChecked(ctx context.Context, pageID uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return err
	}

	q := `UPDATE pages SET last_checked_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, q, pageID)
	return err
}
