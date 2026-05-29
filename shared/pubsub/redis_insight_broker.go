package pubsub

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	insightCacheKey      = "insight:cache:"
	insightChannelKey    = "insight:channel:"
	insightCacheRedisTTL = 5 * time.Minute
)

// RedisInsightBroker is a Redis-backed InsightBroker for multi-instance deployments.
//
// One-shot semantics:
//   - On cache hit (GET insight:cache:<checkID> succeeds): deliver payload directly,
//     close channel, and do NOT subscribe to the Redis pub/sub channel.
//   - On cache miss: subscribe to Redis pub/sub channel, await one message, deliver,
//     then close the goroutine (unsubscribe from Redis).
//
// This preserves the MemoryInsightBroker contract: insight completion is terminal —
// there is exactly one event per check, and it must be delivered exactly once.
//
// Channel lifecycle: channels are NOT closed by the broker after delivery to avoid
// send-on-closed races. Callers terminate via context or unsubscribe func.
type RedisInsightBroker struct {
	rdb *redis.Client
	ctx context.Context

	mu        sync.RWMutex
	listeners map[string][]chan []byte
}

// NewRedisInsightBroker creates a Redis-backed InsightBroker and starts the
// background fan-out goroutine. Shutdown is controlled via ctx.
func NewRedisInsightBroker(ctx context.Context, rdb *redis.Client) *RedisInsightBroker {
	b := &RedisInsightBroker{
		rdb:       rdb,
		ctx:       ctx,
		listeners: make(map[string][]chan []byte),
	}
	go b.listenLoop()
	return b
}

// Subscribe registers a one-shot listener for the given checkID.
//
// Cache hit path: payload delivered immediately, no listener registered (one-shot).
// Cache miss path: listener registered; the background goroutine delivers on next PUBLISH.
func (b *RedisInsightBroker) Subscribe(checkID string) (<-chan []byte, func()) {
	ch := make(chan []byte, 1)

	// Check cache first. On cache hit: deliver and return — do NOT register as listener.
	if data, err := b.rdb.Get(b.ctx, insightCacheKey+checkID).Bytes(); err == nil {
		select {
		case ch <- data:
		default:
		}
		return ch, func() {}
	}

	// Cache miss: register as listener for the live PUBLISH.
	b.mu.Lock()
	b.listeners[checkID] = append(b.listeners[checkID], ch)
	b.mu.Unlock()

	return ch, func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		list := b.listeners[checkID]
		for i, c := range list {
			if c == ch {
				b.listeners[checkID] = append(list[:i], list[i+1:]...)
				break
			}
		}
		if len(b.listeners[checkID]) == 0 {
			delete(b.listeners, checkID)
		}
		// Channel intentionally NOT closed — callers use context/select for termination.
	}
}

// Publish caches the payload in Redis and publishes to the Redis channel.
func (b *RedisInsightBroker) Publish(checkID string, payload []byte) {
	b.rdb.Set(b.ctx, insightCacheKey+checkID, payload, insightCacheRedisTTL)
	b.rdb.Publish(b.ctx, insightChannelKey+checkID, payload)
}

// listenLoop maintains the Redis PSUBSCRIBE connection. Reconnects with
// exponential backoff on error.
func (b *RedisInsightBroker) listenLoop() {
	backoff := time.Second
	const maxBackoff = 30 * time.Second

	for {
		select {
		case <-b.ctx.Done():
			return
		default:
		}

		pubsub := b.rdb.PSubscribe(b.ctx, insightChannelKey+"*")
		if err := b.runInsightReceive(pubsub); err != nil {
			select {
			case <-b.ctx.Done():
				_ = pubsub.Close()
				return
			default:
			}
			_ = pubsub.Close()
			time.Sleep(backoff)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		} else {
			_ = pubsub.Close()
			return
		}
	}
}

// runInsightReceive reads messages from the PubSub connection and delivers them
// to local listeners (one-shot per listener). Listeners are removed after delivery.
func (b *RedisInsightBroker) runInsightReceive(pubsub *redis.PubSub) error {
	ch := pubsub.Channel()
	for {
		select {
		case <-b.ctx.Done():
			return nil
		case msg, ok := <-ch:
			if !ok {
				return errInsightPubSubClosed
			}
			// Extract checkID from "insight:channel:<checkID>".
			checkID := msg.Channel[len(insightChannelKey):]
			payload := []byte(msg.Payload)

			// Deliver to all registered listeners, then remove them (one-shot).
			b.mu.Lock()
			listeners := b.listeners[checkID]
			delete(b.listeners, checkID)
			b.mu.Unlock()

			for _, listener := range listeners {
				select {
				case listener <- payload:
				default: // subscriber is too slow — skip
				}
			}
		}
	}
}

var errInsightPubSubClosed = insightPubSubClosedErr{}

type insightPubSubClosedErr struct{}

func (insightPubSubClosedErr) Error() string { return "insight redis pubsub channel closed" }
