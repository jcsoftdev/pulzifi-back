package bulkdeletepages

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/page/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type BulkDeletePagesRequest struct {
	IDs []string `json:"ids"`
}

type BulkDeletePagesHandler struct {
	repo repositories.PageRepository
}

func NewBulkDeletePagesHandler(repo repositories.PageRepository) *BulkDeletePagesHandler {
	return &BulkDeletePagesHandler{repo: repo}
}

func (h *BulkDeletePagesHandler) Handle(ctx context.Context, ids []uuid.UUID) error {
	return h.repo.BulkDelete(ctx, ids)
}

func (h *BulkDeletePagesHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	var req BulkDeletePagesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.IDs) == 0 {
		http.Error(w, "ids must not be empty", http.StatusBadRequest)
		return
	}

	ids := make([]uuid.UUID, 0, len(req.IDs))
	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "Invalid page ID: "+idStr, http.StatusBadRequest)
			return
		}
		ids = append(ids, id)
	}

	if err := h.Handle(r.Context(), ids); err != nil {
		logger.Error("Failed to bulk delete pages", zap.Error(err))
		http.Error(w, "Failed to delete pages", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
