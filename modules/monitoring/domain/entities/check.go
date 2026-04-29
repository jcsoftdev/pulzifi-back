package entities

import (
	"time"

	"github.com/google/uuid"
)

// Check represents a page monitoring check result
type Check struct {
	ID                  uuid.UUID
	PageID              uuid.UUID
	SectionID           *uuid.UUID // optional link to a monitored_section; nil = full-page check
	ParentCheckID       *uuid.UUID // links section checks back to the parent full-page check
	Status              string     // success, error
	ScreenshotURL       string
	HTMLSnapshotURL     string
	ContentHash         string
	ChangeDetected      bool
	ChangeType          string
	ErrorMessage        string
	DurationMs          int
	ScreenshotHash      string // SHA-256 of screenshot bytes for pixel comparison
	VisionChangeSummary string // AI-generated change description from vision model
	ContentBlockHash    string // SHA-256 of content blocks for structural comparison
	ContentDiffJSON     string // JSON-serialized ContentDiff for the frontend
	DiffImageURL        string // URL of the diff image highlighting changed pixels
	CheckedAt           time.Time
}

// NewCheck creates a new check
func NewCheck(pageID uuid.UUID, status string, changeDetected bool) *Check {
	return &Check{
		ID:             uuid.New(),
		PageID:         pageID,
		Status:         status,
		ChangeDetected: changeDetected,
		CheckedAt:      time.Now(),
	}
}
