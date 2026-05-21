package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
)

// Compile-time interface check.
var _ services.TrialConverter = (*MockTrialConverter)(nil)

// MockTrialConverter is a test double for services.TrialConverter.
type MockTrialConverter struct {
	ConvertErr   error
	ConvertCalls int
	LastOrgID    uuid.UUID
}

// Convert records the orgID and returns the configured fake error.
func (m *MockTrialConverter) Convert(_ context.Context, orgID uuid.UUID) error {
	m.ConvertCalls++
	m.LastOrgID = orgID
	return m.ConvertErr
}
