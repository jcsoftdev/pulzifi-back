package refresh_token

import (
	"context"
	"errors"

	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

var (
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrExpiredRefreshToken = errors.New("refresh token expired")
	ErrRevokedRefreshToken = errors.New("refresh token revoked")
)

type Handler struct {
	refreshTokenRepo repositories.RefreshTokenRepository
	userRepo         repositories.UserRepository
	tokenService     services.TokenService
}

func NewHandler(
	refreshTokenRepo repositories.RefreshTokenRepository,
	userRepo repositories.UserRepository,
	tokenService services.TokenService,
) *Handler {
	return &Handler{
		refreshTokenRepo: refreshTokenRepo,
		userRepo:         userRepo,
		tokenService:     tokenService,
	}
}

func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	// Find refresh token in database
	refreshToken, err := h.refreshTokenRepo.FindByToken(ctx, req.RefreshToken)
	if err != nil {
		logger.Warn("Refresh token not found", zap.Error(err))
		return nil, ErrInvalidRefreshToken
	}

	// Validate refresh token
	if refreshToken.IsRevoked {
		logger.Warn("Refresh token is revoked", zap.String("user_id", refreshToken.UserID.String()))
		return nil, ErrRevokedRefreshToken
	}

	if refreshToken.IsExpired() {
		logger.Warn("Refresh token is expired", zap.String("user_id", refreshToken.UserID.String()))
		return nil, ErrExpiredRefreshToken
	}

	// Get user information
	user, err := h.userRepo.GetByID(ctx, refreshToken.UserID)
	if err != nil {
		logger.Error("Failed to find user", zap.Error(err))
		return nil, errors.New("user not found")
	}

	// Generate new access token
	newAccessToken, err := h.tokenService.GenerateAccessToken(ctx, user.ID, user.Email)
	if err != nil {
		logger.Error("Failed to generate access token", zap.Error(err))
		return nil, err
	}

	// Generate new refresh token (token rotation)
	newRefreshTokenStr, err := h.tokenService.GenerateRefreshToken(ctx, user.ID)
	if err != nil {
		logger.Error("Failed to generate new refresh token", zap.Error(err))
		return nil, err
	}

	// Revoke old refresh token
	if err := h.refreshTokenRepo.Revoke(ctx, req.RefreshToken); err != nil {
		logger.Error("Failed to revoke old refresh token", zap.Error(err))
		// Continue anyway, as we've generated new tokens
	}

	// Store new refresh token
	newRefreshToken := entities.NewRefreshToken(
		user.ID,
		newRefreshTokenStr,
		h.tokenService.GetRefreshTokenExpiration(),
	)

	if err := h.refreshTokenRepo.Create(ctx, newRefreshToken); err != nil {
		logger.Error("Failed to store new refresh token", zap.Error(err))
		return nil, err
	}

	logger.Info("Token refreshed successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email),
	)

	return &Response{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshTokenStr,
		TokenType:    "Bearer",
		ExpiresIn:    int64(h.tokenService.GetTokenExpiration().Seconds()),
	}, nil
}
