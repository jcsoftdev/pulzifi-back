package getprofile

import (
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
)

// SnapshotDetail carries a snapshot's key fields for the profile detail view.
type SnapshotDetail struct {
	ID             uuid.UUID              `json:"id"`
	CapturedAt     time.Time              `json:"captured_at"`
	Status         string                 `json:"status"`
	Error          string                 `json:"error,omitempty"`
	FollowersCount int                    `json:"followers_count"`
	PostsCount     int                    `json:"posts_count"`
	Data           *entities.ProfileData  `json:"data,omitempty"`
}

// ProfileDetail is the full profile representation including the latest snapshot.
type ProfileDetail struct {
	ID                   uuid.UUID  `json:"id"`
	WorkspaceID          uuid.UUID  `json:"workspace_id"`
	Platform             string     `json:"platform"`
	Handle               string     `json:"handle"`
	DisplayName          string     `json:"display_name"`
	AvatarURL            string     `json:"avatar_url"`
	IsActive             bool       `json:"is_active"`
	CheckIntervalMinutes int        `json:"check_interval_minutes"`
	NextCheckAt          *time.Time `json:"next_check_at"`
	LastCheckedAt        *time.Time `json:"last_checked_at"`
	ConsecutiveFailures  int        `json:"consecutive_failures"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// Response is the output DTO for get_profile.
type Response struct {
	Profile        *ProfileDetail  `json:"profile"`
	LatestSnapshot *SnapshotDetail `json:"latest_snapshot"`
}
