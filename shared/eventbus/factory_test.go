package eventbus_test

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq" // register postgres driver for sql.Open
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
)

func TestNewFromConfig_MemoryProvider(t *testing.T) {
	tests := []struct {
		name     string
		provider string
	}{
		{"empty string defaults to memory", ""},
		{"explicit memory", "memory"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bus, err := eventbus.NewFromConfig(tt.provider, 100, 8, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if bus == nil {
				t.Fatal("bus is nil")
			}
			// memory provider returns the singleton *EventBus
			if _, ok := bus.(*eventbus.EventBus); !ok {
				t.Errorf("expected *eventbus.EventBus, got %T", bus)
			}
		})
	}
}

func TestNewFromConfig_OutboxProvider_NilDB_ReturnsError(t *testing.T) {
	// The factory must reject outbox with nil db.
	_, err := eventbus.NewFromConfig("outbox", 100, 8, nil)
	if err == nil {
		t.Fatal("expected error for outbox with nil db")
	}
}

func TestNewFromConfig_OutboxProvider_ReturnsTxPublisher(t *testing.T) {
	// Verifies the factory returns a TxPublisher when provider=outbox and a
	// non-nil *sql.DB is provided. sql.Open does not connect eagerly, so we
	// can pass the resulting *sql.DB without a running postgres instance.
	// The returned bus must implement TxPublisher (confirms it is an outboxAdapter).
	db, err := sql.Open("postgres", "postgres://localhost/test?sslmode=disable")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer func() { _ = db.Close() }()

	bus, err := eventbus.NewFromConfig("outbox", 100, 8, db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := bus.(eventbus.TxPublisher); !ok {
		t.Errorf("expected bus to implement TxPublisher, got %T", bus)
	}
}

func TestNewFromConfig_UnknownProvider(t *testing.T) {
	_, err := eventbus.NewFromConfig("kafka", 100, 8, nil)
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}
