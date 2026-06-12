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

// PruneMemberships soft-deletes all active organization memberships for userID.
// Called after CascadeSolelyOwnedOrgs so solely-owned orgs are already gone;
// this cleans up remaining member/co-owner records (OQ-1, DAO-13 member path).
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
	return nil
}
