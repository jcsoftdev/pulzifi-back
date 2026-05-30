package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBus_Publish_InsertsRow(t *testing.T) {
	store := newFakeStore()
	bus := newBusWithStore(store, DefaultConfig())

	id := uuid.New()
	payload, _ := json.Marshal(map[string]string{"id": id.String(), "tenant": "acme"})

	if err := bus.Publish("test.topic", id.String(), payload); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	if len(store.rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(store.rows))
	}
	var r *Row
	for _, v := range store.rows {
		r = v
	}
	if r.Topic != "test.topic" {
		t.Errorf("topic = %q, want %q", r.Topic, "test.topic")
	}
}

func TestBus_PublishTx_InsertsRowInTransaction(t *testing.T) {
	store := newFakeStore()
	bus := newBusWithStore(store, DefaultConfig())

	id := uuid.New()
	payload, _ := json.Marshal(map[string]string{"id": id.String()})

	// PublishTx with a nil tx works via Insert (the fake ignores tx).
	if err := bus.PublishTx(context.Background(), (*sql.Tx)(nil), "tx.topic", id.String(), payload); err != nil {
		t.Fatalf("PublishTx: %v", err)
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	if len(store.rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(store.rows))
	}
}

func TestBus_Subscribe_RelayDispatchesHandler(t *testing.T) {
	store := newFakeStore()
	cfg := DefaultConfig()
	cfg.PollInterval = 10 * time.Millisecond
	bus := newBusWithStore(store, cfg)

	received := make(chan []byte, 1)
	_ = bus.Subscribe("dispatch.topic", func(_ string, payload []byte) {
		received <- payload
	})

	id := uuid.New()
	payload, _ := json.Marshal(map[string]string{"id": id.String()})
	store.rows[id] = &Row{
		ID:            id,
		Topic:         "dispatch.topic",
		Payload:       payload,
		Status:        "pending",
		NextAttemptAt: time.Now().Add(-1 * time.Second),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	bus.StartRelay(ctx)

	select {
	case got := <-received:
		if string(got) != string(payload) {
			t.Errorf("payload mismatch")
		}
	case <-ctx.Done():
		t.Fatal("handler was not called before timeout")
	}
}
