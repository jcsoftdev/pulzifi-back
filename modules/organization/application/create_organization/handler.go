package create_organization

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/entities"
	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/events"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// EventPublisher interface for publishing domain events
type EventPublisher interface {
	PublishOrganizationCreated(ctx context.Context, event *events.OrganizationCreated) error
}

// CreateOrganizationHandler handles the creation of a new organization
type CreateOrganizationHandler struct {
	repo      repositories.OrganizationRepository
	service   *services.OrganizationService
	db        *sql.DB
	publisher EventPublisher
}

// NewCreateOrganizationHandler creates a new handler instance
func NewCreateOrganizationHandler(
	repo repositories.OrganizationRepository,
	service *services.OrganizationService,
	db *sql.DB,
	publisher EventPublisher,
) *CreateOrganizationHandler {
	return &CreateOrganizationHandler{
		repo:      repo,
		service:   service,
		db:        db,
		publisher: publisher,
	}
}

// Handle executes the create organization use case
func (h *CreateOrganizationHandler) Handle(
	ctx context.Context,
	req *Request,
	userID uuid.UUID,
) (*Response, error) {

	// Validate organization name
	if err := h.service.ValidateOrganizationName(req.Name); err != nil {
		logger.Error("Invalid organization name", zap.Error(err))
		return nil, err
	}

	// Validate subdomain
	if err := h.service.ValidateSubdomain(req.Subdomain); err != nil {
		logger.Error("Invalid subdomain", zap.Error(err))
		return nil, err
	}

	// Check if subdomain already exists
	count, err := h.repo.CountBySubdomain(ctx, req.Subdomain)
	if err != nil {
		logger.Error("Failed to check subdomain availability", zap.Error(err))
		return nil, err
	}
	if count > 0 {
		logger.Warn("Subdomain already exists", zap.String("subdomain", req.Subdomain))
		return nil, &domainerrors.SubdomainAlreadyExistsError{Subdomain: req.Subdomain}
	}

	// Generate schema name
	schemaName := h.service.GenerateSchemaName(req.Subdomain)

	// Create organization entity
	org := entities.NewOrganization(req.Name, req.Subdomain, schemaName, userID)

	// Start transaction
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("Failed to begin transaction", zap.Error(err))
		return nil, err
	}

	// TODO: Implement the following in persistence layer:
	// 1. Create organization in public.organizations table
	// 2. Create tenant schema (using create_tenant_schema() function)
	// 3. Add user as owner in organization_members table

	// For now, just persist the organization
	if err := h.repo.Create(ctx, org); err != nil {
		tx.Rollback()
		logger.Error("Failed to create organization", zap.Error(err))
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		logger.Error("Failed to commit transaction", zap.Error(err))
		return nil, err
	}

	// Publish organization created event
	createdEvent := &events.OrganizationCreated{
		ID:          org.ID,
		Name:        org.Name,
		Subdomain:   org.Subdomain,
		SchemaName:  org.SchemaName,
		OwnerUserID: org.OwnerUserID,
		CreatedAt:   org.CreatedAt,
	}

	if err := h.publisher.PublishOrganizationCreated(ctx, createdEvent); err != nil {
		logger.Error("Failed to publish organization created event", zap.Error(err))
		// Don't fail the request if event publishing fails
	}

	return &Response{
		ID:         org.ID,
		Name:       org.Name,
		Subdomain:  org.Subdomain,
		SchemaName: org.SchemaName,
		CreatedAt:  org.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}
