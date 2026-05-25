package createcheckoutsession

// Request is the input DTO for the CreateCheckoutSession use case.
type Request struct {
	OrgID        string // UUID string of the organisation
	OrgEmail     string // used to create a Stripe customer when one does not exist yet
	OrgName      string // used to create a Stripe customer when one does not exist yet
	PlanID       string // internal plan code (e.g. "starter", "pro", "enterprise")
	BillingCycle string // "monthly" | "yearly"

	// Config URLs — passed from shared/config by the HTTP layer.
	SuccessURL string
	CancelURL  string
}
