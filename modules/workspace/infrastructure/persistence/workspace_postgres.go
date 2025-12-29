package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/value_objects"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// WorkspacePostgresRepository implements WorkspaceRepository using PostgreSQL (tenant schema)
type WorkspacePostgresRepository struct {
	db     *sql.DB
	tenant string // Schema name for this tenant
}

// NewWorkspacePostgresRepository creates a new PostgreSQL repository for tenant schema
func NewWorkspacePostgresRepository(db *sql.DB, tenant string) *WorkspacePostgresRepository {
	return &WorkspacePostgresRepository{
		db:     db,
		tenant: tenant,
	}
}

// Create stores a new workspace in tenant schema
func (r *WorkspacePostgresRepository) Create(ctx context.Context, workspace *entities.Workspace) error {
	query := `
		INSERT INTO workspaces (id, name, type, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	// Set search path to tenant schema
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		logger.Error("Failed to set search path", zap.Error(err))
		return err
	}

	_, err := r.db.ExecContext(ctx, query,
		workspace.ID,
		workspace.Name,
		workspace.Type,
		workspace.CreatedBy,
		workspace.CreatedAt,
		workspace.UpdatedAt,
	)

	if err != nil {
		logger.Error("Failed to create workspace", zap.Error(err))
		return err
	}

	return nil
}

// GetByID retrieves a workspace from tenant schema
func (r *WorkspacePostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Workspace, error) {
	query := `
		SELECT id, name, type, created_by, created_at, updated_at, deleted_at
		FROM workspaces
		WHERE id = $1 AND deleted_at IS NULL
	`

	// Set search path to tenant schema
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		logger.Error("Failed to set search path", zap.Error(err))
		return nil, err
	}

	var workspace entities.Workspace
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&workspace.ID,
		&workspace.Name,
		&workspace.Type,
		&workspace.CreatedBy,
		&workspace.CreatedAt,
		&workspace.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Error("Failed to get workspace by ID", zap.Error(err))
		return nil, err
	}

	if deletedAt.Valid {
		workspace.DeletedAt = &deletedAt.Time
	}

	return &workspace, nil
}

// List retrieves all workspaces in tenant schema
func (r *WorkspacePostgresRepository) List(ctx context.Context) ([]*entities.Workspace, error) {
	query := `
		SELECT id, name, type, created_by, created_at, updated_at, deleted_at
		FROM workspaces
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`

	// Set search path to tenant schema
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		logger.Error("Failed to set search path", zap.Error(err))
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		logger.Error("Failed to list workspaces", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var workspaces []*entities.Workspace
	for rows.Next() {
		var workspace entities.Workspace
		var deletedAt sql.NullTime

		if err := rows.Scan(
			&workspace.ID,
			&workspace.Name,
			&workspace.Type,
			&workspace.CreatedBy,
			&workspace.CreatedAt,
			&workspace.UpdatedAt,
			&deletedAt,
		); err != nil {
			logger.Error("Failed to scan workspace", zap.Error(err))
			return nil, err
		}

		if deletedAt.Valid {
			workspace.DeletedAt = &deletedAt.Time
		}

		workspaces = append(workspaces, &workspace)
	}

	return workspaces, nil
}

// ListByCreator retrieves all workspaces created by a user
func (r *WorkspacePostgresRepository) ListByCreator(ctx context.Context, createdBy uuid.UUID) ([]*entities.Workspace, error) {
	query := `
		SELECT id, name, type, created_by, created_at, updated_at, deleted_at
		FROM workspaces
		WHERE created_by = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	// Set search path to tenant schema
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		logger.Error("Failed to set search path", zap.Error(err))
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, createdBy)
	if err != nil {
		logger.Error("Failed to list workspaces", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var workspaces []*entities.Workspace
	for rows.Next() {
		var workspace entities.Workspace
		var deletedAt sql.NullTime

		if err := rows.Scan(
			&workspace.ID,
			&workspace.Name,
			&workspace.Type,
			&workspace.CreatedBy,
			&workspace.CreatedAt,
			&workspace.UpdatedAt,
			&deletedAt,
		); err != nil {
			logger.Error("Failed to scan workspace", zap.Error(err))
			return nil, err
		}

		if deletedAt.Valid {
			workspace.DeletedAt = &deletedAt.Time
		}

		workspaces = append(workspaces, &workspace)
	}

	return workspaces, nil
}

// Update modifies an existing workspace
func (r *WorkspacePostgresRepository) Update(ctx context.Context, workspace *entities.Workspace) error {
	query := `
		UPDATE workspaces
		SET name = $1, type = $2, updated_at = $3
		WHERE id = $4
	`

	// Set search path to tenant schema
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		logger.Error("Failed to set search path", zap.Error(err))
		return err
	}

	workspace.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		workspace.Name,
		workspace.Type,
		workspace.UpdatedAt,
		workspace.ID,
	)

	if err != nil {
		logger.Error("Failed to update workspace", zap.Error(err))
		return err
	}

	return nil
}

