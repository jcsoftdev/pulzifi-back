package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Bus implements eventbus.MessageBus and eventbus.TxPublisher backed by the
// transactional outbox. Handlers registered via Subscribe are driven by the
// internal Relay; they are NOT called directly from Publish/PublishTx.
type Bus struct {
	store Store
	relay *Relay
	cfg   Config
}

// NewBus constructs a Bus using the given database and config.
// The relay is constructed but NOT started; call StartRelay(ctx) from the
// worker process after all subscribers are registered.
func NewBus(db *sql.DB, cfg Config) *Bus {
	store := NewPostgresStore(db)
	relay := NewRelay(store, cfg)
	return &Bus{store: store, relay: relay, cfg: cfg}
}

// newBusWithStore constructs a Bus with a pre-built Store (for testing).
func newBusWithStore(store Store, cfg Config) *Bus {
	relay := NewRelay(store, cfg)
	return &Bus{store: store, relay: relay, cfg: cfg}
}

// Publish inserts the event into the outbox inside its own short transaction
// (durable but not atomically enlisted with a caller-owned tx).
func (b *Bus) Publish(topic, key string, payload []byte) error {
	row := newRow(topic, key, payload)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return b.store.InsertStandalone(ctx, row)
}

// PublishTx inserts the event into the outbox enlisted in the caller's
// transaction. Atomicity is guaranteed by the caller's tx.
func (b *Bus) PublishTx(ctx context.Context, tx *sql.Tx, topic, key string, payload []byte) error {
	row := newRow(topic, key, payload)
	return b.store.Insert(ctx, tx, row)
}

// Subscribe registers a handler that will be called by the Relay when events
// matching topic are dispatched. Handlers are invoked synchronously inside the
// relay's dispatch loop.
//
// The handler parameter accepts the same func(string, []byte) signature as
// eventbus.EventHandler so callers can pass values interchangeably.
func (b *Bus) Subscribe(topic string, handler func(string, []byte)) error {
	b.relay.Subscribe(topic, Handler(handler))
	return nil
}

// Close stops the relay and waits up to 5 seconds for the in-flight dispatch
// to drain.
func (b *Bus) Close() {
	done := make(chan struct{})
	go func() {
		b.relay.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		logger.Warn("outbox bus: relay did not stop within grace period")
	}
}

// StartRelay launches the relay polling loop. Call this from the worker
// process after all subscribers are registered.
func (b *Bus) StartRelay(ctx context.Context) {
	logger.Info("outbox relay: starting",
		zap.Duration("poll_interval", b.cfg.PollInterval),
		zap.Int("batch_size", b.cfg.BatchSize),
	)
	b.relay.Start(ctx)
}

// newRow builds an outbox Row from Publish/PublishTx parameters. It extracts
// the tenant field from the payload if it contains a "tenant" JSON key so the
// column is populated for observability queries.
func newRow(topic, key string, payload []byte) Row {
	id := uuid.New()
	tenant := extractTenant(payload)
	// Override id from payload if present (end-to-end dedupe with DomainEvent.ID).
	if pid := extractID(payload); pid != uuid.Nil {
		id = pid
	}
	return Row{
		ID:      id,
		Topic:   topic,
		MsgKey:  key,
		Payload: payload,
		Tenant:  tenant,
	}
}

func extractTenant(payload []byte) string {
	var v struct {
		Tenant string `json:"tenant"`
	}
	_ = json.Unmarshal(payload, &v)
	return v.Tenant
}

func extractID(payload []byte) uuid.UUID {
	var v struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(payload, &v); err != nil {
		return uuid.Nil
	}
	id, err := uuid.Parse(v.ID)
	if err != nil {
		return uuid.Nil
	}
	return id
}
