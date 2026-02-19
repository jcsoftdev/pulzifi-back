package listinsights

import (
	"context"
	"encoding/json"
	"net/http"

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

func (h *ListInsightsHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	// Support filtering by check_id (preferred when viewing a specific check)
	// or by page_id (to list all insights for a page)
	checkIDStr := r.URL.Query().Get("check_id")
	if checkIDStr != "" {
		checkID, err := uuid.Parse(checkIDStr)
		if err != nil {
			http.Error(w, "Invalid check ID", http.StatusBadRequest)
			return
		}
		resp, err := h.HandleByCheckID(r.Context(), checkID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
		return
	}

	pageIDStr := r.URL.Query().Get("page_id")
	if pageIDStr == "" {
		http.Error(w, "page_id or check_id query parameter is required", http.StatusBadRequest)
		return
	}

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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
