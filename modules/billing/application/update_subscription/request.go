package updatesubscription

// Request carries the input for the in-place subscription change use case.
// OrgID identifies the caller's organization (resolved upstream from the
// subdomain). PlanID is the internal plan code (starter|pro). BillingCycle
// selects monthly vs yearly. Preview=true returns only the prorated amount
// without mutating the subscription.
type Request struct {
	OrgID        string
	PlanID       string
	BillingCycle string
	Preview      bool
}
