package createpage

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/page/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/page/domain/repositories"
)

type CreatePageHandler struct {
	repo repositories.PageRepository
}

func NewCreatePageHandler(repo repositories.PageRepository) *CreatePageHandler {
	return &CreatePageHandler{repo: repo}
}

func (h *CreatePageHandler) Handle(ctx context.Context, req *CreatePageRequest, createdBy uuid.UUID) (*CreatePageResponse, error) {
	// Create page entity
	page := &entities.Page{
		ID:          uuid.New(),
		WorkspaceID: req.WorkspaceID,
		Name:        req.Name,
		URL:         req.URL,
		CheckCount:  0,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save to database
	if err := h.repo.Create(ctx, page); err != nil {
		return nil, err
	}

	// Return response
	return &CreatePageResponse{
		ID:          page.ID,
		WorkspaceID: page.WorkspaceID,
		Name:        page.Name,
		URL:         page.URL,
		CreatedBy:   page.CreatedBy,
		CreatedAt:   page.CreatedAt,
	}, nil
}

// HTTP Handler wrapper
func (h *CreatePageHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	var req CreatePageRequest

	// Parse JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.WorkspaceID == uuid.Nil || req.Name == "" || req.URL == "" {
		http.Error(w, "workspace_id, name, and url are required", http.StatusBadRequest)
		return
	}

	// For now, use a placeholder user ID
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
