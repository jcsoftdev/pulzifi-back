package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/team/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/team/domain/repositories"
)

type teamMemberPostgresRepository struct {
	db *sql.DB
}

func NewTeamMemberPostgresRepository(db *sql.DB) repositories.TeamMemberRepository {
	return &teamMemberPostgresRepository{db: db}
}

func (r *teamMemberPostgresRepository) GetOrganizationIDBySubdomain(ctx context.Context, subdomain string) (uuid.UUID, error) {
	query := `SELECT id FROM public.organizations WHERE subdomain = $1 AND deleted_at IS NULL`
	var orgID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, subdomain).Scan(&orgID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("organization not found for subdomain %s: %w", subdomain, err)
	}
	return orgID, nil
}

func (r *teamMemberPostgresRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*entities.TeamMember, error) {
	query := `
		SELECT
			om.id, om.organization_id, om.user_id, om.role, om.invited_by, om.joined_at,
			u.first_name, u.last_name, u.email, u.avatar_url
		FROM public.organization_members om
		INNER JOIN public.users u ON om.user_id = u.id
		WHERE om.organization_id = $1 AND om.deleted_at IS NULL AND u.deleted_at IS NULL
		ORDER BY om.joined_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*entities.TeamMember
	for rows.Next() {
		m := &entities.TeamMember{}
		err := rows.Scan(
			&m.ID, &m.OrganizationID, &m.UserID, &m.Role, &m.InvitedBy, &m.JoinedAt,
			&m.FirstName, &m.LastName, &m.Email, &m.AvatarURL,
		)
		if err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *teamMemberPostgresRepository) GetByID(ctx context.Context, memberID uuid.UUID) (*entities.TeamMember, error) {
	query := `
		SELECT
			om.id, om.organization_id, om.user_id, om.role, om.invited_by, om.joined_at,
			u.first_name, u.last_name, u.email, u.avatar_url
		FROM public.organization_members om
		INNER JOIN public.users u ON om.user_id = u.id
		WHERE om.id = $1 AND om.deleted_at IS NULL
	`
	m := &entities.TeamMember{}
	err := r.db.QueryRowContext(ctx, query, memberID).Scan(
		&m.ID, &m.OrganizationID, &m.UserID, &m.Role, &m.InvitedBy, &m.JoinedAt,
		&m.FirstName, &m.LastName, &m.Email, &m.AvatarURL,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *teamMemberPostgresRepository) GetByUserAndOrg(ctx context.Context, orgID, userID uuid.UUID) (*entities.TeamMember, error) {
	query := `
		SELECT
			om.id, om.organization_id, om.user_id, om.role, om.invited_by, om.joined_at,
			u.first_name, u.last_name, u.email, u.avatar_url
		FROM public.organization_members om
		INNER JOIN public.users u ON om.user_id = u.id
		WHERE om.organization_id = $1 AND om.user_id = $2 AND om.deleted_at IS NULL
	`
	m := &entities.TeamMember{}
	err := r.db.QueryRowContext(ctx, query, orgID, userID).Scan(
		&m.ID, &m.OrganizationID, &m.UserID, &m.Role, &m.InvitedBy, &m.JoinedAt,
		&m.FirstName, &m.LastName, &m.Email, &m.AvatarURL,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *teamMemberPostgresRepository) FindUserByEmail(ctx context.Context, email string) (*entities.TeamMember, error) {
	query := `
		SELECT id, first_name, last_name, email, avatar_url
		FROM public.users
		WHERE email = $1 AND deleted_at IS NULL
	`
	m := &entities.TeamMember{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&m.UserID, &m.FirstName, &m.LastName, &m.Email, &m.AvatarURL,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *teamMemberPostgresRepository) AddMember(ctx context.Context, orgID, userID uuid.UUID, role string, invitedBy *uuid.UUID) (*entities.TeamMember, error) {
	id := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO public.organization_members (id, organization_id, user_id, role, invited_by, joined_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query, id, orgID, userID, role, invitedBy, now)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *teamMemberPostgresRepository) UpdateRole(ctx context.Context, memberID uuid.UUID, role string) error {
	query := `UPDATE public.organization_members SET role = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, role, memberID)
	return err
}

func (r *teamMemberPostgresRepository) Remove(ctx context.Context, memberID uuid.UUID) error {
	query := `UPDATE public.organization_members SET deleted_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), memberID)
	return err
}
