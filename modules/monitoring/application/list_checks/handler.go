package listchecks

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type ListChecksHandler struct {
	repo repositories.CheckRepository
}

func NewListChecksHandler(repo repositories.CheckRepository) *ListChecksHandler {
	return &ListChecksHandler{repo: repo}
}

func (h *ListChecksHandler) Handle(ctx context.Context, pageID uuid.UUID) (*ListChecksResponse, error) {
	checks, err := h.repo.ListByPage(ctx, pageID)
	if err != nil {
		return nil, err
	}

	response := &ListChecksResponse{
		Checks: make([]*CheckResponse, len(checks)),
	}

	for i, check := range checks {
		response.Checks[i] = &CheckResponse{
			ID:             check.ID,
			PageID:         check.PageID,
			Status:         check.Status,
			ScreenshotURL:  check.ScreenshotURL,
			ChangeDetected: check.ChangeDetected,
			ErrorMessage:   check.ErrorMessage,
			CheckedAt:      check.CheckedAt,
		}
	}

	return response, nil
}

func (h *ListChecksHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	pageIDStr := chi.URLParam(r, "pageId")
	pageID, err := uuid.Parse(pageIDStr)
	if err != nil {
		http.Error(w, "Invalid page ID", http.StatusBadRequest)
		return
	}

	resp, err := h.Handle(r.Context(), pageID)
	if err != nil {
		logger.Error("Failed to list checks", zap.Error(err))
		http.Error(w, "Failed to list checks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
