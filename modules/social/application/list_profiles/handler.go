package listprofiles

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories"
)

// Handler handles the list-social-profiles use case.
type Handler struct {
	profiles repositories.ProfileRepository
}

// NewHandler creates a new Handler.
func NewHandler(profiles repositories.ProfileRepository) *Handler {
	return &Handler{profiles: profiles}
}

// Handle returns all social profiles for the given workspace.
func (h *Handler) Handle(ctx context.Context, workspaceID uuid.UUID) ([]*ProfileSummary, error) {
	profiles, err := h.profiles.ListByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	result := make([]*ProfileSummary, 0, len(profiles))
	for _, p := range profiles {
		result = append(result, toSummary(p))
	}
	return result, nil
}

func toSummary(p *entities.SocialProfile) *ProfileSummary {
	return &ProfileSummary{
		ID:                   p.ID,
		WorkspaceID:          p.WorkspaceID,
		Platform:             string(p.Platform),
		Handle:               p.Handle,
		DisplayName:          p.DisplayName,
		AvatarURL:            p.AvatarURL,
		IsActive:             p.IsActive,
		CheckIntervalMinutes: p.CheckIntervalMinutes,
		NextCheckAt:          p.NextCheckAt,
		LastCheckedAt:        p.LastCheckedAt,
		ConsecutiveFailures:  p.ConsecutiveFailures,
		CreatedAt:            p.CreatedAt,
		UpdatedAt:            p.UpdatedAt,
	}
}
