package countunreadalerts

import (
	"context"

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
