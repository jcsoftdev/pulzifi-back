package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/entities"
)

// RegistrationRequestRepository defines operations for persisting registration requests
type RegistrationRequestRepository interface {
	// Create stores a new registration request
	Create(ctx context.Context, req *entities.RegistrationRequest) error

	// GetByID retrieves a registration request by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.RegistrationRequest, error)

	// GetByUserID retrieves a registration request by user ID
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entities.RegistrationRequest, error)

	// ListPending retrieves all pending registration requests
	ListPending(ctx context.Context, limit, offset int) ([]*entities.RegistrationRequest, error)

	// UpdateStatus updates the status of a registration request
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, reviewedBy uuid.UUID) error
}
