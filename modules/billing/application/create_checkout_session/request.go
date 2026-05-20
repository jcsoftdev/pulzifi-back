package createcheckoutsession

// Request is the input DTO for the CreateCheckoutSession use case.
type Request struct {
	OrgID        string // UUID string of the organisation
	OrgEmail     string // used to create a Stripe customer when one does not exist yet
	OrgName      string // used to create a Stripe customer when one does not exist yet
	PlanID       string // internal plan identifier (maps to stripe_price_id_*)
	BillingCycle string // "monthly" | "yearly"

	// Caller must resolve and pass these from the plans table.
	// They are passed in rather than looked up here to keep the use case free of
	// a plan repository dependency (lookup happens in the HTTP layer).
	StripePriceIDMonthly string
	StripePriceIDYearly  string

	// Config URLs — passed from shared/config by the HTTP layer.
	SuccessURL string
	CancelURL  string
}
