package provisionorganization

import "time"

// Response is the output of a successful provision_organization call.
//
// Field names intentionally mirror register.Response so both flows surface the
// same shape to the frontend (organization_subdomain + trial_ends_at).
type Response struct {
	OrganizationSubdomain string    `json:"subdomain"`
	TrialEndsAt           time.Time `json:"trial_ends_at"`
}
