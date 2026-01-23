package get_monitoring_config

import (
	"time"

	"github.com/google/uuid"
)

type GetMonitoringConfigResponse struct {
	ID              uuid.UUID `json:"id"`
	PageID          uuid.UUID `json:"page_id"`
	CheckFrequency  string    `json:"check_frequency"`
	ScheduleType    string    `json:"schedule_type"`
	Timezone        string    `json:"timezone"`
	BlockAdsCookies bool      `json:"block_ads_cookies"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
