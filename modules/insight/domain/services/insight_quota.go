package services

import (
	"context"
	"errors"
)

// ErrInsightQuotaExhausted is returned by GenerateInsightsHandler when the
// tenant has already produced ai_insights_allowed insights in the current
// billing period. Callers MUST treat this as a non-fatal "skip" — log + return
// nil to upstream, not a 5xx propagation. The cap is hard (no overage billing).
var ErrInsightQuotaExhausted = errors.New("insight: monthly AI insight quota exhausted for tenant")

// InsightQuotaSnapshot is the per-tenant view of the AI insight allowance for
// the active billing period. Unlimited is derived from the sentinel sentinel
// value INT max written by the PlanAssigner when the plan's
// ai_insights_allowed_monthly is NULL (Enterprise).
type InsightQuotaSnapshot struct {
	Used      int
	Allowed   int
	Unlimited bool
}

// InsightQuotaReader gates AI insight generation by exposing the current
// usage counter and atomically incrementing it after each successful insight
// persistence. Implementations are tenant-aware: tenant is the schema name.
//
// Read MUST never error when the underlying schema lacks the new columns
// (deploy precedes tenant migration) — return Unlimited:true so generation
// proceeds, then quota activates automatically after migration runs.
//
// Increment MUST use a single SQL UPDATE … RETURNING so concurrent insight
// generation in the same period serializes at the row level.
type InsightQuotaReader interface {
	Read(ctx context.Context, tenant string) (InsightQuotaSnapshot, error)
	Increment(ctx context.Context, tenant string) (post InsightQuotaSnapshot, err error)
}
