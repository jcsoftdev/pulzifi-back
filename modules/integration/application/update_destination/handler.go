package updatedestination

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
)

type Request struct {
	ID      uuid.UUID      `json:"-"`       // set from URL param, never the body
	Target  map[string]any `json:"target"`  // optional; nil = no change
	Events  []string       `json:"events"`  // optional; nil = no change
	Enabled *bool          `json:"enabled"` // optional; nil = no change
}

type Response struct {
	Destination *entities.Destination
}

type Handler struct {
	repo repositories.DestinationRepository
}

func NewHandler(repo repositories.DestinationRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Handle(ctx context.Context, req Request) (*Response, error) {
	existing, err := h.repo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("destination not found")
	}

	if req.Target != nil {
		existing.Target = req.Target
	}
	if req.Events != nil {
		if len(req.Events) == 0 {
			return nil, errors.New("events cannot be empty")
		}
		existing.Events = req.Events
	}
	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}

	if err := h.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return &Response{Destination: existing}, nil
}
