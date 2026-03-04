package markallalerts

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jcsoftdev/pulzifi-back/modules/alert/domain/repositories"
)

type MarkAllAlertsReadHandler struct {
	repo repositories.AlertRepository
}

func NewMarkAllAlertsReadHandler(repo repositories.AlertRepository) *MarkAllAlertsReadHandler {
	return &MarkAllAlertsReadHandler{repo: repo}
}

func (h *MarkAllAlertsReadHandler) Handle(ctx context.Context) error {
	return h.repo.MarkAllAsRead(ctx)
}

func (h *MarkAllAlertsReadHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.Handle(r.Context()); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to mark alerts as read"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "all alerts marked as read"})
}
