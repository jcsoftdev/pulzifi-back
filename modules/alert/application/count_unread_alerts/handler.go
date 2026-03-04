package countunreadalerts

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jcsoftdev/pulzifi-back/modules/alert/domain/repositories"
)

type CountUnreadAlertsHandler struct {
	repo repositories.AlertRepository
}

func NewCountUnreadAlertsHandler(repo repositories.AlertRepository) *CountUnreadAlertsHandler {
	return &CountUnreadAlertsHandler{repo: repo}
}

func (h *CountUnreadAlertsHandler) Handle(ctx context.Context) (*CountUnreadAlertsResponse, error) {
	count, err := h.repo.CountUnread(ctx)
	if err != nil {
		return nil, err
	}
	return &CountUnreadAlertsResponse{
		HasNotifications:  count > 0,
		NotificationCount: count,
	}, nil
}

func (h *CountUnreadAlertsHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	resp, err := h.Handle(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to count unread alerts"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
