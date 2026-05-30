package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
)

// UsageTrackingPostgresRepository implements repositories.UsageTrackingRepository.
// It uses database.WithTenant for every query (tenant-scoped).
type UsageTrackingPostgresRepository struct {
	db     *sql.DB
	tenant string
}

// NewUsageTrackingPostgresRepository creates a tenant-scoped usage tracking repository.
func NewUsageTrackingPostgresRepository(db *sql.DB, tenant string) repositories.UsageTrackingRepository {
	return &UsageTrackingPostgresRepository{db: db, tenant: tenant}
}

func (r *UsageTrackingPostgresRepository) FindCurrent(ctx context.Context, today time.Time) (*entities.UsageTracking, error) {
	ut := &entities.UsageTracking{}
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx, `
			SELECT period_start, period_end,
			       checks_allowed, checks_used,
			       ai_insights_allowed, ai_insights_used,
			       next_refill_at
			FROM usage_tracking
			WHERE period_start <= $1::date AND period_end >= $1::date
			ORDER BY period_end DESC
			LIMIT 1
		`, today).Scan(
			&ut.PeriodStart, &ut.PeriodEnd,
			&ut.ChecksAllowed, &ut.ChecksUsed,
			&ut.AIInsightsAllowed, &ut.AIInsightsUsed,
			&ut.NextRefillAt,
		)
		if err == sql.ErrNoRows {
			return sql.ErrNoRows
		}
		return err
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return ut, nil
}

func (r *UsageTrackingPostgresRepository) Insert(ctx context.Context, ut *entities.UsageTracking) error {
	nextRefill := ut.PeriodEnd.AddDate(0, 0, 1)
	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO usage_tracking (period_start, period_end, checks_allowed, checks_used, ai_insights_allowed, ai_insights_used, last_refill_at, next_refill_at, created_at, updated_at)
			VALUES ($1, $2, $3, 0, $5, 0, NOW(), $4, NOW(), NOW())
			ON CONFLICT DO NOTHING
		`, ut.PeriodStart, ut.PeriodEnd, ut.ChecksAllowed, nextRefill, ut.AIInsightsAllowed)
		return err
	})
}

func (r *UsageTrackingPostgresRepository) LatestPeriodEnd(ctx context.Context) (time.Time, error) {
	var t time.Time
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, `SELECT period_end FROM usage_tracking ORDER BY period_end DESC LIMIT 1`).Scan(&t)
	})
	return t, err
}

func (r *UsageTrackingPostgresRepository) SyncChecksAllowed(ctx context.Context, checksAllowed int) error {
	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			UPDATE usage_tracking
			SET checks_allowed = $1, updated_at = NOW()
			WHERE period_start <= CURRENT_DATE AND period_end >= CURRENT_DATE
		`, checksAllowed)
		return err
	})
}

func (r *UsageTrackingPostgresRepository) CountChecks(ctx context.Context) (int, int, int, error) {
	var total, success, failed int
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		tx.QueryRowContext(ctx, `
			SELECT COUNT(*),
			       COUNT(*) FILTER (WHERE status = 'success'),
			       COUNT(*) FILTER (WHERE status = 'error')
			FROM checks
		`).Scan(&total, &success, &failed) //nolint:errcheck
		return nil
	})
	return total, success, failed, err
}

func (r *UsageTrackingPostgresRepository) CountPages(ctx context.Context) (int, error) {
	var n int
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM pages WHERE deleted_at IS NULL`).Scan(&n) //nolint:errcheck
		return nil
	})
	return n, err
}

func (r *UsageTrackingPostgresRepository) CountWorkspaces(ctx context.Context) (int, error) {
	var n int
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM workspaces WHERE deleted_at IS NULL`).Scan(&n) //nolint:errcheck
		return nil
	})
	return n, err
}

func (r *UsageTrackingPostgresRepository) CountAlerts(ctx context.Context) (int, error) {
	var n int
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM alerts`).Scan(&n) //nolint:errcheck
		return nil
	})
	return n, err
}
