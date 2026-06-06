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

	// UpsertStripePricing writes the Stripe price ID and cached amount for one
	// billing cycle of an existing plan. cycle is "monthly" | "yearly".
	// amountCents is the Stripe unit_amount; currency is the lowercase ISO code.
	// Returns ErrPlanNotFound when no plan with that code exists. The other
	// cycle's columns are left untouched.
	UpsertStripePricing(ctx context.Context, planCode, cycle, priceID string, amountCents int64, currency string) error
}
