package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
)

// ChangeFilter controls which changes are returned by ListChanges.
type ChangeFilter struct {
	// ProfileID, when set, restricts results to a single profile.
	ProfileID *uuid.UUID

	// WorkspaceID, when set, restricts results to all profiles in a workspace.
	// Ignored when ProfileID is set.
	WorkspaceID *uuid.UUID

	// Limit caps the number of results. 0 means no cap (implementations may
	// apply a server-side default).
	Limit int

	// Offset enables page-based pagination.
	Offset int
}

// ChangeRepository defines persistence operations for SocialChange.
type ChangeRepository interface {
	// Save persists a new SocialChange.
	Save(ctx context.Context, change *entities.SocialChange) error

	// List returns changes matching the filter in descending created_at order.
	List(ctx context.Context, filter ChangeFilter) ([]*entities.SocialChange, error)

	// GetByID retrieves a change by its UUID. Returns nil when not found.
	GetByID(ctx context.Context, id uuid.UUID) (*entities.SocialChange, error)
}
