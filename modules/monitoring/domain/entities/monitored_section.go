package entities

import (
	"time"

	"github.com/google/uuid"
)

// SectionRect holds pixel coordinates of a section in the original viewport space.
type SectionRect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// MonitoredSection represents a specific region/element of a page to monitor independently.
type MonitoredSection struct {
	ID              uuid.UUID
	PageID          uuid.UUID
	Name            string
	CSSSelector     string
	XPathSelector   string
	SelectorOffsets *SelectorOffsets
	Rect            *SectionRect
	ViewportWidth   int
	SortOrder       int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewMonitoredSection creates a new monitored section.
func NewMonitoredSection(pageID uuid.UUID, name, cssSelector, xpathSelector string, offsets *SelectorOffsets, rect *SectionRect, viewportWidth, sortOrder int) *MonitoredSection {
	return &MonitoredSection{
		ID:              uuid.New(),
		PageID:          pageID,
		Name:            name,
		CSSSelector:     cssSelector,
		XPathSelector:   xpathSelector,
		SelectorOffsets: offsets,
		Rect:            rect,
		ViewportWidth:   viewportWidth,
		SortOrder:       sortOrder,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}
