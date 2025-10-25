package errors

import "fmt"

// OrganizationNotFoundError occurs when an organization is not found
type OrganizationNotFoundError struct {
	OrganizationID string
}

func (e *OrganizationNotFoundError) Error() string {
	return fmt.Sprintf("organization not found: %s", e.OrganizationID)
}

// SubdomainAlreadyExistsError occurs when a subdomain is already taken
type SubdomainAlreadyExistsError struct {
	Subdomain string
}

func (e *SubdomainAlreadyExistsError) Error() string {
	return fmt.Sprintf("subdomain already exists: %s", e.Subdomain)
}

// InvalidOrganizationDataError occurs when organization data is invalid
type InvalidOrganizationDataError struct {
	Message string
}

func (e *InvalidOrganizationDataError) Error() string {
	return fmt.Sprintf("invalid organization data: %s", e.Message)
}

// OrganizationAlreadyDeletedError occurs when trying to operate on a deleted organization
type OrganizationAlreadyDeletedError struct {
	OrganizationID string
}

func (e *OrganizationAlreadyDeletedError) Error() string {
	return fmt.Sprintf("organization already deleted: %s", e.OrganizationID)
}
