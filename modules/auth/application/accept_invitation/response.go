package acceptinvitation

// Response carries the IDs the HTTP layer needs to issue a session cookie
// and a cross-subdomain nonce after a successful acceptance.
type Response struct {
	UserID       string `json:"user_id"`
	OrgID        string `json:"org_id"`
	OrgSubdomain string `json:"org_subdomain"`
}
