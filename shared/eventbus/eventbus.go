package eventbus

import (
	"sync"

	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)



// MessageBus defines the interface for publishing and subscribing to events
// This abstraction allows swapping between In-Memory (MVP) and Kafka (Production) implementations
type MessageBus interface {
	Publish(topic string, key string, payload []byte) error
	Subscribe(topic string, handler EventHandler) error
	Close()
}

// EventHandler is a function that handles an event
type EventHandler func(topic string, payload []byte)

// EventBus is a simple in-memory event bus implementing MessageBus
type EventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

var (
	instance *EventBus
	once     sync.Once
)

// GetInstance returns the singleton instance of EventBus
func GetInstance() *EventBus {
	once.Do(func() {
		instance = &EventBus{
			handlers: make(map[string][]EventHandler),
		}
	})
	return instance
}

// Publish sends an event to all handlers for a topic (In-Memory implementation)
func (eb *EventBus) Publish(topic string, key string, payload []byte) error {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	handlers, ok := eb.handlers[topic]
	if !ok {
		logger.Debug("No handlers for topic (In-Memory)", zap.String("topic", topic))
		return nil
	}

	logger.Info("Publishing event (In-Memory)", zap.String("topic", topic), zap.Int("handler_count", len(handlers)))

	for _, handler := range handlers {
		go func(h EventHandler) {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Panic in event handler", zap.Any("recover", r))
				}
			}()
			h(topic, payload)
		}(handler)
	}
	return nil
}

// Subscribe adds a handler for a topic (In-Memory implementation)
func (eb *EventBus) Subscribe(topic string, handler EventHandler) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers[topic] = append(eb.handlers[topic], handler)
	logger.Info("Subscribed to topic (In-Memory)", zap.String("topic", topic))
	return nil
}

// Close cleans up resources (No-op for In-Memory)
func (eb *EventBus) Close() {
	// Nothing to close for in-memory bus
}
