package inmem

import (
	"context"
	"sync"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/repositories"
)

// compile-time interface check
var _ repositories.WebhookEventRepository = (*WebhookEventRepo)(nil)

// WebhookEventRepo is a thread-safe in-memory implementation of WebhookEventRepository.
// It is intended for use in unit tests only.
type WebhookEventRepo struct {
	mu     sync.Mutex
	events map[string]*entities.WebhookEvent // keyed by EventID
}

// NewWebhookEventRepo returns an empty WebhookEventRepo.
func NewWebhookEventRepo() *WebhookEventRepo {
	return &WebhookEventRepo{
		events: make(map[string]*entities.WebhookEvent),
	}
}

// Save inserts the event if the event_id does not already exist.
// Returns (true, nil) on first insert; (false, nil) on duplicate.
func (r *WebhookEventRepo) Save(ctx context.Context, event *entities.WebhookEvent) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.events[event.EventID]; exists {
		return false, nil
	}
	cp := *event
	r.events[event.EventID] = &cp
	return true, nil
}

// MarkProcessed sets processed_at and status on the stored event.
func (r *WebhookEventRepo) MarkProcessed(ctx context.Context, eventID string, status entities.WebhookEventStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	ev, ok := r.events[eventID]
	if !ok {
		return nil // no-op; event not found (should not happen in normal flow)
	}
	now := time.Now()
	ev.ProcessedAt = &now
	ev.Status = status
	return nil
}
