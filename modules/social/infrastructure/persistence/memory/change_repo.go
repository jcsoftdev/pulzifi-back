package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories"
)

// ChangeRepo is an in-memory implementation of repositories.ChangeRepository.
type ChangeRepo struct {
	mu      sync.RWMutex
	changes map[uuid.UUID]*entities.SocialChange
	// profileToWorkspace is populated by tests that need workspace-level filtering.
	// In production this is handled by a JOIN with social_profiles.
	ProfileToWorkspace map[uuid.UUID]uuid.UUID
}

func NewChangeRepo() *ChangeRepo {
	return &ChangeRepo{
		changes:            make(map[uuid.UUID]*entities.SocialChange),
		ProfileToWorkspace: make(map[uuid.UUID]uuid.UUID),
	}
}

func (r *ChangeRepo) Save(ctx context.Context, change *entities.SocialChange) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *change
	r.changes[change.ID] = &cp
	return nil
}

func (r *ChangeRepo) List(ctx context.Context, filter repositories.ChangeFilter) ([]*entities.SocialChange, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*entities.SocialChange
	for _, c := range r.changes {
		if filter.ProfileID != nil && c.ProfileID != *filter.ProfileID {
			continue
		}
		if filter.WorkspaceID != nil && filter.ProfileID == nil {
			wsID, ok := r.ProfileToWorkspace[c.ProfileID]
			if !ok || wsID != *filter.WorkspaceID {
				continue
			}
		}
		cp := *c
		result = append(result, &cp)
	}
	// Sort descending by created_at
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	// Apply offset + limit
	if filter.Offset > 0 {
		if filter.Offset >= len(result) {
			return []*entities.SocialChange{}, nil
		}
		result = result[filter.Offset:]
	}
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, nil
}

func (r *ChangeRepo) GetByID(ctx context.Context, id uuid.UUID) (*entities.SocialChange, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.changes[id]
	if !ok {
		return nil, nil //nolint:nilnil // nil means not found
	}
	cp := *c
	return &cp, nil
}
