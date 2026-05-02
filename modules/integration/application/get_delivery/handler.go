package getdelivery

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
)

type Request struct {
	ID uuid.UUID
}

type Response struct {
	Delivery *entities.Delivery
}

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
	d, err := h.repo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	return &Response{Delivery: d}, nil
}
