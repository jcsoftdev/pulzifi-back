package get_organization

import (
	"context"

	"github.com/google/uuid"
	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// GetOrganizationHandler handles retrieving an organization
type GetOrganizationHandler struct {
	repo repositories.OrganizationRepository
}

// NewGetOrganizationHandler creates a new handler instance
func NewGetOrganizationHandler(repo repositories.OrganizationRepository) *GetOrganizationHandler {
	return &GetOrganizationHandler{
		repo: repo,
	}
}

// Handle executes the get organization use case
func (h *GetOrganizationHandler) Handle(
	ctx context.Context,
	organizationID uuid.UUID,
) (*Response, error) {

	org, err := h.repo.GetByID(ctx, organizationID)
	if err != nil {
		logger.Error("Failed to get organization", zap.Error(err), zap.String("organization_id", organizationID.String()))
		return nil, err
	}

	if org == nil {
		logger.Warn("Organization not found", zap.String("organization_id", organizationID.String()))
		return nil, &domainerrors.OrganizationNotFoundError{OrganizationID: organizationID.String()}
	}

	if org.IsDeleted() {
		logger.Warn("Organization is deleted", zap.String("organization_id", organizationID.String()))
		return nil, &domainerrors.OrganizationAlreadyDeletedError{OrganizationID: organizationID.String()}
	}

	return &Response{
		ID:        org.ID,
		Name:      org.Name,
		Subdomain: org.Subdomain,
		SchemaName: org.SchemaName,
		OwnerUserID: org.OwnerUserID,
		CreatedAt: org.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: org.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}
