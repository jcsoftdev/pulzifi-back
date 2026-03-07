package listchecks

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type ListChecksHandler struct {
	repo repositories.CheckRepository
}

func NewListChecksHandler(repo repositories.CheckRepository) *ListChecksHandler {
	return &ListChecksHandler{repo: repo}
}

func (h *ListChecksHandler) Handle(ctx context.Context, pageID uuid.UUID) (*ListChecksResponse, error) {
	// Fetch parent checks (section_id IS NULL).
	parentChecks, err := h.repo.ListByPage(ctx, pageID)
	if err != nil {
		return nil, err
	}

	// Fetch all section checks for the page and group by parent_check_id.
	sectionChecks, err := h.repo.ListSectionChecksByPage(ctx, pageID)
	if err != nil {
		return nil, err
	}

	sectionsByParent := make(map[uuid.UUID][]*entities.Check)
	for _, sc := range sectionChecks {
		if sc.ParentCheckID != nil {
			sectionsByParent[*sc.ParentCheckID] = append(sectionsByParent[*sc.ParentCheckID], sc)
		}
	}

	return buildResponseWithSections(parentChecks, sectionsByParent), nil
}

// HandleBySection returns checks filtered by section. sectionID nil means full-page checks only.
func (h *ListChecksHandler) HandleBySection(ctx context.Context, pageID uuid.UUID, sectionID *uuid.UUID) (*ListChecksResponse, error) {
	checks, err := h.repo.ListByPageAndSection(ctx, pageID, sectionID)
	if err != nil {
		return nil, err
	}

	return buildResponse(checks), nil
}

func toCheckResponse(check *entities.Check) *CheckResponse {
	return &CheckResponse{
		ID:              check.ID,
		PageID:          check.PageID,
		SectionID:       check.SectionID,
		ParentCheckID:   check.ParentCheckID,
		Status:          check.Status,
		ScreenshotURL:   check.ScreenshotURL,
		HTMLSnapshotURL: check.HTMLSnapshotURL,
		ChangeDetected:  check.ChangeDetected,
		ChangeType:      check.ChangeType,
		ErrorMessage:    check.ErrorMessage,
		CheckedAt:       check.CheckedAt,
	}
}

func buildResponse(checks []*entities.Check) *ListChecksResponse {
	response := &ListChecksResponse{
		Checks: make([]*CheckResponse, len(checks)),
	}
	for i, check := range checks {
		response.Checks[i] = toCheckResponse(check)
	}
	return response
}

func buildResponseWithSections(parentChecks []*entities.Check, sectionsByParent map[uuid.UUID][]*entities.Check) *ListChecksResponse {
	response := &ListChecksResponse{
		Checks: make([]*CheckResponse, len(parentChecks)),
	}
	for i, check := range parentChecks {
		cr := toCheckResponse(check)
		if sections, ok := sectionsByParent[check.ID]; ok {
			cr.Sections = make([]*CheckResponse, len(sections))
			for j, sc := range sections {
				cr.Sections[j] = toCheckResponse(sc)
			}
		}
		response.Checks[i] = cr
	}
	return response
}

func (h *ListChecksHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	pageIDStr := chi.URLParam(r, "pageId")
	pageID, err := uuid.Parse(pageIDStr)
	if err != nil {
		http.Error(w, "Invalid page ID", http.StatusBadRequest)
		return
	}

	var resp *ListChecksResponse

	// Optional section_id filter
	sectionIDStr := r.URL.Query().Get("section_id")
	if sectionIDStr != "" {
		sectionID, err := uuid.Parse(sectionIDStr)
		if err != nil {
			http.Error(w, "Invalid section_id", http.StatusBadRequest)
			return
		}
		resp, err = h.HandleBySection(r.Context(), pageID, &sectionID)
		if err != nil {
			logger.Error("Failed to list checks by section", zap.Error(err))
			http.Error(w, "Failed to list checks", http.StatusInternalServerError)
			return
		}
	} else {
		resp, err = h.Handle(r.Context(), pageID)
		if err != nil {
			logger.Error("Failed to list checks", zap.Error(err))
			http.Error(w, "Failed to list checks", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
