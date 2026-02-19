package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
)

// MockUserRepository is a mock implementation of repositories.UserRepository.
// Each method delegates to a configurable function hook. If the hook is nil the
// method returns the corresponding default field values (e.g. CreateErr).
type MockUserRepository struct {
	// Default return values
	CreateErr                    error
	GetByIDUser                  *entities.User
	GetByIDErr                   error
	GetByEmailUser               *entities.User
	GetByEmailErr                error
	UpdateErr                    error
	DeleteErr                    error
	ExistsByEmailResult          bool
	ExistsByEmailErr             error
	GetUserFirstOrganizationResult *string
	GetUserFirstOrganizationErr  error
	UpdateStatusErr              error
	ListByStatusResult           []*entities.User
	ListByStatusErr              error

	// Function hooks – when set they take precedence over the default fields.
	CreateFn                    func(ctx context.Context, user *entities.User) error
	GetByIDFn                   func(ctx context.Context, id uuid.UUID) (*entities.User, error)
	GetByEmailFn                func(ctx context.Context, email string) (*entities.User, error)
	UpdateFn                    func(ctx context.Context, user *entities.User) error
	DeleteFn                    func(ctx context.Context, id uuid.UUID) error
	ExistsByEmailFn             func(ctx context.Context, email string) (bool, error)
	GetUserFirstOrganizationFn  func(ctx context.Context, userID uuid.UUID) (*string, error)
	UpdateStatusFn              func(ctx context.Context, id uuid.UUID, status string) error
	ListByStatusFn              func(ctx context.Context, status string, limit, offset int) ([]*entities.User, error)

	// Call tracking
	CreateCalls    int
	GetByIDCalls   int
	GetByEmailCalls int
}

func (m *MockUserRepository) Create(ctx context.Context, user *entities.User) error {
	m.CreateCalls++
	if m.CreateFn != nil {
		return m.CreateFn(ctx, user)
	}
	return m.CreateErr
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	m.GetByIDCalls++
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return m.GetByIDUser, m.GetByIDErr
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	m.GetByEmailCalls++
	if m.GetByEmailFn != nil {
		return m.GetByEmailFn(ctx, email)
	}
	return m.GetByEmailUser, m.GetByEmailErr
}

func (m *MockUserRepository) Update(ctx context.Context, user *entities.User) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, user)
	}
	return m.UpdateErr
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return m.DeleteErr
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if m.ExistsByEmailFn != nil {
		return m.ExistsByEmailFn(ctx, email)
	}
	return m.ExistsByEmailResult, m.ExistsByEmailErr
}

func (m *MockUserRepository) GetUserFirstOrganization(ctx context.Context, userID uuid.UUID) (*string, error) {
	if m.GetUserFirstOrganizationFn != nil {
		return m.GetUserFirstOrganizationFn(ctx, userID)
	}
	return m.GetUserFirstOrganizationResult, m.GetUserFirstOrganizationErr
}

func (m *MockUserRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	if m.UpdateStatusFn != nil {
		return m.UpdateStatusFn(ctx, id, status)
	}
	return m.UpdateStatusErr
}

func (m *MockUserRepository) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.User, error) {
	if m.ListByStatusFn != nil {
		return m.ListByStatusFn(ctx, status, limit, offset)
	}
	return m.ListByStatusResult, m.ListByStatusErr
}
