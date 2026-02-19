package getdashboardstats

import (
	"context"
	"encoding/json"
	"net/http"

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

func (h *GetDashboardStatsHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	resp, err := h.Handle(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
