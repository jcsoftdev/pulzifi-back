package get_monitoring_config

import (
	"time"

	"github.com/google/uuid"
)

type SelectorOffsetsDTO struct {
	Top    int `json:"top"`
	Right  int `json:"right"`
	Bottom int `json:"bottom"`
	Left   int `json:"left"`
}

type GetMonitoringConfigResponse struct {
	ID                     uuid.UUID           `json:"id"`
	PageID                 uuid.UUID           `json:"page_id"`
	CheckFrequency         string              `json:"check_frequency"`
	ScheduleType           string              `json:"schedule_type"`
	Timezone               string              `json:"timezone"`
	BlockAdsCookies        bool                `json:"block_ads_cookies"`
	EnabledInsightTypes    []string            `json:"enabled_insight_types"`
	EnabledAlertConditions []string            `json:"enabled_alert_conditions"`
	CustomAlertCondition   string              `json:"custom_alert_condition"`
	SelectorType           string              `json:"selector_type"`
	CSSSelector            string              `json:"css_selector"`
	XPathSelector          string              `json:"xpath_selector"`
	SelectorOffsets        *SelectorOffsetsDTO `json:"selector_offsets,omitempty"`
	CreatedAt              time.Time           `json:"created_at"`
	UpdatedAt              time.Time           `json:"updated_at"`
}
