package listchecks

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type CheckResponse struct {
	ID              uuid.UUID        `json:"id"`
	PageID          uuid.UUID        `json:"page_id"`
	SectionID       *uuid.UUID       `json:"section_id,omitempty"`
	ParentCheckID   *uuid.UUID       `json:"parent_check_id,omitempty"`
	Status          string           `json:"status"`
	ScreenshotURL   string           `json:"screenshot_url"`
	HTMLSnapshotURL string           `json:"html_snapshot_url"`
	ChangeDetected  bool             `json:"change_detected"`
	ChangeType      string           `json:"change_type"`
	ErrorMessage    string           `json:"error_message,omitempty"`
	// ContentDiff is the stored content-diff jsonb embedded as a real JSON object
	// (not a quoted string) so the frontend can read content_diff.has_changes /
	// content_diff.diffs directly without re-parsing.
	ContentDiff json.RawMessage `json:"content_diff,omitempty"`
	DiffImageURL    string           `json:"diff_image_url,omitempty"`
	CheckedAt       time.Time        `json:"checked_at"`
	Sections        []*CheckResponse `json:"sections,omitempty"`
}

type ListChecksResponse struct {
	Checks []*CheckResponse `json:"checks"`
}
