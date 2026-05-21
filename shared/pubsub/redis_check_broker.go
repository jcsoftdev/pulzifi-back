package pubsub

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	checkCacheKey      = "check:cache:"
	checkChannelKey    = "check:channel:"
	checkCacheRedisTTL = 2 * time.Minute
)

// RedisCheckBroker is a Redis-backed CheckBroker for multi-instance deployments.
//
// Architecture:
//   - One background goroutine drives a PSUBSCRIBE("check:channel:*") connection
//     and fans messages into a local listener map.
//   - Subscribe registers the listener FIRST, then checks Redis GET check:cache:<pageID>
//     for replay (so no live events are missed between the GET and the registration).
//   - Publish sets check:cache:<pageID> EX 120 and PUBLISHes to check:channel:<pageID>.
//   - Reconnect uses exponential backoff (1s → 30s) on PubSub receive error.
//
// Channel lifecycle: channels are NOT closed by the broker. Callers must drain via
// their own context cancellation. This avoids send-on-closed-channel races.
type RedisCheckBroker struct {
	rdb *redis.Client
	ctx context.Context

	mu        sync.RWMutex
	listeners map[string][]chan []byte
}

// NewRedisCheckBroker creates a Redis-backed CheckBroker and starts the
// background fan-out goroutine. Shutdown is controlled via ctx.
func NewRedisCheckBroker(ctx context.Context, rdb *redis.Client) *RedisCheckBroker {
	b := &RedisCheckBroker{
		rdb:       rdb,
		ctx:       ctx,
		listeners: make(map[string][]chan []byte),
	}
	go b.listenLoop()
	return b
}

// Subscribe registers a listener for the given pageID.
// On cache hit: delivers cached payload AND registers listener for future events.
// On cache miss: registers listener only.
// The listener is registered BEFORE checking the cache so no live events are lost.
//
// The returned channel is never closed by the broker — callers must use their own
// context cancellation and the provided unsubscribe func to signal termination.
func (b *RedisCheckBroker) Subscribe(pageID string) (<-chan []byte, func()) {
	ch := make(chan []byte, 2)

	b.mu.Lock()
	b.listeners[pageID] = append(b.listeners[pageID], ch)
	b.mu.Unlock()

	// Replay cache AFTER registering so no live PUBLISH is missed between the GET and the registration.
	if data, err := b.rdb.Get(b.ctx, checkCacheKey+pageID).Bytes(); err == nil {
		select {
		case ch <- data:
		default:
		}
	}

	return ch, func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		list := b.listeners[pageID]
		for i, c := range list {
			if c == ch {
				b.listeners[pageID] = append(list[:i], list[i+1:]...)
				break
			}
		}
		if len(b.listeners[pageID]) == 0 {
			delete(b.listeners, pageID)
		}
		// Channel is intentionally NOT closed here to avoid send-on-closed races
		// with the background goroutine. SSE handlers terminate via context cancellation.
	}
}

// Publish caches the payload in Redis and publishes to the Redis channel.
// The background goroutine receives the message and fans it to local listeners.
func (b *RedisCheckBroker) Publish(pageID string, payload []byte) {
	b.rdb.Set(b.ctx, checkCacheKey+pageID, payload, checkCacheRedisTTL)
	b.rdb.Publish(b.ctx, checkChannelKey+pageID, payload)
}

// listenLoop maintains the Redis PSUBSCRIBE connection and fans messages
// into local listener channels. On error, reconnects with exponential backoff.
func (b *RedisCheckBroker) listenLoop() {
	backoff := time.Second
	const maxBackoff = 30 * time.Second

	for {
		select {
		case <-b.ctx.Done():
			return
		default:
		}

		pubsub := b.rdb.PSubscribe(b.ctx, checkChannelKey+"*")
		if err := b.runReceive(pubsub); err != nil {
			select {
			case <-b.ctx.Done():
				pubsub.Close()
				return
			default:
			}
			pubsub.Close()
			// Exponential backoff before reconnect.
			time.Sleep(backoff)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		} else {
			pubsub.Close()
			return
		}
	}
}

// runReceive reads messages from the PubSub connection and delivers them to
// local listeners. Returns non-nil error on receive failure (triggers reconnect).
func (b *RedisCheckBroker) runReceive(pubsub *redis.PubSub) error {
	ch := pubsub.Channel()
	for {
		select {
		case <-b.ctx.Done():
			return nil
		case msg, ok := <-ch:
			if !ok {
				return errPubSubClosed
			}
			// Extract pageID from "check:channel:<pageID>".
			pageID := msg.Channel[len(checkChannelKey):]
			payload := []byte(msg.Payload)

			b.mu.RLock()
			listeners := make([]chan []byte, len(b.listeners[pageID]))
			copy(listeners, b.listeners[pageID])
			b.mu.RUnlock()

			for _, listener := range listeners {
				select {
				case listener <- payload:
				default: // subscriber is too slow — skip
				}
			}
		}
	}
}

// errPubSubClosed is returned when the PubSub channel closes unexpectedly.
var errPubSubClosed = pubSubClosedErr{}

type pubSubClosedErr struct{}

func (pubSubClosedErr) Error() string { return "redis pubsub channel closed" }
