package entities

import (
	"time"

	"github.com/google/uuid"
)

const (
	RegistrationStatusPending  = "pending"
	RegistrationStatusApproved = "approved"
	RegistrationStatusRejected = "rejected"
)

// RegistrationRequest represents a pending registration request
type RegistrationRequest struct {
	ID                     uuid.UUID
	UserID                 uuid.UUID
	OrganizationName       string
	OrganizationSubdomain  string
	Status                 string
	ReviewedBy             *uuid.UUID
	ReviewedAt             *time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// NewRegistrationRequest creates a new registration request
func NewRegistrationRequest(userID uuid.UUID, orgName, orgSubdomain string) *RegistrationRequest {
	return &RegistrationRequest{
		ID:                    uuid.New(),
		UserID:                userID,
		OrganizationName:      orgName,
		OrganizationSubdomain: orgSubdomain,
		Status:                RegistrationStatusPending,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}
}
