package mocks

import (
	"context"

	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
)

// MockRegistrationRequestWriter is a test double for authservices.RegistrationRequestWriter.
type MockRegistrationRequestWriter struct {
	CreateErr                      error
	ExistsPendingBySubdomainResult bool
	ExistsPendingBySubdomainErr    error

	CreateCalls int
}

func (m *MockRegistrationRequestWriter) Create(ctx context.Context, req *authservices.PendingRegistration) error {
	m.CreateCalls++
	return m.CreateErr
}

func (m *MockRegistrationRequestWriter) ExistsPendingBySubdomain(ctx context.Context, subdomain string) (bool, error) {
	return m.ExistsPendingBySubdomainResult, m.ExistsPendingBySubdomainErr
}
