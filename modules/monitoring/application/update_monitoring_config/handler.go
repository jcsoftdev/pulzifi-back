package updatemonitoringconfig

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
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
		checkFrequency := "Off"
		scheduleType := "all_time"
		timezone := "UTC"
		blockAdsCookies := true

		// Override defaults with provided values
		if req.CheckFrequency != nil {
			checkFrequency = normalizeCheckFrequency(*req.CheckFrequency)
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

		enabledInsightTypes := []string{"marketing", "market_analysis"}
		if len(req.EnabledInsightTypes) > 0 {
			enabledInsightTypes = req.EnabledInsightTypes
		}
		enabledAlertConditions := []string{"any_changes"}
		if len(req.EnabledAlertConditions) > 0 {
			enabledAlertConditions = req.EnabledAlertConditions
		}
		customAlertCondition := ""
		if req.CustomAlertCondition != nil {
			customAlertCondition = *req.CustomAlertCondition
		}

		// Create new config using the constructor from entities
		config = &entities.MonitoringConfig{
			ID:                    uuid.New(),
			PageID:                pageID,
			CheckFrequency:        checkFrequency,
			ScheduleType:          scheduleType,
			Timezone:              timezone,
			BlockAdsCookies:       blockAdsCookies,
			EnabledInsightTypes:   enabledInsightTypes,
			EnabledAlertConditions: enabledAlertConditions,
			CustomAlertCondition:  customAlertCondition,
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
		}

		// Create in database
		if err := h.repo.Create(ctx, config); err != nil {
			return nil, err
		}

		if config.CheckFrequency != "Off" {
			shouldDispatch = true
			if h.scheduler != nil {
				if err := h.scheduler.TriggerPageCheck(ctx, h.tenant, pageID); err != nil {
					logger.Error("UpdateMonitoringConfigHandler: Failed to trigger immediate check", zap.String("page_id", pageID.String()), zap.Error(err))
				}
				// Do NOT call WakeUp here — TriggerPageCheck already dispatches the check.
				// Calling WakeUp causes the scheduler to also pick up the page (race with async UpdateLastChecked),
				// resulting in duplicate executions.
			} else {
				if err := h.repo.MarkPageDueNow(ctx, pageID); err != nil {
					logger.Error("UpdateMonitoringConfigHandler: Failed to mark page due now", zap.String("page_id", pageID.String()), zap.Error(err))
				}
			}
		}
	} else {
		// Config exists, update only provided fields
		logger.Info("UpdateMonitoringConfigHandler: Config found, updating", zap.Any("current_config", config))
		if req.CheckFrequency != nil {
			normalizedFrequency := normalizeCheckFrequency(*req.CheckFrequency)
			logger.Info("UpdateMonitoringConfigHandler: Updating CheckFrequency", zap.String("new_frequency", normalizedFrequency))
			config.CheckFrequency = normalizedFrequency
			if config.CheckFrequency != "Off" {
				shouldDispatch = true
				if h.scheduler != nil {
					if err := h.scheduler.TriggerPageCheck(ctx, h.tenant, pageID); err != nil {
						logger.Error("UpdateMonitoringConfigHandler: Failed to trigger immediate check", zap.String("page_id", pageID.String()), zap.Error(err))
					}
					// Do NOT call WakeUp here — TriggerPageCheck already dispatches the check.
				} else {
					if err := h.repo.MarkPageDueNow(ctx, pageID); err != nil {
						logger.Error("UpdateMonitoringConfigHandler: Failed to mark page due now", zap.String("page_id", pageID.String()), zap.Error(err))
					}
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

		if len(req.EnabledInsightTypes) > 0 {
			config.EnabledInsightTypes = req.EnabledInsightTypes
		}
		if len(req.EnabledAlertConditions) > 0 {
			config.EnabledAlertConditions = req.EnabledAlertConditions
		}
		if req.CustomAlertCondition != nil {
			config.CustomAlertCondition = *req.CustomAlertCondition
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
		ID:                    config.ID,
		PageID:                config.PageID,
		CheckFrequency:        config.CheckFrequency,
		ScheduleType:          config.ScheduleType,
		Timezone:              config.Timezone,
		BlockAdsCookies:       config.BlockAdsCookies,
		EnabledInsightTypes:   config.EnabledInsightTypes,
		EnabledAlertConditions: config.EnabledAlertConditions,
		CustomAlertCondition:  config.CustomAlertCondition,
		UpdatedAt:             config.UpdatedAt,
	}, nil
}

func normalizeCheckFrequency(input string) string {
	normalized := strings.ToLower(strings.TrimSpace(input))
	normalized = strings.ReplaceAll(normalized, "_", " ")
	normalized = strings.ReplaceAll(normalized, "-", " ")
	normalized = strings.Join(strings.Fields(normalized), " ")

	switch normalized {
	case "off", "disabled", "none":
		return "Off"
	case "5m", "5 min", "5 mins", "every 5 minutes", "every 5m":
		return "5m"
	case "10m", "10 min", "10 mins", "every 10 minutes", "every 10m":
		return "10m"
	case "15m", "15 min", "15 mins", "every 15 minutes", "every 15m":
		return "15m"
	case "30m", "30 min", "30 mins", "every 30 minutes", "every 30m":
		return "30m"
	case "1h", "1 hr", "1 hour", "every hour", "every 1 hour", "every 1hr":
		return "1h"
	case "2h", "2 hr", "2 hours", "every 2 hours", "every 2hr":
		return "2h"
	case "4h", "4 hr", "4 hours", "every 4 hours", "every 4hr":
		return "4h"
	case "6h", "6 hr", "6 hours", "every 6 hours", "every 6hr":
		return "6h"
	case "12h", "12 hr", "12 hours", "every 12 hours", "every 12hr":
		return "12h"
	case "24h", "24 hr", "1d", "1 day", "daily", "every day":
		return "24h"
	case "168h", "7d", "7 days", "weekly", "every week":
		return "168h"
	default:
		return input
	}
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
