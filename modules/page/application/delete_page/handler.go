package deletepage

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/page/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type DeletePageHandler struct {
	repo repositories.PageRepository
}

func NewDeletePageHandler(repo repositories.PageRepository) *DeletePageHandler {
	return &DeletePageHandler{repo: repo}
}

func (h *DeletePageHandler) Handle(ctx context.Context, id uuid.UUID) error {
	return h.repo.Delete(ctx, id)
}

func (h *DeletePageHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid page ID", http.StatusBadRequest)
		return
	}

	if err := h.Handle(r.Context(), id); err != nil {
		logger.Error("Failed to delete page", zap.Error(err))
		http.Error(w, "Failed to delete page", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
