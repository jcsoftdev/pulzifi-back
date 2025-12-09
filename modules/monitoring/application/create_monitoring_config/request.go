package create_monitoring_config

import "github.com/google/uuid"

type CreateMonitoringConfigRequest struct {
	PageID          uuid.UUID `json:"page_id"`
	CheckFrequency  string    `json:"check_frequency"` // "5m", "1h", "1d"
	ScheduleType    string    `json:"schedule_type"`   // "continuous", "scheduled"
	Timezone        string    `json:"timezone"`        // "UTC", "America/New_York", etc
	BlockAdsCookies bool      `json:"block_ads_cookies"`
}
