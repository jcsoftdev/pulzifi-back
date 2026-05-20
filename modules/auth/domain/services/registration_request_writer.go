package services

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PendingRegistration is auth's own representation of a pending registration request.
// It intentionally does not import admin/domain/entities.
type PendingRegistration struct {
	ID                    uuid.UUID
	UserID                uuid.UUID
	OrganizationName      string
	OrganizationSubdomain string
	Status                string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// NewPendingRegistration builds a PendingRegistration ready to be passed to
// RegistrationRequestWriter.Create. Using this constructor keeps the register
// use case free of admin/domain/entities.
func NewPendingRegistration(userID uuid.UUID, orgName, orgSubdomain string) *PendingRegistration {
	now := time.Now()
	return &PendingRegistration{
		ID:                    uuid.New(),
		UserID:                userID,
		OrganizationName:      orgName,
		OrganizationSubdomain: orgSubdomain,
		Status:                "pending",
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}

// RegistrationRequestWriter is auth's port for creating and querying pending
// registration requests. Implemented by authwiring in cmd/wiring/auth/.
type RegistrationRequestWriter interface {
	// Create stores a new pending registration request.
	Create(ctx context.Context, req *PendingRegistration) error

	// ExistsPendingBySubdomain reports whether a pending request already exists
	// for the given subdomain.
	ExistsPendingBySubdomain(ctx context.Context, subdomain string) (bool, error)
}
