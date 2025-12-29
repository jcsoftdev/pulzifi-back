package delete_workspace

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// DeleteWorkspaceHandler handles workspace deletion
type DeleteWorkspaceHandler struct {
	repo repositories.WorkspaceRepository
	db   *sql.DB
}

// NewDeleteWorkspaceHandler creates a new handler
func NewDeleteWorkspaceHandler(repo repositories.WorkspaceRepository, db *sql.DB) *DeleteWorkspaceHandler {
	return &DeleteWorkspaceHandler{
		repo: repo,
		db:   db,
	}
}

// Handle executes the delete workspace use case
func (h *DeleteWorkspaceHandler) Handle(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID) error {
	// Verify workspace exists
	workspace, err := h.repo.GetByID(ctx, workspaceID)
	if err != nil {
		logger.Error("Failed to get workspace", zap.Error(err))
		return err
	}

	if workspace == nil {
		logger.Warn("Workspace not found", zap.String("workspace_id", workspaceID.String()))
		return ErrWorkspaceNotFound
	}

	// Check user's permission via workspace_members
	member, err := h.repo.GetMember(ctx, workspaceID, userID)
	if err != nil {
		logger.Error("Failed to get workspace member", zap.Error(err))
		return err
	}

	if member == nil {
		logger.Warn("User is not a member of workspace",
			zap.String("workspace_id", workspaceID.String()),
			zap.String("user_id", userID.String()),
		)
		return ErrWorkspaceNotOwned
	}

	// Verify user has delete permission (only owners can delete)
	if !member.Role.CanDelete() {
		logger.Warn("User does not have delete permission",
			zap.String("workspace_id", workspaceID.String()),
			zap.String("user_id", userID.String()),
			zap.String("role", member.Role.String()),
		)
		return ErrWorkspaceNotOwned
	}

	// Delete workspace (soft delete)
	if err := h.repo.Delete(ctx, workspaceID); err != nil {
		logger.Error("Failed to delete workspace", zap.Error(err))
		return err
	}

	logger.Info("Workspace deleted successfully",
		zap.String("workspace_id", workspaceID.String()),
		zap.String("user_id", userID.String()),
	)

	return nil
}
