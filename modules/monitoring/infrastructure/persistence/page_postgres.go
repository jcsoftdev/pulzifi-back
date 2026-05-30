package persistence

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
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
	q := `UPDATE pages SET last_checked_at = NOW() WHERE id = $1`
	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q, pageID)
		return err
	})
}
