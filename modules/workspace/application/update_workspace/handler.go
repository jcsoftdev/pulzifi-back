package updateworkspace

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/repositories"
)

type UpdateWorkspaceHandler struct {
	repo repositories.WorkspaceRepository
}

func NewUpdateWorkspaceHandler(repo repositories.WorkspaceRepository) *UpdateWorkspaceHandler {
	return &UpdateWorkspaceHandler{repo: repo}
}

func (h *UpdateWorkspaceHandler) Handle(ctx context.Context, id uuid.UUID, req *UpdateWorkspaceRequest) (*UpdateWorkspaceResponse, error) {
	// Get existing workspace
	workspace, err := h.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update only provided fields
	if req.Name != nil {
		workspace.Name = *req.Name
	}
	if req.Type != nil {
		workspace.Type = *req.Type
	}
	if req.Tags != nil {
		workspace.Tags = *req.Tags
	}

	// Update timestamp
	workspace.UpdatedAt = time.Now()

	// Save to database
	if err := h.repo.Update(ctx, workspace); err != nil {
		return nil, err
	}

	// Return response
	return &UpdateWorkspaceResponse{
		ID:        workspace.ID,
		Name:      workspace.Name,
		Type:      workspace.Type,
		Tags:      workspace.Tags,
		CreatedBy: workspace.CreatedBy,
		CreatedAt: workspace.CreatedAt,
		UpdatedAt: workspace.UpdatedAt,
	}, nil
}
