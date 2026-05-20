package getworkspace

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/repositories"
)

type GetWorkspaceHandler struct {
	repo repositories.WorkspaceRepository
}

func NewGetWorkspaceHandler(repo repositories.WorkspaceRepository) *GetWorkspaceHandler {
	return &GetWorkspaceHandler{
		repo: repo,
	}
}

func (h *GetWorkspaceHandler) Handle(ctx context.Context, id uuid.UUID) (*GetWorkspaceResponse, error) {
	workspace, err := h.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	return &GetWorkspaceResponse{
		ID:        workspace.ID,
		Name:      workspace.Name,
		Type:      workspace.Type,
		Tags:      workspace.Tags,
		CreatedBy: workspace.CreatedBy,
		CreatedAt: workspace.CreatedAt,
	}, nil
}
