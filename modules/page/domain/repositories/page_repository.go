package repositories

import (
"context"
"github.com/google/uuid"
"github.com/jcsoftdev/pulzifi-back/modules/page/domain/entities"
)

// PageRepository defines persistence operations
type PageRepository interface {
Create(ctx context.Context, page *entities.Page) error
GetByID(ctx context.Context, id uuid.UUID) (*entities.Page, error)
ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entities.Page, error)
Update(ctx context.Context, page *entities.Page) error
Delete(ctx context.Context, id uuid.UUID) error
}
