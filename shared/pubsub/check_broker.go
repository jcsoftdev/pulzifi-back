package pubsub

import (
	"sync"
	"time"
)

// checkCacheTTL is how long a published payload is kept for late subscribers.
const checkCacheTTL = 2 * time.Minute

// CheckBroker is a lightweight pub/sub broker that routes check-status
// notifications to SSE subscribers, keyed by page ID.
//
// It keeps a short-lived replay cache so that if the frontend's EventSource
// connects after a "pending" event was published, the cached payload is
// delivered immediately. Unlike InsightBroker, subscribers are always
// registered as listeners (even on cache hit) because the SSE handler is
// long-lived and needs both the cached "pending" event AND future
// "success"/"error" events.
type CheckBroker struct {
	mu        sync.Mutex
	listeners map[string][]chan []byte // keyed by pageID
	cache     map[string]cachedPayload
}

func NewCheckBroker() *CheckBroker {
	b := &CheckBroker{
		listeners: make(map[string][]chan []byte),
		cache:     make(map[string]cachedPayload),
	}
	go b.evictLoop()
	return b
}

// evictLoop periodically removes expired entries from the replay cache.
func (b *CheckBroker) evictLoop() {
	ticker := time.NewTicker(checkCacheTTL)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		b.mu.Lock()
		for k, v := range b.cache {
			if now.After(v.exp) {
				delete(b.cache, k)
			}
		}
		b.mu.Unlock()
	}
}

// Subscribe registers a listener for the given pageID. It returns a receive
// channel and an unsubscribe function that must be deferred by the caller.
//
// If a payload was recently published for this pageID, it is written into
// the channel immediately. The subscriber is always registered as a listener
// so it also receives future events (e.g. "pending" → "success" sequence).
func (b *CheckBroker) Subscribe(pageID string) (<-chan []byte, func()) {
	ch := make(chan []byte, 2)
	b.mu.Lock()

	// Always register as listener — long-lived SSE connections need future events too.
	b.listeners[pageID] = append(b.listeners[pageID], ch)

	// Replay cached payload if available.
	if cached, ok := b.cache[pageID]; ok && time.Now().Before(cached.exp) {
		ch <- cached.data
	}

	b.mu.Unlock()

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
		close(ch)
	}
}

// Publish sends payload to every subscriber waiting on pageID and stores it
// in the replay cache for late subscribers.
func (b *CheckBroker) Publish(pageID string, payload []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Cache for late subscribers.
	b.cache[pageID] = cachedPayload{data: payload, exp: time.Now().Add(checkCacheTTL)}

	for _, ch := range b.listeners[pageID] {
		select {
		case ch <- payload:
		default: // subscriber is too slow — channel is buffered(2), skip
		}
	}
}
