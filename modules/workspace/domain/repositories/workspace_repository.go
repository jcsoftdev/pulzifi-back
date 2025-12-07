package repositories

import (
"context"
"github.com/google/uuid"
"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/entities"
)

// WorkspaceRepository defines persistence operations
type WorkspaceRepository interface {
Create(ctx context.Context, workspace *entities.Workspace) error
GetByID(ctx context.Context, id uuid.UUID) (*entities.Workspace, error)
ListByCreator(ctx context.Context, createdBy uuid.UUID) ([]*entities.Workspace, error)
Update(ctx context.Context, workspace *entities.Workspace) error
Delete(ctx context.Context, id uuid.UUID) error
}
