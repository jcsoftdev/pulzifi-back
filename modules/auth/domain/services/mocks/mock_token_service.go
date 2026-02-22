package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
)

type MockTokenService struct {
	GenerateAccessTokenResult  string
	GenerateAccessTokenErr     error
	GenerateRefreshTokenResult string
	GenerateRefreshTokenErr    error
	ValidateTokenResult        *services.TokenClaims
	ValidateTokenErr           error
	TokenExpiration            time.Duration
	RefreshTokenExpiration     time.Time

	GenerateAccessTokenFn  func(ctx context.Context, userID uuid.UUID, email string) (string, error)
	GenerateRefreshTokenFn func(ctx context.Context, userID uuid.UUID) (string, error)

	GenerateAccessTokenCalls  int
	GenerateRefreshTokenCalls int
}

func (m *MockTokenService) GenerateAccessToken(ctx context.Context, userID uuid.UUID, email string) (string, error) {
	m.GenerateAccessTokenCalls++
	if m.GenerateAccessTokenFn != nil {
		return m.GenerateAccessTokenFn(ctx, userID, email)
	}
	return m.GenerateAccessTokenResult, m.GenerateAccessTokenErr
}

func (m *MockTokenService) GenerateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	m.GenerateRefreshTokenCalls++
	if m.GenerateRefreshTokenFn != nil {
		return m.GenerateRefreshTokenFn(ctx, userID)
	}
	return m.GenerateRefreshTokenResult, m.GenerateRefreshTokenErr
}

func (m *MockTokenService) ValidateToken(_ context.Context, _ string) (*services.TokenClaims, error) {
	return m.ValidateTokenResult, m.ValidateTokenErr
}

func (m *MockTokenService) GetTokenExpiration() time.Duration {
	if m.TokenExpiration == 0 {
		return 15 * time.Minute
	}
	return m.TokenExpiration
}

func (m *MockTokenService) GetRefreshTokenExpiration() time.Time {
	if m.RefreshTokenExpiration.IsZero() {
		return time.Now().Add(7 * 24 * time.Hour)
	}
	return m.RefreshTokenExpiration
}
