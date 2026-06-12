package authwiring

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
)

// membershipPrunerAdapter implements authservices.UserCleanup via raw SQL.
// It soft-deletes all remaining organization memberships for a user after the
// org cascade has already handled solely-owned orgs (OQ-1, design §5).
type membershipPrunerAdapter struct {
	db *sql.DB
}

// compile-time assertion
var _ authservices.UserCleanup = (*membershipPrunerAdapter)(nil)

// NewMembershipPrunerAdapter constructs the adapter.
func NewMembershipPrunerAdapter(db *sql.DB) authservices.UserCleanup {
	return &membershipPrunerAdapter{db: db}
}

// PruneMemberships soft-deletes all active organization memberships for userID
// and hard-deletes the user's role grants from public.user_roles.
//
// Invariant: this is called after CascadeSolelyOwnedOrgs, so solely-owned orgs
// are already gone. Any surviving 'owner' row in organization_members belongs to
// a co-owned org — another active owner row still exists there, so pruning this
// user's row cannot create an ownerless org (OQ-1, DAO-13 member path).
//
// user_roles cleanup is necessary because UserRepository.Delete performs a
// soft-delete (sets deleted_at) rather than a hard-delete, so the
// "user_roles.user_id ON DELETE CASCADE" FK never fires. Without this explicit
// DELETE, role rows would orphan for soft-deleted users (design §5, DAO-14).
func (a *membershipPrunerAdapter) PruneMemberships(ctx context.Context, userID uuid.UUID) error {
	_, err := a.db.ExecContext(ctx, `
		UPDATE public.organization_members
		   SET deleted_at = NOW()
		 WHERE user_id    = $1
		   AND deleted_at IS NULL
	`, userID)
	if err != nil {
		return fmt.Errorf("membership_pruner_adapter: prune memberships: %w", err)
	}

	// Remove role grants for the user. Hard-delete is correct here: role rows
	// have no soft-delete column and carry no audit value after account deletion.
	_, err = a.db.ExecContext(ctx, `
		DELETE FROM public.user_roles
		 WHERE user_id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("membership_pruner_adapter: delete user_roles: %w", err)
	}

	return nil
}
