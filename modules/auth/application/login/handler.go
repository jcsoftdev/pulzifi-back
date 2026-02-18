package login

import (
	"context"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type Handler struct {
	authService services.AuthService
	userRepo    repositories.UserRepository
	sessionRepo repositories.SessionRepository
	sessionTTL  time.Duration
}

func NewHandler(
	authService services.AuthService,
	userRepo repositories.UserRepository,
	sessionRepo repositories.SessionRepository,
	sessionTTL time.Duration,
) *Handler {
	return &Handler{
		authService: authService,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		sessionTTL:  sessionTTL,
	}
}

func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	user, err := h.authService.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		logger.Warn("Authentication failed", zap.String("email", req.Email), zap.Error(err))
		return nil, err
	}

	session := entities.NewSession(user.ID, h.sessionTTL)
	if err := h.sessionRepo.Create(ctx, session); err != nil {
		logger.Error("Failed to create session", zap.Error(err))
		return nil, err
	}

	logger.Info("User logged in successfully", zap.String("email", user.Email), zap.String("id", user.ID.String()))

	// Get user's first organization
	tenant, err := h.userRepo.GetUserFirstOrganization(ctx, user.ID)
	if err != nil {
		logger.Error("Failed to get user first organization", zap.Error(err))
		// Don't fail login if we can't get organization, just log the error
	}

	return &Response{
		SessionID: session.ID,
		ExpiresIn: int64(h.sessionTTL.Seconds()),
		Tenant:    tenant,
	}, nil
}
