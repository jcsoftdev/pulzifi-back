package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/email/domain/entities"
)

// EmailRepository defines the interface for email persistence
type EmailRepository interface {
	// Save persists an email
	Save(ctx context.Context, email *entities.Email) error

	// GetByID retrieves an email by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Email, error)

	// GetByTo retrieves emails by recipient
	GetByTo(ctx context.Context, to string, limit int) ([]*entities.Email, error)

	// Update updates an existing email
	Update(ctx context.Context, email *entities.Email) error
}
