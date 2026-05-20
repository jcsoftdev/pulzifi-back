package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/repositories"
)

// Compile-time interface check.
var _ repositories.WebhookEventRepository = (*WebhookEventPostgresRepository)(nil)

// WebhookEventPostgresRepository implements WebhookEventRepository using PostgreSQL.
// The idempotency contract is enforced by the UNIQUE PRIMARY KEY on event_id:
//
//	INSERT ... ON CONFLICT (event_id) DO NOTHING
//
// When the INSERT affects 0 rows, the event is a duplicate and Save returns (false, nil).
type WebhookEventPostgresRepository struct {
	db *sql.DB
}

// NewWebhookEventPostgresRepository returns a new WebhookEventPostgresRepository.
func NewWebhookEventPostgresRepository(db *sql.DB) *WebhookEventPostgresRepository {
	return &WebhookEventPostgresRepository{db: db}
}

// Save inserts the event if the event_id does not already exist.
// Returns (true, nil) on first insertion; (false, nil) if the event_id is a duplicate.
func (r *WebhookEventPostgresRepository) Save(ctx context.Context, event *entities.WebhookEvent) (bool, error) {
	const query = `
		INSERT INTO public.stripe_webhook_events (event_id, event_type, received_at, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (event_id) DO NOTHING
	`
	result, err := r.db.ExecContext(ctx, query,
		event.EventID,
		event.EventType,
		event.ReceivedAt,
		string(event.Status),
	)
	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows > 0, nil
}

// MarkProcessed updates processed_at and status for the given event_id.
func (r *WebhookEventPostgresRepository) MarkProcessed(ctx context.Context, eventID string, status entities.WebhookEventStatus) error {
	const query = `
		UPDATE public.stripe_webhook_events
		SET processed_at = $2, status = $3
		WHERE event_id = $1
	`
	_, err := r.db.ExecContext(ctx, query, eventID, time.Now(), string(status))
	return err
}
