package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/alert/domain/entities"
)

type MockAlertRepository struct {
	CreateErr          error
	GetByIDResult      *entities.Alert
	GetByIDErr         error
	ListByWorkResult   []*entities.Alert
	ListByWorkErr      error
	MarkAsReadErr      error
	DeleteErr          error

	CreateFn func(ctx context.Context, alert *entities.Alert) error

	CreateCalls int
}

func (m *MockAlertRepository) Create(ctx context.Context, alert *entities.Alert) error {
	m.CreateCalls++
	if m.CreateFn != nil {
		return m.CreateFn(ctx, alert)
	}
	return m.CreateErr
}

func (m *MockAlertRepository) GetByID(_ context.Context, _ uuid.UUID) (*entities.Alert, error) {
	return m.GetByIDResult, m.GetByIDErr
}

func (m *MockAlertRepository) ListByWorkspace(_ context.Context, _ uuid.UUID) ([]*entities.Alert, error) {
	return m.ListByWorkResult, m.ListByWorkErr
}

func (m *MockAlertRepository) MarkAsRead(_ context.Context, _ uuid.UUID) error {
	return m.MarkAsReadErr
}

func (m *MockAlertRepository) Delete(_ context.Context, _ uuid.UUID) error {
	return m.DeleteErr
}
