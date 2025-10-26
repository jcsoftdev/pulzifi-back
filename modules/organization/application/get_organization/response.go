package get_organization

import "github.com/google/uuid"

// Response represents the output of getting an organization
type Response struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Subdomain   string    `json:"subdomain"`
	SchemaName  string    `json:"schema_name"`
	OwnerUserID uuid.UUID `json:"owner_user_id"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}
