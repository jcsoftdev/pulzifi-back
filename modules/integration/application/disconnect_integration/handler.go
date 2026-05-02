package disconnectintegration

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
)

type Request struct {
	IntegrationID uuid.UUID
}

type Response struct{}

type DestinationRepoForTenant interface {
	DestinationRepoForTenant(tenant string) repositories.DestinationRepository
}

type Handler struct {
	intRepo  repositories.IntegrationRepository
	destRepo repositories.DestinationRepository
}

func NewHandler(intRepo repositories.IntegrationRepository, destRepo repositories.DestinationRepository) *Handler {
	return &Handler{intRepo: intRepo, destRepo: destRepo}
}

func (h *Handler) Handle(ctx context.Context, req Request) (*Response, error) {
	if req.IntegrationID == uuid.Nil {
		return nil, errors.New("integration_id required")
	}
	if err := h.intRepo.SoftDelete(ctx, req.IntegrationID); err != nil {
		return nil, err
	}
	if err := h.destRepo.DisableByIntegrationID(ctx, req.IntegrationID); err != nil {
		return nil, err
	}
	return &Response{}, nil
}
