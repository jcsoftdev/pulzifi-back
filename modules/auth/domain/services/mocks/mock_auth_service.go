package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
)

type MockAuthService struct {
	AuthenticateResult *entities.User
	AuthenticateErr    error
	ValidateErr        error
	HashResult         string
	HashErr            error
	CheckPermResult    bool
	CheckPermErr       error

	AuthenticateFn    func(ctx context.Context, email, password string) (*entities.User, error)
	ValidateFn        func(ctx context.Context, user *entities.User, password string) error
	HashPasswordFn    func(password string) (string, error)
	CheckPermissionFn func(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)

	AuthenticateCalls int
}

func (m *MockAuthService) Authenticate(ctx context.Context, email, password string) (*entities.User, error) {
	m.AuthenticateCalls++
	if m.AuthenticateFn != nil {
		return m.AuthenticateFn(ctx, email, password)
	}
	return m.AuthenticateResult, m.AuthenticateErr
}

func (m *MockAuthService) ValidateCredentials(ctx context.Context, user *entities.User, password string) error {
	if m.ValidateFn != nil {
		return m.ValidateFn(ctx, user, password)
	}
	return m.ValidateErr
}

func (m *MockAuthService) HashPassword(password string) (string, error) {
	if m.HashPasswordFn != nil {
		return m.HashPasswordFn(password)
	}
	return m.HashResult, m.HashErr
}

func (m *MockAuthService) CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	if m.CheckPermissionFn != nil {
		return m.CheckPermissionFn(ctx, userID, resource, action)
	}
	return m.CheckPermResult, m.CheckPermErr
}
