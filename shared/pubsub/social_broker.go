package pubsub

import (
	"sync"
	"time"
)

// socialCacheTTL keeps a published payload for late SSE subscribers (reconnect window).
const socialCacheTTL = 2 * time.Minute

// SocialBroker is the backend-agnostic interface for social check pub/sub,
// keyed by profile ID. Mirrors CheckBroker; a Redis implementation can be
// added later for multi-instance deployments.
type SocialBroker interface {
	Subscribe(profileID string) (<-chan []byte, func())
	Publish(profileID string, payload []byte)
}

// MemorySocialBroker is an in-process SocialBroker with a short replay cache.
type MemorySocialBroker struct {
	mu        sync.Mutex
	listeners map[string][]chan []byte
	cache     map[string]cachedPayload
}

// NewSocialBroker creates an in-memory SocialBroker and starts cache eviction.
func NewSocialBroker() *MemorySocialBroker {
	b := &MemorySocialBroker{
		listeners: make(map[string][]chan []byte),
		cache:     make(map[string]cachedPayload),
	}
	go b.evictSocialLoop()
	return b
}

func (b *MemorySocialBroker) evictSocialLoop() {
	ticker := time.NewTicker(socialCacheTTL)
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

// Subscribe registers a listener for profileID, replaying any cached payload.
func (b *MemorySocialBroker) Subscribe(profileID string) (<-chan []byte, func()) {
	ch := make(chan []byte, 2)
	b.mu.Lock()
	b.listeners[profileID] = append(b.listeners[profileID], ch)
	if cached, ok := b.cache[profileID]; ok && time.Now().Before(cached.exp) {
		ch <- cached.data
	}
	b.mu.Unlock()

	return ch, func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		list := b.listeners[profileID]
		for i, c := range list {
			if c == ch {
				b.listeners[profileID] = append(list[:i], list[i+1:]...)
				break
			}
		}
		if len(b.listeners[profileID]) == 0 {
			delete(b.listeners, profileID)
		}
		close(ch)
	}
}

// Publish sends payload to all listeners and caches it for late subscribers.
func (b *MemorySocialBroker) Publish(profileID string, payload []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cache[profileID] = cachedPayload{data: payload, exp: time.Now().Add(socialCacheTTL)}
	for _, ch := range b.listeners[profileID] {
		select {
		case ch <- payload:
		default: // slow subscriber — skip
		}
	}
}
