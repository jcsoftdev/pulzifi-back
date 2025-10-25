package entities

import (
	"time"

	"github.com/google/uuid"
)

// Organization represents an organization entity
type Organization struct {
	ID          uuid.UUID  `db:"id"`
	Name        string     `db:"name"`
	Subdomain   string     `db:"subdomain"`
	SchemaName  string     `db:"schema_name"`
	OwnerUserID uuid.UUID  `db:"owner_user_id"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
}

// IsDeleted checks if the organization is soft-deleted
func (o *Organization) IsDeleted() bool {
	return o.DeletedAt != nil
}

// NewOrganization creates a new organization instance
func NewOrganization(name, subdomain, schemaName string, ownerUserID uuid.UUID) *Organization {
	now := time.Now()
	return &Organization{
		ID:          uuid.New(),
		Name:        name,
		Subdomain:   subdomain,
		SchemaName:  schemaName,
		OwnerUserID: ownerUserID,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   nil,
	}
}
