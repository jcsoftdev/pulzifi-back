package inmem

import (
	"context"
	"encoding/json"
	"sort"
	"sync"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/repositories"
)

var jsonUnmarshal = json.Unmarshal

func sortByReceived(events []*entities.WebhookEvent) {
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].ReceivedAt.Before(events[j].ReceivedAt)
	})
}

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

// FindDeferredByCustomer scans the in-memory store for deferred events whose
// raw_payload's data.object.customer field matches the given customer ID.
// Returns events ordered by ReceivedAt ascending.
func (r *WebhookEventRepo) FindDeferredByCustomer(ctx context.Context, customerID string) ([]*entities.WebhookEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*entities.WebhookEvent
	for _, ev := range r.events {
		if ev.Status != entities.WebhookEventDeferred {
			continue
		}
		if extractPayloadCustomer(ev.RawPayload) != customerID {
			continue
		}
		cp := *ev
		out = append(out, &cp)
	}
	sortByReceived(out)
	return out, nil
}

func extractPayloadCustomer(payload []byte) string {
	if len(payload) == 0 {
		return ""
	}
	var wrapper struct {
		Data struct {
			Object struct {
				Customer string `json:"customer"`
			} `json:"object"`
		} `json:"data"`
	}
	if err := jsonUnmarshal(payload, &wrapper); err != nil {
		return ""
	}
	return wrapper.Data.Object.Customer
}
