# PubSub Package (`shared/pubsub/`)

SSE (Server-Sent Events) brokers for real-time push notifications.

## Files

- `check_broker.go` — Page check status notifications (2-minute cache TTL)
- `insight_broker.go` — Insight generation completion notifications (5-minute cache TTL)

## Exported API

### CheckBroker (`check_broker.go`)
- `CheckBroker` — Pub/sub broker keyed by page ID
- `NewCheckBroker() *CheckBroker` — Creates broker, starts background cache eviction
- `Subscribe(pageID) (<-chan []byte, func())` — Returns receive channel (buffer 2) + unsubscribe function. **Always registers as listener** even on cache hit. Replays cached payload if available.
- `Publish(pageID, payload)` — Sends to all listeners, caches for late subscribers (2-min TTL). Non-blocking.

### InsightBroker (`insight_broker.go`)
- `InsightBroker` — Pub/sub broker keyed by check ID
- `NewInsightBroker() *InsightBroker` — Creates broker, starts background cache eviction
- `Subscribe(checkID) (<-chan []byte, func())` — Returns receive channel (buffer 1) + unsubscribe function. **One-shot delivery:** if cached payload exists, delivers immediately without registering as listener.
- `Publish(checkID, payload)` — Sends to all listeners, caches for late subscribers (5-min TTL). Non-blocking.

## Key Behavioral Difference

| Broker | On Cache Hit | Use Case |
|--------|-------------|----------|
| `CheckBroker` | Replays cache AND registers listener | Long-lived SSE connections need both cached "pending" and future "success"/"error" events |
| `InsightBroker` | Delivers cache, does NOT register listener | One-shot: insight generation result is final, no further updates needed |

## Usage

- `CheckBroker` — Used by monitoring module SSE endpoint (`GET /monitoring/checks/page/{pageId}/stream`)
- `InsightBroker` — Used by insight module SSE endpoint (`GET /insights/sse`)

## Notes

- Both brokers use non-blocking sends (slow subscribers are skipped)
- Background eviction loop runs on each broker's TTL interval to clean expired cache entries
- Unsubscribe function removes the channel from the listener list

## Architecture Improvements

### Redis-Backed Brokers for Horizontal Scaling
Both brokers are **node-local** in-memory — SSE clients connected to instance A will not receive events published on instance B. For multi-instance deployments:
1. Replace in-memory maps with Redis Pub/Sub channels (`SUBSCRIBE`/`PUBLISH`)
2. Keep the local channel fan-out (Go channels for SSE connections), but source events from Redis instead of direct in-memory publish
3. Use Redis key-value with TTL for the cache layer (replaces the in-memory cache maps)
4. Define a `Broker` interface so in-memory and Redis implementations are swappable

### Cache Consistency
The cache replay mechanism (delivering cached payloads to late subscribers) works well for single-instance but needs Redis `GET`/`SETEX` for consistency across instances.

### Slow Subscriber Handling
Currently non-blocking sends silently drop messages to slow subscribers. Consider:
- Logging dropped messages for observability
- Increasing channel buffer sizes for bursty workloads
- Adding backpressure metrics (count of dropped messages per subscriber)
