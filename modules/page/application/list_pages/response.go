package listpages

import (
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/page/domain/entities"
)

type PageResponse struct {
	ID                   uuid.UUID  `json:"id"`
	WorkspaceID          uuid.UUID  `json:"workspace_id"`
	Name                 string     `json:"name"`
	URL                  string     `json:"url"`
	ThumbnailURL         string     `json:"thumbnail_url,omitempty"`
	LastCheckedAt        *time.Time `json:"last_checked_at,omitempty"`
	LastChangeDetectedAt *time.Time `json:"last_change_detected_at,omitempty"`
	CheckCount           int        `json:"check_count"`
	CreatedBy            uuid.UUID  `json:"created_by"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	Tags                 []string   `json:"tags,omitempty"`
	CheckFrequency       string     `json:"check_frequency,omitempty"`
	DetectedChanges      int        `json:"detected_changes"`
}

type ListPagesResponse struct {
	Pages []PageResponse `json:"pages"`
}

func ToPageResponse(page *entities.Page) PageResponse {
	return PageResponse{
		ID:                   page.ID,
		WorkspaceID:          page.WorkspaceID,
		Name:                 page.Name,
		URL:                  page.URL,
		ThumbnailURL:         page.ThumbnailURL,
		LastCheckedAt:        page.LastCheckedAt,
		LastChangeDetectedAt: page.LastChangeDetectedAt,
		CheckCount:           page.CheckCount,
		CreatedBy:            page.CreatedBy,
		CreatedAt:            page.CreatedAt,
		UpdatedAt:            page.UpdatedAt,
		Tags:                 page.Tags,
		CheckFrequency:       page.CheckFrequency,
		DetectedChanges:      page.DetectedChanges,
	}
}
