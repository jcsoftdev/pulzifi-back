package createpage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/page/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/page/domain/repositories"
)

type CreatePageHandler struct {
	repo repositories.PageRepository
}

func NewCreatePageHandler(repo repositories.PageRepository) *CreatePageHandler {
	return &CreatePageHandler{repo: repo}
}

func (h *CreatePageHandler) Handle(ctx context.Context, req *CreatePageRequest, createdBy uuid.UUID) (*CreatePageResponse, error) {
	// Create page entity
	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}

	page := &entities.Page{
		ID:              uuid.New(),
		WorkspaceID:     req.WorkspaceID,
		Name:            req.Name,
		URL:             req.URL,
		CheckCount:      0,
		Tags:            tags,
		CheckFrequency:  "Off",
		DetectedChanges: 0,
		CreatedBy:       createdBy,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Save to database
	if err := h.repo.Create(ctx, page); err != nil {
		return nil, err
	}

	return &CreatePageResponse{
		ID:                   page.ID,
		WorkspaceID:          page.WorkspaceID,
		Name:                 page.Name,
		URL:                  page.URL,
		ThumbnailURL:         page.ThumbnailURL,
		LastCheckedAt:        page.LastCheckedAt,
		LastChangeDetectedAt: page.LastChangeDetectedAt,
		CheckCount:           page.CheckCount,
		Tags:                 page.Tags,
		CheckFrequency:       page.CheckFrequency,
		DetectedChanges:      page.DetectedChanges,
		CreatedBy:            page.CreatedBy,
		CreatedAt:            page.CreatedAt,
		UpdatedAt:            page.UpdatedAt,
	}, nil
}
