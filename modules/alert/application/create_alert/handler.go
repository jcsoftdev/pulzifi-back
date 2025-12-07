package createalert

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/alert/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/alert/domain/repositories"
)

type CreateAlertHandler struct {
	repo repositories.AlertRepository
}

func NewCreateAlertHandler(repo repositories.AlertRepository) *CreateAlertHandler {
	return &CreateAlertHandler{repo: repo}
}

func (h *CreateAlertHandler) Handle(ctx context.Context, req *CreateAlertRequest) (*CreateAlertResponse, error) {
	// Create alert entity
	alert := &entities.Alert{
		ID:          uuid.New(),
		WorkspaceID: req.WorkspaceID,
		PageID:      req.PageID,
		CheckID:     req.CheckID,
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
	}

	// Save to database
	if err := h.repo.Create(ctx, alert); err != nil {
		return nil, err
	}

	// Return response
	return &CreateAlertResponse{
		ID:          alert.ID,
		WorkspaceID: alert.WorkspaceID,
		PageID:      alert.PageID,
		CheckID:     alert.CheckID,
		Type:        alert.Type,
		Title:       alert.Title,
		Description: alert.Description,
		Metadata:    alert.Metadata,
		CreatedAt:   alert.CreatedAt,
	}, nil
}

// HTTP Handler wrapper
func (h *CreateAlertHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	var req CreateAlertRequest

	// Parse JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.WorkspaceID == uuid.Nil || req.PageID == uuid.Nil || req.CheckID == uuid.Nil {
		http.Error(w, "workspace_id, page_id, and check_id are required", http.StatusBadRequest)
		return
	}
	if req.Type == "" || req.Title == "" {
		http.Error(w, "type and title are required", http.StatusBadRequest)
		return
	}

	// Execute use case
	resp, err := h.Handle(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
