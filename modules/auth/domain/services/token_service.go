package services

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TokenClaims struct {
	UserID      uuid.UUID
	Email       string
	Roles       []string
	Permissions []string
}

type TokenService interface {
	GenerateAccessToken(ctx context.Context, userID uuid.UUID, email string) (string, error)
	GenerateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error)
	// GenerateTokenPairForUser issues an access+refresh token pair for the given user
	// and returns the access token's lifetime in seconds. It looks up the user's email
	// internally so callers (login, invitation accept, etc.) only need the userID.
	GenerateTokenPairForUser(ctx context.Context, userID uuid.UUID) (accessToken string, refreshToken string, expiresIn int64, err error)
	// SaveRefreshToken persists a refresh token string for the given user. Convenience
	// method so callers (e.g. the BFF session helper) don't need a direct dependency on
	// the RefreshTokenRepository or RefreshToken entity construction.
	SaveRefreshToken(ctx context.Context, userID uuid.UUID, refreshToken string) error
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
	GetTokenExpiration() time.Duration
	GetRefreshTokenExpiration() time.Duration
}
