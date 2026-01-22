package updatemonitoringconfig

type UpdateMonitoringConfigRequest struct {
	CheckFrequency  *string `json:"check_frequency,omitempty"`
	ScheduleType    *string `json:"schedule_type,omitempty"`
	Timezone        *string `json:"timezone,omitempty"`
	BlockAdsCookies *bool   `json:"block_ads_cookies,omitempty"`
}
