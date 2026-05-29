package authwiring

import (
	"context"

	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	orgrepos "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/repositories"
	orgservices "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
)

// organizationDirectoryAdapter implements authservices.OrganizationDirectory by
// composing organization's repository and service. Lives in cmd/wiring/auth so
// the auth module itself never imports from the organization module.
type organizationDirectoryAdapter struct {
	repo    orgrepos.OrganizationRepository
	service *orgservices.OrganizationService
}

// NewOrganizationDirectoryAdapter creates an OrganizationDirectory backed by
// organization's repository and domain service.
func NewOrganizationDirectoryAdapter(
	repo orgrepos.OrganizationRepository,
	service *orgservices.OrganizationService,
) authservices.OrganizationDirectory {
	return &organizationDirectoryAdapter{repo: repo, service: service}
}

func (a *organizationDirectoryAdapter) ValidateOrganizationName(name string) error {
	return a.service.ValidateOrganizationName(name)
}

func (a *organizationDirectoryAdapter) ValidateSubdomain(subdomain string) error {
	return a.service.ValidateSubdomain(subdomain)
}

func (a *organizationDirectoryAdapter) CountBySubdomain(ctx context.Context, subdomain string) (int, error) {
	return a.repo.CountBySubdomain(ctx, subdomain)
}
