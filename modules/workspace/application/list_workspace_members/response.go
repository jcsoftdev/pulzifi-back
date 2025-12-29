package list_workspace_members

import (
	"time"

	"github.com/google/uuid"
)

type ListWorkspaceMembersResponse struct {
	Members []WorkspaceMemberDTO `json:"members"`
}

type WorkspaceMemberDTO struct {
	UserID    uuid.UUID  `json:"user_id"`
	Role      string     `json:"role"`
	InvitedBy *uuid.UUID `json:"invited_by,omitempty"`
	InvitedAt time.Time  `json:"invited_at"`
}
