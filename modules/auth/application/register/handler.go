package register

import (
	"context"
	"strings"

	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Handler orchestrates self-serve registration on a 14-day no-card trial.
//
// The previous admin-approval gate is gone: the new user is created with
// status="approved" and the TrialProvisioner atomically creates the org,
// membership, ADMIN role, trial plan row, and provisions the tenant schema.
type Handler struct {
	userRepo         repositories.UserRepository
	trialProvisioner authservices.TrialProvisioner
	orgDirectory     authservices.OrganizationDirectory
	trialDays        int
}

// NewHandler creates a new handler instance.
func NewHandler(
	userRepo repositories.UserRepository,
	trialProvisioner authservices.TrialProvisioner,
	orgDirectory authservices.OrganizationDirectory,
	trialDays int,
) *Handler {
	return &Handler{
		userRepo:         userRepo,
		trialProvisioner: trialProvisioner,
		orgDirectory:     orgDirectory,
		trialDays:        trialDays,
	}
}

// Handle executes the self-serve register use case.
func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	// Validate organization name
	if err := h.orgDirectory.ValidateOrganizationName(req.OrganizationName); err != nil {
		return nil, errors.NewUserError("INVALID_ORG_NAME", err.Error())
	}

	// Validate and normalize subdomain
	subdomain := strings.TrimSpace(strings.ToLower(req.OrganizationSubdomain))
	if err := h.orgDirectory.ValidateSubdomain(subdomain); err != nil {
		return nil, errors.NewUserError("INVALID_SUBDOMAIN", err.Error())
	}

	// Check subdomain uniqueness against existing organizations
	count, err := h.orgDirectory.CountBySubdomain(ctx, subdomain)
	if err != nil {
		logger.Error("Failed to check subdomain uniqueness", zap.Error(err))
		return nil, err
	}
	if count > 0 {
		return nil, errors.NewUserError("SUBDOMAIN_TAKEN", "subdomain is already in use")
	}

	// Check if user already exists
	exists, err := h.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		logger.Error("Failed to check if user exists", zap.Error(err))
		return nil, err
	}
	if exists {
		logger.Warn("User already exists", zap.String("email", req.Email))
		return nil, errors.NewUserError("USER_ALREADY_EXISTS", "user already exists with this email")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to hash password", zap.Error(err))
		return nil, err
	}

	// Create user entity — self-serve onboarding goes straight to approved.
	user := entities.NewUser(req.Email, string(hashedPassword), req.FirstName, req.LastName)
	user.Status = entities.UserStatusApproved

	// Persist user
	if err := h.userRepo.Create(ctx, user); err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		return nil, err
	}

	// Atomically provision org + trial plan + tenant schema.
	out, err := h.trialProvisioner.Provision(ctx, authservices.TrialProvisionInput{
		UserID:                user.ID,
		OrganizationName:      req.OrganizationName,
		OrganizationSubdomain: subdomain,
		TrialDays:             h.trialDays,
	})
	if err != nil {
		logger.Error("Failed to provision trial organization", zap.Error(err))
		return nil, err
	}

	logger.Info("User self-serve trial registered",
		zap.String("email", user.Email),
		zap.String("id", user.ID.String()),
		zap.String("org_subdomain", out.OrganizationSubdomain),
		zap.Time("trial_ends_at", out.TrialEndsAt),
	)

	return &Response{
		UserID:                user.ID,
		Email:                 user.Email,
		FirstName:             user.FirstName,
		LastName:              user.LastName,
		Status:                user.Status,
		Message:               "Trial started",
		OrganizationID:        out.OrganizationID,
		OrganizationSubdomain: out.OrganizationSubdomain,
		TrialEndsAt:           out.TrialEndsAt,
	}, nil
}
