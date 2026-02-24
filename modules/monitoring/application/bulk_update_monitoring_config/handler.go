package bulkupdatemonitoringconfig

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	updatemonitoringconfig "github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/update_monitoring_config"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type BulkUpdateMonitoringConfigRequest struct {
	PageIDs        []string `json:"page_ids"`
	CheckFrequency string   `json:"check_frequency"`
}

type BulkUpdateMonitoringConfigHandler struct {
	repo repositories.MonitoringConfigRepository
}

func NewBulkUpdateMonitoringConfigHandler(repo repositories.MonitoringConfigRepository) *BulkUpdateMonitoringConfigHandler {
	return &BulkUpdateMonitoringConfigHandler{repo: repo}
}

func (h *BulkUpdateMonitoringConfigHandler) Handle(ctx context.Context, pageIDs []uuid.UUID, frequency string) error {
	normalized := updatemonitoringconfig.NormalizeCheckFrequency(frequency)
	return h.repo.BulkUpdateFrequency(ctx, pageIDs, normalized)
}

func (h *BulkUpdateMonitoringConfigHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	var req BulkUpdateMonitoringConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.PageIDs) == 0 {
		http.Error(w, "page_ids must not be empty", http.StatusBadRequest)
		return
	}
	if req.CheckFrequency == "" {
		http.Error(w, "check_frequency is required", http.StatusBadRequest)
		return
	}

	pageIDs := make([]uuid.UUID, 0, len(req.PageIDs))
	for _, idStr := range req.PageIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "Invalid page ID: "+idStr, http.StatusBadRequest)
			return
		}
		pageIDs = append(pageIDs, id)
	}

	if err := h.Handle(r.Context(), pageIDs, req.CheckFrequency); err != nil {
		logger.Error("Failed to bulk update monitoring config", zap.Error(err))
		http.Error(w, "Failed to update check frequency", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
