package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
)

// MonitoringConfigRepository defines operations for managing monitoring configs
type MonitoringConfigRepository interface {
	Create(ctx context.Context, config *entities.MonitoringConfig) error
	GetByPageID(ctx context.Context, pageID uuid.UUID) (*entities.MonitoringConfig, error)
	Update(ctx context.Context, config *entities.MonitoringConfig) error
}
