# EventBus Package (`shared/eventbus/`)

In-memory pub/sub event bus, designed to be swappable for Kafka in production.

## Files

- `eventbus.go` — Event bus interface and in-memory implementation

## Exported API

### Interfaces
- `MessageBus` — Event bus abstraction:
  - `Publish(topic, key string, payload []byte) error`
  - `Subscribe(topic string, handler EventHandler) error`
  - `Close()`

### Types
- `EventHandler` — `func(topic string, payload []byte)`

### Structs
- `EventBus` — In-memory implementation with thread-safe handler map (`sync.RWMutex`)

### Functions
- `GetInstance() *EventBus` — Returns singleton (lazy-initialized with `sync.Once`)

### Methods (`*EventBus`)
- `Publish(topic, key, payload) error` — Sends to all topic handlers. Each handler runs in a separate goroutine with panic recovery. `key` parameter unused in in-memory impl (exists for Kafka compatibility).
- `Subscribe(topic, handler) error` — Registers handler for topic
- `Close()` — No-op for in-memory bus

## Usage

Currently only used by the `organization` module:
- Publisher: `modules/organization/infrastructure/messaging/publisher.go`
- Subscriber: `modules/organization/infrastructure/messaging/subscriber.go`
- Event: `organization.created`

## Notes

- Singleton pattern — call `GetInstance()` to get the shared bus
- Handlers execute asynchronously in goroutines
- Panics in handlers are recovered and logged (do not crash the process)

## Architecture Improvements

### Kafka Migration
The `MessageBus` interface is already designed for a Kafka adapter swap. To implement:
1. Create `kafka_bus.go` implementing `MessageBus` with `confluent-kafka-go` or `segmentio/kafka-go`
2. Use the `key` parameter (currently unused in in-memory impl) for Kafka partition routing
3. Implement `Close()` to flush and disconnect
4. Wire via dependency injection in `cmd/server/main.go` based on config flag (e.g., `EVENT_BUS_PROVIDER=kafka|memory`)

### Multi-Instance Scaling
The in-memory singleton is **node-local** — events published on one instance are invisible to other instances. This blocks horizontal scaling for any feature that relies on event-driven communication. Kafka (or Redis Pub/Sub as a lighter alternative) is required for multi-instance deployments.

### Durability
Events are lost on process restart. For critical workflows (e.g., tenant provisioning), consider:
- Adding an outbox table pattern for guaranteed delivery
- Implementing dead-letter queue for failed handler executions
- Adding retry logic (currently fire-and-forget with panic recovery only)
