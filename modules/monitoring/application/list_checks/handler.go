package listchecks

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/repositories"
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
	resp := &CheckResponse{
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
		ContentDiff:     check.ContentDiffJSON,
		DiffImageURL:    check.DiffImageURL,
		CheckedAt:       check.CheckedAt,
	}
	return resp
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
