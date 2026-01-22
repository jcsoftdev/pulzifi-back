package createworkspace

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	authmw "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/middleware"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/value_objects"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
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
		Tags:      req.Tags,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to database
	if err := h.repo.Create(ctx, workspace); err != nil {
		return nil, err
	}

	// Add creator as owner to workspace_members
	member := entities.NewWorkspaceMember(
		workspace.ID,
		createdBy,
		value_objects.RoleOwner,
		nil, // No inviter for the creator
	)

	if err := h.repo.AddMember(ctx, member); err != nil {
		logger.Error("Failed to add creator as owner", zap.Error(err))
		// Note: Workspace is created but member not added - should handle this in production
		return nil, err
	}

	// Return response
	return &CreateWorkspaceResponse{
		ID:        workspace.ID,
		Name:      workspace.Name,
		Type:      workspace.Type,
		Tags:      workspace.Tags,
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

	// Get user ID from context (set by auth middleware)
	userIDStr, ok := r.Context().Value(authmw.UserIDKey).(string)
	if !ok {
		logger.Error("User ID not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	createdBy, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Error("Invalid user ID", zap.Error(err))
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

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
