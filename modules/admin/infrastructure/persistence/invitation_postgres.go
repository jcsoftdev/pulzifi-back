package persistence

import (
	"context"
	"database/sql"
	stderrors "errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/entities"
	domerrs "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories"
)

// InvitationPostgres is a PostgreSQL-backed implementation of
// repositories.InvitationRepository. All invitation rows live in the public
// schema (`registration_requests` table) and are distinguished from
// self-service registration rows by `invitation_token IS NOT NULL`.
type InvitationPostgres struct {
	db *sql.DB
}

// NewInvitationPostgres constructs a new InvitationPostgres repository.
func NewInvitationPostgres(db *sql.DB) *InvitationPostgres {
	return &InvitationPostgres{db: db}
}

const invitationCols = `id, email, invitation_token, status, invited_by, expires_at,
    email_sent_at, email_error, revoked_by, revoked_at, resent_count, last_resent_at,
    user_id, organization_name, organization_subdomain, created_at, updated_at`

// scanInvitation decodes a single registration_requests row into an Invitation
// entity. Accepts any scanner (pgx Row, sql.Row, sql.Rows).
func (r *InvitationPostgres) scanInvitation(scanner interface {
	Scan(...interface{}) error
}) (*entities.Invitation, error) {
	var inv entities.Invitation
	var (
		email        sql.NullString
		token        sql.NullString
		invitedBy    sql.NullString
		expiresAt    sql.NullTime
		userID       sql.NullString
		orgName      sql.NullString
		subd         sql.NullString
		emailErrCol  sql.NullString
		revokedBy    sql.NullString
		emailSentAt  sql.NullTime
		revokedAt    sql.NullTime
		lastResentAt sql.NullTime
		statusStr    string
	)
	if err := scanner.Scan(
		&inv.ID, &email, &token, &statusStr, &invitedBy, &expiresAt,
		&emailSentAt, &emailErrCol, &revokedBy, &revokedAt, &inv.ResentCount, &lastResentAt,
		&userID, &orgName, &subd, &inv.CreatedAt, &inv.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if email.Valid {
		inv.Email = email.String
	}
	if token.Valid {
		inv.Token = token.String
	}
	inv.Status = entities.InvitationStatus(statusStr)
	if invitedBy.Valid {
		if id, err := uuid.Parse(invitedBy.String); err == nil {
			inv.InvitedBy = id
		}
	}
	if expiresAt.Valid {
		inv.ExpiresAt = expiresAt.Time
	}
	if emailSentAt.Valid {
		t := emailSentAt.Time
		inv.EmailSentAt = &t
	}
	if emailErrCol.Valid {
		s := emailErrCol.String
		inv.EmailError = &s
	}
	if revokedBy.Valid {
		if id, err := uuid.Parse(revokedBy.String); err == nil {
			inv.RevokedBy = &id
		}
	}
	if revokedAt.Valid {
		t := revokedAt.Time
		inv.RevokedAt = &t
	}
	if lastResentAt.Valid {
		t := lastResentAt.Time
		inv.LastResentAt = &t
	}
	if userID.Valid {
		if id, err := uuid.Parse(userID.String); err == nil {
			inv.AcceptedUserID = &id
		}
	}
	if orgName.Valid {
		s := orgName.String
		inv.OrgName = &s
	}
	if subd.Valid {
		s := subd.String
		inv.OrgSubdomain = &s
	}
	return &inv, nil
}

// Create inserts a new pending invitation row, enforcing per-inviter and global
// daily caps under a transaction-scoped advisory lock keyed on the lowercased
// email. Stale pending rows whose expiry has passed are transitioned to
// 'expired' first so the partial-unique index does not falsely block reuse.
func (r *InvitationPostgres) Create(ctx context.Context, in repositories.CreateInvitationInput) (*entities.Invitation, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	emailLower := strings.ToLower(in.Email)

	if _, err := tx.ExecContext(ctx,
		`SELECT pg_advisory_xact_lock(hashtextextended('invite:' || $1, 0))`,
		emailLower,
	); err != nil {
		return nil, fmt.Errorf("advisory lock: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE public.registration_requests
		SET status='expired', updated_at=now()
		WHERE email=$1
		  AND invitation_token IS NOT NULL
		  AND status='pending'
		  AND expires_at < now()
	`, emailLower); err != nil {
		return nil, fmt.Errorf("transition expired: %w", err)
	}

	var inviterCount int
	if err := tx.QueryRowContext(ctx, `
		SELECT count(*) FROM public.registration_requests
		WHERE invited_by = $1 AND created_at > now() - interval '24 hours'
	`, in.InvitedBy).Scan(&inviterCount); err != nil {
		return nil, fmt.Errorf("inviter cap query: %w", err)
	}
	if inviterCount >= in.DailyCapPerInviter {
		return nil, domerrs.ErrDailyCapExceeded
	}

	var globalCount int
	if err := tx.QueryRowContext(ctx, `
		SELECT count(*) FROM public.registration_requests
		WHERE invitation_token IS NOT NULL AND created_at > now() - interval '24 hours'
	`).Scan(&globalCount); err != nil {
		return nil, fmt.Errorf("global cap query: %w", err)
	}
	if globalCount >= in.DailyCapGlobal {
		return nil, domerrs.ErrDailyCapExceeded
	}

	id := uuid.New()
	now := time.Now().UTC()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO public.registration_requests
		    (id, status, email, invitation_token, invited_by, expires_at, created_at, updated_at)
		VALUES ($1, 'pending', $2, $3, $4, $5, $6, $6)
	`, id, emailLower, in.Token, in.InvitedBy, in.ExpiresAt, now)
	if err != nil {
		var pqErr *pq.Error
		if stderrors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, domerrs.ErrCannotInviteEmail
		}
		return nil, fmt.Errorf("insert invitation: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

// GetByToken retrieves an invitation by its token. Returns ErrInvitationNotFound
// when no row matches.
func (r *InvitationPostgres) GetByToken(ctx context.Context, token string) (*entities.Invitation, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+invitationCols+` FROM public.registration_requests WHERE invitation_token = $1`, token)
	inv, err := r.scanInvitation(row)
	if stderrors.Is(err, sql.ErrNoRows) {
		return nil, domerrs.ErrInvitationNotFound
	}
	if err != nil {
		return nil, err
	}
	return inv, nil
}

// GetByID retrieves an invitation by its primary key (only invitation rows).
func (r *InvitationPostgres) GetByID(ctx context.Context, id uuid.UUID) (*entities.Invitation, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+invitationCols+` FROM public.registration_requests WHERE id = $1 AND invitation_token IS NOT NULL`, id)
	inv, err := r.scanInvitation(row)
	if stderrors.Is(err, sql.ErrNoRows) {
		return nil, domerrs.ErrInvitationNotFound
	}
	if err != nil {
		return nil, err
	}
	return inv, nil
}

// List returns a paginated slice of invitations and the total matching count
// (ignoring limit/offset). Limit is clamped to (0, 200] with default 50.
func (r *InvitationPostgres) List(ctx context.Context, f repositories.ListInvitationsFilter) ([]*entities.Invitation, int, error) {
	if f.Limit <= 0 || f.Limit > 200 {
		f.Limit = 50
	}
	args := []interface{}{}
	where := "invitation_token IS NOT NULL"
	if f.Status != "" {
		args = append(args, f.Status)
		where += fmt.Sprintf(" AND status = $%d", len(args))
	}
	listArgs := append(append([]interface{}{}, args...), f.Limit, f.Offset)
	limitIdx := len(args) + 1
	offsetIdx := len(args) + 2
	rows, err := r.db.QueryContext(ctx,
		fmt.Sprintf(`SELECT %s FROM public.registration_requests WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
			invitationCols, where, limitIdx, offsetIdx),
		listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []*entities.Invitation
	for rows.Next() {
		inv, err := r.scanInvitation(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, inv)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	totalArgs := append([]interface{}{}, args...)
	if err := r.db.QueryRowContext(ctx,
		fmt.Sprintf(`SELECT count(*) FROM public.registration_requests WHERE %s`, where), totalArgs...,
	).Scan(&total); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// AcceptInvitation atomically materialises the invited user, organization, and
// owner membership, calls the supplied ProvisionFunc to create the tenant
// schema, and marks the invitation accepted. The schema name is derived from
// the subdomain by replacing hyphens with underscores.
//
// Returns ErrInvitationNotFound if the token is absent, expired, or already
// non-pending; ErrCannotInviteEmail if a user with the same email already
// exists; ErrSubdomainTaken if the subdomain is already used; and
// ErrSchemaProvisioning (wrapped) if ProvisionFunc fails.
func (r *InvitationPostgres) AcceptInvitation(ctx context.Context, in repositories.AcceptInvitationInput) (*repositories.AcceptInvitationOutput, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var inviteID uuid.UUID
	var email string
	var status string
	var expiresAt time.Time
	if err := tx.QueryRowContext(ctx,
		`SELECT id, email, status, expires_at FROM public.registration_requests WHERE invitation_token = $1 FOR UPDATE`,
		in.Token,
	).Scan(&inviteID, &email, &status, &expiresAt); err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return nil, domerrs.ErrInvitationNotFound
		}
		return nil, err
	}
	if status != string(entities.InvitationPending) || time.Now().After(expiresAt) {
		return nil, domerrs.ErrInvitationNotFound
	}

	userID := uuid.New()
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO public.users (id, email, password_hash, first_name, last_name, status, email_verified)
		VALUES ($1, $2, $3, $4, $5, 'approved', TRUE)
	`, userID, email, in.PasswordHash, in.FirstName, in.LastName); err != nil {
		var pqErr *pq.Error
		if stderrors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, domerrs.ErrCannotInviteEmail
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}

	orgID := uuid.New()
	schemaName := strings.ReplaceAll(strings.ToLower(in.OrgSubdomain), "-", "_")
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO public.organizations (id, name, subdomain, schema_name, owner_user_id, schema_provisioning_failed_at)
		VALUES ($1, $2, $3, $4, $5, now())
	`, orgID, in.OrgName, in.OrgSubdomain, schemaName, userID); err != nil {
		var pqErr *pq.Error
		if stderrors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, domerrs.ErrSubdomainTaken
		}
		return nil, fmt.Errorf("insert org: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO public.organization_members (organization_id, user_id, role, invited_by, invitation_status)
		VALUES ($1, $2, 'OWNER', $3, 'active')
	`, orgID, userID, userID); err != nil {
		return nil, fmt.Errorf("insert member: %w", err)
	}

	if in.ProvisionFunc != nil {
		if err := in.ProvisionFunc(schemaName); err != nil {
			return nil, fmt.Errorf("%w: %v", domerrs.ErrSchemaProvisioning, err)
		}
	}

	if _, err = tx.ExecContext(ctx,
		`UPDATE public.organizations SET schema_provisioning_failed_at = NULL WHERE id = $1`, orgID,
	); err != nil {
		return nil, err
	}

	if _, err = tx.ExecContext(ctx, `
		UPDATE public.registration_requests
		SET status='accepted', accepted_at=now() AT TIME ZONE 'UTC',
		    user_id=$1, organization_name=$2, organization_subdomain=$3, updated_at=now()
		WHERE id=$4
	`, userID, in.OrgName, in.OrgSubdomain, inviteID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	committed = true

	return &repositories.AcceptInvitationOutput{
		UserID:        userID,
		OrgID:         orgID,
		OrgSubdomain:  in.OrgSubdomain,
		OrgSchemaName: schemaName,
	}, nil
}

// Revoke marks a pending invitation as revoked. Returns
// ErrInvitationAlreadyDecided if the invitation is missing, already accepted,
// already revoked, or not an invitation row.
func (r *InvitationPostgres) Revoke(ctx context.Context, id uuid.UUID, revokedBy uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE public.registration_requests
		SET status='revoked', revoked_by=$1, revoked_at=now(), updated_at=now()
		WHERE id=$2 AND status='pending' AND invitation_token IS NOT NULL
	`, revokedBy, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domerrs.ErrInvitationAlreadyDecided
	}
	return nil
}

// UpdateEmailDeliveryStatus records the outcome of the asynchronous email send
// attempt for an invitation. Either sentAt or sendErr should be set.
func (r *InvitationPostgres) UpdateEmailDeliveryStatus(ctx context.Context, id uuid.UUID, sentAt *time.Time, sendErr *string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE public.registration_requests
		SET email_sent_at=$1, email_error=$2, updated_at=now()
		WHERE id=$3
	`, sentAt, sendErr, id)
	return err
}

// RotateForResend issues a new token and expiry for an existing pending
// invitation, incrementing resent_count. Returns ErrInvitationAlreadyDecided if
// the row is no longer pending.
func (r *InvitationPostgres) RotateForResend(ctx context.Context, id uuid.UUID, newToken string, newExpiresAt time.Time) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE public.registration_requests
		SET invitation_token=$1, expires_at=$2, resent_count=resent_count+1, last_resent_at=now(), updated_at=now()
		WHERE id=$3 AND status='pending' AND invitation_token IS NOT NULL
	`, newToken, newExpiresAt, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domerrs.ErrInvitationAlreadyDecided
	}
	return nil
}
