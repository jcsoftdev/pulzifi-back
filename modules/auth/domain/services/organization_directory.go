package services

import "context"

// OrganizationDirectory is auth's port for the minimal organization operations
// it needs: subdomain validation and uniqueness checks.
// Implemented by authwiring in cmd/wiring/auth/.
type OrganizationDirectory interface {
	// ValidateOrganizationName returns an error if name is empty, too short, or
	// too long.
	ValidateOrganizationName(name string) error

	// ValidateSubdomain returns an error if subdomain format is invalid.
	ValidateSubdomain(subdomain string) error

	// CountBySubdomain returns the number of approved organizations that already
	// use the given subdomain.
	CountBySubdomain(ctx context.Context, subdomain string) (int, error)
}
