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

// Handler handles user registration
type Handler struct {
	userRepo     repositories.UserRepository
	regReqWriter authservices.RegistrationRequestWriter
	orgDirectory authservices.OrganizationDirectory
}

// NewHandler creates a new handler instance
func NewHandler(
	userRepo repositories.UserRepository,
	regReqWriter authservices.RegistrationRequestWriter,
	orgDirectory authservices.OrganizationDirectory,
) *Handler {
	return &Handler{
		userRepo:     userRepo,
		regReqWriter: regReqWriter,
		orgDirectory: orgDirectory,
	}
}

// Handle executes the register use case
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

	// Check subdomain uniqueness against existing (approved) organizations
	count, err := h.orgDirectory.CountBySubdomain(ctx, subdomain)
	if err != nil {
		logger.Error("Failed to check subdomain uniqueness", zap.Error(err))
		return nil, err
	}
	if count > 0 {
		return nil, errors.NewUserError("SUBDOMAIN_TAKEN", "subdomain is already in use")
	}

	// Check subdomain uniqueness against pending registration requests
	pendingExists, err := h.regReqWriter.ExistsPendingBySubdomain(ctx, subdomain)
	if err != nil {
		logger.Error("Failed to check pending subdomain", zap.Error(err))
		return nil, err
	}
	if pendingExists {
		return nil, errors.NewUserError("SUBDOMAIN_PENDING", "subdomain is already pending registration approval")
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

	// Create user entity (status: pending)
	user := entities.NewUser(req.Email, string(hashedPassword), req.FirstName, req.LastName)

	// Persist user
	if err := h.userRepo.Create(ctx, user); err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		return nil, err
	}

	// Create registration request using auth's own type
	regReq := authservices.NewPendingRegistration(user.ID, req.OrganizationName, subdomain)
	if err := h.regReqWriter.Create(ctx, regReq); err != nil {
		logger.Error("Failed to create registration request", zap.Error(err))
		return nil, err
	}

	logger.Info("User registration submitted",
		zap.String("email", user.Email),
		zap.String("id", user.ID.String()),
		zap.String("org_subdomain", subdomain),
	)

	return &Response{
		UserID:    user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Status:    user.Status,
		Message:   "Registration submitted, awaiting approval",
	}, nil
}
