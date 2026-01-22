package updatemonitoringconfig

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type UpdateMonitoringConfigHandler struct {
	repo repositories.MonitoringConfigRepository
}

func NewUpdateMonitoringConfigHandler(repo repositories.MonitoringConfigRepository) *UpdateMonitoringConfigHandler {
	return &UpdateMonitoringConfigHandler{repo: repo}
}

func (h *UpdateMonitoringConfigHandler) Handle(ctx context.Context, pageID uuid.UUID, req *UpdateMonitoringConfigRequest) (*UpdateMonitoringConfigResponse, error) {
	// Get existing config
	config, err := h.repo.GetByPageID(ctx, pageID)
	if err != nil {
		return nil, err
	}

	// If config doesn't exist, create a new one with default values
	if config == nil {
		// Set defaults for new config
		checkFrequency := "30m"
		scheduleType := "interval"
		timezone := "UTC"
		blockAdsCookies := true

		// Override defaults with provided values
		if req.CheckFrequency != nil {
			checkFrequency = *req.CheckFrequency
		}
		if req.ScheduleType != nil {
			scheduleType = *req.ScheduleType
		}
		if req.Timezone != nil {
			timezone = *req.Timezone
		}
		if req.BlockAdsCookies != nil {
			blockAdsCookies = *req.BlockAdsCookies
		}

		// Create new config using the constructor from entities
		config = &entities.MonitoringConfig{
			ID:              uuid.New(),
			PageID:          pageID,
			CheckFrequency:  checkFrequency,
			ScheduleType:    scheduleType,
			Timezone:        timezone,
			BlockAdsCookies: blockAdsCookies,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		// Create in database
		if err := h.repo.Create(ctx, config); err != nil {
			return nil, err
		}
	} else {
		// Config exists, update only provided fields
		if req.CheckFrequency != nil {
			config.CheckFrequency = *req.CheckFrequency
		}

		if req.ScheduleType != nil {
			config.ScheduleType = *req.ScheduleType
		}

		if req.Timezone != nil {
			config.Timezone = *req.Timezone
		}

		if req.BlockAdsCookies != nil {
			config.BlockAdsCookies = *req.BlockAdsCookies
		}

		config.UpdatedAt = time.Now()

		// Save changes
		if err := h.repo.Update(ctx, config); err != nil {
			return nil, err
		}
	}

	// Return response
	return &UpdateMonitoringConfigResponse{
		ID:              config.ID,
		PageID:          config.PageID,
		CheckFrequency:  config.CheckFrequency,
		ScheduleType:    config.ScheduleType,
		Timezone:        config.Timezone,
		BlockAdsCookies: config.BlockAdsCookies,
		UpdatedAt:       config.UpdatedAt,
	}, nil
}

// HTTP Handler wrapper
func (h *UpdateMonitoringConfigHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	// Parse JSON body
	var req UpdateMonitoringConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get page ID from URL
	pageIDStr := chi.URLParam(r, "pageId")
	pageID, err := uuid.Parse(pageIDStr)
	if err != nil {
		logger.Error("Invalid page ID", zap.Error(err))
		http.Error(w, "invalid page_id", http.StatusBadRequest)
		return
	}

	// Execute handler
	response, err := h.Handle(r.Context(), pageID, &req)
	if err != nil {
		logger.Error("Failed to update monitoring config", zap.Error(err))
		http.Error(w, "failed to update monitoring config", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
