package list_workspace_members

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type ListWorkspaceMembersHandler struct {
	repo repositories.WorkspaceRepository
}

func NewListWorkspaceMembersHandler(repo repositories.WorkspaceRepository) *ListWorkspaceMembersHandler {
	return &ListWorkspaceMembersHandler{repo: repo}
}

func (h *ListWorkspaceMembersHandler) Handle(
	ctx context.Context,
	workspaceID uuid.UUID,
	requesterID uuid.UUID,
) (*ListWorkspaceMembersResponse, error) {
	// Verify workspace exists
	workspace, err := h.repo.GetByID(ctx, workspaceID)
	if err != nil {
		logger.Error("Failed to get workspace", zap.Error(err))
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	// Check if requester is a member (only members can see member list)
	requesterMember, err := h.repo.GetMember(ctx, workspaceID, requesterID)
	if err != nil {
		logger.Error("Failed to get requester member", zap.Error(err))
		return nil, err
	}
	if requesterMember == nil {
		return nil, ErrNotWorkspaceMember
	}

	// Get all members
	members, err := h.repo.ListMembers(ctx, workspaceID)
	if err != nil {
		logger.Error("Failed to list workspace members", zap.Error(err))
		return nil, err
	}

	// Convert to DTOs
	memberDTOs := make([]WorkspaceMemberDTO, 0, len(members))
	for _, member := range members {
		memberDTOs = append(memberDTOs, WorkspaceMemberDTO{
			UserID:    member.UserID,
			Role:      member.Role.String(),
			InvitedBy: member.InvitedBy,
			InvitedAt: member.InvitedAt,
		})
	}

	return &ListWorkspaceMembersResponse{
		Members: memberDTOs,
	}, nil
}
