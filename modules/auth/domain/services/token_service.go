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
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
	GetTokenExpiration() time.Duration
}
