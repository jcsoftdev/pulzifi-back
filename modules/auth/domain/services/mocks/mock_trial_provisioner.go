package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
)

// MockTrialProvisioner is a mock of services.TrialProvisioner used by tests.
type MockTrialProvisioner struct {
	ProvisionResult *authservices.TrialProvisionOutput
	ProvisionErr    error
	ProvisionCalls  int
	LastInput       authservices.TrialProvisionInput
}

// Provision records the call and returns the configured fake result. When no
// ProvisionResult is set, a deterministic default is produced so callers can
// reason about the response shape in success-path tests.
func (m *MockTrialProvisioner) Provision(ctx context.Context, in authservices.TrialProvisionInput) (*authservices.TrialProvisionOutput, error) {
	m.ProvisionCalls++
	m.LastInput = in
	if m.ProvisionErr != nil {
		return nil, m.ProvisionErr
	}
	if m.ProvisionResult != nil {
		return m.ProvisionResult, nil
	}
	days := in.TrialDays
	if days <= 0 {
		days = 14
	}
	return &authservices.TrialProvisionOutput{
		OrganizationID:        uuid.New(),
		OrganizationSubdomain: in.OrganizationSubdomain,
		SchemaName:            "tenant_" + in.OrganizationSubdomain,
		TrialEndsAt:           time.Now().Add(time.Duration(days) * 24 * time.Hour),
	}, nil
}
