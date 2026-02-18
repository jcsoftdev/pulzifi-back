package get_monitoring_config

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/repositories"
)

type GetMonitoringConfigHandler struct {
	repo repositories.MonitoringConfigRepository
}

func NewGetMonitoringConfigHandler(repo repositories.MonitoringConfigRepository) *GetMonitoringConfigHandler {
	return &GetMonitoringConfigHandler{repo: repo}
}

func (h *GetMonitoringConfigHandler) Handle(ctx context.Context, pageID uuid.UUID) (*GetMonitoringConfigResponse, error) {
	config, err := h.repo.GetByPageID(ctx, pageID)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, nil
	}

	return &GetMonitoringConfigResponse{
		ID:                    config.ID,
		PageID:                config.PageID,
		CheckFrequency:        config.CheckFrequency,
		ScheduleType:          config.ScheduleType,
		Timezone:              config.Timezone,
		BlockAdsCookies:       config.BlockAdsCookies,
		EnabledInsightTypes:   config.EnabledInsightTypes,
		EnabledAlertConditions: config.EnabledAlertConditions,
		CustomAlertCondition:  config.CustomAlertCondition,
		CreatedAt:             config.CreatedAt,
		UpdatedAt:             config.UpdatedAt,
	}, nil
}

func (h *GetMonitoringConfigHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	pageIDStr := chi.URLParam(r, "pageId")
	pageID, err := uuid.Parse(pageIDStr)
	if err != nil {
		http.Error(w, "Invalid page ID", http.StatusBadRequest)
		return
	}

	resp, err := h.Handle(r.Context(), pageID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if resp == nil {
		// Return 404 but maybe with a default config structure if client expects it?
		// For now 404 is correct if not found.
		http.Error(w, "Monitoring config not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
