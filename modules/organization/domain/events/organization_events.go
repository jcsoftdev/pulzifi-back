package events

import (
	"time"

	"github.com/google/uuid"
)

// OrganizationCreated is a domain event published when an organization is created
type OrganizationCreated struct {
	ID          uuid.UUID
	Name        string
	Subdomain   string
	SchemaName  string
	OwnerUserID uuid.UUID
	CreatedAt   time.Time
}

// EventType returns the event type identifier
func (e *OrganizationCreated) EventType() string {
	return "organization.created"
}

// OrganizationDeleted is a domain event published when an organization is deleted
type OrganizationDeleted struct {
	ID        uuid.UUID
	Subdomain string
	DeletedAt time.Time
}

// EventType returns the event type identifier
func (e *OrganizationDeleted) EventType() string {
	return "organization.deleted"
}

// OrganizationUpdated is a domain event published when an organization is updated
type OrganizationUpdated struct {
	ID        uuid.UUID
	Name      string
	UpdatedAt time.Time
}

// EventType returns the event type identifier
func (e *OrganizationUpdated) EventType() string {
	return "organization.updated"
}
