package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
)

type DestinationRepository interface {
	Create(ctx context.Context, d *entities.Destination) error
	Update(ctx context.Context, d *entities.Destination) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Destination, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListByScope(ctx context.Context, scope entities.ScopeType, scopeID uuid.UUID) ([]*entities.Destination, error)
	// ResolveForEvent returns destinations that should fire for an event,
	// applying scope override (page > workspace > org) per service_type.
	ResolveForEvent(ctx context.Context, eventType string, orgID uuid.UUID, workspaceID, pageID *uuid.UUID) ([]*entities.Destination, error)
	DisableByIntegrationID(ctx context.Context, integrationID uuid.UUID) error
}
