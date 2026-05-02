package bulkretrydeliveries

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
)

type Request struct {
	IDs []uuid.UUID
}

type Response struct {
	Retried int `json:"retried"`
	Skipped int `json:"skipped"`
	Failed  int `json:"failed"`
}

type Handler struct {
	repo repositories.DeliveryRepository
}

func NewHandler(repo repositories.DeliveryRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle iterates request IDs and calls repo.Retry per row.
// Each call is independent (autocommit). Partial-success is acceptable —
// counts returned to the UI may not sum to len(IDs) on a server crash mid-loop.
// Each Retry is idempotent; re-running this is safe.
func (h *Handler) Handle(ctx context.Context, req Request) (*Response, error) {
	resp := &Response{}
	for _, id := range req.IDs {
		if id == uuid.Nil {
			resp.Failed++
			continue
		}
		err := h.repo.Retry(ctx, id)
		switch {
		case err == nil:
			resp.Retried++
		case errors.Is(err, repositories.ErrDeliveryNotInDeadState):
			resp.Skipped++
		default:
			resp.Failed++
		}
	}
	return resp, nil
}
