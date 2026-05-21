package mocks

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/dashboard/domain/entities"
)

// MockDashboardRepository is a hand-rolled mock for repositories.DashboardRepository.
type MockDashboardRepository struct {
	GetStatsResult *entities.DashboardStats
	GetStatsErr    error

	GetStatsFn func(ctx context.Context) (*entities.DashboardStats, error)

	GetStatsCalls int
}

func (m *MockDashboardRepository) GetStats(ctx context.Context) (*entities.DashboardStats, error) {
	m.GetStatsCalls++
	if m.GetStatsFn != nil {
		return m.GetStatsFn(ctx)
	}
	return m.GetStatsResult, m.GetStatsErr
}
