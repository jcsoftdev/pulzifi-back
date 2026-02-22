package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
)

type MockNotificationPreferenceRepository struct {
	CreateErr                error
	GetByIDResult            *entities.NotificationPreference
	GetByIDErr               error
	GetByUserAndWorkspaceRes *entities.NotificationPreference
	GetByUserAndWorkspaceErr error
	GetByUserAndPageResult   *entities.NotificationPreference
	GetByUserAndPageErr      error
	UpdateErr                error
	DeleteByIDErr            error
	GetEmailEnabledResult    []*entities.NotificationPreference
	GetEmailEnabledErr       error

	CreateCalls int
}

func (m *MockNotificationPreferenceRepository) Create(_ context.Context, _ *entities.NotificationPreference) error {
	m.CreateCalls++
	return m.CreateErr
}

func (m *MockNotificationPreferenceRepository) GetByID(_ context.Context, _ uuid.UUID) (*entities.NotificationPreference, error) {
	return m.GetByIDResult, m.GetByIDErr
}

func (m *MockNotificationPreferenceRepository) GetByUserAndWorkspace(_ context.Context, _, _ uuid.UUID) (*entities.NotificationPreference, error) {
	return m.GetByUserAndWorkspaceRes, m.GetByUserAndWorkspaceErr
}

func (m *MockNotificationPreferenceRepository) GetByUserAndPage(_ context.Context, _, _ uuid.UUID) (*entities.NotificationPreference, error) {
	return m.GetByUserAndPageResult, m.GetByUserAndPageErr
}

func (m *MockNotificationPreferenceRepository) Update(_ context.Context, _ *entities.NotificationPreference) error {
	return m.UpdateErr
}

func (m *MockNotificationPreferenceRepository) DeleteByID(_ context.Context, _ uuid.UUID) error {
	return m.DeleteByIDErr
}

func (m *MockNotificationPreferenceRepository) GetEmailEnabledByPage(_ context.Context, _ uuid.UUID) ([]*entities.NotificationPreference, error) {
	return m.GetEmailEnabledResult, m.GetEmailEnabledErr
}
