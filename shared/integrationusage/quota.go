package integrationusage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var ErrQuotaExceeded = errors.New("integration usage: monthly quota exceeded")

type Tracker struct{ db *sql.DB }

func NewTracker(db *sql.DB) *Tracker { return &Tracker{db: db} }

// AllowedFunc returns the org's allowed count for the period.
type AllowedFunc func(context.Context, uuid.UUID) (int, error)

// CheckAndIncrement atomically increments count_used for the current calendar
// month UTC, creating a fresh row if one doesn't exist for this period.
//
// Returns:
//   - nil on success (quota incremented)
//   - ErrQuotaExceeded when the org has already reached its limit
//   - wrapped error on DB failure or allowedFor error
//
// Atomicity: ON CONFLICT … DO UPDATE … WHERE atomically applies the
// increment OR rejects via WHERE filter. RETURNING is empty exactly when
// the WHERE-filtered update didn't apply (= quota exceeded).
func (t *Tracker) CheckAndIncrement(
	ctx context.Context,
	orgID uuid.UUID,
	serviceType string,
	allowedFor AllowedFunc,
) error {
	period := firstOfMonthUTC(time.Now())

	allowed, err := allowedFor(ctx, orgID)
	if err != nil {
		return fmt.Errorf("integrationusage: allowedFor: %w", err)
	}
	if allowed <= 0 {
		return ErrQuotaExceeded
	}

	var (
		countUsed    int
		countAllowed int
	)
	err = t.db.QueryRowContext(ctx, `
		INSERT INTO integration_send_quotas (org_id, service_type, period_start, count_allowed, count_used)
		VALUES ($1, $2, $3, $4, 1)
		ON CONFLICT (org_id, service_type, period_start)
		DO UPDATE SET count_used = integration_send_quotas.count_used + 1
		WHERE integration_send_quotas.count_used < integration_send_quotas.count_allowed
		RETURNING count_used, count_allowed
	`, orgID, serviceType, period, allowed).Scan(&countUsed, &countAllowed)

	if errors.Is(err, sql.ErrNoRows) {
		// RETURNING was empty → WHERE-filtered the UPDATE → quota was exceeded
		return ErrQuotaExceeded
	}
	if err != nil {
		return fmt.Errorf("integrationusage: upsert: %w", err)
	}
	return nil
}

func firstOfMonthUTC(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), 1, 0, 0, 0, 0, time.UTC)
}
