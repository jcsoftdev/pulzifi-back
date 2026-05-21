package getquotas

// Response holds quota information for the current billing period.
type Response struct {
	ChecksUsed        int         `json:"checks_used"`
	ChecksAllowed     int         `json:"checks_allowed"`
	NextRefillAt      interface{} `json:"next_refill_at"`
	StoragePeriodDays int         `json:"storage_period_days"`
	Message           string      `json:"message"`
}
