package provisionorganization

import (
	"context"
	"strings"

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
//  3. Return — surface subdomain + trial expiry to the caller.
type Handler struct {
	trialProvisioner authservices.TrialProvisioner
	membershipChecker authservices.OrganizationMembershipChecker
}

// NewHandler creates a ready-to-use Handler.
func NewHandler(
	trialProvisioner authservices.TrialProvisioner,
	membershipChecker authservices.OrganizationMembershipChecker,
) *Handler {
	return &Handler{
		trialProvisioner:  trialProvisioner,
		membershipChecker: membershipChecker,
	}
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

	return &Response{
		OrganizationSubdomain: out.OrganizationSubdomain,
		TrialEndsAt:           out.TrialEndsAt,
	}, nil
}
