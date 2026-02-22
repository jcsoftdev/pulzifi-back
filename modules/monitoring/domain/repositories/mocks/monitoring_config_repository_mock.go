package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
)

type MockMonitoringConfigRepository struct {
	CreateErr            error
	GetByPageIDResult    *entities.MonitoringConfig
	GetByPageIDErr       error
	UpdateErr            error
	GetDueTasksResult    []entities.SnapshotTask
	GetDueTasksErr       error
	GetPageURLResult     string
	GetPageURLErr        error
	UpdateLastCheckedErr error
	MarkPageDueNowErr   error

	CreateFn func(ctx context.Context, config *entities.MonitoringConfig) error

	CreateCalls          int
	UpdateCalls          int
	MarkPageDueNowCalls  int
}

func (m *MockMonitoringConfigRepository) Create(ctx context.Context, config *entities.MonitoringConfig) error {
	m.CreateCalls++
	if m.CreateFn != nil {
		return m.CreateFn(ctx, config)
	}
	return m.CreateErr
}

func (m *MockMonitoringConfigRepository) GetByPageID(_ context.Context, _ uuid.UUID) (*entities.MonitoringConfig, error) {
	return m.GetByPageIDResult, m.GetByPageIDErr
}

func (m *MockMonitoringConfigRepository) Update(_ context.Context, _ *entities.MonitoringConfig) error {
	m.UpdateCalls++
	return m.UpdateErr
}

func (m *MockMonitoringConfigRepository) GetDueSnapshotTasks(_ context.Context) ([]entities.SnapshotTask, error) {
	return m.GetDueTasksResult, m.GetDueTasksErr
}

func (m *MockMonitoringConfigRepository) GetPageURL(_ context.Context, _ uuid.UUID) (string, error) {
	return m.GetPageURLResult, m.GetPageURLErr
}

func (m *MockMonitoringConfigRepository) UpdateLastCheckedAt(_ context.Context, _ uuid.UUID) error {
	return m.UpdateLastCheckedErr
}

func (m *MockMonitoringConfigRepository) MarkPageDueNow(_ context.Context, _ uuid.UUID) error {
	m.MarkPageDueNowCalls++
	return m.MarkPageDueNowErr
}
