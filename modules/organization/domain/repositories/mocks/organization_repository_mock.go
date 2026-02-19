package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/entities"
)

// MockOrganizationRepository is a mock implementation of
// repositories.OrganizationRepository.
type MockOrganizationRepository struct {
	CreateErr            error
	GetByIDResult        *entities.Organization
	GetByIDErr           error
	GetBySubdomainResult *entities.Organization
	GetBySubdomainErr    error
	ListResult           []*entities.Organization
	ListErr              error
	UpdateErr            error
	DeleteErr            error
	CountBySubdomainResult int
	CountBySubdomainErr  error

	CountBySubdomainFn func(ctx context.Context, subdomain string) (int, error)
}

func (m *MockOrganizationRepository) Create(ctx context.Context, organization *entities.Organization) error {
	return m.CreateErr
}

func (m *MockOrganizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Organization, error) {
	return m.GetByIDResult, m.GetByIDErr
}

func (m *MockOrganizationRepository) GetBySubdomain(ctx context.Context, subdomain string) (*entities.Organization, error) {
	return m.GetBySubdomainResult, m.GetBySubdomainErr
}

func (m *MockOrganizationRepository) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Organization, error) {
	return m.ListResult, m.ListErr
}

func (m *MockOrganizationRepository) Update(ctx context.Context, organization *entities.Organization) error {
	return m.UpdateErr
}

func (m *MockOrganizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.DeleteErr
}

func (m *MockOrganizationRepository) CountBySubdomain(ctx context.Context, subdomain string) (int, error) {
	if m.CountBySubdomainFn != nil {
		return m.CountBySubdomainFn(ctx, subdomain)
	}
	return m.CountBySubdomainResult, m.CountBySubdomainErr
}
