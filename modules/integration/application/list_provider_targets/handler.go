package listprovidertargets

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
)

type Request struct {
	OrgID       uuid.UUID
	ServiceType string
}

type Response struct {
	Targets []entities.Target
}

type Handler struct {
	intRepo  repositories.IntegrationRepository
	registry services.ProviderRegistry
}

func NewHandler(intRepo repositories.IntegrationRepository, registry services.ProviderRegistry) *Handler {
	return &Handler{intRepo: intRepo, registry: registry}
}

func (h *Handler) Handle(ctx context.Context, req Request) (*Response, error) {
	if req.ServiceType == "" {
		return nil, errors.New("service_type required")
	}
	if req.OrgID == uuid.Nil {
		return nil, errors.New("org_id required")
	}

	integ, err := h.intRepo.GetByOrgAndService(ctx, req.OrgID, req.ServiceType)
	if err != nil {
		return nil, err
	}
	if integ == nil {
		return nil, errors.New("no active integration for service")
	}

	client, ok := h.registry.Get(req.ServiceType)
	if !ok {
		return nil, errors.New("unknown provider")
	}

	targets, err := client.ListTargets(ctx, integ)
	if err != nil {
		return nil, err
	}
	return &Response{Targets: targets}, nil
}
