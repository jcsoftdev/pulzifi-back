package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
)

// ProfileRepository defines persistence operations for SocialProfile.
// All implementations must be tenant-aware: the concrete adapter receives
// the tenant schema name in its constructor and calls
// middleware.GetSetSearchPathSQL before every query.
type ProfileRepository interface {
	// Create persists a new SocialProfile. Returns ErrProfileAlreadyExists
	// when the (workspace_id, platform, handle) unique constraint fires.
	Create(ctx context.Context, profile *entities.SocialProfile) error

	// GetByID retrieves a profile by its UUID. Returns ErrProfileNotFound
	// when no row matches.
	GetByID(ctx context.Context, id uuid.UUID) (*entities.SocialProfile, error)

	// ListByWorkspace returns all profiles belonging to the given workspace.
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entities.SocialProfile, error)

	// Update persists changes to an existing profile (is_active, interval,
	// next_check_at, last_checked_at, consecutive_failures, display_name,
	// avatar_url, updated_at).
	Update(ctx context.Context, profile *entities.SocialProfile) error

	// Delete removes a profile. Associated snapshots and changes are removed
	// by ON DELETE CASCADE at the database level.
	Delete(ctx context.Context, id uuid.UUID) error

	// CountActiveByWorkspace returns the number of active profiles in a workspace.
	// Used at creation time to enforce plan profile limits.
	CountActiveByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int, error)

	// ListDue returns profiles with is_active=true and next_check_at <= now,
	// using SELECT ... FOR UPDATE SKIP LOCKED to prevent duplicate processing
	// across concurrent worker instances. limit caps the batch size.
	ListDue(ctx context.Context, now time.Time, limit int) ([]*entities.SocialProfile, error)
}
