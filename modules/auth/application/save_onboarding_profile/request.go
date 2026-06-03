package saveonboardingprofile

import "github.com/google/uuid"

// Request carries the caller's onboarding answers.
// UserID MUST come from the auth middleware context — never from the request body.
type Request struct {
	UserID               uuid.UUID `json:"-"`
	CompanySize          string    `json:"company_size"`
	BusinessType         string    `json:"business_type"`
	CompetitorChallenges []string  `json:"competitor_challenges"`
	WebsiteURL           string    `json:"website_url"`
}
