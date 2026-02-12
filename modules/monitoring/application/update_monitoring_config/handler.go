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
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/scheduler"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type UpdateMonitoringConfigHandler struct {
	repo      repositories.MonitoringConfigRepository
	eventBus  *eventbus.EventBus
	tenant    string
	scheduler *scheduler.Scheduler
}

func NewUpdateMonitoringConfigHandler(repo repositories.MonitoringConfigRepository, eventBus *eventbus.EventBus, tenant string, scheduler *scheduler.Scheduler) *UpdateMonitoringConfigHandler {
	return &UpdateMonitoringConfigHandler{
		repo:      repo,
		eventBus:  eventBus,
		tenant:    tenant,
		scheduler: scheduler,
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
			if h.scheduler != nil {
				h.scheduler.WakeUp()
			}
		}
	} else {
		// Config exists, update only provided fields
		logger.Info("UpdateMonitoringConfigHandler: Config found, updating", zap.Any("current_config", config))
		if req.CheckFrequency != nil {
			logger.Info("UpdateMonitoringConfigHandler: Updating CheckFrequency", zap.String("new_frequency", *req.CheckFrequency))
			config.CheckFrequency = *req.CheckFrequency
			if config.CheckFrequency != "Off" {
				shouldDispatch = true
				if h.scheduler != nil {
					h.scheduler.WakeUp()
				}
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
	// REFACTOR: The API should NOT dispatch execution directly. The Scheduler is responsible for this.
	// We just log that the config was updated.
	if shouldDispatch {
		logger.Info("UpdateMonitoringConfigHandler: Config updated, Scheduler will pick up next run", zap.String("page_id", pageID.String()))
		// We can optionally trigger a scheduler "poke" if we want immediate feedback, but the requirement says "API does not dispatch executions".
		// So we do nothing here regarding dispatch.
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
