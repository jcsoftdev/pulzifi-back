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
	authService      services.AuthService
	userRepo         repositories.UserRepository
	refreshTokenRepo repositories.RefreshTokenRepository
	tokenService     services.TokenService
}

func NewHandler(
	authService services.AuthService,
	userRepo repositories.UserRepository,
	refreshTokenRepo repositories.RefreshTokenRepository,
	tokenService services.TokenService,
) *Handler {
	return &Handler{
		authService:      authService,
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		tokenService:     tokenService,
	}
}

func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	user, err := h.authService.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		logger.Warn("Authentication failed", zap.String("email", req.Email), zap.Error(err))
		return nil, err
	}

	accessToken, refreshTokenStr, expiresIn, err := h.tokenService.GenerateTokenPairForUser(ctx, user.ID)
	if err != nil {
		logger.Error("Failed to generate token pair", zap.Error(err))
		return nil, err
	}

	refreshTokenExpiry := time.Now().Add(h.tokenService.GetRefreshTokenExpiration())
	refreshToken := entities.NewRefreshToken(user.ID, refreshTokenStr, refreshTokenExpiry)
	if err := h.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		logger.Error("Failed to store refresh token", zap.Error(err))
		return nil, err
	}

	tenant, err := h.userRepo.GetUserFirstOrganization(ctx, user.ID)
	if err != nil {
		logger.Error("Failed to get user first organization", zap.Error(err))
	}

	logger.Info("User logged in successfully", zap.String("email", user.Email))

	return &Response{
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		ExpiresIn:    expiresIn,
		Tenant:       tenant,
	}, nil
}
