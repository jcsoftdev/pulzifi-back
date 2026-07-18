package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/report/domain/entities"
)

// MockReportRepository is a hand-rolled mock of repositories.ReportRepository.
type MockReportRepository struct {
	CreateErr        error
	GetByIDResult    *entities.Report
	GetByIDErr       error
	ListByPageResult []*entities.Report
	ListByPageErr    error
	ListResult       []*entities.Report
	ListErr          error
	DeleteErr        error
}

func (m *MockReportRepository) Create(_ context.Context, _ *entities.Report) error {
	return m.CreateErr
}

func (m *MockReportRepository) GetByID(_ context.Context, _ uuid.UUID) (*entities.Report, error) {
	return m.GetByIDResult, m.GetByIDErr
}

func (m *MockReportRepository) ListByPage(_ context.Context, _ uuid.UUID) ([]*entities.Report, error) {
	return m.ListByPageResult, m.ListByPageErr
}

func (m *MockReportRepository) List(_ context.Context) ([]*entities.Report, error) {
	return m.ListResult, m.ListErr
}

func (m *MockReportRepository) Delete(_ context.Context, _ uuid.UUID) error {
	return m.DeleteErr
}
