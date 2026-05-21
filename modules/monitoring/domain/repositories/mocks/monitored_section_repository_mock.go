package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
)

// MockMonitoredSectionRepository is a hand-rolled test double for repositories.MonitoredSectionRepository.
type MockMonitoredSectionRepository struct {
	CreateErr           error
	GetByIDResult       *entities.MonitoredSection
	GetByIDErr          error
	ListByPageIDResult  []*entities.MonitoredSection
	ListByPageIDErr     error
	UpdateErr           error
	DeleteErr           error
	ReplaceAllErr       error
	ReplaceAllFn        func(ctx context.Context, pageID uuid.UUID, sections []*entities.MonitoredSection) error

	CreateCalls     int
	DeleteCalls     int
	ReplaceAllCalls int
	DeleteLastID    uuid.UUID
}

func (m *MockMonitoredSectionRepository) Create(ctx context.Context, section *entities.MonitoredSection) error {
	m.CreateCalls++
	return m.CreateErr
}

func (m *MockMonitoredSectionRepository) GetByID(_ context.Context, _ uuid.UUID) (*entities.MonitoredSection, error) {
	return m.GetByIDResult, m.GetByIDErr
}

func (m *MockMonitoredSectionRepository) ListByPageID(_ context.Context, _ uuid.UUID) ([]*entities.MonitoredSection, error) {
	return m.ListByPageIDResult, m.ListByPageIDErr
}

func (m *MockMonitoredSectionRepository) Update(_ context.Context, _ *entities.MonitoredSection) error {
	return m.UpdateErr
}

func (m *MockMonitoredSectionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	m.DeleteCalls++
	m.DeleteLastID = id
	return m.DeleteErr
}

func (m *MockMonitoredSectionRepository) ReplaceAll(ctx context.Context, pageID uuid.UUID, sections []*entities.MonitoredSection) error {
	m.ReplaceAllCalls++
	if m.ReplaceAllFn != nil {
		return m.ReplaceAllFn(ctx, pageID, sections)
	}
	return m.ReplaceAllErr
}
