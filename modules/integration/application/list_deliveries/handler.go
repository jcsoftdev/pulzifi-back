package listdeliveries

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
)

// Request describes a paginated query for deliveries belonging to a destination.
// status filter deferred to Phase 2 (requires new repo method).
type Request struct {
	DestinationID uuid.UUID
	Limit         int // default 50, max 200
	Offset        int // default 0
}

type Response struct {
	Deliveries []*entities.Delivery
}

type Handler struct {
	repo repositories.DeliveryRepository
}

func NewHandler(repo repositories.DeliveryRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Handle(ctx context.Context, req Request) (*Response, error) {
	if req.DestinationID == uuid.Nil {
		return nil, errors.New("destination_id required")
	}
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Limit > 200 {
		req.Limit = 200
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	deliveries, err := h.repo.ListByDestination(ctx, req.DestinationID, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}
	return &Response{Deliveries: deliveries}, nil
}
