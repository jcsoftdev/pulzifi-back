package authwiring

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
)

// OnboardingProfileAdapter implements authservices.OrganizationOnboardingWriter
// and authservices.OrganizationOrgFinder using raw SQL against public.organizations.
// public.organizations is in the PUBLIC schema — no search_path required.
type OnboardingProfileAdapter struct {
	db *sql.DB
}

// NewOnboardingProfileAdapter creates the adapter.
func NewOnboardingProfileAdapter(db *sql.DB) *OnboardingProfileAdapter {
	return &OnboardingProfileAdapter{db: db}
}

var _ authservices.OrganizationOnboardingWriter = (*OnboardingProfileAdapter)(nil)
var _ authservices.OrganizationOrgFinder = (*OnboardingProfileAdapter)(nil)

// GetUserOrgID returns the primary organization ID for the given user.
// Returns (nil, nil) if the user has no org.
func (a *OnboardingProfileAdapter) GetUserOrgID(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error) {
	var orgID uuid.UUID
	err := a.db.QueryRowContext(ctx, `
		SELECT om.organization_id
		  FROM public.organization_members om
		  JOIN public.organizations o ON o.id = om.organization_id
		 WHERE om.user_id = $1
		   AND o.deleted_at IS NULL
		 ORDER BY om.created_at ASC
		 LIMIT 1
	`, userID).Scan(&orgID)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &orgID, nil
}

// SaveOnboardingProfile upserts the 4 onboarding answers and sets
// onboarding_completed_at on the organization row.
func (a *OnboardingProfileAdapter) SaveOnboardingProfile(ctx context.Context, input authservices.OnboardingProfileInput) error {
	challengesJSON, err := json.Marshal(input.CompetitorChallenges)
	if err != nil {
		return err
	}

	res, err := a.db.ExecContext(ctx, `
		UPDATE public.organizations
		   SET company_size            = $1,
		       business_type           = $2,
		       competitor_challenges   = $3,
		       website_url             = $4,
		       onboarding_completed_at = $5,
		       updated_at              = NOW()
		 WHERE id = $6
	`,
		nullableString(input.CompanySize),
		nullableString(input.BusinessType),
		challengesJSON,
		nullableString(input.WebsiteURL),
		input.OnboardingCompletedAt,
		input.OrgID,
	)
	if err != nil {
		return err
	}

	// Guard against a silent no-op: if the org row no longer exists, the caller
	// must not be told onboarding completed successfully.
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("save onboarding profile: organization %s not found", input.OrgID)
	}
	return nil
}

// nullableString converts an empty string to nil (for nullable VARCHAR columns).
func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
