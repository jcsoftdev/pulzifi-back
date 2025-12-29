package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/value_objects"
)

type WorkspaceMember struct {
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	Role        value_objects.WorkspaceRole
	InvitedBy   *uuid.UUID
	InvitedAt   time.Time
}

func NewWorkspaceMember(
	workspaceID uuid.UUID,
	userID uuid.UUID,
	role value_objects.WorkspaceRole,
	invitedBy *uuid.UUID,
) *WorkspaceMember {
	return &WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        role,
		InvitedBy:   invitedBy,
		InvitedAt:   time.Now(),
	}
}
