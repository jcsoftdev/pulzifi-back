package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
)

type RefreshTokenRepository interface {
	// Create stores a new refresh token
	Create(ctx context.Context, refreshToken *entities.RefreshToken) error

	// FindByToken retrieves a refresh token by its token string
	FindByToken(ctx context.Context, token string) (*entities.RefreshToken, error)

	// FindByUserID retrieves all refresh tokens for a user
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.RefreshToken, error)

	// Revoke marks a refresh token as revoked
	Revoke(ctx context.Context, token string) error

	// RevokeAllByUserID revokes all refresh tokens for a user
	RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error

	// DeleteExpired removes expired refresh tokens
	DeleteExpired(ctx context.Context) error
}
