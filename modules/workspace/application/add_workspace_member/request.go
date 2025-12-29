package add_workspace_member

import (
	"time"

	"github.com/google/uuid"
)

type AddWorkspaceMemberRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
}

type AddWorkspaceMemberResponse struct {
	WorkspaceID uuid.UUID `json:"workspace_id"`
	UserID      uuid.UUID `json:"user_id"`
	Role        string    `json:"role"`
	InvitedBy   uuid.UUID `json:"invited_by"`
	InvitedAt   time.Time `json:"invited_at"`
}
