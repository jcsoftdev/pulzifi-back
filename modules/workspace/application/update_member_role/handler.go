package update_member_role

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/value_objects"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type UpdateMemberRoleRequest struct {
	Role string `json:"role"`
}

type UpdateMemberRoleHandler struct {
	repo repositories.WorkspaceRepository
}

func NewUpdateMemberRoleHandler(repo repositories.WorkspaceRepository) *UpdateMemberRoleHandler {
	return &UpdateMemberRoleHandler{repo: repo}
}

func (h *UpdateMemberRoleHandler) Handle(
	ctx context.Context,
	workspaceID uuid.UUID,
	requesterID uuid.UUID,
	targetUserID uuid.UUID,
	req *UpdateMemberRoleRequest,
) error {
	// Validate role
	role, err := value_objects.NewWorkspaceRole(req.Role)
	if err != nil {
		return ErrInvalidRole
	}

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
		logger.Warn("User does not have permission to update member roles",
			zap.String("workspace_id", workspaceID.String()),
			zap.String("requester_id", requesterID.String()),
			zap.String("role", requesterMember.Role.String()),
		)
		return ErrInsufficientPermissions
	}

	// Cannot change own role
	if requesterID == targetUserID {
		return ErrCannotChangeOwnRole
	}

	// Check if target user is a member
	targetMember, err := h.repo.GetMember(ctx, workspaceID, targetUserID)
	if err != nil {
		logger.Error("Failed to get target member", zap.Error(err))
		return err
	}
	if targetMember == nil {
		return ErrNotWorkspaceMember
	}

	// Update member role
	if err := h.repo.UpdateMemberRole(ctx, workspaceID, targetUserID, role); err != nil {
		logger.Error("Failed to update workspace member role", zap.Error(err))
		return err
	}

	logger.Info("Member role updated in workspace",
		zap.String("workspace_id", workspaceID.String()),
		zap.String("target_user_id", targetUserID.String()),
		zap.String("new_role", role.String()),
		zap.String("requester_id", requesterID.String()),
	)

	return nil
}
