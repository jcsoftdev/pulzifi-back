package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

// pqUndefinedColumn is the SQLState returned by Postgres when a SELECT
// references a column that does not exist. Used so the quota reader can
// gracefully degrade to "unlimited" between a deploy and the moment the
// tenant migration adds ai_insights_* columns to usage_tracking.
const pqUndefinedColumn = "42703"

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

// Read returns the current period's ai insight usage/cap. When the column
// does not yet exist (tenant migration pending) → returns Unlimited:true and
// logs a warning so generation continues unimpeded.
func (r *QuotaPostgresRepository) Read(ctx context.Context, tenant string) (services.InsightQuotaSnapshot, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(tenant)); err != nil {
		return services.InsightQuotaSnapshot{}, err
	}

	const q = `
		SELECT COALESCE(ai_insights_used, 0), COALESCE(ai_insights_allowed, 0)
		  FROM usage_tracking
		 WHERE period_start <= $1 AND period_end >= $1
		 ORDER BY period_start DESC
		 LIMIT 1
	`

	var snap services.InsightQuotaSnapshot
	err := r.db.QueryRowContext(ctx, q, time.Now()).Scan(&snap.Used, &snap.Allowed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No active period — let generation proceed; the next call to
			// get_quotas (which already auto-creates a period) will seed
			// counters from the plan.
			return services.InsightQuotaSnapshot{Unlimited: true}, nil
		}
		if isUndefinedColumn(err) {
			logger.Warn("insight quota: ai_insights_* columns missing on tenant — treating as unlimited until migration runs",
				zap.String("tenant", tenant))
			return services.InsightQuotaSnapshot{Unlimited: true}, nil
		}
		return services.InsightQuotaSnapshot{}, err
	}
	snap.Unlimited = snap.Allowed >= unlimitedSentinel
	return snap, nil
}

// Increment atomically bumps ai_insights_used by 1 on the active period row
// and returns the post-update snapshot. Uses UPDATE … RETURNING so concurrent
// goroutines serialize on the row lock — no read-modify-write race.
func (r *QuotaPostgresRepository) Increment(ctx context.Context, tenant string) (services.InsightQuotaSnapshot, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(tenant)); err != nil {
		return services.InsightQuotaSnapshot{}, err
	}

	const q = `
		UPDATE usage_tracking
		   SET ai_insights_used = ai_insights_used + 1,
		       updated_at       = NOW()
		 WHERE period_start <= $1 AND period_end >= $1
		 RETURNING ai_insights_used, ai_insights_allowed
	`

	var snap services.InsightQuotaSnapshot
	err := r.db.QueryRowContext(ctx, q, time.Now()).Scan(&snap.Used, &snap.Allowed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No active period — treat as unlimited so the handler does not
			// abort a successful insight just because the period row hasn't
			// been seeded yet.
			return services.InsightQuotaSnapshot{Unlimited: true}, nil
		}
		if isUndefinedColumn(err) {
			logger.Warn("insight quota: ai_insights_* columns missing on tenant during increment — no-op",
				zap.String("tenant", tenant))
			return services.InsightQuotaSnapshot{Unlimited: true}, nil
		}
		return services.InsightQuotaSnapshot{}, err
	}
	snap.Unlimited = snap.Allowed >= unlimitedSentinel
	return snap, nil
}

func isUndefinedColumn(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return string(pqErr.Code) == pqUndefinedColumn
	}
	return false
}
