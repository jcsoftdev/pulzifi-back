package updateprofile

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/services"
)

// Minimum cadence is 12h to keep scraping costs bounded.
var allowedPresets = map[int]bool{
	720:  true,
	1440: true,
}

// Handler handles the update-social-profile use case.
type Handler struct {
	profiles   repositories.ProfileRepository
	planLimits services.PlanLimits
}

// NewHandler creates a new Handler.
func NewHandler(profiles repositories.ProfileRepository, planLimits services.PlanLimits) *Handler {
	return &Handler{profiles: profiles, planLimits: planLimits}
}

// Handle applies partial updates to a social profile.
func (h *Handler) Handle(ctx context.Context, id uuid.UUID, req *Request) (*Response, error) {
	profile, err := h.profiles.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// --- Apply interval change ---
	if req.CheckIntervalMinutes != nil {
		newInterval := *req.CheckIntervalMinutes
		if !allowedPresets[newInterval] {
			return nil, fmt.Errorf(
				"invalid check_interval_minutes %d: must be one of 720, 1440",
				newInterval,
			)
		}
		// Validate interval vs plan (REQ-PROFILE-08, same logic as create)
		checksPerDay, err := h.planLimits.GetChecksPerDay(ctx, profile.WorkspaceID.String())
		if err != nil {
			return nil, fmt.Errorf("fetching plan checks per day: %w", err)
		}
		if checksPerDay > 0 {
			activeCount, err := h.profiles.CountActiveByWorkspace(ctx, profile.WorkspaceID)
			if err != nil {
				return nil, fmt.Errorf("counting active profiles: %w", err)
			}
			// Pessimistic: count this profile as active even if currently inactive.
			projected := activeCount * (1440 / newInterval)
			if projected > checksPerDay {
				return nil, fmt.Errorf(
					"interval %d min would exceed plan capacity (%d checks/day with %d active profiles)",
					newInterval, checksPerDay, activeCount,
				)
			}
		}
		profile.CheckIntervalMinutes = newInterval
	}

	// --- Apply active state change ---
	if req.IsActive != nil {
		profile.IsActive = *req.IsActive
		if !*req.IsActive {
			// Deactivate: remove from scheduler (REQ-PROFILE-09)
			profile.NextCheckAt = nil
		} else {
			// Re-activate: reset next_check_at to now (REQ-PROFILE-09)
			now := time.Now().UTC()
			profile.NextCheckAt = &now
		}
	}

	profile.UpdatedAt = time.Now().UTC()

	if err := h.profiles.Update(ctx, profile); err != nil {
		return nil, err
	}

	return toResponse(profile), nil
}

func toResponse(p *entities.SocialProfile) *Response {
	return &Response{
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
