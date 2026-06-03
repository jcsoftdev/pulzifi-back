package mocks

import (
	"context"

	"github.com/google/uuid"
	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
)

// MockOrganizationOnboardingWriter is a test double for
// authservices.OrganizationOnboardingWriter.
type MockOrganizationOnboardingWriter struct {
	SaveErr    error
	SaveCalls  int
	LastInput  authservices.OnboardingProfileInput
}

func (m *MockOrganizationOnboardingWriter) SaveOnboardingProfile(ctx context.Context, input authservices.OnboardingProfileInput) error {
	m.SaveCalls++
	m.LastInput = input
	return m.SaveErr
}

// MockOrganizationOrgFinder is a test double for authservices.OrganizationOrgFinder.
type MockOrganizationOrgFinder struct {
	OrgID     *uuid.UUID
	FindErr   error
	FindCalls int
}

func (m *MockOrganizationOrgFinder) GetUserOrgID(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error) {
	m.FindCalls++
	return m.OrgID, m.FindErr
}
