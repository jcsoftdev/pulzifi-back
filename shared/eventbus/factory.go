package eventbus

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jcsoftdev/pulzifi-back/shared/eventbus/outbox"
)

// outboxAdapter wraps *outbox.Bus and adapts the Subscribe signature to satisfy
// the MessageBus interface. The outbox.Handler type is defined as
// func(string, []byte), identical in shape to EventHandler; an explicit
// conversion bridges the named-type difference without importing outbox from
// within the eventbus package (which would create a cycle the other way).
type outboxAdapter struct {
	bus *outbox.Bus
}

// Publish delegates to the outbox bus.
func (a *outboxAdapter) Publish(topic, key string, payload []byte) error {
	return a.bus.Publish(topic, key, payload)
}

// Subscribe adapts EventHandler (named type) to outbox.Handler (same shape).
func (a *outboxAdapter) Subscribe(topic string, handler EventHandler) error {
	return a.bus.Subscribe(topic, func(t string, p []byte) { handler(t, p) })
}

// Close delegates to the outbox bus.
func (a *outboxAdapter) Close() {
	a.bus.Close()
}

// PublishTx satisfies TxPublisher by delegating to the outbox bus.
func (a *outboxAdapter) PublishTx(ctx context.Context, tx *sql.Tx, topic, key string, payload []byte) error {
	return a.bus.PublishTx(ctx, tx, topic, key, payload)
}

// StartRelay exposes the relay start method for the worker process.
func (a *outboxAdapter) StartRelay(ctx context.Context) {
	a.bus.StartRelay(ctx)
}

// OutboxBus is the concrete type returned by NewFromConfig when provider is
// "outbox". Callers that need to start the relay (worker process) can type-
// assert to *OutboxBus and call StartRelay.
type OutboxBus = outboxAdapter

// NewFromConfig constructs a MessageBus from application configuration.
//
// When provider is "memory" (the default), it returns the in-memory singleton
// so existing tests and production behaviour are unchanged.
//
// When provider is "outbox", it returns an *outboxAdapter (aliased as
// *OutboxBus) backed by the given db. The relay is NOT started here; the
// worker process must type-assert and call bus.(*eventbus.OutboxBus).StartRelay(ctx).
//
// db is only required when provider == "outbox"; it may be nil for "memory".
func NewFromConfig(provider string, batchSize, maxAttempts int, db *sql.DB) (MessageBus, error) {
	switch provider {
	case "", "memory":
		return GetInstance(), nil
	case "outbox":
		if db == nil {
			return nil, fmt.Errorf("eventbus: outbox provider requires a non-nil *sql.DB")
		}
		cfg := outbox.DefaultConfig()
		cfg.BatchSize = batchSize
		cfg.MaxAttempts = maxAttempts
		return &outboxAdapter{bus: outbox.NewBus(db, cfg)}, nil
	default:
		return nil, fmt.Errorf("eventbus: unknown provider %q (valid: memory, outbox)", provider)
	}
}
