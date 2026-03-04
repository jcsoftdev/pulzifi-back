package updatemonitoringconfig

import (
	"time"

	"github.com/google/uuid"
)

type UpdateMonitoringConfigResponse struct {
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
	UpdatedAt              time.Time           `json:"updated_at"`
	QuotaExceeded          bool                `json:"quota_exceeded"`
}
