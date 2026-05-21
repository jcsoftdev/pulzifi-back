package getmetrics

// CheckMetrics is the checks sub-section.
type CheckMetrics struct {
	Total   int `json:"total"`
	Success int `json:"success"`
	Failed  int `json:"failed"`
}

// Response holds all usage metrics for a tenant.
type Response struct {
	Checks        CheckMetrics `json:"checks"`
	Pages         int          `json:"pages"`
	Workspaces    int          `json:"workspaces"`
	Alerts        int          `json:"alerts"`
	ChecksUsed    int          `json:"checks_used"`
	ChecksAllowed int          `json:"checks_allowed"`
}
