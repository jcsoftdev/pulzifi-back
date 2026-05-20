package markallalerts

import (
	"context"

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
