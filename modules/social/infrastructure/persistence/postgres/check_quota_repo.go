package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/social/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
)

// Compile-time interface check.
var _ services.CheckQuota = (*CheckQuotaPostgresRepository)(nil)

// CheckQuotaPostgresRepository implements services.CheckQuota using PostgreSQL.
// It uses an atomic upsert (INSERT … ON CONFLICT … WHERE) that mirrors the
// pattern from shared/integrationusage for race-safe daily quota enforcement.
type CheckQuotaPostgresRepository struct {
	db     *sql.DB
	tenant string
}

// NewCheckQuotaPostgresRepository creates a new tenant-scoped CheckQuota implementation.
func NewCheckQuotaPostgresRepository(db *sql.DB, tenant string) *CheckQuotaPostgresRepository {
	return &CheckQuotaPostgresRepository{db: db, tenant: tenant}
}

// Consume atomically debits one check from the daily pool for the given date.
//
// limit=0 means unlimited: Consume always succeeds.
// Returns domainerrors.ErrQuotaExceeded (and Allowed=false) when the pool is exhausted.
func (r *CheckQuotaPostgresRepository) Consume(ctx context.Context, date time.Time, limit int) (services.QuotaResult, error) {
	usageDate := date.UTC().Format("2006-01-02")

	// Unlimited plan: just increment, no cap.
	if limit == 0 {
		used, err := r.increment(ctx, usageDate)
		if err != nil {
			return services.QuotaResult{}, fmt.Errorf("check quota: consume unlimited: %w", err)
		}
		return services.QuotaResult{Allowed: true, Used: used, Limit: 0}, nil
	}

	// Atomic upsert with WHERE guard: RETURNING is empty exactly when quota is exceeded.
	//
	// Note: SET LOCAL search_path is applied by database.WithTenant. The table
	// social_check_usage lives in the tenant schema and is accessed without
	// schema qualification because search_path is already set.
	q := `
		INSERT INTO social_check_usage (usage_date, checks_used)
		VALUES ($1, 1)
		ON CONFLICT (usage_date) DO UPDATE
		  SET checks_used = social_check_usage.checks_used + 1
		WHERE social_check_usage.checks_used < $2
		RETURNING checks_used`

	var checksUsed int
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, q, usageDate, limit).Scan(&checksUsed)
	})
	if errors.Is(err, sql.ErrNoRows) {
		// RETURNING was empty → WHERE clause blocked the update → quota exceeded.
		return services.QuotaResult{Allowed: false, Used: limit, Limit: limit}, domainerrors.ErrQuotaExceeded
	}
	if err != nil {
		return services.QuotaResult{}, fmt.Errorf("check quota: consume: %w", err)
	}

	return services.QuotaResult{Allowed: true, Used: checksUsed, Limit: limit}, nil
}

// Compensate decrements the daily counter by one, reversing a Consume call
// that was followed by a provider 5xx error.
// Idempotent: will not decrement below zero.
func (r *CheckQuotaPostgresRepository) Compensate(ctx context.Context, date time.Time) error {
	usageDate := date.UTC().Format("2006-01-02")

	q := `
		UPDATE social_check_usage
		SET checks_used = GREATEST(checks_used - 1, 0)
		WHERE usage_date = $1`

	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q, usageDate)
		if err != nil {
			return fmt.Errorf("check quota: compensate: %w", err)
		}
		return nil
	})
}

// GetUsageForDate returns the current checks_used for a given date.
// Returns 0 when no row exists yet (no checks run today).
// Used for the quota status endpoint.
func (r *CheckQuotaPostgresRepository) GetUsageForDate(ctx context.Context, date time.Time) (int, error) {
	usageDate := date.UTC().Format("2006-01-02")
	q := `SELECT checks_used FROM social_check_usage WHERE usage_date = $1`

	var checksUsed int
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, q, usageDate).Scan(&checksUsed)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("check quota: get usage: %w", err)
	}
	return checksUsed, nil
}

// increment unconditionally increments checks_used for unlimited plans and returns
// the new count. Uses a simple upsert with no WHERE guard.
func (r *CheckQuotaPostgresRepository) increment(ctx context.Context, usageDate string) (int, error) {
	q := `
		INSERT INTO social_check_usage (usage_date, checks_used)
		VALUES ($1, 1)
		ON CONFLICT (usage_date) DO UPDATE
		  SET checks_used = social_check_usage.checks_used + 1
		RETURNING checks_used`

	var checksUsed int
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, q, usageDate).Scan(&checksUsed)
	})
	return checksUsed, err
}
