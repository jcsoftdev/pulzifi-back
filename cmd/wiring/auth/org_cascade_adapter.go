package authwiring

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	deleteorganization "github.com/jcsoftdev/pulzifi-back/modules/organization/application/delete_organization"
	orgservices "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"

	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
)

// deleteOrgInvoker is the minimal interface this adapter needs from the use case.
// Using an interface lets tests inject a stub without depending on the full handler.
type deleteOrgInvoker interface {
	Handle(ctx context.Context, req *deleteorganization.Request) (*deleteorganization.Response, error)
}

// orgCascadeAdapter implements authservices.OrgCascade by delegating to the
// organization delete use case. Lives in cmd/wiring/auth so the auth module
// never imports from the organization module (DAO-16, design D9).
type orgCascadeAdapter struct {
	handler deleteOrgInvoker
	repo    orgservices.OrgDeletionRepo
}

// compile-time assertion
var _ authservices.OrgCascade = (*orgCascadeAdapter)(nil)

// NewOrgCascadeAdapter builds the adapter.
// handler: the delete_organization use case (or a test stub).
// repo: used to find solely-owned orgs for the given user.
func NewOrgCascadeAdapter(
	handler deleteOrgInvoker,
	repo orgservices.OrgDeletionRepo,
) authservices.OrgCascade {
	return &orgCascadeAdapter{handler: handler, repo: repo}
}

// CascadeSolelyOwnedOrgs finds orgs where userID is the sole owner and runs the
// full deletion cascade for each. Stops and returns ErrBillingActive (mapped to
// auth's domain sentinel) on the first billing block — no further orgs processed
// and the user is NOT deleted (design §5, DAO-13, Scenario 14).
func (a *orgCascadeAdapter) CascadeSolelyOwnedOrgs(ctx context.Context, userID uuid.UUID) error {
	orgs, err := a.repo.FindSolelyOwnedOrgs(ctx, userID)
	if err != nil {
		return fmt.Errorf("org_cascade_adapter: find solely owned orgs: %w", err)
	}

	for _, org := range orgs {
		_, handleErr := a.handler.Handle(ctx, &deleteorganization.Request{
			OrgID:     org.ID,
			ActorID:   userID,
			ActorType: "owner",
		})
		if handleErr != nil {
			if errors.Is(handleErr, orgservices.ErrBillingActive) {
				// Translate org-domain sentinel into auth-domain sentinel so the
				// auth handler never imports organization packages (DAO-16).
				return authservices.ErrBillingActive
			}
			return fmt.Errorf("org_cascade_adapter: delete org %s: %w", org.ID, handleErr)
		}
	}

	return nil
}
