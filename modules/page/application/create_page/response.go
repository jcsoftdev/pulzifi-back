package createpage

import (
	"time"

	"github.com/google/uuid"
)

type CreatePageResponse struct {
	ID                   uuid.UUID `json:"id"`
	WorkspaceID          uuid.UUID `json:"workspace_id"`
	Name                 string    `json:"name"`
	URL                  string    `json:"url"`
	ThumbnailURL         string    `json:"thumbnail_url,omitempty"`
	LastCheckedAt        *time.Time `json:"last_checked_at,omitempty"`
	LastChangeDetectedAt *time.Time `json:"last_change_detected_at,omitempty"`
	CheckCount           int        `json:"check_count"`
	Tags                 []string   `json:"tags"`
	CheckFrequency       string     `json:"check_frequency"`
	DetectedChanges      int        `json:"detected_changes"`
	CreatedBy            uuid.UUID  `json:"created_by"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}
