package create_monitoring_config

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/repositories"
)

type CreateMonitoringConfigHandler struct {
	repo repositories.MonitoringConfigRepository
}

func NewCreateMonitoringConfigHandler(repo repositories.MonitoringConfigRepository) *CreateMonitoringConfigHandler {
	return &CreateMonitoringConfigHandler{repo: repo}
}

func (h *CreateMonitoringConfigHandler) Handle(ctx context.Context, req *CreateMonitoringConfigRequest) (*CreateMonitoringConfigResponse, error) {
	config := entities.NewMonitoringConfig(req.PageID, req.CheckFrequency, req.ScheduleType, req.Timezone)
	config.BlockAdsCookies = req.BlockAdsCookies

	if err := h.repo.Create(ctx, config); err != nil {
		return nil, err
	}

	return &CreateMonitoringConfigResponse{
		ID:              config.ID,
		PageID:          config.PageID,
		CheckFrequency:  config.CheckFrequency,
		ScheduleType:    config.ScheduleType,
		Timezone:        config.Timezone,
		BlockAdsCookies: config.BlockAdsCookies,
		CreatedAt:       config.CreatedAt,
	}, nil
}

// HTTP Handler wrapper
func (h *CreateMonitoringConfigHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	var req CreateMonitoringConfigRequest

	// Parse JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.PageID == [16]byte{} || req.CheckFrequency == "" {
		http.Error(w, "page_id and check_frequency are required", http.StatusBadRequest)
		return
	}

	// Execute use case
	resp, err := h.Handle(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
