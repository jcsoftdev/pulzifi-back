package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
)

// MockPasswordResetRepository is a hand-rolled mock of repositories.PasswordResetRepository.
type MockPasswordResetRepository struct {
	StoreErr   error
	FindResult *repositories.PasswordReset
	FindErr    error
	ConsumeErr error
}

func (m *MockPasswordResetRepository) Store(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ string, _ time.Time) error {
	return m.StoreErr
}

func (m *MockPasswordResetRepository) Find(_ context.Context, _ string) (*repositories.PasswordReset, error) {
	return m.FindResult, m.FindErr
}

func (m *MockPasswordResetRepository) Consume(_ context.Context, _ string, _ uuid.UUID, _ string) error {
	return m.ConsumeErr
}
