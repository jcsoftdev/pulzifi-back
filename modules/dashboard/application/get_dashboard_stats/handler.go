package getdashboardstats

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/dashboard/domain/repositories"
)

type GetDashboardStatsHandler struct {
	repo repositories.DashboardRepository
}

func NewGetDashboardStatsHandler(repo repositories.DashboardRepository) *GetDashboardStatsHandler {
	return &GetDashboardStatsHandler{repo: repo}
}

func (h *GetDashboardStatsHandler) Handle(ctx context.Context) (*GetDashboardStatsResponse, error) {
	stats, err := h.repo.GetStats(ctx)
	if err != nil {
		return nil, err
	}
	return buildResponse(stats), nil
}
