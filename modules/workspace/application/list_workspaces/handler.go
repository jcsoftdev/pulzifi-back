package listworkspaces

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/repositories"
)

type ListWorkspacesHandler struct {
	repo repositories.WorkspaceRepository
}

func NewListWorkspacesHandler(repo repositories.WorkspaceRepository) *ListWorkspacesHandler {
	return &ListWorkspacesHandler{repo: repo}
}

func (h *ListWorkspacesHandler) Handle(ctx context.Context) (*ListWorkspacesResponse, error) {
	// Get all workspaces from repository
	workspaces, err := h.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	// Convert entities to response DTOs
	response := &ListWorkspacesResponse{
		Workspaces: make([]Workspace, len(workspaces)),
	}

	for i, ws := range workspaces {
		response.Workspaces[i] = Workspace{
			ID:        ws.ID,
			Name:      ws.Name,
			Type:      ws.Type,
			Tags:      ws.Tags,
			CreatedBy: ws.CreatedBy,
			CreatedAt: ws.CreatedAt,
		}
	}

	return response, nil
}
