package login

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type Handler struct {
	authService  services.AuthService
	tokenService services.TokenService
	userRepo     repositories.UserRepository
}

func NewHandler(authService services.AuthService, tokenService services.TokenService, userRepo repositories.UserRepository) *Handler {
	return &Handler{
		authService:  authService,
		tokenService: tokenService,
		userRepo:     userRepo,
	}
}

func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	user, err := h.authService.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		logger.Warn("Authentication failed", zap.String("email", req.Email), zap.Error(err))
		return nil, err
	}

	accessToken, err := h.tokenService.GenerateAccessToken(ctx, user.ID, user.Email)
	if err != nil {
		logger.Error("Failed to generate access token", zap.Error(err))
		return nil, err
	}

	refreshToken, err := h.tokenService.GenerateRefreshToken(ctx, user.ID)
	if err != nil {
		logger.Error("Failed to generate refresh token", zap.Error(err))
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
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(h.tokenService.GetTokenExpiration().Seconds()),
		Tenant:       tenant,
	}, nil
}
