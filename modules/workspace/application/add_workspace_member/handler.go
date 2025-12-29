package add_workspace_member

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/value_objects"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type AddWorkspaceMemberHandler struct {
	repo repositories.WorkspaceRepository
}

func NewAddWorkspaceMemberHandler(repo repositories.WorkspaceRepository) *AddWorkspaceMemberHandler {
	return &AddWorkspaceMemberHandler{repo: repo}
}

func (h *AddWorkspaceMemberHandler) Handle(
	ctx context.Context,
	workspaceID uuid.UUID,
	inviterID uuid.UUID,
	req *AddWorkspaceMemberRequest,
) (*AddWorkspaceMemberResponse, error) {
	// Verify workspace exists
	workspace, err := h.repo.GetByID(ctx, workspaceID)
	if err != nil {
		logger.Error("Failed to get workspace", zap.Error(err))
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	// Check inviter's permission
	inviterMember, err := h.repo.GetMember(ctx, workspaceID, inviterID)
	if err != nil {
		logger.Error("Failed to get inviter member", zap.Error(err))
		return nil, err
	}
	if inviterMember == nil {
		return nil, ErrNotWorkspaceMember
	}
	if !inviterMember.Role.CanInvite() {
		logger.Warn("User does not have invite permission",
			zap.String("workspace_id", workspaceID.String()),
			zap.String("inviter_id", inviterID.String()),
			zap.String("role", inviterMember.Role.String()),
		)
		return nil, ErrInsufficientPermissions
	}

	// Check if user is already a member
	existingMember, err := h.repo.GetMember(ctx, workspaceID, req.UserID)
	if err != nil {
		logger.Error("Failed to check existing member", zap.Error(err))
		return nil, err
	}
	if existingMember != nil {
		return nil, ErrMemberAlreadyExists
	}

	// Validate and create role
	role, err := value_objects.NewWorkspaceRole(req.Role)
	if err != nil {
		logger.Error("Invalid role", zap.String("role", req.Role))
		return nil, err
	}

	// Create member
	member := entities.NewWorkspaceMember(
		workspaceID,
		req.UserID,
		role,
		&inviterID,
	)

	// Add member to workspace
	if err := h.repo.AddMember(ctx, member); err != nil {
		logger.Error("Failed to add workspace member", zap.Error(err))
		return nil, err
	}

	logger.Info("Member added to workspace",
		zap.String("workspace_id", workspaceID.String()),
		zap.String("user_id", req.UserID.String()),
		zap.String("role", role.String()),
		zap.String("invited_by", inviterID.String()),
	)

	return &AddWorkspaceMemberResponse{
		WorkspaceID: member.WorkspaceID,
		UserID:      member.UserID,
		Role:        member.Role.String(),
		InvitedBy:   inviterID,
		InvitedAt:   member.InvitedAt,
	}, nil
}
