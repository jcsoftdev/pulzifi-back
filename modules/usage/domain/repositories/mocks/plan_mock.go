package mocks

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/entities"
)

// MockPlanRepository is a hand-rolled mock of repositories.PlanRepository.
type MockPlanRepository struct {
	ListActiveResult []*entities.Plan
	ListActiveErr    error
	GetByCodeResult  *entities.Plan
	GetByCodeErr     error
}

func (m *MockPlanRepository) ListActive(_ context.Context) ([]*entities.Plan, error) {
	return m.ListActiveResult, m.ListActiveErr
}

func (m *MockPlanRepository) GetByCode(_ context.Context, _ string) (*entities.Plan, error) {
	return m.GetByCodeResult, m.GetByCodeErr
}
