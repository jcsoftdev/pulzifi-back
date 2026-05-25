package authwiring

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
)

// ── Compile-time interface check ──────────────────────────────────────────────

var _ authservices.OrganizationMembershipChecker = (*MembershipChecker)(nil)

// MembershipChecker implements authservices.OrganizationMembershipChecker via
// a raw SQL query on the public schema. It lives here so the auth module never
// imports organization infrastructure packages directly.
type MembershipChecker struct {
	db *sql.DB
}

// NewMembershipChecker constructs a MembershipChecker backed by the given
// *sql.DB. The db must be connected to the public schema (no SET search_path
// required — the query fully-qualifies the table).
func NewMembershipChecker(db *sql.DB) authservices.OrganizationMembershipChecker {
	return &MembershipChecker{db: db}
}

// HasAnyMembership reports whether userID has at least one non-deleted row in
// public.organization_members. Soft-deleted rows (deleted_at IS NOT NULL) are
// excluded so a previously-removed member is treated as having no membership.
func (c *MembershipChecker) HasAnyMembership(ctx context.Context, userID uuid.UUID) (bool, error) {
	var exists bool
	err := c.db.QueryRowContext(ctx,
		`SELECT EXISTS(
			SELECT 1
			FROM public.organization_members
			WHERE user_id = $1
			  AND deleted_at IS NULL
		)`,
		userID,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
