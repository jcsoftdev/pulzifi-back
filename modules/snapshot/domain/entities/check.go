package entities

import (
	"time"

	"github.com/google/uuid"
)

// Check is the snapshot module's view of a monitoring check result.
// It is owned by this module and does not import monitoring entities.
type Check struct {
	ID                  uuid.UUID
	PageID              uuid.UUID
	SectionID           *uuid.UUID
	ParentCheckID       *uuid.UUID
	Status              string
	ScreenshotURL       string
	HTMLSnapshotURL     string
	ContentHash         string
	ChangeDetected      bool
	ChangeType          string
	ErrorMessage        string
	DurationMs          int
	ScreenshotHash      string
	VisionChangeSummary string
	ContentBlockHash    string
	ContentDiffJSON     string
	DiffImageURL        string
	CheckedAt           time.Time
}

// NewCheck creates a new Check with a fresh ID.
func NewCheck(pageID uuid.UUID, status string, changeDetected bool) *Check {
	return &Check{
		ID:             uuid.New(),
		PageID:         pageID,
		Status:         status,
		ChangeDetected: changeDetected,
		CheckedAt:      time.Now(),
	}
}

// SelectorOffsets holds pixel offsets for element bounding-box adjustments.
type SelectorOffsets struct {
	Top    int `json:"top"`
	Right  int `json:"right"`
	Bottom int `json:"bottom"`
	Left   int `json:"left"`
}

// MonitoringConfig is the snapshot module's view of a page's monitoring config.
type MonitoringConfig struct {
	BlockAdsCookies        bool
	EnabledInsightTypes    []string
	EnabledAlertConditions []string
	SelectorType           string // "full_page", "element", "sections"
	CSSSelector            string
	XPathSelector          string
	SelectorOffsets        *SelectorOffsets
	IgnoreSelectors        []string
}

// MonitoredSection is the snapshot module's view of a monitored page section.
type MonitoredSection struct {
	ID              uuid.UUID
	CSSSelector     string
	XPathSelector   string
	SelectorOffsets *SelectorOffsets
}

// NotificationPreference holds the data the snapshot module needs to send alert emails.
type NotificationPreference struct {
	UserID      uuid.UUID
	ChangeTypes []string
}
