package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
)

type PermissionRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Permission, error)
	GetByName(ctx context.Context, name string) (*entities.Permission, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*entities.Permission, error)
	HasPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)
}
