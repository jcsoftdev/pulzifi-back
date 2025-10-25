package create_organization

import "github.com/google/uuid"

// Response represents the output of creating an organization
type Response struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Subdomain string    `json:"subdomain"`
	SchemaName string   `json:"schema_name"`
	CreatedAt string    `json:"created_at"`
}
