package createcheck

import (
	"github.com/google/uuid"
)

type CreateCheckRequest struct {
	PageID          uuid.UUID `json:"page_id" binding:"required"`
	Status          string    `json:"status" binding:"required"`
	ChangeDetected  bool      `json:"change_detected"`
	ChangeType      string    `json:"change_type"`
	ScreenshotURL   string    `json:"screenshot_url"`
	HTMLSnapshotURL string    `json:"html_snapshot_url"`
	ErrorMessage    string    `json:"error_message"`
	DurationMs      int       `json:"duration_ms"`
}
