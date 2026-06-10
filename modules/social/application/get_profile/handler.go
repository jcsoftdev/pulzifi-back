package getprofile

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories"
)

// Handler handles the get-social-profile use case.
type Handler struct {
	profiles  repositories.ProfileRepository
	snapshots repositories.SnapshotRepository
}

// NewHandler creates a new Handler.
func NewHandler(profiles repositories.ProfileRepository, snapshots repositories.SnapshotRepository) *Handler {
	return &Handler{profiles: profiles, snapshots: snapshots}
}

// Handle returns the profile plus its latest snapshot (nil if no snapshot yet).
func (h *Handler) Handle(ctx context.Context, id uuid.UUID) (*Response, error) {
	profile, err := h.profiles.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	snapshot, err := h.snapshots.GetLatestByProfile(ctx, id)
	if err != nil {
		return nil, err
	}

	return &Response{
		Profile:        toProfileDetail(profile),
		LatestSnapshot: toSnapshotDetail(snapshot),
	}, nil
}

func toProfileDetail(p *entities.SocialProfile) *ProfileDetail {
	return &ProfileDetail{
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

func toSnapshotDetail(s *entities.SocialSnapshot) *SnapshotDetail {
	if s == nil {
		return nil
	}
	return &SnapshotDetail{
		ID:             s.ID,
		CapturedAt:     s.CapturedAt,
		Status:         string(s.Status),
		Error:          s.Error,
		FollowersCount: s.FollowersCount,
		PostsCount:     s.PostsCount,
		Data:           s.Data,
	}
}
