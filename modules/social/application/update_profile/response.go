package updateprofile

import (
	"time"

	"github.com/google/uuid"
)

// Response is the output DTO for an updated social profile.
type Response struct {
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
