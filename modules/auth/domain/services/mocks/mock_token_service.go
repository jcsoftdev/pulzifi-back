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
	RefreshTokenExpiration     time.Duration
	SaveRefreshTokenErr        error

	GenerateAccessTokenFn      func(ctx context.Context, userID uuid.UUID, email string) (string, error)
	GenerateRefreshTokenFn     func(ctx context.Context, userID uuid.UUID) (string, error)
	GenerateTokenPairForUserFn func(ctx context.Context, userID uuid.UUID) (string, string, int64, error)
	SaveRefreshTokenFn         func(ctx context.Context, userID uuid.UUID, refreshToken string) error

	GenerateAccessTokenCalls      int
	GenerateRefreshTokenCalls     int
	GenerateTokenPairForUserCalls int
	SaveRefreshTokenCalls         int
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

func (m *MockTokenService) GenerateTokenPairForUser(ctx context.Context, userID uuid.UUID) (string, string, int64, error) {
	m.GenerateTokenPairForUserCalls++
	if m.GenerateTokenPairForUserFn != nil {
		return m.GenerateTokenPairForUserFn(ctx, userID)
	}
	if m.GenerateAccessTokenErr != nil {
		return "", "", 0, m.GenerateAccessTokenErr
	}
	if m.GenerateRefreshTokenErr != nil {
		return "", "", 0, m.GenerateRefreshTokenErr
	}
	exp := m.TokenExpiration
	if exp == 0 {
		exp = 15 * time.Minute
	}
	return m.GenerateAccessTokenResult, m.GenerateRefreshTokenResult, int64(exp.Seconds()), nil
}

func (m *MockTokenService) SaveRefreshToken(ctx context.Context, userID uuid.UUID, refreshToken string) error {
	m.SaveRefreshTokenCalls++
	if m.SaveRefreshTokenFn != nil {
		return m.SaveRefreshTokenFn(ctx, userID, refreshToken)
	}
	return m.SaveRefreshTokenErr
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

func (m *MockTokenService) GetRefreshTokenExpiration() time.Duration {
	if m.RefreshTokenExpiration == 0 {
		return 7 * 24 * time.Hour
	}
	return m.RefreshTokenExpiration
}
