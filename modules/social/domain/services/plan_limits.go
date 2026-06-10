package services

import "context"

// PlanLimits is the port for reading the social-media plan constraints for an
// organisation. The concrete adapter lives in cmd/wiring/social/plan_lookup.go
// and queries public.plans + public.organization_plans via raw SQL.
type PlanLimits interface {
	// GetProfilesLimit returns the maximum number of social profiles allowed for
	// the organisation identified by workspaceID (looked up via the tenant schema).
	// Returns 0 for unlimited (NULL in the DB maps to 0 here).
	// Returns -1 when the feature is disabled for the plan (social_profiles_limit=0 in DB).
	GetProfilesLimit(ctx context.Context, workspaceID string) (int, error)

	// GetChecksPerDay returns the maximum number of social checks per day for the
	// organisation. Returns 0 for unlimited (NULL in the DB maps to 0 here).
	// Returns -1 when the feature is disabled (social_checks_per_day=0 in DB).
	GetChecksPerDay(ctx context.Context, workspaceID string) (int, error)

	// GetChecksUsedToday returns the number of social checks already consumed
	// today by the tenant (used by get_quota_status — read-only, no consume).
	GetChecksUsedToday(ctx context.Context) (int, error)
}
