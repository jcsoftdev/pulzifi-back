package getquotas

// Response holds quota information for the current billing period.
//
// AI insight fields mirror the checks pair. AIInsightsUnlimited is true when
// the cap is the INT max sentinel — frontend should render "Unlimited" instead
// of a numeric remaining value in that case.
type Response struct {
	ChecksUsed        int         `json:"checks_used"`
	ChecksAllowed     int         `json:"checks_allowed"`
	NextRefillAt      interface{} `json:"next_refill_at"`
	StoragePeriodDays int         `json:"storage_period_days"`

	AIInsightsUsed      int  `json:"ai_insights_used"`
	AIInsightsAllowed   int  `json:"ai_insights_allowed"`
	AIInsightsRemaining int  `json:"ai_insights_remaining"`
	AIInsightsUnlimited bool `json:"ai_insights_unlimited"`

	Message string `json:"message"`
}
