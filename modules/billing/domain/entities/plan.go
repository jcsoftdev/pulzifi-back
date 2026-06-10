package entities

// CatalogPlan holds the data the billing plan-catalog API exposes.
// nil pointer fields mean "unlimited" (no cap for that dimension).
type CatalogPlan struct {
	ID                       string
	Code                     string
	Name                     string
	Description              string
	PriceMonthlyCents        int
	PriceYearlyCents         int
	Currency                 string
	ChecksAllowedMonthly     *int
	StoragePeriodDays        int
	AIInsightsAllowedMonthly *int
	MaxPages                 *int
	MaxWorkspaces            *int
	IsPopular                bool
}

// IsUnlimitedChecks reports whether this plan has no monthly check cap.
func (p *CatalogPlan) IsUnlimitedChecks() bool { return p.ChecksAllowedMonthly == nil }
