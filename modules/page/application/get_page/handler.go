package getpage

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/page/domain/repositories"
)

type GetPageHandler struct {
	repo repositories.PageRepository
}

func NewGetPageHandler(repo repositories.PageRepository) *GetPageHandler {
	return &GetPageHandler{repo: repo}
}

func (h *GetPageHandler) Handle(ctx context.Context, id uuid.UUID) (*GetPageResponse, error) {
	page, err := h.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if page == nil {
		return nil, nil
	}

	return &GetPageResponse{
		ID:                   page.ID,
		WorkspaceID:          page.WorkspaceID,
		Name:                 page.Name,
		URL:                  page.URL,
		ThumbnailURL:         page.ThumbnailURL,
		LastCheckedAt:        page.LastCheckedAt,
		LastChangeDetectedAt: page.LastChangeDetectedAt,
		CheckCount:           page.CheckCount,
		CheckFrequency:       page.CheckFrequency,
		DetectedChanges:      page.DetectedChanges,
		Tags:                 page.Tags,
		CreatedAt:            page.CreatedAt,
		UpdatedAt:            page.UpdatedAt,
	}, nil
}
