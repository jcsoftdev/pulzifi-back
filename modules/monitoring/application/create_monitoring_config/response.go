package create_monitoring_config

import (
	"time"

	"github.com/google/uuid"
)

type CreateMonitoringConfigResponse struct {
	ID                    uuid.UUID `json:"id"`
	PageID                uuid.UUID `json:"page_id"`
	CheckFrequency        string    `json:"check_frequency"`
	ScheduleType          string    `json:"schedule_type"`
	Timezone              string    `json:"timezone"`
	BlockAdsCookies       bool      `json:"block_ads_cookies"`
	EnabledInsightTypes   []string  `json:"enabled_insight_types"`
	EnabledAlertConditions []string `json:"enabled_alert_conditions"`
	CustomAlertCondition  string    `json:"custom_alert_condition"`
	CreatedAt             time.Time `json:"created_at"`
}
