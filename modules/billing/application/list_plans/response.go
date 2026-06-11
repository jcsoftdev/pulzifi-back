package listplans

// PlanResponse is the per-plan DTO returned to the frontend.
// Nil integer pointers mean "unlimited" for that dimension.
type PlanResponse struct {
	ID                       string `json:"id"`
	Code                     string `json:"code"`
	Name                     string `json:"name"`
	Description              string `json:"description"`
	PriceMonthlyCents        int    `json:"price_monthly_cents"`
	PriceYearlyCents         int    `json:"price_yearly_cents"`
	Currency                 string `json:"currency"`
	ChecksAllowedMonthly     *int   `json:"checks_allowed_monthly"`
	StoragePeriodDays        int    `json:"storage_period_days"`
	AIInsightsAllowedMonthly *int   `json:"ai_insights_allowed_monthly"`
	MaxPages                 *int   `json:"max_pages"`
	MaxWorkspaces            *int   `json:"max_workspaces"`
	IsPopular                bool   `json:"is_popular"`
}

// Response wraps the plan list.
type Response struct {
	Plans []PlanResponse `json:"plans"`
}
