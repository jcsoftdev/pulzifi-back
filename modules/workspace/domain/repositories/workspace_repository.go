package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/value_objects"
)

// WorkspaceRepository defines persistence operations
type WorkspaceRepository interface {
	Create(ctx context.Context, workspace *entities.Workspace) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Workspace, error)
	List(ctx context.Context) ([]*entities.Workspace, error)
	ListByCreator(ctx context.Context, createdBy uuid.UUID) ([]*entities.Workspace, error)
	Update(ctx context.Context, workspace *entities.Workspace) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Member management
	AddMember(ctx context.Context, member *entities.WorkspaceMember) error
	GetMember(ctx context.Context, workspaceID, userID uuid.UUID) (*entities.WorkspaceMember, error)
	ListMembers(ctx context.Context, workspaceID uuid.UUID) ([]*entities.WorkspaceMember, error)
	ListByMember(ctx context.Context, userID uuid.UUID) ([]*entities.Workspace, error)
	UpdateMemberRole(ctx context.Context, workspaceID, userID uuid.UUID, role value_objects.WorkspaceRole) error
	RemoveMember(ctx context.Context, workspaceID, userID uuid.UUID) error
}
