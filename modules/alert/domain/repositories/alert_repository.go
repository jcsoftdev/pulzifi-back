package repositories

import (
"context"
"github.com/google/uuid"
"github.com/jcsoftdev/pulzifi-back/modules/alert/domain/entities"
)

type AlertRepository interface {
	Create(ctx context.Context, alert *entities.Alert) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Alert, error)
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entities.Alert, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}
