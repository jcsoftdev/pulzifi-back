package mocks

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
)

// compile-time interface check
var _ services.PlanAssigner = (*MockPlanAssigner)(nil)

// MockPlanAssigner is a hand-rolled test double for services.PlanAssigner.
type MockPlanAssigner struct {
	AssignErr error
	AssignFn  func(ctx context.Context, in services.AssignInput) error

	// Call tracking
	AssignCalls  int
	LastAssignIn services.AssignInput
}

func (m *MockPlanAssigner) Assign(ctx context.Context, in services.AssignInput) error {
	m.AssignCalls++
	m.LastAssignIn = in
	if m.AssignFn != nil {
		return m.AssignFn(ctx, in)
	}
	return m.AssignErr
}
