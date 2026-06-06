package createdestination

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
)

type Request struct {
	IntegrationID *uuid.UUID         `json:"integration_id"`
	ServiceType   string             `json:"service_type"`
	ScopeType     entities.ScopeType `json:"scope_type"`
	ScopeID       uuid.UUID          `json:"scope_id"`
	Target        map[string]any     `json:"target"`
	Events        []string           `json:"events"`
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
	if req.ServiceType == "" {
		return nil, errors.New("service_type required")
	}
	if req.ScopeType != entities.ScopeOrg &&
		req.ScopeType != entities.ScopeWorkspace &&
		req.ScopeType != entities.ScopePage {
		return nil, errors.New("invalid scope_type")
	}
	if req.ScopeID == uuid.Nil {
		return nil, errors.New("scope_id required")
	}
	if len(req.Events) == 0 {
		return nil, errors.New("at least one event required")
	}
	if req.Target == nil {
		return nil, errors.New("target required")
	}

	d := &entities.Destination{
		ID:            uuid.New(),
		IntegrationID: req.IntegrationID,
		ServiceType:   req.ServiceType,
		ScopeType:     req.ScopeType,
		ScopeID:       req.ScopeID,
		Target:        req.Target,
		Events:        req.Events,
		Enabled:       true,
	}

	if err := h.repo.Create(ctx, d); err != nil {
		return nil, err
	}
	return &Response{Destination: d}, nil
}
