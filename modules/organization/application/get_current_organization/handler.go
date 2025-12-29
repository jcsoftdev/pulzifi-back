package get_current_organization

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/repositories"
)

// Handler handles the get current organization use case
type Handler struct {
	orgRepo repositories.OrganizationRepository
}

// NewHandler creates a new get current organization handler
func NewHandler(orgRepo repositories.OrganizationRepository) *Handler {
	return &Handler{
		orgRepo: orgRepo,
	}
}

// Response represents the current organization response
type Response struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Company   *string `json:"company,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// Handle executes the get current organization use case
func (h *Handler) Handle(ctx context.Context, subdomain string) (*Response, error) {
	org, err := h.orgRepo.GetBySubdomain(ctx, subdomain)
	if err != nil {
		return nil, err
	}

	if org == nil {
		return nil, nil
	}

	return h.toResponse(org), nil
}

func (h *Handler) toResponse(org *entities.Organization) *Response {
	// For now, company is same as name
	company := org.Name

	return &Response{
		ID:        org.ID.String(),
		Name:      org.Name,
		Company:   &company,
		CreatedAt: org.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: org.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
