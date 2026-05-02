package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
)

type IntegrationRepository interface {
	Create(ctx context.Context, i *entities.Integration) error
	Update(ctx context.Context, i *entities.Integration) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Integration, error)
	GetByOrgAndService(ctx context.Context, orgID uuid.UUID, serviceType string) (*entities.Integration, error)
	ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*entities.Integration, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}
