package remove_workspace_member

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/value_objects"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type RemoveWorkspaceMemberHandler struct {
	repo repositories.WorkspaceRepository
}

func NewRemoveWorkspaceMemberHandler(repo repositories.WorkspaceRepository) *RemoveWorkspaceMemberHandler {
	return &RemoveWorkspaceMemberHandler{repo: repo}
}

func (h *RemoveWorkspaceMemberHandler) Handle(
	ctx context.Context,
	workspaceID uuid.UUID,
	requesterID uuid.UUID,
	memberToRemoveID uuid.UUID,
) error {
	// Verify workspace exists
	workspace, err := h.repo.GetByID(ctx, workspaceID)
	if err != nil {
		logger.Error("Failed to get workspace", zap.Error(err))
		return err
	}
	if workspace == nil {
		return ErrWorkspaceNotFound
	}

	// Check requester's permission
	requesterMember, err := h.repo.GetMember(ctx, workspaceID, requesterID)
	if err != nil {
		logger.Error("Failed to get requester member", zap.Error(err))
		return err
	}
	if requesterMember == nil {
		return ErrNotWorkspaceMember
	}
	if !requesterMember.Role.CanManageMembers() {
		logger.Warn("User does not have permission to remove members",
			zap.String("workspace_id", workspaceID.String()),
			zap.String("requester_id", requesterID.String()),
			zap.String("role", requesterMember.Role.String()),
		)
		return ErrInsufficientPermissions
	}

	// Cannot remove yourself
	if requesterID == memberToRemoveID {
		return ErrCannotRemoveSelf
	}

	// Check if member to remove exists
	memberToRemove, err := h.repo.GetMember(ctx, workspaceID, memberToRemoveID)
	if err != nil {
		logger.Error("Failed to get member to remove", zap.Error(err))
		return err
	}
	if memberToRemove == nil {
		return ErrMemberNotFound
	}

	// Cannot remove owner
	if memberToRemove.Role == value_objects.RoleOwner {
		return ErrCannotRemoveOwner
	}

	// Remove member
	if err := h.repo.RemoveMember(ctx, workspaceID, memberToRemoveID); err != nil {
		logger.Error("Failed to remove workspace member", zap.Error(err))
		return err
	}

	logger.Info("Member removed from workspace",
		zap.String("workspace_id", workspaceID.String()),
		zap.String("removed_user_id", memberToRemoveID.String()),
		zap.String("requester_id", requesterID.String()),
	)

	return nil
}
