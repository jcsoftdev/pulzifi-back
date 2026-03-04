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
	CountUnread(ctx context.Context) (int, error)
	CountAll(ctx context.Context) (int, error)
	ListAll(ctx context.Context, limit int) ([]*entities.AlertWithPage, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context) error
	Delete(ctx context.Context, id uuid.UUID) error
}
