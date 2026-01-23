package updatepage

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/page/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type UpdatePageHandler struct {
	repo repositories.PageRepository
}

func NewUpdatePageHandler(repo repositories.PageRepository) *UpdatePageHandler {
	return &UpdatePageHandler{repo: repo}
}

func (h *UpdatePageHandler) Handle(ctx context.Context, id uuid.UUID, req *UpdatePageRequest) (*UpdatePageResponse, error) {
	page, err := h.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if page == nil {
		return nil, nil // Or specific error
	}

	if req.Name != "" {
		page.Name = req.Name
	}
	if req.URL != "" {
		page.URL = req.URL
	}
	if req.Tags != nil {
		page.Tags = req.Tags
	}

	if err := h.repo.Update(ctx, page); err != nil {
		return nil, err
	}

	return &UpdatePageResponse{
		ID:   page.ID,
		Name: page.Name,
		URL:  page.URL,
	}, nil
}

func (h *UpdatePageHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid page ID", http.StatusBadRequest)
		return
	}

	var req UpdatePageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.Handle(r.Context(), id, &req)
	if err != nil {
		logger.Error("Failed to update page", zap.Error(err))
		http.Error(w, "Failed to update page", http.StatusInternalServerError)
		return
	}
	if resp == nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
