package entities

import (
	"time"

	"github.com/google/uuid"
)

const (
	InvitationStatusPending = "pending"
	InvitationStatusActive  = "active"
)

// TeamMember represents an organization-level member with user profile info
type TeamMember struct {
	ID               uuid.UUID
	OrganizationID   uuid.UUID
	UserID           uuid.UUID
	Role             string // "OWNER", "ADMIN", "MEMBER"
	InvitedBy        *uuid.UUID
	JoinedAt         time.Time
	InvitationStatus string // "pending", "active"
	// Denormalized user fields (joined from public.users)
	FirstName string
	LastName  string
	Email     string
	AvatarURL *string
}

func (m *TeamMember) FullName() string {
	return m.FirstName + " " + m.LastName
}
