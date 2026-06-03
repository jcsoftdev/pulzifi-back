package services

import (
	"context"

	"github.com/google/uuid"
)

// OrgContext is the combined org payload needed for /me. Returning a single
// composed struct from one interface call avoids three round-trips to different
// repos (organizations + organization_plans + feature_flags).
type OrgContext struct {
	ID                   uuid.UUID
	Name                 string
	Subdomain            string
	PlanCode             string         // "" if no active plan (free tier)
	FeatureFlags         map[string]any // {} if column empty
	OnboardingCompleted  bool           // true when onboarding_completed_at IS NOT NULL
}

// OrgContextLookup is implemented at the composition root (cmd/wiring/integration/)
// to avoid auth importing other modules.
type OrgContextLookup interface {
	// Lookup returns the OrgContext for the user's primary org.
	// Returns (nil, nil) if user has no org (treated as legacy single-org flat shape upstream).
	Lookup(ctx context.Context, userID uuid.UUID) (*OrgContext, error)
}
