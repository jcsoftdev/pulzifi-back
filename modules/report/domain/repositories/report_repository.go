package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/report/domain/entities"
)

type ReportRepository interface {
	Create(ctx context.Context, report *entities.Report) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Report, error)
	ListByPage(ctx context.Context, pageID uuid.UUID) ([]*entities.Report, error)
	List(ctx context.Context) ([]*entities.Report, error)
}
