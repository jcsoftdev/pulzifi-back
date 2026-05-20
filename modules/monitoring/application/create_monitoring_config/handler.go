package create_monitoring_config

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/repositories"
	monitoringservices "github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/services"
)

type CreateMonitoringConfigHandler struct {
	repo      repositories.MonitoringConfigRepository
	scheduler monitoringservices.Scheduler
}

func NewCreateMonitoringConfigHandler(repo repositories.MonitoringConfigRepository, scheduler monitoringservices.Scheduler) *CreateMonitoringConfigHandler {
	return &CreateMonitoringConfigHandler{
		repo:      repo,
		scheduler: scheduler,
	}
}

func (h *CreateMonitoringConfigHandler) Handle(ctx context.Context, req *CreateMonitoringConfigRequest) (*CreateMonitoringConfigResponse, error) {
	config := entities.NewMonitoringConfig(req.PageID, req.CheckFrequency, req.ScheduleType, req.Timezone)
	config.BlockAdsCookies = req.BlockAdsCookies

	if err := h.repo.Create(ctx, config); err != nil {
		return nil, err
	}

	if config.CheckFrequency != "Off" && h.scheduler != nil {
		h.scheduler.WakeUp()
	}

	return &CreateMonitoringConfigResponse{
		ID:                     config.ID,
		PageID:                 config.PageID,
		CheckFrequency:         config.CheckFrequency,
		ScheduleType:           config.ScheduleType,
		Timezone:               config.Timezone,
		BlockAdsCookies:        config.BlockAdsCookies,
		EnabledInsightTypes:    config.EnabledInsightTypes,
		EnabledAlertConditions: config.EnabledAlertConditions,
		CustomAlertCondition:   config.CustomAlertCondition,
		CreatedAt:              config.CreatedAt,
	}, nil
}
