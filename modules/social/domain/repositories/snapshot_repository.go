package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
)

// SnapshotRepository defines persistence operations for SocialSnapshot.
type SnapshotRepository interface {
	// Save persists a new snapshot (success or failed).
	Save(ctx context.Context, snapshot *entities.SocialSnapshot) error

	// GetLatestByProfile returns the most recent snapshot for a profile,
	// or nil when no snapshot exists yet (first-run / baseline case).
	GetLatestByProfile(ctx context.Context, profileID uuid.UUID) (*entities.SocialSnapshot, error)

	// GetByID retrieves a snapshot by its UUID. Returns nil when not found.
	GetByID(ctx context.Context, id uuid.UUID) (*entities.SocialSnapshot, error)
}
