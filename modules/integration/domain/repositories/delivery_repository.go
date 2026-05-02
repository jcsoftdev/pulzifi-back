package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
)

// ErrDeliveryNotInDeadState is returned by Retry when the row doesn't exist
// or isn't in the dead state (and therefore can't be retried).
var ErrDeliveryNotInDeadState = errors.New("delivery not found or not in dead state")

type DeliveryRepository interface {
	Create(ctx context.Context, d *entities.Delivery) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Delivery, error)
	ClaimPending(ctx context.Context, limit int, now time.Time) ([]*entities.Delivery, error)
	MarkDelivered(ctx context.Context, id uuid.UUID, code int, bodySnip string) error
	MarkFailed(ctx context.Context, id uuid.UUID, code *int, body, errMsg string, nextAttempt time.Time, history []entities.Attempt) error
	MarkDead(ctx context.Context, id uuid.UUID, errMsg string, history []entities.Attempt) error
	ListByDestination(ctx context.Context, destinationID uuid.UUID, limit, offset int) ([]*entities.Delivery, error)
	Retry(ctx context.Context, id uuid.UUID) error
}
