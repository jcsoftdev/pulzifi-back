package updatecurrentuser

// OrganizationResponse is the org-context section of the updated user response.
type OrganizationResponse struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Subdomain    string         `json:"subdomain"`
	PlanCode     string         `json:"planCode"`
	FeatureFlags map[string]any `json:"featureFlags"`
}

// Response represents the updated current user profile.
type Response struct {
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	Email        string                `json:"email"`
	Role         string                `json:"role"`
	Status       string                `json:"status"`
	Avatar       *string               `json:"avatar,omitempty"`
	Tenant       *string               `json:"tenant,omitempty"`
	Organization *OrganizationResponse `json:"organization,omitempty"`
	CreatedAt    string                `json:"created_at"`
	UpdatedAt    string                `json:"updated_at"`
}
