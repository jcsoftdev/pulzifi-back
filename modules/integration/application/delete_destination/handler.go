package deletedestination

import (
	"context"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
)

type Request struct {
	ID uuid.UUID
}

type Response struct{}

type Handler struct {
	repo repositories.DestinationRepository
}

func NewHandler(repo repositories.DestinationRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Handle(ctx context.Context, req Request) (*Response, error) {
	if err := h.repo.Delete(ctx, req.ID); err != nil {
		return nil, err
	}
	return &Response{}, nil
}