// Delete soft-deletes a workspace
func (r *WorkspacePostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE workspaces
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	// Set search path to tenant schema
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		logger.Error("Failed to set search path", zap.Error(err))
		return err
	}

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		logger.Error("Failed to delete workspace", zap.Error(err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", zap.Error(err))
		return err
	}

	if rowsAffected == 0 {
		logger.Warn("Workspace not found for deletion", zap.String("id", id.String()))
		return nil
	}

	return nil
}

// AddMember adds a user to a workspace
func (r *WorkspacePostgresRepository) AddMember(ctx context.Context, member *entities.WorkspaceMember) error {
	query := `
		INSERT INTO workspace_members (workspace_id, user_id, role, invited_by, invited_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	// Set search path to tenant schema
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		logger.Error("Failed to set search path", zap.Error(err))
		return err
	}

	_, err := r.db.ExecContext(ctx, query,
		member.WorkspaceID,
		member.UserID,
		member.Role.String(),
		member.InvitedBy,
		member.InvitedAt,
	)

	if err != nil {
		logger.Error("Failed to add workspace member", zap.Error(err))
		return err
	}

	return nil
}

// GetMember retrieves a workspace member
func (r *WorkspacePostgresRepository) GetMember(ctx context.Context, workspaceID, userID uuid.UUID) (*entities.WorkspaceMember, error) {
	query := `
		SELECT workspace_id, user_id, role, invited_by, invited_at
		FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2 AND removed_at IS NULL
	`

	// Set search path to tenant schema
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		logger.Error("Failed to set search path", zap.Error(err))
		return nil, err
	}

	var member entities.WorkspaceMember
	var roleStr string
	var invitedBy sql.NullString

	err := r.db.QueryRowContext(ctx, query, workspaceID, userID).Scan(
		&member.WorkspaceID,
		&member.UserID,
		&roleStr,
		&invitedBy,
		&member.InvitedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Error("Failed to get workspace member", zap.Error(err))
		return nil, err
	}

	role, err := value_objects.NewWorkspaceRole(roleStr)
	if err != nil {
		logger.Error("Invalid role in database", zap.String("role", roleStr))
		return nil, err
	}
	member.Role = role

	if invitedBy.Valid {
		invitedByUUID, err := uuid.Parse(invitedBy.String)
		if err == nil {
			member.InvitedBy = &invitedByUUID
		}
	}

	return &member, nil
}

// ListMembers retrieves all members of a workspace
func (r *WorkspacePostgresRepository) ListMembers(ctx context.Context, workspaceID uuid.UUID) ([]*entities.WorkspaceMember, error) {
	query := `
		SELECT workspace_id, user_id, role, invited_by, invited_at
		FROM workspace_members
		WHERE workspace_id = $1 AND removed_at IS NULL
		ORDER BY invited_at ASC
	`

	// Set search path to tenant schema
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		logger.Error("Failed to set search path", zap.Error(err))
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, workspaceID)
	if err != nil {
		logger.Error("Failed to list workspace members", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var members []*entities.WorkspaceMember
	for rows.Next() {
		var member entities.WorkspaceMember
		var roleStr string
		var invitedBy sql.NullString

		if err := rows.Scan(
			&member.WorkspaceID,
			&member.UserID,
			&roleStr,
			&invitedBy,
			&member.InvitedAt,
		); err != nil {
			logger.Error("Failed to scan workspace member", zap.Error(err))
			return nil, err
		}

		role, err := value_objects.NewWorkspaceRole(roleStr)
		if err != nil {
			logger.Error("Invalid role in database", zap.String("role", roleStr))
			continue
		}
		member.Role = role

		if invitedBy.Valid {
			invitedByUUID, err := uuid.Parse(invitedBy.String)
			if err == nil {
				member.InvitedBy = &invitedByUUID
			}
		}

		members = append(members, &member)
	}

	return members, nil
}

// ListByMember retrieves all workspaces where user is a member
func (r *WorkspacePostgresRepository) ListByMember(ctx context.Context, userID uuid.UUID) ([]*entities.Workspace, error) {
	query := `
		SELECT w.id, w.name, w.type, w.created_by, w.created_at, w.updated_at, w.deleted_at
		FROM workspaces w
		INNER JOIN workspace_members wm ON w.id = wm.workspace_id
		WHERE wm.user_id = $1 AND w.deleted_at IS NULL AND wm.removed_at IS NULL
		ORDER BY w.created_at DESC
	`

	// Set search path to tenant schema
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		logger.Error("Failed to set search path", zap.Error(err))
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		logger.Error("Failed to list workspaces by member", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var workspaces []*entities.Workspace
	for rows.Next() {
		var workspace entities.Workspace
		var deletedAt sql.NullTime

		if err := rows.Scan(
			&workspace.ID,
			&workspace.Name,
			&workspace.Type,
			&workspace.CreatedBy,
			&workspace.CreatedAt,
			&workspace.UpdatedAt,
			&deletedAt,
		); err != nil {
			logger.Error("Failed to scan workspace", zap.Error(err))
			return nil, err
		}

		if deletedAt.Valid {
			workspace.DeletedAt = &deletedAt.Time
		}

		workspaces = append(workspaces, &workspace)
	}

	return workspaces, nil
}

// UpdateMemberRole updates the role of a workspace member
func (r *WorkspacePostgresRepository) UpdateMemberRole(ctx context.Context, workspaceID, userID uuid.UUID, role value_objects.WorkspaceRole) error {
	query := `
		UPDATE workspace_members
		SET role = $1
		WHERE workspace_id = $2 AND user_id = $3
	`

	// Set search path to tenant schema
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		logger.Error("Failed to set search path", zap.Error(err))
		return err
	}

	result, err := r.db.ExecContext(ctx, query, role.String(), workspaceID, userID)
	if err != nil {
		logger.Error("Failed to update member role", zap.Error(err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", zap.Error(err))
		return err
	}

	if rowsAffected == 0 {
		logger.Warn("Member not found for role update",
			zap.String("workspace_id", workspaceID.String()),
			zap.String("user_id", userID.String()))
		return nil
	}

	return nil
}

// RemoveMember removes a user from a workspace (soft delete)
func (r *WorkspacePostgresRepository) RemoveMember(ctx context.Context, workspaceID, userID uuid.UUID) error {
	query := `
		UPDATE workspace_members
		SET removed_at = $1
		WHERE workspace_id = $2 AND user_id = $3 AND removed_at IS NULL
	`

	// Set search path to tenant schema
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		logger.Error("Failed to set search path", zap.Error(err))
		return err
	}

	result, err := r.db.ExecContext(ctx, query, time.Now(), workspaceID, userID)
	if err != nil {
		logger.Error("Failed to remove workspace member", zap.Error(err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", zap.Error(err))
		return err
	}

	if rowsAffected == 0 {
		logger.Warn("Member not found for removal",
			zap.String("workspace_id", workspaceID.String()),
			zap.String("user_id", userID.String()))
		return nil
	}

	return nil
}
