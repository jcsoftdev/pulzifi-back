package saveonboardingprofile

import (
	"context"
	"fmt"
	"time"

	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
)

// Handler persists the 4 onboarding answers to the caller's organization and
// marks onboarding as complete. Re-submitting is idempotent (overwrites).
type Handler struct {
	orgFinder authservices.OrganizationOrgFinder
	writer    authservices.OrganizationOnboardingWriter
}

// NewHandler creates a ready-to-use Handler.
func NewHandler(
	orgFinder authservices.OrganizationOrgFinder,
	writer authservices.OrganizationOnboardingWriter,
) *Handler {
	return &Handler{orgFinder: orgFinder, writer: writer}
}

// Handle executes the save_onboarding_profile use case.
func (h *Handler) Handle(ctx context.Context, req Request) (*Response, error) {
	orgID, err := h.orgFinder.GetUserOrgID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if orgID == nil {
		return nil, fmt.Errorf("user has no organization")
	}

	input := authservices.OnboardingProfileInput{
		OrgID:                 *orgID,
		CompanySize:           req.CompanySize,
		BusinessType:          req.BusinessType,
		CompetitorChallenges:  req.CompetitorChallenges,
		WebsiteURL:            req.WebsiteURL,
		OnboardingCompletedAt: time.Now().UTC(),
	}

	if err := h.writer.SaveOnboardingProfile(ctx, input); err != nil {
		return nil, err
	}

	return &Response{OnboardingCompleted: true}, nil
}
