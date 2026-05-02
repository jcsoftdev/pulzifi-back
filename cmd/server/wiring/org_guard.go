package wiring

import (
	"context"

	"github.com/google/uuid"

	orgrepos "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/repositories"
)

// OrgGuard satisfies dispatch_event.OrgGuard.
// IsActive returns true if the org exists and is not soft-deleted.
type OrgGuard struct{ repo orgrepos.OrganizationRepository }

// NewOrgGuard constructs an OrgGuard backed by the given OrganizationRepository.
func NewOrgGuard(repo orgrepos.OrganizationRepository) *OrgGuard {
	return &OrgGuard{repo: repo}
}

// IsActive returns true when the organization exists and has not been deleted.
func (g *OrgGuard) IsActive(ctx context.Context, orgID uuid.UUID) (bool, error) {
	org, err := g.repo.GetByID(ctx, orgID)
	if err != nil {
		return false, err
	}
	if org == nil || org.DeletedAt != nil {
		return false, nil
	}
	return true, nil
}
