package entities

import (
	"time"

	"github.com/google/uuid"
	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"
)

// SocialProfile represents a monitored social media profile.
type SocialProfile struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID

	// Platform is the social network this profile belongs to.
	Platform valueobjects.Platform

	// Handle is the username / @handle on the platform (e.g. "nasa").
	Handle string

	// DisplayName is the human-readable name shown on the profile (may be empty
	// until the first successful check populates it).
	DisplayName string

	// AvatarURL is the stored (MinIO/Cloudinary) URL of the profile picture.
	AvatarURL string

	// IsActive controls whether the scheduler includes this profile in due-profile
	// queries. Setting to false removes it from the scheduler.
	IsActive bool

	// CheckIntervalMinutes is the desired polling cadence.
	// Allowed presets: 720, 1440 (validated at create/update time). Minimum cadence is 12h.
	CheckIntervalMinutes int

	// NextCheckAt is the time at which the scheduler should next run a check.
	// Set to now() on creation and after each successful check.
	NextCheckAt *time.Time

	// LastCheckedAt is the time of the most recent completed check (success or failure).
	LastCheckedAt *time.Time

	// ConsecutiveFailures tracks back-to-back fetch errors.
	// Reset to 0 on success. When it reaches 5, the profile is deactivated.
	ConsecutiveFailures int

	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewSocialProfile constructs a new SocialProfile with safe defaults.
// NextCheckAt is set to now so the scheduler picks it up immediately.
func NewSocialProfile(
	workspaceID uuid.UUID,
	platform valueobjects.Platform,
	handle string,
	checkIntervalMinutes int,
) *SocialProfile {
	now := time.Now().UTC()
	return &SocialProfile{
		ID:                   uuid.New(),
		WorkspaceID:          workspaceID,
		Platform:             platform,
		Handle:               handle,
		IsActive:             true,
		CheckIntervalMinutes: checkIntervalMinutes,
		NextCheckAt:          &now,
		ConsecutiveFailures:  0,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}
