package getpage

import (
	"time"

	"github.com/google/uuid"
)

type GetPageResponse struct {
	ID                   uuid.UUID `json:"id"`
	WorkspaceID          uuid.UUID `json:"workspace_id"`
	Name                 string    `json:"name"`
	URL                  string    `json:"url"`
	ThumbnailURL         string    `json:"thumbnail_url,omitempty"`
	LastCheckedAt        *time.Time `json:"last_checked_at,omitempty"`
	LastChangeDetectedAt *time.Time `json:"last_change_detected_at,omitempty"`
	CheckCount           int        `json:"check_count"`
	CheckFrequency       string     `json:"check_frequency"`
	DetectedChanges      int        `json:"detected_changes"`
	Tags                 []string   `json:"tags"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}
