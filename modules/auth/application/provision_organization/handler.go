package provisionorganization

import (
	"context"
	"strings"
	"time"

	autherrors "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/errors"
	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

const defaultTrialDays = 14

// Handler orchestrates org provisioning for an already-authenticated OAuth user
// who has not yet created an organization.
//
// The flow is:
//  1. Guard — ensure the user does not already belong to an org (idempotency).
//  2. Delegate — call TrialProvisioner to atomically create the org, membership,
//     OWNER role, trial plan row, and tenant schema.
//  3. Optionally persist the 4 onboarding answers when they are supplied.
//  4. Return — surface subdomain + trial expiry to the caller.
type Handler struct {
	trialProvisioner  authservices.TrialProvisioner
	membershipChecker authservices.OrganizationMembershipChecker
	// onboardingWriter is optional; when nil the profile step is skipped.
	onboardingWriter authservices.OrganizationOnboardingWriter
	// orgFinder is optional; required only when onboardingWriter is set.
	orgFinder authservices.OrganizationOrgFinder
}

// NewHandler creates a ready-to-use Handler without onboarding profile support.
func NewHandler(
	trialProvisioner authservices.TrialProvisioner,
	membershipChecker authservices.OrganizationMembershipChecker,
) *Handler {
	return &Handler{
		trialProvisioner:  trialProvisioner,
		membershipChecker: membershipChecker,
	}
}

// WithOnboardingProfile wires the optional onboarding profile persistence.
func (h *Handler) WithOnboardingProfile(
	writer authservices.OrganizationOnboardingWriter,
	finder authservices.OrganizationOrgFinder,
) *Handler {
	h.onboardingWriter = writer
	h.orgFinder = finder
	return h
}

// Handle executes the provision_organization use case.
//
// Returns:
//   - (*Response, nil) on success
//   - (nil, autherrors.ErrAlreadyProvisioned) when the user already has an org
//   - (nil, err) for any infrastructure or provisioner failure
func (h *Handler) Handle(ctx context.Context, req Request) (*Response, error) {
	// Idempotency guard — reject if user already has a membership.
	hasMembership, err := h.membershipChecker.HasAnyMembership(ctx, req.UserID)
	if err != nil {
		logger.Error("Failed to check org membership", zap.String("user_id", req.UserID.String()), zap.Error(err))
		return nil, err
	}
	if hasMembership {
		return nil, autherrors.ErrAlreadyProvisioned
	}

	// Normalise subdomain before handing off to the provisioner.
	subdomain := strings.TrimSpace(strings.ToLower(req.Subdomain))

	out, err := h.trialProvisioner.Provision(ctx, authservices.TrialProvisionInput{
		UserID:                req.UserID,
		OrganizationName:      req.OrgName,
		OrganizationSubdomain: subdomain,
		TrialDays:             defaultTrialDays,
	})
	if err != nil {
		logger.Error("Failed to provision trial organization",
			zap.String("user_id", req.UserID.String()),
			zap.String("subdomain", subdomain),
			zap.Error(err),
		)
		return nil, err
	}

	logger.Info("OAuth user provisioned organization",
		zap.String("user_id", req.UserID.String()),
		zap.String("org_subdomain", out.OrganizationSubdomain),
		zap.Time("trial_ends_at", out.TrialEndsAt),
	)

	// Optionally persist onboarding profile answers when they were supplied.
	hasAnswers := req.CompanySize != "" || req.BusinessType != "" ||
		len(req.CompetitorChallenges) > 0 || req.WebsiteURL != ""

	if hasAnswers && h.onboardingWriter != nil && h.orgFinder != nil {
		orgID, findErr := h.orgFinder.GetUserOrgID(ctx, req.UserID)
		if findErr != nil {
			// Non-fatal: log and continue — org was created, only profile is missing.
			logger.Warn("provision_organization: failed to find org for profile save (non-fatal)",
				zap.String("user_id", req.UserID.String()),
				zap.Error(findErr),
			)
		} else if orgID != nil {
			profileInput := authservices.OnboardingProfileInput{
				OrgID:                 *orgID,
				CompanySize:           req.CompanySize,
				BusinessType:          req.BusinessType,
				CompetitorChallenges:  req.CompetitorChallenges,
				WebsiteURL:            req.WebsiteURL,
				OnboardingCompletedAt: time.Now().UTC(),
			}
			if writeErr := h.onboardingWriter.SaveOnboardingProfile(ctx, profileInput); writeErr != nil {
				// Non-fatal: org was provisioned; profile can be re-saved later.
				logger.Warn("provision_organization: failed to save onboarding profile (non-fatal)",
					zap.String("user_id", req.UserID.String()),
					zap.Error(writeErr),
				)
			}
		}
	}

	return &Response{
		OrganizationSubdomain: out.OrganizationSubdomain,
		TrialEndsAt:           out.TrialEndsAt,
	}, nil
}
