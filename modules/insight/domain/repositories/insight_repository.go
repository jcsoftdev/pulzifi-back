package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/entities"
)

type InsightRepository interface {
	Create(ctx context.Context, insight *entities.Insight) error
	ListByPageID(ctx context.Context, pageID uuid.UUID) ([]*entities.Insight, error)
	ListByCheckID(ctx context.Context, checkID uuid.UUID) ([]*entities.Insight, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Insight, error)
}
