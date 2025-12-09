package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
)

// NotificationPreferenceRepository defines operations for managing notification preferences
type NotificationPreferenceRepository interface {
	Create(ctx context.Context, pref *entities.NotificationPreference) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.NotificationPreference, error)
	GetByUserAndWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) (*entities.NotificationPreference, error)
	GetByUserAndPage(ctx context.Context, userID, pageID uuid.UUID) (*entities.NotificationPreference, error)
	Update(ctx context.Context, pref *entities.NotificationPreference) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}
