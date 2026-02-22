package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
)

type MockRefreshTokenRepository struct {
	CreateErr          error
	FindByTokenResult  *entities.RefreshToken
	FindByTokenErr     error
	FindByUserIDResult []*entities.RefreshToken
	FindByUserIDErr    error
	RevokeErr          error
	RevokeAllErr       error
	DeleteExpiredErr   error

	CreateFn func(ctx context.Context, refreshToken *entities.RefreshToken) error

	CreateCalls int
}

func (m *MockRefreshTokenRepository) Create(ctx context.Context, refreshToken *entities.RefreshToken) error {
	m.CreateCalls++
	if m.CreateFn != nil {
		return m.CreateFn(ctx, refreshToken)
	}
	return m.CreateErr
}

func (m *MockRefreshTokenRepository) FindByToken(_ context.Context, _ string) (*entities.RefreshToken, error) {
	return m.FindByTokenResult, m.FindByTokenErr
}

func (m *MockRefreshTokenRepository) FindByUserID(_ context.Context, _ uuid.UUID) ([]*entities.RefreshToken, error) {
	return m.FindByUserIDResult, m.FindByUserIDErr
}

func (m *MockRefreshTokenRepository) Revoke(_ context.Context, _ string) error {
	return m.RevokeErr
}

func (m *MockRefreshTokenRepository) RevokeAllByUserID(_ context.Context, _ uuid.UUID) error {
	return m.RevokeAllErr
}

func (m *MockRefreshTokenRepository) DeleteExpired(_ context.Context) error {
	return m.DeleteExpiredErr
}
