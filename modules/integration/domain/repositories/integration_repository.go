package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
)

type IntegrationRepository interface {
	Create(ctx context.Context, integration *entities.Integration) error
	List(ctx context.Context) ([]*entities.Integration, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Integration, error)
	GetByServiceType(ctx context.Context, serviceType string) (*entities.Integration, error)
	ListByServiceType(ctx context.Context, serviceType string) ([]*entities.Integration, error)
	Update(ctx context.Context, integration *entities.Integration) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}
