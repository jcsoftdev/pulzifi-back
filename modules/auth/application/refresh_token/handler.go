package refresh_token

import (
	"context"
	"errors"

	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/cache"
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
	// Check if we have a cached response for this token (grace period for concurrent requests)
	if cachedResponse, err := cache.GetRefreshTokenCache(ctx, req.RefreshToken); err == nil {
		logger.InfoWithContext(ctx, "Returning cached refresh token response",
			zap.String("reason", "concurrent_request_within_grace_period"),
		)
		return &Response{
			AccessToken:  cachedResponse.AccessToken,
			RefreshToken: cachedResponse.RefreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    cachedResponse.ExpiresIn,
			Tenant:       cachedResponse.Tenant,
		}, nil
	}

	// Find refresh token in database
	refreshToken, err := h.refreshTokenRepo.FindByToken(ctx, req.RefreshToken)
	if err != nil {
		logger.WarnWithContext(ctx, "Refresh token not found", zap.Error(err), zap.String("reason", "token_not_in_database"))
		return nil, ErrInvalidRefreshToken
	}

	// Validate refresh token
	if refreshToken.IsRevoked {
		logger.WarnWithContext(ctx, "Refresh token is revoked",
			zap.String("user_id", refreshToken.UserID.String()),
			zap.String("reason", "token_revoked"),
		)
		return nil, ErrRevokedRefreshToken
	}

	if refreshToken.IsExpired() {
		logger.WarnWithContext(ctx, "Refresh token is expired",
			zap.String("user_id", refreshToken.UserID.String()),
			zap.String("reason", "token_expired"),
		)
		return nil, ErrExpiredRefreshToken
	}

	// Get user information
	user, err := h.userRepo.GetByID(ctx, refreshToken.UserID)
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to find user",
			zap.Error(err),
			zap.String("user_id", refreshToken.UserID.String()),
		)
		return nil, errors.New("user not found")
	}

	// Generate new access token
	newAccessToken, err := h.tokenService.GenerateAccessToken(ctx, user.ID, user.Email)
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to generate access token",
			zap.Error(err),
			zap.String("user_id", user.ID.String()),
		)
		return nil, err
	}

	// Generate new refresh token (token rotation)
	newRefreshTokenStr, err := h.tokenService.GenerateRefreshToken(ctx, user.ID)
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to generate new refresh token",
			zap.Error(err),
			zap.String("user_id", user.ID.String()),
		)
		return nil, err
	}

	// Revoke old refresh token
	if err := h.refreshTokenRepo.Revoke(ctx, req.RefreshToken); err != nil {
		logger.ErrorWithContext(ctx, "Failed to revoke old refresh token",
			zap.Error(err),
			zap.String("user_id", user.ID.String()),
		)
		// Continue anyway, as we've generated new tokens
	}

	// Store new refresh token
	newRefreshToken := entities.NewRefreshToken(
		user.ID,
		newRefreshTokenStr,
		h.tokenService.GetRefreshTokenExpiration(),
	)

	if err := h.refreshTokenRepo.Create(ctx, newRefreshToken); err != nil {
		logger.ErrorWithContext(ctx, "Failed to store new refresh token",
			zap.Error(err),
			zap.String("user_id", user.ID.String()),
		)
		return nil, err
	}

	logger.InfoWithContext(ctx, "Token refreshed successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email),
	)

	// Get user's first organization (tenant)
	tenant, err := h.userRepo.GetUserFirstOrganization(ctx, user.ID)
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to get user first organization", zap.Error(err))
		// Don't fail refresh if we can't get organization
	}

	response := &Response{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshTokenStr,
		TokenType:    "Bearer",
		ExpiresIn:    int64(h.tokenService.GetTokenExpiration().Seconds()),
		Tenant:       "",
	}

	if tenant != nil {
		response.Tenant = *tenant
	}

	// Cache the response for grace period to handle concurrent requests
	cacheData := &cache.RefreshTokenCache{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		ExpiresIn:    response.ExpiresIn,
		Tenant:       response.Tenant,
	}
	if err := cache.SetRefreshTokenCache(ctx, req.RefreshToken, cacheData); err != nil {
		logger.WarnWithContext(ctx, "Failed to cache refresh token response",
			zap.Error(err),
			zap.String("user_id", user.ID.String()),
		)
		// Don't fail the request if caching fails
	}

	return response, nil
}
