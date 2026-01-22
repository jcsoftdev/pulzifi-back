package updateworkspace

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
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
		CreatedBy: workspace.CreatedBy,
		CreatedAt: workspace.CreatedAt,
		UpdatedAt: workspace.UpdatedAt,
	}, nil
}

// HTTP Handler wrapper
func (h *UpdateWorkspaceHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	var req UpdateWorkspaceRequest

	// Parse JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate at least one field is provided
	if req.Name == nil && req.Type == nil {
		http.Error(w, "at least one field (name or type) is required", http.StatusBadRequest)
		return
	}

	// Get workspace ID from URL
	workspaceIDStr := chi.URLParam(r, "id")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		logger.Error("Invalid workspace ID", zap.Error(err))
		http.Error(w, "invalid workspace ID", http.StatusBadRequest)
		return
	}

	// Execute use case
	resp, err := h.Handle(r.Context(), workspaceID, &req)
	if err != nil {
		logger.Error("Failed to update workspace", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
