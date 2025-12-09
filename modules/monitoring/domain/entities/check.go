package entities

import (
	"time"

	"github.com/google/uuid"
)

// Check represents a page monitoring check result
type Check struct {
	ID               uuid.UUID
	PageID           uuid.UUID
	Status           string // success, error
	ScreenshotURL    string
	HTMLSnapshotURL  string
	ChangeDetected   bool
	ChangeType       string
	ErrorMessage     string
	DurationMs       int
	CheckedAt        time.Time
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
