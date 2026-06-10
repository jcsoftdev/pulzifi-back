package createprofile

import (
	"context"
	"errors"
	"fmt"

	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/social/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/services"
	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"
)

// allowedPresets are the valid check_interval_minutes values (REQ-QUOTA-INTERVAL-02).
// Minimum cadence is 12h to keep scraping costs bounded.
var allowedPresets = map[int]bool{
	720:  true, // every 12h
	1440: true, // daily
}

// Handler orchestrates the create-social-profile use case.
type Handler struct {
	profiles   repositories.ProfileRepository
	planLimits services.PlanLimits
}

// NewHandler creates a new Handler.
func NewHandler(
	profiles repositories.ProfileRepository,
	planLimits services.PlanLimits,
) *Handler {
	return &Handler{profiles: profiles, planLimits: planLimits}
}

// Handle validates input, enforces plan constraints, and persists the new profile.
func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	// --- 1. Validate platform ---
	platform := valueobjects.Platform(req.Platform)
	if err := platform.Validate(); err != nil {
		return nil, err
	}

	// --- 2. Validate handle ---
	if req.Handle == "" {
		return nil, errors.New("handle must not be empty")
	}

	// --- 3. Validate interval preset (REQ-PROFILE-05) ---
	if !allowedPresets[req.CheckIntervalMinutes] {
		return nil, fmt.Errorf(
			"invalid check_interval_minutes %d: must be one of 720, 1440",
			req.CheckIntervalMinutes,
		)
	}

	// --- 4. Enforce plan profile limit (REQ-PROFILE-03) ---
	profilesLimit, err := h.planLimits.GetProfilesLimit(ctx, req.WorkspaceID.String())
	if err != nil {
		return nil, fmt.Errorf("fetching plan profiles limit: %w", err)
	}
	// -1 = feature disabled for plan
	if profilesLimit == -1 {
		return nil, domainerrors.ErrProfileLimitReached
	}
	// 0 = unlimited; skip count check
	if profilesLimit > 0 {
		activeCount, err := h.profiles.CountActiveByWorkspace(ctx, req.WorkspaceID)
		if err != nil {
			return nil, fmt.Errorf("counting active profiles: %w", err)
		}
		if activeCount >= profilesLimit {
			return nil, domainerrors.ErrProfileLimitReached
		}
	}

	// --- 5. Validate interval vs plan checks/day (REQ-PROFILE-06, REQ-QUOTA-INTERVAL-01) ---
	checksPerDay, err := h.planLimits.GetChecksPerDay(ctx, req.WorkspaceID.String())
	if err != nil {
		return nil, fmt.Errorf("fetching plan checks per day: %w", err)
	}
	// 0 = unlimited; skip validation. -1 = feature off (covered by profiles limit above).
	if checksPerDay > 0 {
		activeCount, err := h.profiles.CountActiveByWorkspace(ctx, req.WorkspaceID)
		if err != nil {
			return nil, fmt.Errorf("counting active profiles: %w", err)
		}
		// Pessimistic count: include the profile being created (REQ-QUOTA-INTERVAL-04)
		projected := (activeCount + 1) * (1440 / req.CheckIntervalMinutes)
		if projected > checksPerDay {
			return nil, buildIntervalError(activeCount+1, checksPerDay)
		}
	}

	// --- 6. Create and persist ---
	profile := entities.NewSocialProfile(
		req.WorkspaceID,
		platform,
		req.Handle,
		req.CheckIntervalMinutes,
	)

	if err := h.profiles.Create(ctx, profile); err != nil {
		return nil, err
	}

	return profileToResponse(profile), nil
}

// buildIntervalError returns an error that names the cheapest valid preset or
// states that no preset fits (REQ-QUOTA-INTERVAL-03).
func buildIntervalError(newActiveCount, checksPerDay int) error {
	// Try presets from cheapest (fewest checks/day) to most frequent
	presets := []int{1440, 720}
	for _, preset := range presets {
		demand := newActiveCount * (1440 / preset)
		if demand <= checksPerDay {
			return fmt.Errorf(
				"interval %d min would exceed plan capacity (%d checks/day with %d active profiles); "+
					"cheapest feasible preset: %d min (%d checks/day)",
				// Placeholder for the rejected interval; caller has it in the request.
				preset, checksPerDay, newActiveCount, preset, demand,
			)
		}
	}
	return fmt.Errorf(
		"no interval preset is feasible for %d active profiles on a plan with %d checks/day",
		newActiveCount, checksPerDay,
	)
}

func profileToResponse(p *entities.SocialProfile) *Response {
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
		ConsecutiveFailures:  p.ConsecutiveFailures,
		CreatedAt:            p.CreatedAt,
		UpdatedAt:            p.UpdatedAt,
	}
}

