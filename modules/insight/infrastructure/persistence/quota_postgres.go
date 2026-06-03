package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
)

// unlimitedSentinel mirrors the value PlanAssigner writes when the active
// plan's ai_insights_allowed_monthly is NULL (Enterprise). Anything equal to
// or above this is considered "no cap".
const unlimitedSentinel = 2147483647

// Compile-time interface check.
var _ services.InsightQuotaReader = (*QuotaPostgresRepository)(nil)

// QuotaPostgresRepository implements services.InsightQuotaReader against
// each tenant's usage_tracking row for the active billing period.
//
// Tenant arrives per-call (not stored on the struct) because a single
// QuotaPostgresRepository instance is shared across all tenants; the wiring
// adapter passes the schema name from snapshot.InsightRequest.SchemaName.
type QuotaPostgresRepository struct {
	db *sql.DB
}

// NewQuotaPostgresRepository constructs a process-wide reader. Safe to share.
func NewQuotaPostgresRepository(db *sql.DB) *QuotaPostgresRepository {
	return &QuotaPostgresRepository{db: db}
}

// Read returns the current period's ai insight usage/cap. No active period →
// Unlimited:true so generation proceeds (get_quotas seeds the period on next read).
func (r *QuotaPostgresRepository) Read(ctx context.Context, tenant string) (services.InsightQuotaSnapshot, error) {
	const q = `
		SELECT ai_insights_used, ai_insights_allowed
		  FROM usage_tracking
		 WHERE period_start <= $1 AND period_end >= $1
		 ORDER BY period_start DESC
		 LIMIT 1
	`

	var snap services.InsightQuotaSnapshot
	err := database.WithTenant(ctx, r.db, tenant, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx, q, time.Now()).Scan(&snap.Used, &snap.Allowed)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				snap = services.InsightQuotaSnapshot{Unlimited: true}
				return nil
			}
			return err
		}
		snap.Unlimited = snap.Allowed >= unlimitedSentinel
		return nil
	})
	if err != nil {
		return services.InsightQuotaSnapshot{}, err
	}
	return snap, nil
}

// Increment atomically bumps ai_insights_used by 1 on the active period row
// and returns the post-update snapshot. Uses UPDATE … RETURNING so concurrent
// goroutines serialize on the row lock — no read-modify-write race.
func (r *QuotaPostgresRepository) Increment(ctx context.Context, tenant string) (services.InsightQuotaSnapshot, error) {
	const q = `
		UPDATE usage_tracking
		   SET ai_insights_used = ai_insights_used + 1,
		       updated_at       = NOW()
		 WHERE period_start <= $1 AND period_end >= $1
		 RETURNING ai_insights_used, ai_insights_allowed
	`

	var snap services.InsightQuotaSnapshot
	err := database.WithTenant(ctx, r.db, tenant, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx, q, time.Now()).Scan(&snap.Used, &snap.Allowed)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				snap = services.InsightQuotaSnapshot{Unlimited: true}
				return nil
			}
			return err
		}
		snap.Unlimited = snap.Allowed >= unlimitedSentinel
		return nil
	})
	if err != nil {
		return services.InsightQuotaSnapshot{}, err
	}
	return snap, nil
}
