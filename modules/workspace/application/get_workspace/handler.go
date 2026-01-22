package getworkspace

import (
	"context"
	"database/sql"
	"net/http"

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
		if err == sql.ErrNoRows {
			return nil, ErrWorkspaceNotFound
		}
		return nil, err
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

func (h *GetWorkspaceHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	// Not used directly as we parse ID in module
}
