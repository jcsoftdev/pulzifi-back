package createworkspace

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/repositories"
)

type CreateWorkspaceHandler struct {
	repo repositories.WorkspaceRepository
}

func NewCreateWorkspaceHandler(repo repositories.WorkspaceRepository) *CreateWorkspaceHandler {
	return &CreateWorkspaceHandler{repo: repo}
}

func (h *CreateWorkspaceHandler) Handle(ctx context.Context, req *CreateWorkspaceRequest, createdBy uuid.UUID) (*CreateWorkspaceResponse, error) {
	// Create workspace entity
	workspace := &entities.Workspace{
		ID:        uuid.New(),
		Name:      req.Name,
		Type:      req.Type,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to database
	if err := h.repo.Create(ctx, workspace); err != nil {
		return nil, err
	}

	// Return response
	return &CreateWorkspaceResponse{
		ID:        workspace.ID,
		Name:      workspace.Name,
		Type:      workspace.Type,
		CreatedBy: workspace.CreatedBy,
		CreatedAt: workspace.CreatedAt,
	}, nil
}

// HTTP Handler wrapper
func (h *CreateWorkspaceHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	var req CreateWorkspaceRequest

	// Parse JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" || req.Type == "" {
		http.Error(w, "name and type are required", http.StatusBadRequest)
		return
	}

	// For now, use a placeholder user ID (from JWT/auth in real scenario)
	createdBy := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	// Execute use case
	resp, err := h.Handle(r.Context(), &req, createdBy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
