package listchecks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/repositories"
)

type ListChecksHandler struct {
	repo          repositories.CheckRepository
	mediationCfg  URLMediationConfig
}

func NewListChecksHandler(repo repositories.CheckRepository) *ListChecksHandler {
	return &ListChecksHandler{repo: repo}
}

// NewListChecksHandlerWithMediation creates a handler that applies URL ownership
// checks and optional presigning when building check responses.
func NewListChecksHandlerWithMediation(repo repositories.CheckRepository, tenant string, presigner SnapshotURLPresigner, bucketPrivate bool, presignTTL time.Duration) *ListChecksHandler {
	return &ListChecksHandler{
		repo: repo,
		mediationCfg: URLMediationConfig{
			Tenant:        tenant,
			BucketPrivate: bucketPrivate,
			PresignTTL:    presignTTL,
			Presigner:     presigner,
		},
	}
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

	return buildResponseWithSections(ctx, parentChecks, sectionsByParent, h.mediationCfg), nil
}

// HandleBySection returns checks filtered by section. sectionID nil means full-page checks only.
func (h *ListChecksHandler) HandleBySection(ctx context.Context, pageID uuid.UUID, sectionID *uuid.UUID) (*ListChecksResponse, error) {
	checks, err := h.repo.ListByPageAndSection(ctx, pageID, sectionID)
	if err != nil {
		return nil, err
	}

	return buildResponse(ctx, checks, h.mediationCfg), nil
}

func toCheckResponse(ctx context.Context, check *entities.Check, cfg URLMediationConfig) *CheckResponse {
	resp := &CheckResponse{
		ID:              check.ID,
		PageID:          check.PageID,
		SectionID:       check.SectionID,
		ParentCheckID:   check.ParentCheckID,
		Status:          check.Status,
		ScreenshotURL:   mediateURL(ctx, check.ScreenshotURL, cfg),
		HTMLSnapshotURL: mediateURL(ctx, check.HTMLSnapshotURL, cfg),
		ChangeDetected:  check.ChangeDetected,
		ChangeType:      check.ChangeType,
		ErrorMessage:    check.ErrorMessage,
		ContentDiff:     check.ContentDiffJSON,
		DiffImageURL:    mediateURL(ctx, check.DiffImageURL, cfg),
		CheckedAt:       check.CheckedAt,
	}
	return resp
}

func buildResponse(ctx context.Context, checks []*entities.Check, cfg URLMediationConfig) *ListChecksResponse {
	response := &ListChecksResponse{
		Checks: make([]*CheckResponse, len(checks)),
	}
	for i, check := range checks {
		response.Checks[i] = toCheckResponse(ctx, check, cfg)
	}
	return response
}

func buildResponseWithSections(ctx context.Context, parentChecks []*entities.Check, sectionsByParent map[uuid.UUID][]*entities.Check, cfg URLMediationConfig) *ListChecksResponse {
	response := &ListChecksResponse{
		Checks: make([]*CheckResponse, len(parentChecks)),
	}
	for i, check := range parentChecks {
		cr := toCheckResponse(ctx, check, cfg)
		if sections, ok := sectionsByParent[check.ID]; ok {
			cr.Sections = make([]*CheckResponse, len(sections))
			for j, sc := range sections {
				cr.Sections[j] = toCheckResponse(ctx, sc, cfg)
			}
		}
		response.Checks[i] = cr
	}
	return response
}
