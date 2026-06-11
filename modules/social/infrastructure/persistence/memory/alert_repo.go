package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories"
)

// Compile-time interface check.
var _ repositories.AlertRepository = (*AlertRepository)(nil)

// AlertRepository is an in-memory AlertRepository for tests.
type AlertRepository struct {
	mu     sync.Mutex
	alerts []*entities.SocialAlert
}

// NewAlertRepository creates an empty in-memory AlertRepository.
func NewAlertRepository() *AlertRepository {
	return &AlertRepository{}
}

func (r *AlertRepository) Save(_ context.Context, alert *entities.SocialAlert) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.alerts = append(r.alerts, alert)
	return nil
}

func (r *AlertRepository) ListByWorkspace(_ context.Context, workspaceID uuid.UUID, limit, offset int) ([]*entities.SocialAlert, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*entities.SocialAlert
	for _, a := range r.alerts {
		if a.WorkspaceID == workspaceID {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	if offset >= len(out) {
		return nil, nil
	}
	end := len(out)
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	return out[offset:end], nil
}

func (r *AlertRepository) MarkRead(_ context.Context, alertID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, a := range r.alerts {
		if a.ID == alertID {
			a.IsRead = true
		}
	}
	return nil
}

func (r *AlertRepository) UnreadCount(_ context.Context, workspaceID uuid.UUID) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for _, a := range r.alerts {
		if a.WorkspaceID == workspaceID && !a.IsRead {
			count++
		}
	}
	return count, nil
}
