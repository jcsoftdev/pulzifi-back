package mocks

import (
	"context"
)

// MockOrganizationDirectory is a test double for authservices.OrganizationDirectory.
type MockOrganizationDirectory struct {
	ValidateOrganizationNameErr error
	ValidateSubdomainErr        error
	CountBySubdomainResult      int
	CountBySubdomainErr         error
}

func (m *MockOrganizationDirectory) ValidateOrganizationName(name string) error {
	return m.ValidateOrganizationNameErr
}

func (m *MockOrganizationDirectory) ValidateSubdomain(subdomain string) error {
	return m.ValidateSubdomainErr
}

func (m *MockOrganizationDirectory) CountBySubdomain(ctx context.Context, subdomain string) (int, error) {
	return m.CountBySubdomainResult, m.CountBySubdomainErr
}
