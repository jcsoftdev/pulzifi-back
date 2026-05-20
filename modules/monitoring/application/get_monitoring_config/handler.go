package get_monitoring_config

import (
	"context"

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

	var selectorOffsetsDTO *SelectorOffsetsDTO
	if config.SelectorOffsets != nil {
		selectorOffsetsDTO = &SelectorOffsetsDTO{
			Top:    config.SelectorOffsets.Top,
			Right:  config.SelectorOffsets.Right,
			Bottom: config.SelectorOffsets.Bottom,
			Left:   config.SelectorOffsets.Left,
		}
	}

	return &GetMonitoringConfigResponse{
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
		IgnoreSelectors:        config.IgnoreSelectors,
		CreatedAt:              config.CreatedAt,
		UpdatedAt:              config.UpdatedAt,
	}, nil
}
