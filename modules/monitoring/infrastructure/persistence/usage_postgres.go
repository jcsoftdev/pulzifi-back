package persistence

import (
	"context"
	"database/sql"
	"fmt"
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

	if err := r.ensureCurrentPeriod(ctx); err != nil {
		return false, fmt.Errorf("ensure billing period: %w", err)
	}

	q := `SELECT EXISTS (
		SELECT 1 FROM usage_tracking
		WHERE period_start <= $1 AND period_end >= $1
		AND checks_used < checks_allowed
	)`

	var hasQuota bool
	err := r.db.QueryRowContext(ctx, q, time.Now()).Scan(&hasQuota)
	if err != nil {
		return false, err
	}
	return hasQuota, nil
}

func (r *UsagePostgresRepository) LogUsage(ctx context.Context, pageID, checkID uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return err
	}

	if err := r.ensureCurrentPeriod(ctx); err != nil {
		return fmt.Errorf("ensure billing period: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qUpdate := `UPDATE usage_tracking
		SET checks_used = checks_used + 1
		WHERE period_start <= $1 AND period_end >= $1`
	if _, err := tx.ExecContext(ctx, qUpdate, time.Now()); err != nil {
		return err
	}

	qInsert := `INSERT INTO usage_logs (page_id, check_id, checks_consumed) VALUES ($1, $2, 1)`
	if _, err := tx.ExecContext(ctx, qInsert, pageID, checkID); err != nil {
		return err
	}

	return tx.Commit()
}

// ensureCurrentPeriod checks if a usage_tracking row exists for the current billing period.
// If not, it creates one based on the org's plan and the plan's started_at anchor day.
func (r *UsagePostgresRepository) ensureCurrentPeriod(ctx context.Context) error {
	now := time.Now()

	// Check if a row already exists for today
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS (
		SELECT 1 FROM usage_tracking WHERE period_start <= $1::date AND period_end >= $1::date
	)`, now).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	// Look up plan info
	planQuery := `
		SELECT p.checks_allowed_monthly, op.started_at
		FROM public.organizations o
		JOIN public.organization_plans op ON op.organization_id = o.id
			AND op.status = 'active' AND op.deleted_at IS NULL
		JOIN public.plans p ON p.id = op.plan_id
		WHERE o.schema_name = $1
		ORDER BY op.started_at DESC
		LIMIT 1
	`
	var checksAllowed int
	var startedAt time.Time
	if err := r.db.QueryRowContext(ctx, planQuery, r.tenant).Scan(&checksAllowed, &startedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil // no plan, nothing to create
		}
		return err
	}

	anchorDay := startedAt.Day()
	periodStart, periodEnd := billingPeriodForDate(now, anchorDay)
	nextRefill := periodEnd.AddDate(0, 0, 1)

	insertQ := `
		INSERT INTO usage_tracking (period_start, period_end, checks_allowed, checks_used, last_refill_at, next_refill_at, created_at, updated_at)
		VALUES ($1, $2, $3, 0, NOW(), $4, NOW(), NOW())
		ON CONFLICT DO NOTHING
	`
	_, err = r.db.ExecContext(ctx, insertQ, periodStart, periodEnd, checksAllowed, nextRefill)
	return err
}

func billingPeriodForDate(today time.Time, anchorDay int) (start, end time.Time) {
	year, month, day := today.Date()

	lastDay := daysInMonth(year, month)
	clampedAnchor := anchorDay
	if clampedAnchor > lastDay {
		clampedAnchor = lastDay
	}

	if day >= clampedAnchor {
		start = time.Date(year, month, clampedAnchor, 0, 0, 0, 0, time.UTC)
	} else {
		prevMonth := month - 1
		prevYear := year
		if prevMonth < 1 {
			prevMonth = 12
			prevYear--
		}
		prevLastDay := daysInMonth(prevYear, prevMonth)
		prevAnchor := anchorDay
		if prevAnchor > prevLastDay {
			prevAnchor = prevLastDay
		}
		start = time.Date(prevYear, prevMonth, prevAnchor, 0, 0, 0, 0, time.UTC)
	}

	nextMonth := start.Month() + 1
	nextYear := start.Year()
	if nextMonth > 12 {
		nextMonth = 1
		nextYear++
	}
	nextLastDay := daysInMonth(nextYear, nextMonth)
	nextAnchor := anchorDay
	if nextAnchor > nextLastDay {
		nextAnchor = nextLastDay
	}
	end = time.Date(nextYear, nextMonth, nextAnchor, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)

	return start, end
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
