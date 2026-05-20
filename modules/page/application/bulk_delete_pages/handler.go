package bulkdeletepages

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/page/domain/repositories"
)

type BulkDeletePagesRequest struct {
	IDs []string `json:"ids"`
}

type BulkDeletePagesHandler struct {
	repo repositories.PageRepository
}

func NewBulkDeletePagesHandler(repo repositories.PageRepository) *BulkDeletePagesHandler {
	return &BulkDeletePagesHandler{repo: repo}
}

func (h *BulkDeletePagesHandler) Handle(ctx context.Context, ids []uuid.UUID) error {
	return h.repo.BulkDelete(ctx, ids)
}
