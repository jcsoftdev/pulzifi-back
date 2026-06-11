package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
)

// MockAlertRepository is a hand-rolled mock for repositories.AlertRepository.
type MockAlertRepository struct {
	SaveErr           error
	ListResult        []*entities.SocialAlert
	ListErr           error
	MarkReadErr       error
	UnreadCountResult int
	UnreadCountErr    error

	SaveFn        func(ctx context.Context, alert *entities.SocialAlert) error
	ListFn        func(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]*entities.SocialAlert, error)
	MarkReadFn    func(ctx context.Context, alertID uuid.UUID) error
	UnreadCountFn func(ctx context.Context, workspaceID uuid.UUID) (int, error)

	SaveCalls        int
	ListCalls        int
	MarkReadCalls    int
	UnreadCountCalls int
	SavedAlerts      []*entities.SocialAlert
}

func (m *MockAlertRepository) Save(ctx context.Context, alert *entities.SocialAlert) error {
	m.SaveCalls++
	m.SavedAlerts = append(m.SavedAlerts, alert)
	if m.SaveFn != nil {
		return m.SaveFn(ctx, alert)
	}
	return m.SaveErr
}

func (m *MockAlertRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]*entities.SocialAlert, error) {
	m.ListCalls++
	if m.ListFn != nil {
		return m.ListFn(ctx, workspaceID, limit, offset)
	}
	return m.ListResult, m.ListErr
}

func (m *MockAlertRepository) MarkRead(ctx context.Context, alertID uuid.UUID) error {
	m.MarkReadCalls++
	if m.MarkReadFn != nil {
		return m.MarkReadFn(ctx, alertID)
	}
	return m.MarkReadErr
}

func (m *MockAlertRepository) UnreadCount(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	m.UnreadCountCalls++
	if m.UnreadCountFn != nil {
		return m.UnreadCountFn(ctx, workspaceID)
	}
	return m.UnreadCountResult, m.UnreadCountErr
}
