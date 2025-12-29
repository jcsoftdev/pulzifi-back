package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
)

// UserRepository defines operations for persisting users
type UserRepository interface {
	// Create stores a new user
	Create(ctx context.Context, user *entities.User) error

	// GetByID retrieves a user by their ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error)

	// GetByEmail retrieves a user by their email
	GetByEmail(ctx context.Context, email string) (*entities.User, error)

	// Update modifies an existing user
	Update(ctx context.Context, user *entities.User) error

	// Delete soft-deletes a user
	Delete(ctx context.Context, id uuid.UUID) error

	// ExistsByEmail checks if a user with the given email exists
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// GetUserFirstOrganization gets the first organization subdomain for a user
	GetUserFirstOrganization(ctx context.Context, userID uuid.UUID) (*string, error)
}
