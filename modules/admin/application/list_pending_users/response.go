package listpendingusers

import (
	"time"

	"github.com/google/uuid"
)

// PendingUserResponse represents a single pending user in the list
type PendingUserResponse struct {
	RequestID             uuid.UUID `json:"request_id"`
	UserID                uuid.UUID `json:"user_id"`
	Email                 string    `json:"email"`
	FirstName             string    `json:"first_name"`
	LastName              string    `json:"last_name"`
	OrganizationName      string    `json:"organization_name"`
	OrganizationSubdomain string    `json:"organization_subdomain"`
	CreatedAt             time.Time `json:"created_at"`
}

// Response contains the list of pending users
type Response struct {
	PendingUsers []PendingUserResponse `json:"pending_users"`
}
