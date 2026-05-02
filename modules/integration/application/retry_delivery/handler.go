package retrydelivery

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
)

type Request struct {
	ID uuid.UUID
}

type Response struct{}

type Handler struct {
	repo repositories.DeliveryRepository
}

func NewHandler(repo repositories.DeliveryRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Handle(ctx context.Context, req Request) (*Response, error) {
	if req.ID == uuid.Nil {
		return nil, errors.New("delivery id required")
	}
	if err := h.repo.Retry(ctx, req.ID); err != nil {
		return nil, err
	}
	return &Response{}, nil
}
