package register

// Request contains the registration request data
type Request struct {
	Email                 string `json:"email" binding:"required,email"`
	Password              string `json:"password" binding:"required,min=8"`
	FirstName             string `json:"first_name" binding:"required"`
	LastName              string `json:"last_name" binding:"required"`
	OrganizationName      string `json:"organization_name" binding:"required"`
	OrganizationSubdomain string `json:"organization_subdomain" binding:"required"`
	// SelectedPlan is the paid plan code chosen on the pricing page (e.g.
	// "starter"). Empty for the free/trial signup. Used to personalize the
	// welcome email — the actual subscription is created via Stripe checkout.
	SelectedPlan string `json:"selected_plan"`
}
