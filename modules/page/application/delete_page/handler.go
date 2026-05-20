package deletepage

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/page/domain/repositories"
)

type DeletePageHandler struct {
	repo repositories.PageRepository
}

func NewDeletePageHandler(repo repositories.PageRepository) *DeletePageHandler {
	return &DeletePageHandler{repo: repo}
}

func (h *DeletePageHandler) Handle(ctx context.Context, id uuid.UUID) error {
	return h.repo.Delete(ctx, id)
}
