package listpages

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/page/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type ListPagesHandler struct {
	repo repositories.PageRepository
}

func NewListPagesHandler(repo repositories.PageRepository) *ListPagesHandler {
	return &ListPagesHandler{repo: repo}
}

func (h *ListPagesHandler) Handle(ctx context.Context, workspaceID uuid.UUID) (*ListPagesResponse, error) {
	pages, err := h.repo.ListByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	pageResponses := make([]PageResponse, 0, len(pages))
	for _, p := range pages {
		pageResponses = append(pageResponses, ToPageResponse(p))
	}

	return &ListPagesResponse{Pages: pageResponses}, nil
}

// HTTP Handler wrapper
func (h *ListPagesHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	// Get workspace_id from query params
	workspaceIDStr := r.URL.Query().Get("workspace_id")
	if workspaceIDStr == "" {
		logger.ErrorWithContext(r.Context(), "workspace_id query parameter is required")
		http.Error(w, "workspace_id query parameter is required", http.StatusBadRequest)
		return
	}

	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		logger.ErrorWithContext(r.Context(), "Invalid workspace_id format",
			zap.String("workspace_id", workspaceIDStr),
			zap.Error(err),
		)
		http.Error(w, "invalid workspace_id", http.StatusBadRequest)
		return
	}

	// Execute handler
	response, err := h.Handle(r.Context(), workspaceID)
	if err != nil {
		logger.ErrorWithContext(r.Context(), "Failed to list pages",
			zap.String("workspace_id", workspaceID.String()),
			zap.Error(err),
		)
		http.Error(w, "failed to list pages\n", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
