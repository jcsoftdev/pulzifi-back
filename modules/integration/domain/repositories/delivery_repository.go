package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
)

type DeliveryRepository interface {
	Create(ctx context.Context, d *entities.Delivery) error
	ClaimPending(ctx context.Context, limit int, now time.Time) ([]*entities.Delivery, error)
	MarkDelivered(ctx context.Context, id uuid.UUID, code int, bodySnip string) error
	MarkFailed(ctx context.Context, id uuid.UUID, code *int, body, errMsg string, nextAttempt time.Time) error
	MarkDead(ctx context.Context, id uuid.UUID, errMsg string) error
	ListByDestination(ctx context.Context, destinationID uuid.UUID, limit, offset int) ([]*entities.Delivery, error)
	Retry(ctx context.Context, id uuid.UUID) error
}
