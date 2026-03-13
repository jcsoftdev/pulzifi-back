package updatemonitoringconfig

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/orchestrator"
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
	quotaExceeded := false

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

		selectorType := "full_page"
		if req.SelectorType != nil {
			selectorType = *req.SelectorType
		}
		cssSelector := ""
		if req.CSSSelector != nil {
			cssSelector = *req.CSSSelector
		}
		xpathSelector := ""
		if req.XPathSelector != nil {
			xpathSelector = *req.XPathSelector
		}
		var selectorOffsets *entities.SelectorOffsets
		if req.SelectorOffsets != nil {
			selectorOffsets = &entities.SelectorOffsets{
				Top:    req.SelectorOffsets.Top,
				Right:  req.SelectorOffsets.Right,
				Bottom: req.SelectorOffsets.Bottom,
				Left:   req.SelectorOffsets.Left,
			}
		}

		// Create new config using the constructor from entities
		config = &entities.MonitoringConfig{
			ID:                     uuid.New(),
			PageID:                 pageID,
			CheckFrequency:         checkFrequency,
			ScheduleType:           scheduleType,
			Timezone:               timezone,
			BlockAdsCookies:        blockAdsCookies,
			EnabledInsightTypes:    enabledInsightTypes,
			EnabledAlertConditions: enabledAlertConditions,
			CustomAlertCondition:   customAlertCondition,
			SelectorType:           selectorType,
			CSSSelector:            cssSelector,
			XPathSelector:          xpathSelector,
			SelectorOffsets:        selectorOffsets,
			CreatedAt:              time.Now(),
			UpdatedAt:              time.Now(),
		}

		// Create in database — the scheduler will pick up the page on its
		// next tick (last_checked_at is NULL, so it is immediately "due").
		if err := h.repo.Create(ctx, config); err != nil {
			return nil, err
		}

		if config.CheckFrequency != "Off" {
			// Wake the scheduler so it re-evaluates next run time and picks
			// up this newly created config promptly.
			if h.scheduler != nil {
				h.scheduler.WakeUp()
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
				// Only dispatch immediately if the page is overdue under the new frequency.
				// If the user increased the interval (e.g. 5m -> 1h) and the last check
				// was recent, skip the immediate check and let the scheduler handle it.
				shouldDispatch = h.isPageDueForCheck(ctx, pageID, normalizedFrequency)
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

		if req.SelectorType != nil {
			config.SelectorType = *req.SelectorType
		}
		if req.CSSSelector != nil {
			config.CSSSelector = *req.CSSSelector
		}
		if req.XPathSelector != nil {
			config.XPathSelector = *req.XPathSelector
		}
		if req.SelectorOffsets != nil {
			config.SelectorOffsets = &entities.SelectorOffsets{
				Top:    req.SelectorOffsets.Top,
				Right:  req.SelectorOffsets.Right,
				Bottom: req.SelectorOffsets.Bottom,
				Left:   req.SelectorOffsets.Left,
			}
		}

		config.UpdatedAt = time.Now()

		// Pre-claim: set last_checked_at = NOW() BEFORE saving the config so the
		// scheduler cannot see the page as "due" during the gap between config
		// save and TriggerPageCheck dispatch.
		if shouldDispatch {
			if err := h.repo.UpdateLastCheckedAt(ctx, pageID); err != nil {
				logger.Error("UpdateMonitoringConfigHandler: Failed to pre-claim page", zap.String("page_id", pageID.String()), zap.Error(err))
			}
		}

		// Save the config. TriggerPageCheck queries check_frequency != 'Off',
		// so the config must be persisted before dispatch.
		if err := h.repo.Update(ctx, config); err != nil {
			return nil, err
		}

		// Dispatch the immediate check. The page is already claimed (last_checked_at = NOW)
		// so the scheduler's GetDueSnapshotTasks will skip it.
		if shouldDispatch {
			if h.scheduler != nil {
				if err := h.scheduler.TriggerPageCheck(ctx, h.tenant, pageID); err != nil {
					if errors.Is(err, orchestrator.ErrQuotaExceeded) {
						quotaExceeded = true
						logger.Warn("UpdateMonitoringConfigHandler: Quota exceeded", zap.String("page_id", pageID.String()))
					} else {
						logger.Error("UpdateMonitoringConfigHandler: Failed to trigger immediate check", zap.String("page_id", pageID.String()), zap.Error(err))
					}
				}
			} else {
				if err := h.repo.MarkPageDueNow(ctx, pageID); err != nil {
					logger.Error("UpdateMonitoringConfigHandler: Failed to mark page due now", zap.String("page_id", pageID.String()), zap.Error(err))
				}
			}
		}
	}

	// Build selector offsets DTO
	var selectorOffsetsDTO *SelectorOffsetsDTO
	if config.SelectorOffsets != nil {
		selectorOffsetsDTO = &SelectorOffsetsDTO{
			Top:    config.SelectorOffsets.Top,
			Right:  config.SelectorOffsets.Right,
			Bottom: config.SelectorOffsets.Bottom,
			Left:   config.SelectorOffsets.Left,
		}
	}

	// Return response
	return &UpdateMonitoringConfigResponse{
		ID:                     config.ID,
		PageID:                 config.PageID,
		CheckFrequency:         config.CheckFrequency,
		ScheduleType:           config.ScheduleType,
		Timezone:               config.Timezone,
		BlockAdsCookies:        config.BlockAdsCookies,
		EnabledInsightTypes:    config.EnabledInsightTypes,
		EnabledAlertConditions: config.EnabledAlertConditions,
		CustomAlertCondition:   config.CustomAlertCondition,
		SelectorType:           config.SelectorType,
		CSSSelector:            config.CSSSelector,
		XPathSelector:          config.XPathSelector,
		SelectorOffsets:        selectorOffsetsDTO,
		UpdatedAt:              config.UpdatedAt,
		QuotaExceeded:          quotaExceeded,
	}, nil
}

// isPageDueForCheck returns true only if the page has been checked before AND
// is overdue under the given frequency. Never-checked pages are left for the
// scheduler to pick up at the natural interval.
func (h *UpdateMonitoringConfigHandler) isPageDueForCheck(ctx context.Context, pageID uuid.UUID, frequency string) bool {
	interval, ok := entities.ResolveFrequency(frequency)
	if !ok {
		return false
	}

	lastCheckedAt, err := h.repo.GetLastCheckedAt(ctx, pageID)
	if err != nil || lastCheckedAt == nil {
		// Never checked — let the scheduler handle the first check
		return false
	}

	elapsed := time.Since(*lastCheckedAt)
	isDue := elapsed >= interval
	if !isDue {
		logger.Info("UpdateMonitoringConfigHandler: Page not overdue under new frequency, skipping immediate check",
			zap.String("page_id", pageID.String()),
			zap.String("frequency", frequency),
			zap.Duration("elapsed", elapsed),
			zap.Duration("interval", interval))
	}
	return isDue
}

// NormalizeCheckFrequency converts various frequency input formats to a canonical form.
func NormalizeCheckFrequency(input string) string {
	return normalizeCheckFrequency(input)
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
