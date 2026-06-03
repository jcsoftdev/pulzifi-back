package provisionorganization

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// subdomainPattern allows lowercase letters, digits, and hyphens (3–63 chars,
// must start and end with an alphanumeric character).
var subdomainPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9\-]{1,61}[a-z0-9]$`)

// Request carries the input for the provision_organization use case.
//
// UserID is always populated from the auth middleware context — it MUST NOT
// come from the request body.  OrgName and Subdomain come from the JSON body.
// The onboarding profile fields are optional; when provided they are persisted
// atomically with the org creation in the OAuth flow.
type Request struct {
	UserID    uuid.UUID `json:"-"`
	OrgName   string    `json:"org_name"`
	Subdomain string    `json:"subdomain"`

	// Onboarding profile (optional — provided by OAuth users who haven't completed it)
	CompanySize          string   `json:"company_size,omitempty"`
	BusinessType         string   `json:"business_type,omitempty"`
	CompetitorChallenges []string `json:"competitor_challenges,omitempty"`
	WebsiteURL           string   `json:"website_url,omitempty"`
}

// Validate performs lightweight structural checks before calling any
// infrastructure.  Heavy uniqueness checks are delegated to TrialProvisioner.
func (r *Request) Validate() error {
	if r.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}
	if strings.TrimSpace(r.OrgName) == "" {
		return fmt.Errorf("org_name is required")
	}
	subdomain := strings.TrimSpace(r.Subdomain)
	if subdomain == "" {
		return fmt.Errorf("subdomain is required")
	}
	if len(subdomain) < 3 {
		return fmt.Errorf("subdomain must be at least 3 characters")
	}
	if len(subdomain) > 63 {
		return fmt.Errorf("subdomain must be at most 63 characters")
	}
	if !subdomainPattern.MatchString(subdomain) {
		return fmt.Errorf("subdomain may only contain lowercase letters, digits, and hyphens and must start/end with an alphanumeric character")
	}
	return nil
}
