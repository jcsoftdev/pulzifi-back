package createpage

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	authmw "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/middleware"
	"github.com/jcsoftdev/pulzifi-back/modules/page/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/page/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type CreatePageHandler struct {
	repo repositories.PageRepository
}

func NewCreatePageHandler(repo repositories.PageRepository) *CreatePageHandler {
	return &CreatePageHandler{repo: repo}
}

func (h *CreatePageHandler) Handle(ctx context.Context, req *CreatePageRequest, createdBy uuid.UUID) (*CreatePageResponse, error) {
	// Create page entity
	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}

	page := &entities.Page{
		ID:              uuid.New(),
		WorkspaceID:     req.WorkspaceID,
		Name:            req.Name,
		URL:             req.URL,
		CheckCount:      0,
		Tags:            tags,
		CheckFrequency:  "Off",
		DetectedChanges: 0,
		CreatedBy:       createdBy,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Save to database
	if err := h.repo.Create(ctx, page); err != nil {
		return nil, err
	}

	// Create default monitoring config
	// Note: This should ideally be done in a transaction or a separate service
	// For now we'll just return the page with default values

	// Return response with all fields
	return &CreatePageResponse{
		ID:                   page.ID,
		WorkspaceID:          page.WorkspaceID,
		Name:                 page.Name,
		URL:                  page.URL,
		ThumbnailURL:         page.ThumbnailURL,
		LastCheckedAt:        page.LastCheckedAt,
		LastChangeDetectedAt: page.LastChangeDetectedAt,
		CheckCount:           page.CheckCount,
		Tags:                 page.Tags,
		CheckFrequency:       page.CheckFrequency,
		DetectedChanges:      page.DetectedChanges,
		CreatedBy:            page.CreatedBy,
		CreatedAt:            page.CreatedAt,
		UpdatedAt:            page.UpdatedAt,
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
