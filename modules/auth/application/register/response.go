package register

import (
	"time"

	"github.com/google/uuid"
)

// Response contains the registration response data.
//
// The self-serve trial flow now returns the freshly-provisioned organization
// subdomain plus the trial expiry so the frontend can redirect the user to
// their tenant and surface a countdown without an extra round-trip.
type Response struct {
	UserID                uuid.UUID `json:"user_id"`
	Email                 string    `json:"email"`
	FirstName             string    `json:"first_name"`
	LastName              string    `json:"last_name"`
	Status                string    `json:"status"`
	Message               string    `json:"message"`
	OrganizationID        uuid.UUID `json:"organization_id"`
	OrganizationSubdomain string    `json:"organization_subdomain"`
	TrialEndsAt           time.Time `json:"trial_ends_at"`
}
