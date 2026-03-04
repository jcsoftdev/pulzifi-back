package entities

import (
	"time"

	"github.com/google/uuid"
)

// SelectorOffsets defines pixel offsets to adjust the element bounding box.
type SelectorOffsets struct {
	Top    int `json:"top"`
	Right  int `json:"right"`
	Bottom int `json:"bottom"`
	Left   int `json:"left"`
}

type MonitoringConfig struct {
	ID                     uuid.UUID
	PageID                 uuid.UUID
	CheckFrequency         string
	ScheduleType           string
	Timezone               string
	BlockAdsCookies        bool
	EnabledInsightTypes    []string
	EnabledAlertConditions []string
	CustomAlertCondition   string
	SelectorType           string          // "full_page" (default) or "element"
	CSSSelector            string
	XPathSelector          string
	SelectorOffsets        *SelectorOffsets // pixel offsets for element bounding box
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

func NewMonitoringConfig(pageID uuid.UUID, frequency, scheduleType, timezone string) *MonitoringConfig {
	return &MonitoringConfig{
		ID:                    uuid.New(),
		PageID:                pageID,
		CheckFrequency:        frequency,
		ScheduleType:          scheduleType,
		Timezone:              timezone,
		BlockAdsCookies:       true,
		EnabledInsightTypes:   []string{"marketing", "market_analysis"},
		EnabledAlertConditions: []string{"any_changes"},
		CustomAlertCondition:  "",
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}
}

type SnapshotTask struct {
	PageID uuid.UUID
	URL    string
}
