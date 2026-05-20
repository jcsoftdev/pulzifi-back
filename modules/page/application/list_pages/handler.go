package listpages

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/page/domain/repositories"
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
