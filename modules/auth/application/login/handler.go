package login

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Handler authenticates a user and returns a token pair plus tenant hint.
//
// NOTE: This handler does NOT persist the refresh token. Callers are responsible
// for calling tokenService.SaveRefreshToken(ctx, userID, refreshToken) on the
// pair they actually use, so a single source-of-truth exists between cookies
// and DB. The BFF flow (shared/bff/handler.go) generates and persists its own
// pair via IssueSessionForUser; the legacy /api/v1/auth/login route persists
// the pair returned here directly.
type Handler struct {
	authService      services.AuthService
	userRepo         repositories.UserRepository
	refreshTokenRepo repositories.RefreshTokenRepository // retained for backwards-compatible NewHandler signature; unused
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
