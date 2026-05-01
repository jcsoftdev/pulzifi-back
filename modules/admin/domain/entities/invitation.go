package entities

import (
	"time"

	"github.com/google/uuid"
)

type InvitationStatus string

const (
	InvitationPending  InvitationStatus = "pending"
	InvitationAccepted InvitationStatus = "accepted"
	InvitationRevoked  InvitationStatus = "revoked"
	InvitationExpired  InvitationStatus = "expired"
)

type Invitation struct {
	ID             uuid.UUID
	Email          string
	Token          string
	Status         InvitationStatus
	InvitedBy      uuid.UUID
	ExpiresAt      time.Time
	EmailSentAt    *time.Time
	EmailError     *string
	RevokedBy      *uuid.UUID
	RevokedAt      *time.Time
	ResentCount    int
	LastResentAt   *time.Time
	AcceptedUserID *uuid.UUID
	OrgName        *string
	OrgSubdomain   *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (i *Invitation) IsActive(now time.Time) bool {
	return i.Status == InvitationPending && now.Before(i.ExpiresAt)
}
