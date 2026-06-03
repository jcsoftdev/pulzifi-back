package mocks

import (
	"context"

	"github.com/google/uuid"
	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
)

// MockOrgContextLookup is a test double for authservices.OrgContextLookup.
type MockOrgContextLookup struct {
	Result    *authservices.OrgContext
	LookupErr error
	Calls     int
}

func (m *MockOrgContextLookup) Lookup(ctx context.Context, userID uuid.UUID) (*authservices.OrgContext, error) {
	m.Calls++
	return m.Result, m.LookupErr
}
