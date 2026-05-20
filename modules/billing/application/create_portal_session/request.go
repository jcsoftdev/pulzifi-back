package createportalsession

// Request is the input DTO for the CreatePortalSession use case.
type Request struct {
	OrgID     string // UUID string of the organisation
	ReturnURL string // where Stripe redirects after the user leaves the portal
}
