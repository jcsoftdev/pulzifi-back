package services

import (
	"context"
	"time"
)

// QuotaResult holds the outcome of a quota consume attempt.
type QuotaResult struct {
	// Allowed is true when the consume succeeded (a check credit was deducted).
	Allowed bool

	// Used is the number of checks consumed so far today (after this consume).
	Used int

	// Limit is the plan's daily cap. 0 means unlimited (no cap enforced).
	Limit int
}

// CheckQuota is the port for enforcing the daily check quota.
//
// The concrete adapter lives in infrastructure/persistence/postgres/check_quota_repo.go.
// It uses an atomic SQL upsert (INSERT ... ON CONFLICT DO UPDATE ... WHERE
// checks_used < $limit RETURNING checks_used) — the same pattern as
// shared/integrationusage.
type CheckQuota interface {
	// Consume atomically debits one check from the daily pool for the given date.
	//
	// Returns QuotaResult.Allowed=false (and ErrQuotaExceeded) when the pool is
	// exhausted. The external fetcher MUST NOT be called in that case.
	//
	// limit=0 means unlimited: Consume always succeeds.
	Consume(ctx context.Context, date time.Time, limit int) (QuotaResult, error)

	// Compensate decrements the daily counter by one, reversing a Consume call
	// that was followed by a provider 5xx error. This prevents quota burn when
	// the check never completed successfully.
	//
	// Implementations must be idempotent and must not decrement below zero.
	Compensate(ctx context.Context, date time.Time) error
}
