package managesections

import (
	"time"

	"github.com/google/uuid"
)

// SectionResponse is the API response for a single monitored section.
type SectionResponse struct {
	ID              uuid.UUID          `json:"id"`
	PageID          uuid.UUID          `json:"page_id"`
	Name            string             `json:"name"`
	CSSSelector     string             `json:"css_selector"`
	XPathSelector   string             `json:"xpath_selector"`
	SelectorOffsets *SectionOffsetsDTO `json:"selector_offsets,omitempty"`
	Rect            *SectionRectDTO    `json:"rect,omitempty"`
	ViewportWidth   int                `json:"viewport_width,omitempty"`
	SortOrder       int                `json:"sort_order"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

// ListSectionsResponse wraps a list of sections.
type ListSectionsResponse struct {
	Sections []*SectionResponse `json:"sections"`
}
