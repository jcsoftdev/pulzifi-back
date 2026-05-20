package entities

import "time"

// WebhookEventStatus tracks the processing state of a Stripe webhook event.
type WebhookEventStatus string

const (
	WebhookEventReceived  WebhookEventStatus = "received"
	WebhookEventProcessed WebhookEventStatus = "processed"
	WebhookEventFailed    WebhookEventStatus = "failed"
)

// WebhookEvent is the domain representation of a row in stripe_webhook_events.
// Its primary purpose is idempotency: the event_id is the deduplication key.
type WebhookEvent struct {
	EventID     string
	EventType   string
	ReceivedAt  time.Time
	ProcessedAt *time.Time
	Status      WebhookEventStatus
}
