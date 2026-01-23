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
	"github.com/jcsoftdev/pulzifi-back/shared/kafka"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type UpdateMonitoringConfigHandler struct {
	repo     repositories.MonitoringConfigRepository
	producer *kafka.ProducerClient
	tenant   string
}

func NewUpdateMonitoringConfigHandler(repo repositories.MonitoringConfigRepository, producer *kafka.ProducerClient, tenant string) *UpdateMonitoringConfigHandler {
	return &UpdateMonitoringConfigHandler{
		repo:     repo,
		producer: producer,
		tenant:   tenant,
	}
}

func (h *UpdateMonitoringConfigHandler) Handle(ctx context.Context, pageID uuid.UUID, req *UpdateMonitoringConfigRequest) (*UpdateMonitoringConfigResponse, error) {
	logger.Info("UpdateMonitoringConfigHandler: Start processing", zap.String("page_id", pageID.String()))

	// Get existing config
	config, err := h.repo.GetByPageID(ctx, pageID)
	if err != nil {
		return nil, err
	}

	shouldDispatch := false

	// If config doesn't exist, create a new one with default values
	if config == nil {
		logger.Info("UpdateMonitoringConfigHandler: Config not found, creating new one")
		// Set defaults for new config
		checkFrequency := "Every 1 hour"
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

		if config.CheckFrequency != "Off" {
			shouldDispatch = true
		}
	} else {
		// Config exists, update only provided fields
		logger.Info("UpdateMonitoringConfigHandler: Config found, updating", zap.Any("current_config", config))
		if req.CheckFrequency != nil {
			logger.Info("UpdateMonitoringConfigHandler: Updating CheckFrequency", zap.String("new_frequency", *req.CheckFrequency))
			config.CheckFrequency = *req.CheckFrequency
			if config.CheckFrequency != "Off" {
				shouldDispatch = true
			}
		} else {
			logger.Info("UpdateMonitoringConfigHandler: CheckFrequency not provided in request")
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

	// Dispatch snapshot request if needed
	logger.Info("UpdateMonitoringConfigHandler: Dispatch check", zap.Bool("shouldDispatch", shouldDispatch), zap.Bool("producer_exists", h.producer != nil))
	if shouldDispatch && h.producer != nil {
		pageURL, err := h.repo.GetPageURL(ctx, pageID)
		logger.Info("UpdateMonitoringConfigHandler: Page URL retrieval", zap.String("url", pageURL), zap.Error(err))
		if err == nil && pageURL != "" {
			payload := map[string]interface{}{
				"page_id":     pageID.String(),
				"url":         pageURL,
				"schema_name": h.tenant,
			}
			bytes, _ := json.Marshal(payload)
			err := h.producer.Produce("snapshot-requests", pageID.String(), bytes)
			if err != nil {
				logger.Error("Failed to produce snapshot request", zap.String("page_id", pageID.String()), zap.Error(err))
			} else {
				logger.Info("Dispatched snapshot request due to config update", zap.String("page_id", pageID.String()))
				// Update last checked at
				if err := h.repo.UpdateLastCheckedAt(ctx, pageID); err != nil {
					logger.Error("Failed to update last_checked_at", zap.String("page_id", pageID.String()), zap.Error(err))
				}
			}
		} else if err != nil {
			logger.Error("Failed to get page URL for snapshot dispatch", zap.Error(err))
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
