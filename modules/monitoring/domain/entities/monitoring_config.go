package entities

import (
	"time"

	"github.com/google/uuid"
)

type MonitoringConfig struct {
	ID                    uuid.UUID
	PageID                uuid.UUID
	CheckFrequency        string
	ScheduleType          string
	Timezone              string
	BlockAdsCookies       bool
	EnabledInsightTypes   []string
	EnabledAlertConditions []string
	CustomAlertCondition  string
	CreatedAt             time.Time
	UpdatedAt             time.Time
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
