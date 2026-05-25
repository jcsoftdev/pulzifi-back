package services

import (
	"context"

	"github.com/google/uuid"
)

// OrganizationMembershipChecker is auth's port for checking whether a user
// already belongs to at least one organization. It drives the idempotency
// guard in the OAuth callback and the provision_organization use case.
//
// Implementations live in cmd/wiring/auth/ to keep the auth module decoupled
// from organization infrastructure packages.
type OrganizationMembershipChecker interface {
	HasAnyMembership(ctx context.Context, userID uuid.UUID) (bool, error)
}
