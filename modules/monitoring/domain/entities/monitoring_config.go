package entities

import (
	"time"
	"github.com/google/uuid"
)

type MonitoringConfig struct {
	ID             uuid.UUID
	PageID         uuid.UUID
	CheckFrequency string
	ScheduleType   string
	Timezone       string
	BlockAdsCookies bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewMonitoringConfig(pageID uuid.UUID, frequency, scheduleType, timezone string) *MonitoringConfig {
	return &MonitoringConfig{
		ID:             uuid.New(),
		PageID:         pageID,
		CheckFrequency: frequency,
		ScheduleType:   scheduleType,
		Timezone:       timezone,
		BlockAdsCookies: true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

type SnapshotTask struct {
	PageID uuid.UUID
	URL    string
}
