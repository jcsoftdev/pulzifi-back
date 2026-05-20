package repositories

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
)

// WebhookEventRepository provides idempotent storage for Stripe webhook events.
type WebhookEventRepository interface {
	// Save inserts the event if the event_id does not already exist.
	// Returns (true, nil) when the event was inserted (first delivery).
	// Returns (false, nil) when the event_id was a duplicate (idempotent no-op).
	Save(ctx context.Context, event *entities.WebhookEvent) (inserted bool, err error)

	// MarkProcessed updates processed_at and status for the given event_id.
	MarkProcessed(ctx context.Context, eventID string, status entities.WebhookEventStatus) error
}
