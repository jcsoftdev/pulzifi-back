package listsnapshots

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories"
)

// Handler handles the list-social-snapshots use case.
type Handler struct {
	snapshots repositories.SnapshotRepository
}

// NewHandler creates a new Handler.
func NewHandler(snapshots repositories.SnapshotRepository) *Handler {
	return &Handler{snapshots: snapshots}
}

// Handle returns a paginated list of snapshots for a profile, ordered newest first.
func (h *Handler) Handle(ctx context.Context, profileID uuid.UUID, limit, offset int) (*Response, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	snaps, err := h.snapshots.ListByProfile(ctx, profileID, limit, offset)
	if err != nil {
		return nil, err
	}

	items := make([]SnapshotSummary, 0, len(snaps))
	for _, s := range snaps {
		items = append(items, SnapshotSummary{
			ID:             s.ID,
			ProfileID:      s.ProfileID,
			CapturedAt:     s.CapturedAt,
			Status:         string(s.Status),
			Error:          s.Error,
			FollowersCount: s.FollowersCount,
			PostsCount:     s.PostsCount,
		})
	}

	return &Response{Snapshots: items}, nil
}
