package services

import (
	"context"

	"github.com/google/uuid"
)

// UserCleanup handles non-cascade cleanup for a user being deleted.
// Implemented in cmd/wiring/auth/membership_pruner_adapter.go (OQ-1).
type UserCleanup interface {
	// PruneMemberships soft-deletes all organization memberships for the given
	// user where the user is NOT the sole owner (those orgs were already handled
	// by OrgCascade). This replaces the destructive branch that used to run in
	// the user.deleted subscriber (now neutralized per design D4).
	PruneMemberships(ctx context.Context, userID uuid.UUID) error
}
