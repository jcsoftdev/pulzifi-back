package services

import (
	"regexp"
	"strings"

	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/errors"
)

// OrganizationService provides domain logic for organizations
type OrganizationService struct{}

// NewOrganizationService creates a new organization service
func NewOrganizationService() *OrganizationService {
	return &OrganizationService{}
}

// ValidateOrganizationName validates the organization name
func (s *OrganizationService) ValidateOrganizationName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return &domainerrors.InvalidOrganizationDataError{Message: "name cannot be empty"}
	}
	if len(name) < 2 {
		return &domainerrors.InvalidOrganizationDataError{Message: "name must be at least 2 characters"}
	}
	if len(name) > 255 {
		return &domainerrors.InvalidOrganizationDataError{Message: "name must be at most 255 characters"}
	}
	return nil
}

// ValidateSubdomain validates the subdomain format
func (s *OrganizationService) ValidateSubdomain(subdomain string) error {
	subdomain = strings.TrimSpace(strings.ToLower(subdomain))
	if subdomain == "" {
		return &domainerrors.InvalidOrganizationDataError{Message: "subdomain cannot be empty"}
	}

	// Subdomain must be alphanumeric and hyphens only, 3-63 characters
	if len(subdomain) < 3 {
		return &domainerrors.InvalidOrganizationDataError{Message: "subdomain must be at least 3 characters"}
	}
	if len(subdomain) > 63 {
		return &domainerrors.InvalidOrganizationDataError{Message: "subdomain must be at most 63 characters"}
	}

	// Allow alphanumeric and hyphens, but not starting/ending with hyphen
	subdomainRegex := regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
	if !subdomainRegex.MatchString(subdomain) {
		return &domainerrors.InvalidOrganizationDataError{Message: "subdomain must contain only alphanumeric characters and hyphens, and cannot start or end with a hyphen"}
	}

	return nil
}

// GenerateSchemaName generates a PostgreSQL schema name from subdomain
func (s *OrganizationService) GenerateSchemaName(subdomain string) string {
	// Replace hyphens with underscores for valid PostgreSQL schema name
	schemaName := strings.ReplaceAll(strings.ToLower(subdomain), "-", "_")
	// Ensure it doesn't start with a digit (PostgreSQL requirement)
	if schemaName[0] >= '0' && schemaName[0] <= '9' {
		schemaName = "s_" + schemaName
	}
	return schemaName
}
