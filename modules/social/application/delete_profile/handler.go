package deleteprofile

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories"
)

// Handler handles the delete-social-profile use case.
type Handler struct {
	profiles repositories.ProfileRepository
}

// NewHandler creates a new Handler.
func NewHandler(profiles repositories.ProfileRepository) *Handler {
	return &Handler{profiles: profiles}
}

// Handle deletes the social profile with the given ID.
// Associated snapshots and changes are removed by ON DELETE CASCADE at DB level.
func (h *Handler) Handle(ctx context.Context, id uuid.UUID) error {
	return h.profiles.Delete(ctx, id)
}
