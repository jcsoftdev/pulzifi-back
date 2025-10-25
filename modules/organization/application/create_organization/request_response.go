package create_organization

import "github.com/google/uuid"

// CreateOrganizationRequest represents the input for creating an organization
type CreateOrganizationRequest struct {
	Name      string `json:"name" binding:"required,min=2,max=255"`
	Subdomain string `json:"subdomain" binding:"required,min=3,max=63"`
}

// CreateOrganizationResponse represents the output of creating an organization
type CreateOrganizationResponse struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Subdomain  string    `json:"subdomain"`
	SchemaName string    `json:"schema_name"`
	CreatedAt  string    `json:"created_at"`
}
