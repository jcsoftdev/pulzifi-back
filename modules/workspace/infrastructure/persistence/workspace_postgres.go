package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/entities"
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
