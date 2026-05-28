package entities

import "time"

// WebhookEventStatus tracks the processing state of a Stripe webhook event.
type WebhookEventStatus string

const (
	WebhookEventReceived  WebhookEventStatus = "received"
	WebhookEventProcessed WebhookEventStatus = "processed"
	WebhookEventFailed    WebhookEventStatus = "failed"
	// WebhookEventDeferred marks events that arrived before the org/customer
	// mapping existed locally. These are NOT failures — the payload is kept
	// in raw_payload and replayed when the customer is later linked.
	WebhookEventDeferred WebhookEventStatus = "deferred"
)

// WebhookEvent is the domain representation of a row in stripe_webhook_events.
// Its primary purpose is idempotency: the event_id is the deduplication key.
// RawPayload holds the exact bytes Stripe sent (signature-validated) so the
// event can be replayed by ReconcileFromStripe without round-tripping Stripe.
type WebhookEvent struct {
	EventID     string
	EventType   string
	ReceivedAt  time.Time
	ProcessedAt *time.Time
	Status      WebhookEventStatus
	RawPayload  []byte
}
