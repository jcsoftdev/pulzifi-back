package create_organization

// Request represents the input for creating an organization
type Request struct {
	Name      string `json:"name" binding:"required,min=2,max=255"`
	Subdomain string `json:"subdomain" binding:"required,min=3,max=63"`
}
