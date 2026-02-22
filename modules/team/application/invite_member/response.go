package invitemember

import (
	"time"

	"github.com/google/uuid"
)

type InviteMemberResponse struct {
	ID               uuid.UUID `json:"id"`
	UserID           uuid.UUID `json:"user_id"`
	Role             string    `json:"role"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	Email            string    `json:"email"`
	JoinedAt         time.Time `json:"joined_at"`
	IsNewUser        bool      `json:"is_new_user,omitempty"`
	SetPasswordToken string    `json:"set_password_token,omitempty"`
}
