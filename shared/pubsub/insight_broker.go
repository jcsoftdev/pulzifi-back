package pubsub

import (
	"sync"
	"time"
)

// cacheTTL is how long a published payload is kept for late subscribers.
const cacheTTL = 5 * time.Minute

// cachedPayload holds a payload and its expiry time.
type cachedPayload struct {
	data []byte
	exp  time.Time
}

// InsightBroker is a lightweight pub/sub broker that routes insight-ready
// notifications to SSE subscribers, keyed by check ID.
//
// It also keeps a short-lived replay cache: if a subscriber connects after
// the publish event (common when generation is fast), the cached payload is
// delivered immediately so the client never stalls until timeout.
type InsightBroker struct {
	mu        sync.Mutex
	listeners map[string][]chan []byte
	cache     map[string]cachedPayload
}

func NewInsightBroker() *InsightBroker {
	return &InsightBroker{
		listeners: make(map[string][]chan []byte),
		cache:     make(map[string]cachedPayload),
	}
}

// Subscribe registers a listener for the given checkID. It returns a receive
// channel and an unsubscribe function that must be deferred by the caller.
//
// If a payload was published for this checkID within cacheTTL, it is written
// into the channel immediately so the caller does not wait for a future publish.
func (b *InsightBroker) Subscribe(checkID string) (<-chan []byte, func()) {
	ch := make(chan []byte, 1)
	b.mu.Lock()

	if cached, ok := b.cache[checkID]; ok && time.Now().Before(cached.exp) {
		// Replay: deliver cached payload directly; no need to register as listener.
		ch <- cached.data
		b.mu.Unlock()
	} else {
		b.listeners[checkID] = append(b.listeners[checkID], ch)
		b.mu.Unlock()
	}

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
		close(ch)
	}
}

// Publish sends payload to every subscriber waiting on checkID and stores it
// in the replay cache for late subscribers.
func (b *InsightBroker) Publish(checkID string, payload []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Cache for late subscribers.
	b.cache[checkID] = cachedPayload{data: payload, exp: time.Now().Add(cacheTTL)}

	for _, ch := range b.listeners[checkID] {
		select {
		case ch <- payload:
		default: // subscriber is too slow — channel is buffered(1), skip
		}
	}
}
