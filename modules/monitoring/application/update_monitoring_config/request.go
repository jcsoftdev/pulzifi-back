package updatemonitoringconfig

type UpdateMonitoringConfigRequest struct {
	CheckFrequency         *string  `json:"check_frequency,omitempty"`
	ScheduleType           *string  `json:"schedule_type,omitempty"`
	Timezone               *string  `json:"timezone,omitempty"`
	BlockAdsCookies        *bool    `json:"block_ads_cookies,omitempty"`
	EnabledInsightTypes    []string `json:"enabled_insight_types,omitempty"`
	EnabledAlertConditions []string `json:"enabled_alert_conditions,omitempty"`
	CustomAlertCondition   *string  `json:"custom_alert_condition,omitempty"`
}
