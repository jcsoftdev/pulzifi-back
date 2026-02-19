package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/entities"
)

// MockRegistrationRequestRepository is a mock implementation of
// repositories.RegistrationRequestRepository.
type MockRegistrationRequestRepository struct {
	CreateErr                    error
	GetByIDResult                *entities.RegistrationRequest
	GetByIDErr                   error
	GetByUserIDResult            *entities.RegistrationRequest
	GetByUserIDErr               error
	ListPendingResult            []*entities.RegistrationRequest
	ListPendingErr               error
	UpdateStatusErr              error
	ExistsPendingBySubdomainResult bool
	ExistsPendingBySubdomainErr  error

	CreateFn                    func(ctx context.Context, req *entities.RegistrationRequest) error
	ExistsPendingBySubdomainFn  func(ctx context.Context, subdomain string) (bool, error)

	CreateCalls int
}

func (m *MockRegistrationRequestRepository) Create(ctx context.Context, req *entities.RegistrationRequest) error {
	m.CreateCalls++
	if m.CreateFn != nil {
		return m.CreateFn(ctx, req)
	}
	return m.CreateErr
}

func (m *MockRegistrationRequestRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.RegistrationRequest, error) {
	return m.GetByIDResult, m.GetByIDErr
}

func (m *MockRegistrationRequestRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entities.RegistrationRequest, error) {
	return m.GetByUserIDResult, m.GetByUserIDErr
}

func (m *MockRegistrationRequestRepository) ListPending(ctx context.Context, limit, offset int) ([]*entities.RegistrationRequest, error) {
	return m.ListPendingResult, m.ListPendingErr
}

func (m *MockRegistrationRequestRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, reviewedBy uuid.UUID) error {
	return m.UpdateStatusErr
}

func (m *MockRegistrationRequestRepository) ExistsPendingBySubdomain(ctx context.Context, subdomain string) (bool, error) {
	if m.ExistsPendingBySubdomainFn != nil {
		return m.ExistsPendingBySubdomainFn(ctx, subdomain)
	}
	return m.ExistsPendingBySubdomainResult, m.ExistsPendingBySubdomainErr
}
