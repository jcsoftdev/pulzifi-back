package createcheck

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/repositories"
)

type CreateCheckHandler struct {
	repo repositories.CheckRepository
}

func NewCreateCheckHandler(repo repositories.CheckRepository) *CreateCheckHandler {
	return &CreateCheckHandler{repo: repo}
}

func (h *CreateCheckHandler) Handle(ctx context.Context, req *CreateCheckRequest) (*CreateCheckResponse, error) {
	// Create check entity
	check := &entities.Check{
		PageID:          req.PageID,
		Status:          req.Status,
		ChangeDetected:  req.ChangeDetected,
		ChangeType:      req.ChangeType,
		ScreenshotURL:   req.ScreenshotURL,
		HTMLSnapshotURL: req.HTMLSnapshotURL,
		ErrorMessage:    req.ErrorMessage,
		DurationMs:      req.DurationMs,
	}
	check = entities.NewCheck(req.PageID, req.Status, req.ChangeDetected)
	check.ChangeType = req.ChangeType
	check.ScreenshotURL = req.ScreenshotURL
	check.HTMLSnapshotURL = req.HTMLSnapshotURL
	check.ErrorMessage = req.ErrorMessage
	check.DurationMs = req.DurationMs

	// Save to database
	if err := h.repo.Create(ctx, check); err != nil {
		return nil, err
	}

	// Return response
	return &CreateCheckResponse{
		ID:             check.ID,
		PageID:         check.PageID,
		Status:         check.Status,
		ChangeDetected: check.ChangeDetected,
		ChangeType:     check.ChangeType,
		CheckedAt:      check.CheckedAt,
	}, nil
}

// HTTP Handler wrapper
func (h *CreateCheckHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	var req CreateCheckRequest

	// Parse JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.PageID == [16]byte{} || req.Status == "" {
		http.Error(w, "page_id and status are required", http.StatusBadRequest)
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
