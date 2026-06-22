package intwiring

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
)

// MemberEmailResolver resolves a workspace's active member emails for the
// integration dispatcher's default-email fallback (seed a workspace email
// destination when none is configured). It excludes removed members, soft-deleted
// users, and users who turned off email notifications — mirroring the social
// emailer's recipient query.
type MemberEmailResolver struct {
	db *sql.DB
}

// NewMemberEmailResolver builds a MemberEmailResolver backed by the shared DB pool.
func NewMemberEmailResolver(db *sql.DB) *MemberEmailResolver {
	return &MemberEmailResolver{db: db}
}

// WorkspaceMemberEmails returns the email addresses of all active members of the
// given workspace within the tenant schema.
func (r *MemberEmailResolver) WorkspaceMemberEmails(ctx context.Context, tenant string, workspaceID uuid.UUID) ([]string, error) {
	var emails []string
	err := database.WithTenant(ctx, r.db, tenant, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT u.email
			   FROM workspace_members wm
			   JOIN public.users u ON u.id = wm.user_id
			  WHERE wm.workspace_id = $1
			    AND wm.removed_at IS NULL
			    AND u.deleted_at IS NULL
			    AND COALESCE(u.email_notifications_enabled, TRUE) = TRUE`, workspaceID)
		if err != nil {
			return fmt.Errorf("resolve workspace member emails for %s: %w", workspaceID, err)
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var email string
			if err := rows.Scan(&email); err != nil {
				return err
			}
			if email != "" {
				emails = append(emails, email)
			}
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return emails, nil
}
