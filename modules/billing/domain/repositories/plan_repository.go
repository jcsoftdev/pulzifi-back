package repositories

import (
	"context"
	"errors"
)

// ErrPlanNotFound is returned when a plan code cannot be resolved.
var ErrPlanNotFound = errors.New("billing: plan not found")

// PlanRepository resolves Stripe price IDs from internal plan identifiers.
type PlanRepository interface {
	// FindStripePriceID returns the Stripe price ID for the given plan code
	// (e.g. "starter", "pro", "enterprise") and billing cycle ("monthly" | "yearly").
	// Returns ErrPlanNotFound if no active plan matches the code, or an empty string
	// if the plan exists but has no price configured for that cycle.
	FindStripePriceID(ctx context.Context, planCode, billingCycle string) (string, error)
}
