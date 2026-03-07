package logout

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type Handler struct {
	refreshTokenRepo repositories.RefreshTokenRepository
}

func NewHandler(refreshTokenRepo repositories.RefreshTokenRepository) *Handler {
	return &Handler{
		refreshTokenRepo: refreshTokenRepo,
	}
}

// Handle revokes the given refresh token in the database.
// It is a best-effort operation: even if the token is not found or already
// revoked, the logout is still considered successful so that clients always
// get their cookies cleared regardless.
func (h *Handler) Handle(ctx context.Context, refreshToken string) {
	if refreshToken == "" {
		return
	}

	if err := h.refreshTokenRepo.Revoke(ctx, refreshToken); err != nil {
		// Log but do not fail the logout — the cookie will be cleared anyway.
		logger.WarnWithContext(ctx, "Failed to revoke refresh token on logout",
			zap.Error(err),
		)
	}
}
