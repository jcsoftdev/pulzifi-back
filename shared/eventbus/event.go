package eventbus

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	TopicChangeDetected = "change.detected"
	TopicAlertCreated   = "alert.created"
	TopicInsightReady   = "insight.ready"
	TopicCheckFailed    = "check.failed"
	TopicReportWeekly   = "report.weekly"
)

type DomainEvent struct {
	ID          uuid.UUID       `json:"id"`
	Type        string          `json:"type"`
	OccurredAt  time.Time       `json:"occurred_at"`
	OrgID       uuid.UUID       `json:"org_id"`
	WorkspaceID *uuid.UUID      `json:"workspace_id,omitempty"`
	PageID      *uuid.UUID      `json:"page_id,omitempty"`
	Tenant      string          `json:"tenant"`
	Data        json.RawMessage `json:"data"`
}

func PublishDomainEvent(bus MessageBus, ev DomainEvent) error {
	if ev.ID == uuid.Nil {
		ev.ID = uuid.New()
	}
	if ev.OccurredAt.IsZero() {
		ev.OccurredAt = time.Now().UTC()
	}
	payload, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	return bus.Publish(ev.Type, ev.ID.String(), payload)
}

// TxPublisher is an optional capability implemented by the outbox bus.
// Callers that own a *sql.Tx can atomically enlist an event into the same
// transaction using PublishTx. Plain Publish remains available for callers
// that do not own a transaction.
type TxPublisher interface {
	PublishTx(ctx context.Context, tx *sql.Tx, topic, key string, payload []byte) error
}

// PublishDomainEventTx auto-fills ID/OccurredAt, marshals the event, then
// dispatches it. If bus implements TxPublisher the event is atomically
// enlisted in tx; otherwise it falls back to plain Publish so callers using
// the in-memory bus in tests work without change.
func PublishDomainEventTx(ctx context.Context, tx *sql.Tx, bus MessageBus, ev DomainEvent) error {
	if ev.ID == uuid.Nil {
		ev.ID = uuid.New()
	}
	if ev.OccurredAt.IsZero() {
		ev.OccurredAt = time.Now().UTC()
	}
	payload, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	if txp, ok := bus.(TxPublisher); ok {
		return txp.PublishTx(ctx, tx, ev.Type, ev.ID.String(), payload)
	}
	return bus.Publish(ev.Type, ev.ID.String(), payload)
}

// Idempotent wraps handler so that redelivery of an already-consumed event is
// a no-op. consumer is a stable, unique name for the subscriber (e.g.
// "org.user-deleted"). The deduplication record lives in
// public.outbox_consumed. If db is nil the wrapper is transparent (suitable
// for in-memory bus usage in tests).
func Idempotent(consumer string, db *sql.DB, handler EventHandler) EventHandler {
	return func(topic string, payload []byte) {
		if db == nil {
			handler(topic, payload)
			return
		}
		// Extract event_id from payload (best-effort — skip if unparseable).
		var ev struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(payload, &ev); err != nil || ev.ID == "" {
			handler(topic, payload)
			return
		}
		eventID, err := uuid.Parse(ev.ID)
		if err != nil {
			handler(topic, payload)
			return
		}
		res, err := db.Exec(
			`INSERT INTO public.outbox_consumed (consumer, event_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			consumer, eventID,
		)
		if err != nil {
			// On error let the handler run (prefer at-least-once over loss).
			handler(topic, payload)
			return
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			// Already consumed — skip.
			return
		}
		handler(topic, payload)
	}
}

type DomainEventHandler func(ev DomainEvent)

func SubscribeDomainEvent(bus MessageBus, topic string, h DomainEventHandler) error {
	return bus.Subscribe(topic, func(_ string, payload []byte) {
		var ev DomainEvent
		if err := json.Unmarshal(payload, &ev); err != nil {
			return
		}
		h(ev)
	})
}
