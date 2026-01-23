package getpage

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

type GetPageHandler struct {
	repo repositories.PageRepository
}

func NewGetPageHandler(repo repositories.PageRepository) *GetPageHandler {
	return &GetPageHandler{repo: repo}
}

func (h *GetPageHandler) Handle(ctx context.Context, id uuid.UUID) (*GetPageResponse, error) {
	page, err := h.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if page == nil {
		return nil, nil
	}

	return &GetPageResponse{
		ID:                   page.ID,
		WorkspaceID:          page.WorkspaceID,
		Name:                 page.Name,
		URL:                  page.URL,
		ThumbnailURL:         page.ThumbnailURL,
		LastCheckedAt:        page.LastCheckedAt,
		LastChangeDetectedAt: page.LastChangeDetectedAt,
		CheckCount:           page.CheckCount,
		CheckFrequency:       page.CheckFrequency,
		DetectedChanges:      page.DetectedChanges,
		Tags:                 page.Tags,
		CreatedAt:            page.CreatedAt,
		UpdatedAt:            page.UpdatedAt,
	}, nil
}

func (h *GetPageHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid page ID", http.StatusBadRequest)
		return
	}

	resp, err := h.Handle(r.Context(), id)
	if err != nil {
		logger.Error("Failed to get page", zap.Error(err))
		http.Error(w, "Failed to get page", http.StatusInternalServerError)
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
