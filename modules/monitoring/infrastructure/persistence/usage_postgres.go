package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
)

type UsagePostgresRepository struct {
	db     *sql.DB
	tenant string
}

func NewUsagePostgresRepository(db *sql.DB, tenant string) *UsagePostgresRepository {
	return &UsagePostgresRepository{
		db:     db,
		tenant: tenant,
	}
}

func (r *UsagePostgresRepository) HasQuota(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return false, err
	}

	// Check if active period exists and has quota
	q := `SELECT EXISTS (
		SELECT 1 FROM usage_tracking 
		WHERE period_start <= $1 AND period_end >= $1 
		AND checks_used < checks_allowed
	)`
	
	var hasQuota bool
	err := r.db.QueryRowContext(ctx, q, time.Now()).Scan(&hasQuota)
	if err != nil {
		// If query fails, we assume no quota to be safe, or return error
		return false, err
	}
	return hasQuota, nil
}

func (r *UsagePostgresRepository) LogUsage(ctx context.Context, pageID, checkID uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return err
	}

	// Transaction to update tracking and insert log
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update tracking
	qUpdate := `UPDATE usage_tracking 
		SET checks_used = checks_used + 1 
		WHERE period_start <= $1 AND period_end >= $1`
	if _, err := tx.ExecContext(ctx, qUpdate, time.Now()); err != nil {
		return err
	}

	// Insert log
	qInsert := `INSERT INTO usage_logs (page_id, check_id, checks_consumed) VALUES ($1, $2, 1)`
	if _, err := tx.ExecContext(ctx, qInsert, pageID, checkID); err != nil {
		return err
	}

	return tx.Commit()
}
