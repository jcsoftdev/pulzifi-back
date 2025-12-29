package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
)

type RoleRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Role, error)
	GetByName(ctx context.Context, name string) (*entities.Role, error)
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*entities.Role, error)
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*entities.Permission, error)
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error
}
