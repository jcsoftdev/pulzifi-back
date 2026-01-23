package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
)

// CheckRepository defines operations for managing checks
type CheckRepository interface {
	Create(ctx context.Context, check *entities.Check) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Check, error)
	ListByPage(ctx context.Context, pageID uuid.UUID) ([]*entities.Check, error)
	GetLatestByPage(ctx context.Context, pageID uuid.UUID) (*entities.Check, error)
}
