package listmembers

import (
	"time"

	"github.com/google/uuid"
)

type TeamMemberResponse struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	Role             string     `json:"role"`
	FirstName        string     `json:"first_name"`
	LastName         string     `json:"last_name"`
	Email            string     `json:"email"`
	AvatarURL        *string    `json:"avatar_url,omitempty"`
	InvitedBy        *uuid.UUID `json:"invited_by,omitempty"`
	JoinedAt         time.Time  `json:"joined_at"`
	InvitationStatus string     `json:"invitation_status"`
}

type ListMembersResponse struct {
	Members []*TeamMemberResponse `json:"members"`
}
