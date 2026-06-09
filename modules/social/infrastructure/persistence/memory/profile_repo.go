package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/social/domain/errors"
)

// ProfileRepo is an in-memory implementation of repositories.ProfileRepository.
// It is not safe for concurrent scheduler-style use (FOR UPDATE SKIP LOCKED)
// but is sufficient for unit tests.
type ProfileRepo struct {
	mu       sync.RWMutex
	profiles map[uuid.UUID]*entities.SocialProfile
}

func NewProfileRepo() *ProfileRepo {
	return &ProfileRepo{profiles: make(map[uuid.UUID]*entities.SocialProfile)}
}

func (r *ProfileRepo) Create(ctx context.Context, profile *entities.SocialProfile) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range r.profiles {
		if p.WorkspaceID == profile.WorkspaceID &&
			p.Platform == profile.Platform &&
			p.Handle == profile.Handle {
			return domainerrors.ErrProfileAlreadyExists
		}
	}
	cp := *profile
	r.profiles[profile.ID] = &cp
	return nil
}

func (r *ProfileRepo) GetByID(ctx context.Context, id uuid.UUID) (*entities.SocialProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.profiles[id]
	if !ok {
		return nil, domainerrors.ErrProfileNotFound
	}
	cp := *p
	return &cp, nil
}

func (r *ProfileRepo) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entities.SocialProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*entities.SocialProfile
	for _, p := range r.profiles {
		if p.WorkspaceID == workspaceID {
			cp := *p
			result = append(result, &cp)
		}
	}
	return result, nil
}

func (r *ProfileRepo) Update(ctx context.Context, profile *entities.SocialProfile) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.profiles[profile.ID]; !ok {
		return domainerrors.ErrProfileNotFound
	}
	cp := *profile
	r.profiles[profile.ID] = &cp
	return nil
}

func (r *ProfileRepo) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.profiles[id]; !ok {
		return domainerrors.ErrProfileNotFound
	}
	delete(r.profiles, id)
	return nil
}

func (r *ProfileRepo) CountActiveByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, p := range r.profiles {
		if p.WorkspaceID == workspaceID && p.IsActive {
			count++
		}
	}
	return count, nil
}

func (r *ProfileRepo) ListDue(ctx context.Context, now time.Time, limit int) ([]*entities.SocialProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*entities.SocialProfile
	for _, p := range r.profiles {
		if p.IsActive && p.NextCheckAt != nil && !p.NextCheckAt.After(now) {
			cp := *p
			result = append(result, &cp)
			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}
