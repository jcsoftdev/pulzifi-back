package acceptinvitation

// Request is the public payload posted to the accept-invitation endpoint.
// All fields are required; the HTTP layer validates structure before this
// reaches the handler.
type Request struct {
	FirstName             string `json:"first_name" validate:"required"`
	LastName              string `json:"last_name" validate:"required"`
	Password              string `json:"password" validate:"required,min=8"`
	OrganizationName      string `json:"organization_name" validate:"required"`
	OrganizationSubdomain string `json:"organization_subdomain" validate:"required"`
}
