package services

import (
	"context"

	"github.com/google/uuid"
)

// UsageSnapshot reports the current billing period's check counters for an org.
// Returned by UsageReader so the billing module can compute usage-based
// proration credits without importing the usage module's infrastructure.
type UsageSnapshot struct {
	ChecksUsed    int
	ChecksAllowed int
}

// RemainingRatio returns the fraction of checks NOT yet consumed in the current
// period. Values are clamped to [0, 1]. Used by the usage-based proration path
// to compute the refund credit when a customer upgrades mid-cycle. Returns 0
// when ChecksAllowed is non-positive so the caller never produces a credit
// against an org with no quota row.
func (s UsageSnapshot) RemainingRatio() float64 {
	if s.ChecksAllowed <= 0 {
		return 0
	}
	rem := s.ChecksAllowed - s.ChecksUsed
	if rem <= 0 {
		return 0
	}
	if rem >= s.ChecksAllowed {
		return 1
	}
	return float64(rem) / float64(s.ChecksAllowed)
}

// UsageReader resolves the active usage_tracking row for an org. Implemented in
// cmd/wiring/billing/usage_reader.go to keep cross-module SQL out of the
// domain. Returns a zero-value snapshot (not nil error) when no active row
// exists — the caller must treat that as "no credit possible".
type UsageReader interface {
	ReadActiveUsage(ctx context.Context, orgID uuid.UUID) (UsageSnapshot, error)
}
