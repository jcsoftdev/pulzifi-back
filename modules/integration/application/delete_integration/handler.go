package deleteintegration

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
)

var ErrIntegrationNotFound = errors.New("integration not found")

type Handler struct {
	repo repositories.IntegrationRepository
}

func NewHandler(repo repositories.IntegrationRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Handle(ctx context.Context, id uuid.UUID) error {
	return h.repo.DeleteByID(ctx, id)
}
