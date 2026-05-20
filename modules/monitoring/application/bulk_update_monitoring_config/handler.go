package bulkupdatemonitoringconfig

import (
	"context"

	"github.com/google/uuid"
	updatemonitoringconfig "github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/update_monitoring_config"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/repositories"
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
