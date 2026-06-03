package services

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// OnboardingProfileInput carries the 4 onboarding answers for an organization.
type OnboardingProfileInput struct {
	OrgID                 uuid.UUID
	CompanySize           string
	BusinessType          string
	CompetitorChallenges  []string
	WebsiteURL            string
	OnboardingCompletedAt time.Time
}

// OrganizationOnboardingWriter is auth's port for persisting the post-signup
// onboarding answers to the user's organization row. Implementations live in
// cmd/wiring/auth to avoid cross-module imports.
type OrganizationOnboardingWriter interface {
	// SaveOnboardingProfile upserts the 4 onboarding answers and sets
	// onboarding_completed_at for the given organization.
	SaveOnboardingProfile(ctx context.Context, input OnboardingProfileInput) error
}

// OrganizationOrgFinder is auth's port for resolving the caller's primary org ID.
// Returns (nil, nil) when the user has no org yet.
type OrganizationOrgFinder interface {
	GetUserOrgID(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error)
}
