package listinsights

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/repositories"
)

type ListInsightsHandler struct {
	repo repositories.InsightRepository
}

func NewListInsightsHandler(repo repositories.InsightRepository) *ListInsightsHandler {
	return &ListInsightsHandler{repo: repo}
}

func (h *ListInsightsHandler) Handle(ctx context.Context, pageID uuid.UUID) (*ListInsightsResponse, error) {
	insights, err := h.repo.ListByPageID(ctx, pageID)
	if err != nil {
		return nil, err
	}
	return buildResponse(insights), nil
}

func (h *ListInsightsHandler) HandleByCheckID(ctx context.Context, checkID uuid.UUID) (*ListInsightsResponse, error) {
	insights, err := h.repo.ListByCheckID(ctx, checkID)
	if err != nil {
		return nil, err
	}
	return buildResponse(insights), nil
}

func buildResponse(insights []*entities.Insight) *ListInsightsResponse {
	response := &ListInsightsResponse{
		Insights: make([]*InsightResponse, len(insights)),
	}
	for i, insight := range insights {
		var metadata interface{}
		if len(insight.Metadata) > 0 {
			_ = json.Unmarshal(insight.Metadata, &metadata)
		}
		response.Insights[i] = &InsightResponse{
			ID:          insight.ID,
			PageID:      insight.PageID,
			CheckID:     insight.CheckID,
			InsightType: insight.InsightType,
			Title:       insight.Title,
			Content:     insight.Content,
			Metadata:    metadata,
			CreatedAt:   insight.CreatedAt,
		}
	}
	return response
}
