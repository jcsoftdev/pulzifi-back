package eventbus_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
)

// fakeTxPublisher is a stub implementing both MessageBus and TxPublisher.
type fakeTxPublisher struct {
	txCalls    []txCall
	plainCalls []plainCall
}

type txCall struct {
	topic   string
	key     string
	payload []byte
}

type plainCall struct {
	topic   string
	payload []byte
}

func (f *fakeTxPublisher) Publish(topic, key string, payload []byte) error {
	f.plainCalls = append(f.plainCalls, plainCall{topic: topic, payload: payload})
	return nil
}

func (f *fakeTxPublisher) Subscribe(topic string, _ eventbus.EventHandler) error { return nil }
func (f *fakeTxPublisher) Close()                                                  {}

func (f *fakeTxPublisher) PublishTx(_ context.Context, _ *sql.Tx, topic, key string, payload []byte) error {
	f.txCalls = append(f.txCalls, txCall{topic: topic, key: key, payload: payload})
	return nil
}

func TestPublishDomainEventTx_UsesTxPublisherWhenAvailable(t *testing.T) {
	bus := &fakeTxPublisher{}
	ev := eventbus.DomainEvent{
		Type:   "test.topic",
		OrgID:  uuid.New(),
		Tenant: "t1",
		Data:   json.RawMessage(`{}`),
	}

	if err := eventbus.PublishDomainEventTx(context.Background(), nil, bus, ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(bus.txCalls) != 1 {
		t.Fatalf("expected 1 tx call, got %d", len(bus.txCalls))
	}
	if len(bus.plainCalls) != 0 {
		t.Fatalf("expected 0 plain calls, got %d", len(bus.plainCalls))
	}
	if bus.txCalls[0].topic != "test.topic" {
		t.Errorf("topic = %q, want %q", bus.txCalls[0].topic, "test.topic")
	}
}

func TestPublishDomainEventTx_FallsBackToPlainPublish(t *testing.T) {
	// In-memory EventBus does NOT implement TxPublisher.
	bus := eventbus.GetInstance()

	received := make(chan eventbus.DomainEvent, 1)
	_ = eventbus.SubscribeDomainEvent(bus, "tx.fallback.topic", func(ev eventbus.DomainEvent) {
		received <- ev
	})

	raw, _ := json.Marshal(map[string]string{"k": "v"})
	ev := eventbus.DomainEvent{
		Type:   "tx.fallback.topic",
		OrgID:  uuid.New(),
		Tenant: "t2",
		Data:   raw,
	}

	if err := eventbus.PublishDomainEventTx(context.Background(), nil, bus, ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := <-received
	if got.Type != "tx.fallback.topic" {
		t.Errorf("type = %q, want %q", got.Type, "tx.fallback.topic")
	}
}

func TestPublishDomainEventTx_AutoFillsIDAndOccurredAt(t *testing.T) {
	bus := &fakeTxPublisher{}
	ev := eventbus.DomainEvent{Type: "auto.fill", OrgID: uuid.New(), Tenant: "t"}

	_ = eventbus.PublishDomainEventTx(context.Background(), nil, bus, ev)

	if len(bus.txCalls) == 0 {
		t.Fatal("no tx call recorded")
	}

	var decoded eventbus.DomainEvent
	if err := json.Unmarshal(bus.txCalls[0].payload, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.ID == uuid.Nil {
		t.Error("ID was not auto-filled")
	}
	if decoded.OccurredAt.IsZero() {
		t.Error("OccurredAt was not auto-filled")
	}
}

// --- Idempotent tests (nil db = transparent pass-through) ---

func TestIdempotent_NilDB_AlwaysRunsHandler(t *testing.T) {
	calls := 0
	handler := eventbus.Idempotent("test-consumer", nil, func(_ string, _ []byte) {
		calls++
	})

	payload, _ := json.Marshal(map[string]string{"id": uuid.New().String()})
	handler("topic", payload)
	handler("topic", payload) // second call still runs (no dedup without db)

	if calls != 2 {
		t.Errorf("calls = %d, want 2 (nil db is transparent)", calls)
	}
}

func TestIdempotent_NilDB_BadPayload_StillRunsHandler(t *testing.T) {
	calls := 0
	handler := eventbus.Idempotent("test-consumer", nil, func(_ string, _ []byte) {
		calls++
	})
	handler("topic", []byte("not-json"))

	if calls != 1 {
		t.Errorf("calls = %d, want 1", calls)
	}
}
