package listdestinations

import (
	"context"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
)

type Request struct {
	ScopeType entities.ScopeType
	ScopeID   uuid.UUID
}

type Response struct {
	Destinations []*entities.Destination
}

type Handler struct {
	repo repositories.DestinationRepository
}

func NewHandler(repo repositories.DestinationRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Handle(ctx context.Context, req Request) (*Response, error) {
	destinations, err := h.repo.ListByScope(ctx, req.ScopeType, req.ScopeID)
	if err != nil {
		return nil, err
	}
	return &Response{Destinations: destinations}, nil
}
