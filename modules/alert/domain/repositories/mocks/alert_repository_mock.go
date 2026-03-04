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
	CountUnreadResult  int
	CountUnreadErr     error
	CountAllResult     int
	CountAllErr        error
	ListAllResult      []*entities.AlertWithPage
	ListAllErr         error
	MarkAsReadErr      error
	MarkAllAsReadErr   error
	DeleteErr          error

	CreateFn func(ctx context.Context, alert *entities.Alert) error

	CreateCalls       int
	CountUnreadCalls  int
	CountAllCalls     int
	ListAllCalls      int
	MarkAllAsReadCalls int
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

func (m *MockAlertRepository) CountUnread(_ context.Context) (int, error) {
	m.CountUnreadCalls++
	return m.CountUnreadResult, m.CountUnreadErr
}

func (m *MockAlertRepository) CountAll(_ context.Context) (int, error) {
	m.CountAllCalls++
	return m.CountAllResult, m.CountAllErr
}

func (m *MockAlertRepository) ListAll(_ context.Context, _ int) ([]*entities.AlertWithPage, error) {
	m.ListAllCalls++
	return m.ListAllResult, m.ListAllErr
}

func (m *MockAlertRepository) MarkAsRead(_ context.Context, _ uuid.UUID) error {
	return m.MarkAsReadErr
}

func (m *MockAlertRepository) MarkAllAsRead(_ context.Context) error {
	m.MarkAllAsReadCalls++
	return m.MarkAllAsReadErr
}

func (m *MockAlertRepository) Delete(_ context.Context, _ uuid.UUID) error {
	return m.DeleteErr
}
