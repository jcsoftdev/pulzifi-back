package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
)

// SnapshotRepo is an in-memory implementation of repositories.SnapshotRepository.
type SnapshotRepo struct {
	mu        sync.RWMutex
	snapshots map[uuid.UUID]*entities.SocialSnapshot
}

func NewSnapshotRepo() *SnapshotRepo {
	return &SnapshotRepo{snapshots: make(map[uuid.UUID]*entities.SocialSnapshot)}
}

func (r *SnapshotRepo) Save(ctx context.Context, snapshot *entities.SocialSnapshot) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *snapshot
	r.snapshots[snapshot.ID] = &cp
	return nil
}

func (r *SnapshotRepo) GetLatestByProfile(ctx context.Context, profileID uuid.UUID) (*entities.SocialSnapshot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var latest *entities.SocialSnapshot
	for _, s := range r.snapshots {
		if s.ProfileID != profileID {
			continue
		}
		if latest == nil || s.CapturedAt.After(latest.CapturedAt) {
			latest = s
		}
	}
	if latest == nil {
		return nil, nil //nolint:nilnil // no snapshot yet = baseline case
	}
	cp := *latest
	return &cp, nil
}

func (r *SnapshotRepo) GetByID(ctx context.Context, id uuid.UUID) (*entities.SocialSnapshot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.snapshots[id]
	if !ok {
		return nil, nil //nolint:nilnil // nil means not found
	}
	cp := *s
	return &cp, nil
}
