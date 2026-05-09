# EventBus Package (`shared/eventbus/`)

In-memory pub/sub event bus, designed to be swappable for Kafka in production.

## Files

- `eventbus.go` — `MessageBus` interface and in-memory implementation
- `event.go` — `DomainEvent` struct, topic constants, typed publish/subscribe helpers
- `event_test.go` — Tests for domain event helpers

## Exported API

### Interfaces
- `MessageBus` — Event bus abstraction:
  - `Publish(topic, key string, payload []byte) error`
  - `Subscribe(topic string, handler EventHandler) error`
  - `Close()`

### Types
- `EventHandler` — `func(topic string, payload []byte)`
- `DomainEventHandler` — `func(ev DomainEvent)`
- `DomainEvent` — Structured event: `ID`, `Type`, `OccurredAt`, `OrgID`, `WorkspaceID`, `PageID`, `Tenant`, `Data json.RawMessage`

### Topic Constants
- `TopicChangeDetected = "change.detected"`
- `TopicAlertCreated = "alert.created"`
- `TopicInsightReady = "insight.ready"`
- `TopicCheckFailed = "check.failed"`
- `TopicReportWeekly = "report.weekly"`

### Functions
- `GetInstance() *EventBus` — Returns singleton (lazy-initialized with `sync.Once`)
- `PublishDomainEvent(bus, ev DomainEvent) error` — Marshals and publishes; auto-fills `ID` and `OccurredAt` if zero
- `SubscribeDomainEvent(bus, topic, handler DomainEventHandler) error` — Wraps raw handler with JSON unmarshal

### Methods (`*EventBus`)
- `Publish(topic, key, payload) error` — Sends to all topic handlers asynchronously with panic recovery
- `Subscribe(topic, handler) error` — Registers handler for topic
- `Close()` — No-op for in-memory bus

## Usage

- `organization` module publishes `organization.created` via `infrastructure/messaging/publisher.go`
- `alert` module publishes `TopicAlertCreated` via `SetEventBus`
- `integration` module subscribes to `TopicChangeDetected` and `TopicAlertCreated` to dispatch webhook deliveries

## Notes

- Singleton pattern — call `GetInstance()` to get the shared bus
- Handlers execute asynchronously in goroutines; panics recovered and logged
- `key` parameter is unused in in-memory impl — reserved for Kafka partition routing

## Architecture Improvements

### Kafka Migration
The `MessageBus` interface is already designed for a Kafka adapter swap. To implement:
1. Create `kafka_bus.go` implementing `MessageBus` with `confluent-kafka-go` or `segmentio/kafka-go`
2. Use the `key` parameter (currently unused in in-memory impl) for Kafka partition routing
3. Implement `Close()` to flush and disconnect
4. Wire via dependency injection in `cmd/server/main.go` based on config flag (e.g., `EVENT_BUS_PROVIDER=kafka|memory`)
5. `DomainEvent.OrgID` and `Tenant` fields enable partition routing — no schema changes needed

### Multi-Instance Scaling
The in-memory singleton is **node-local** — events published on one instance are invisible to other instances. This blocks horizontal scaling for any feature that relies on event-driven communication. Kafka (or Redis Pub/Sub as a lighter alternative) is required for multi-instance deployments.

### Durability
Events are lost on process restart. For critical workflows (e.g., tenant provisioning), consider:
- Adding an outbox table pattern for guaranteed delivery
- Implementing dead-letter queue for failed handler executions
- Adding retry logic (currently fire-and-forget with panic recovery only)
