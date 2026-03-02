package pubsub

import (
	"sync"
)

// CheckBroker is a lightweight pub/sub broker that routes check-status
// notifications to SSE subscribers, keyed by page ID.
//
// Unlike InsightBroker it has no replay cache — SSE connections are long-lived
// and EventSource auto-reconnects, so late delivery is not a concern.
type CheckBroker struct {
	mu        sync.Mutex
	listeners map[string][]chan []byte // keyed by pageID
}

func NewCheckBroker() *CheckBroker {
	return &CheckBroker{
		listeners: make(map[string][]chan []byte),
	}
}

// Subscribe registers a listener for the given pageID. It returns a receive
// channel and an unsubscribe function that must be deferred by the caller.
func (b *CheckBroker) Subscribe(pageID string) (<-chan []byte, func()) {
	ch := make(chan []byte, 1)
	b.mu.Lock()
	b.listeners[pageID] = append(b.listeners[pageID], ch)
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
	}
}

// Publish sends payload to every subscriber waiting on pageID.
func (b *CheckBroker) Publish(pageID string, payload []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, ch := range b.listeners[pageID] {
		select {
		case ch <- payload:
		default: // subscriber is too slow — channel is buffered(1), skip
		}
	}
}
