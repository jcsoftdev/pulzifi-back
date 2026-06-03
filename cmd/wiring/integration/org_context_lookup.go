package intwiring

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"

	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
)

// OrgContextLookup composes feature flags + plan code + org identity into a single
// payload for the /me endpoint. Single SQL with LEFT JOINs for one round-trip.
type OrgContextLookup struct{ db *sql.DB }

func NewOrgContextLookup(db *sql.DB) *OrgContextLookup {
	return &OrgContextLookup{db: db}
}

var _ authservices.OrgContextLookup = (*OrgContextLookup)(nil)

// Lookup finds the user's primary org (their first org membership) and returns
// its identity + active plan code + feature flags + onboarding status as a single struct.
// Returns (nil, nil) if the user has no org.
func (l *OrgContextLookup) Lookup(ctx context.Context, userID uuid.UUID) (*authservices.OrgContext, error) {
	var (
		orgID                uuid.UUID
		name                 string
		subdomain            string
		planCode             sql.NullString
		flagsRaw             []byte
		onboardingCompletedAt sql.NullTime
	)
	// organization_members joins to organizations; organization_plans is the active row (if any).
	// Use ORDER BY created_at to pick the user's *first* org (matches Phase 1's GetUserFirstOrganization semantic).
	err := l.db.QueryRowContext(ctx, `
		SELECT o.id, o.name, o.subdomain, p.code, o.feature_flags, o.onboarding_completed_at
		  FROM public.organization_members om
		  JOIN public.organizations o ON o.id = om.organization_id
		  LEFT JOIN public.organization_plans op
		         ON op.organization_id = o.id
		        AND op.status = 'active'
		        AND op.deleted_at IS NULL
		  LEFT JOIN public.plans p ON p.id = op.plan_id
		 WHERE om.user_id = $1
		   AND o.deleted_at IS NULL
		 ORDER BY om.created_at ASC
		 LIMIT 1
	`, userID).Scan(&orgID, &name, &subdomain, &planCode, &flagsRaw, &onboardingCompletedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	flags := map[string]any{}
	if len(flagsRaw) > 0 {
		// Defensive: malformed JSONB shouldn't break /me. Log not needed — auth handler will fail-open.
		_ = json.Unmarshal(flagsRaw, &flags)
	}

	return &authservices.OrgContext{
		ID:                  orgID,
		Name:                name,
		Subdomain:           subdomain,
		PlanCode:            planCode.String, // empty string if no active plan
		FeatureFlags:        flags,
		OnboardingCompleted: onboardingCompletedAt.Valid,
	}, nil
}
