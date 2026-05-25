package mocks

import (
	"context"

	"github.com/google/uuid"
)

// MockOrganizationMembershipChecker is a test double for
// services.OrganizationMembershipChecker used by unit tests.
type MockOrganizationMembershipChecker struct {
	HasAnyMembershipResult bool
	HasAnyMembershipErr    error
	HasAnyMembershipCalls  int
	LastUserID             uuid.UUID
}

// HasAnyMembership records the call and returns the configured fake result.
func (m *MockOrganizationMembershipChecker) HasAnyMembership(ctx context.Context, userID uuid.UUID) (bool, error) {
	m.HasAnyMembershipCalls++
	m.LastUserID = userID
	return m.HasAnyMembershipResult, m.HasAnyMembershipErr
}
