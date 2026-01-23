package listinsights

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
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

	return response, nil
}

func (h *ListInsightsHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	// Assuming insights are scoped to a page.
	// The route might be /insights?page_id=... or /pages/{pageId}/insights
	// Based on the module.go in insight, it was just /insights.
	// But usually we list by page.
	// Let's check query param "page_id".

	pageIDStr := r.URL.Query().Get("page_id")
	if pageIDStr == "" {
		// If no page_id provided, maybe return empty or error?
		// Or maybe implement ListAll? For now let's require page_id as this is for page details.
		http.Error(w, "page_id query parameter is required", http.StatusBadRequest)
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
