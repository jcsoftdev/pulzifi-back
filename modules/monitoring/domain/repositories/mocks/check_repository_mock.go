package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
)

type MockCheckRepository struct {
	CreateErr          error
	GetByIDResult      *entities.Check
	GetByIDErr         error
	ListByPageResult   []*entities.Check
	ListByPageErr      error
	GetLatestResult    *entities.Check
	GetLatestErr       error
	UpdateErr          error
	GetPreviousResult  *entities.Check
	GetPreviousErr     error

	CreateFn func(ctx context.Context, check *entities.Check) error

	CreateCalls int
}

func (m *MockCheckRepository) Create(ctx context.Context, check *entities.Check) error {
	m.CreateCalls++
	if m.CreateFn != nil {
		return m.CreateFn(ctx, check)
	}
	return m.CreateErr
}

func (m *MockCheckRepository) GetByID(_ context.Context, _ uuid.UUID) (*entities.Check, error) {
	return m.GetByIDResult, m.GetByIDErr
}

func (m *MockCheckRepository) ListByPage(_ context.Context, _ uuid.UUID) ([]*entities.Check, error) {
	return m.ListByPageResult, m.ListByPageErr
}

func (m *MockCheckRepository) GetLatestByPage(_ context.Context, _ uuid.UUID) (*entities.Check, error) {
	return m.GetLatestResult, m.GetLatestErr
}

func (m *MockCheckRepository) Update(_ context.Context, _ *entities.Check) error {
	return m.UpdateErr
}

func (m *MockCheckRepository) GetPreviousSuccessfulByPage(_ context.Context, _, _ uuid.UUID) (*entities.Check, error) {
	return m.GetPreviousResult, m.GetPreviousErr
}
