package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/entities"
)

// OrganizationRepository defines operations for persisting organizations
type OrganizationRepository interface {
	// Create stores a new organization
	Create(ctx context.Context, organization *entities.Organization) error

	// GetByID retrieves an organization by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Organization, error)

	// GetBySubdomain retrieves an organization by its subdomain
	GetBySubdomain(ctx context.Context, subdomain string) (*entities.Organization, error)

	// List retrieves all organizations for a user (paginated)
	List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Organization, error)

	// Update modifies an existing organization
	Update(ctx context.Context, organization *entities.Organization) error

	// Delete soft-deletes an organization
	Delete(ctx context.Context, id uuid.UUID) error

	// CountBySubdomain checks if a subdomain already exists
	CountBySubdomain(ctx context.Context, subdomain string) (int, error)
}
