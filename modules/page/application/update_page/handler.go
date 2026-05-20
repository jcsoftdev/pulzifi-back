package updatepage

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/page/domain/repositories"
)

type UpdatePageHandler struct {
	repo repositories.PageRepository
}

func NewUpdatePageHandler(repo repositories.PageRepository) *UpdatePageHandler {
	return &UpdatePageHandler{repo: repo}
}

func (h *UpdatePageHandler) Handle(ctx context.Context, id uuid.UUID, req *UpdatePageRequest) (*UpdatePageResponse, error) {
	page, err := h.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if page == nil {
		return nil, nil
	}

	if req.Name != "" {
		page.Name = req.Name
	}
	if req.URL != "" {
		page.URL = req.URL
	}
	if req.Tags != nil {
		page.Tags = req.Tags
	}

	if err := h.repo.Update(ctx, page); err != nil {
		return nil, err
	}

	return &UpdatePageResponse{
		ID:   page.ID,
		Name: page.Name,
		URL:  page.URL,
	}, nil
}
