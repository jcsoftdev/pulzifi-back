package persistence

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	getuserdetail "github.com/jcsoftdev/pulzifi-back/modules/auth/application/get_user_detail"
	listusers "github.com/jcsoftdev/pulzifi-back/modules/auth/application/list_users"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// AdminUserPostgresRepository implements list_users.AdminUserReader.
// All queries target the public schema — no SET search_path required.
type AdminUserPostgresRepository struct {
	db *sql.DB
}

// NewAdminUserPostgresRepository creates a new repository backed by db.
func NewAdminUserPostgresRepository(db *sql.DB) *AdminUserPostgresRepository {
	return &AdminUserPostgresRepository{db: db}
}

// ListUsers returns a paginated list of users with their SUPER_ADMIN status,
// status, email_verified, and org_count computed in a single SQL query.
func (r *AdminUserPostgresRepository) ListUsers(ctx context.Context, filter listusers.ListFilter, limit, offset int) ([]listusers.AdminUserRow, int, error) {
	// $1 search, $2 orgID ('' = no filter), $3 status ('' = no filter)
	const where = `
		WHERE u.deleted_at IS NULL
		  AND ($1 = '' OR u.email ILIKE '%' || $1 || '%'
		       OR (u.first_name || ' ' || u.last_name) ILIKE '%' || $1 || '%')
		  AND ($3 = '' OR u.status = $3)
		  AND ($2 = '' OR EXISTS (
		        SELECT 1 FROM public.organization_members om
		        WHERE om.user_id = u.id AND om.deleted_at IS NULL
		          AND om.organization_id::text = $2))
	`

	countQuery := `SELECT COUNT(*) FROM public.users u ` + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, filter.Search, filter.OrgID, filter.Status).Scan(&total); err != nil {
		logger.Error("AdminUserRepo: count query failed", zap.Error(err))
		return nil, 0, err
	}

	listQuery := `
		SELECT u.id, u.email, u.first_name, u.last_name, u.status, u.email_verified,
		       EXISTS(
		         SELECT 1 FROM public.user_roles ur
		         JOIN public.roles ro ON ro.id = ur.role_id
		         WHERE ur.user_id = u.id AND ro.name = 'SUPER_ADMIN'
		       ) AS is_super_admin,
		       (SELECT COUNT(*) FROM public.organization_members om2
		          WHERE om2.user_id = u.id AND om2.deleted_at IS NULL) AS org_count
		FROM public.users u ` + where + `
		ORDER BY u.created_at DESC
		LIMIT $4 OFFSET $5
	`
	rows, err := r.db.QueryContext(ctx, listQuery, filter.Search, filter.OrgID, filter.Status, limit, offset)
	if err != nil {
		logger.Error("AdminUserRepo: list query failed", zap.Error(err))
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var result []listusers.AdminUserRow
	for rows.Next() {
		var row listusers.AdminUserRow
		if err := rows.Scan(&row.ID, &row.Email, &row.FirstName, &row.LastName,
			&row.Status, &row.EmailVerified, &row.IsSuperAdmin, &row.OrgCount); err != nil {
			logger.Error("AdminUserRepo: scan failed", zap.Error(err))
			return nil, 0, err
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return result, total, nil
}

// GetUserDetail fetches full profile + org memberships for a single user.
// Implements getuserdetail.Reader.
func (r *AdminUserPostgresRepository) GetUserDetail(ctx context.Context, id uuid.UUID) (*getuserdetail.UserDetail, error) {
	const userQuery = `
		SELECT u.id, u.email, u.first_name, u.last_name, u.status, u.email_verified,
		       EXISTS(SELECT 1 FROM public.user_roles ur
		              JOIN public.roles ro ON ro.id = ur.role_id
		              WHERE ur.user_id = u.id AND ro.name = 'SUPER_ADMIN') AS is_super_admin
		FROM public.users u
		WHERE u.id = $1 AND u.deleted_at IS NULL
	`
	var d getuserdetail.UserDetail
	err := r.db.QueryRowContext(ctx, userQuery, id).Scan(
		&d.ID, &d.Email, &d.FirstName, &d.LastName, &d.Status, &d.EmailVerified, &d.IsSuperAdmin)
	if err == sql.ErrNoRows {
		return nil, getuserdetail.ErrUserNotFound
	}
	if err != nil {
		logger.Error("AdminUserRepo: detail query failed", zap.Error(err))
		return nil, err
	}

	const memQuery = `
		SELECT o.id, o.name, o.subdomain, om.role, om.invitation_status,
		       (o.owner_user_id = $1) AS is_owner
		FROM public.organization_members om
		JOIN public.organizations o ON o.id = om.organization_id
		WHERE om.user_id = $1 AND om.deleted_at IS NULL AND o.deleted_at IS NULL
		ORDER BY om.joined_at ASC
	`
	rows, err := r.db.QueryContext(ctx, memQuery, id)
	if err != nil {
		logger.Error("AdminUserRepo: memberships query failed", zap.Error(err))
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var m getuserdetail.Membership
		if err := rows.Scan(&m.OrgID, &m.OrgName, &m.Subdomain, &m.Role, &m.InvitationStatus, &m.IsOwner); err != nil {
			return nil, err
		}
		d.Memberships = append(d.Memberships, m)
	}
	return &d, rows.Err()
}

// SetStatus updates a user's status to "approved" or "suspended".
// Implements setuserstatus.Writer.
func (r *AdminUserPostgresRepository) SetStatus(ctx context.Context, id uuid.UUID, status string) error {
	const q = `UPDATE public.users SET status = $2, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, q, id, status)
	if err != nil {
		logger.Error("AdminUserRepo: set status failed", zap.Error(err))
	}
	return err
}

// IsOrgOwner returns true when userID is the owner_user_id of the given org.
// Implements setmembershiprole.Writer and removemembership.Remover.
func (r *AdminUserPostgresRepository) IsOrgOwner(ctx context.Context, userID, orgID uuid.UUID) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM public.organizations WHERE id = $2 AND owner_user_id = $1 AND deleted_at IS NULL)`
	var owner bool
	err := r.db.QueryRowContext(ctx, q, userID, orgID).Scan(&owner)
	return owner, err
}

// SetRole updates the role column of an active org membership.
// Implements setmembershiprole.Writer.
func (r *AdminUserPostgresRepository) SetRole(ctx context.Context, userID, orgID uuid.UUID, role string) error {
	const q = `UPDATE public.organization_members SET role = $3
	           WHERE user_id = $1 AND organization_id = $2 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, q, userID, orgID, role)
	if err != nil {
		logger.Error("AdminUserRepo: set role failed", zap.Error(err))
	}
	return err
}

// RemoveMembership soft-deletes an org membership by setting deleted_at = NOW().
// Implements removemembership.Remover.
func (r *AdminUserPostgresRepository) RemoveMembership(ctx context.Context, userID, orgID uuid.UUID) error {
	const q = `UPDATE public.organization_members SET deleted_at = NOW()
	           WHERE user_id = $1 AND organization_id = $2 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, q, userID, orgID)
	if err != nil {
		logger.Error("AdminUserRepo: remove membership failed", zap.Error(err))
	}
	return err
}
