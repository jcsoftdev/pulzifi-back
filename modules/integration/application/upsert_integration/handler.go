package upsertintegration

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
)

var ErrInvalidServiceType = errors.New("invalid service type")

var validServiceTypes = map[string]bool{
	"slack":         true,
	"teams":         true,
	"discord":       true,
	"google_sheets": true,
}

type Request struct {
	ServiceType string                 `json:"service_type"`
	Config      map[string]interface{} `json:"config"`
}

type Response struct {
	ID          string                 `json:"id"`
	ServiceType string                 `json:"service_type"`
	Config      map[string]interface{} `json:"config"`
	Enabled     bool                   `json:"enabled"`
	CreatedAt   string                 `json:"created_at"`
}

type Handler struct {
	repo repositories.IntegrationRepository
}

func NewHandler(repo repositories.IntegrationRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Handle(ctx context.Context, req *Request, createdBy uuid.UUID) (*Response, error) {
	if !validServiceTypes[req.ServiceType] {
		return nil, ErrInvalidServiceType
	}

	existing, err := h.repo.GetByServiceType(ctx, req.ServiceType)
	if err != nil {
		return nil, err
	}

	var integration *entities.Integration
	if existing != nil {
		existing.Config = req.Config
		existing.Enabled = true
		if err := h.repo.Update(ctx, existing); err != nil {
			return nil, err
		}
		integration = existing
	} else {
		integration = entities.NewIntegration(req.ServiceType, req.Config, createdBy)
		if err := h.repo.Create(ctx, integration); err != nil {
			return nil, err
		}
	}

	return &Response{
		ID:          integration.ID.String(),
		ServiceType: integration.ServiceType,
		Config:      integration.Config,
		Enabled:     integration.Enabled,
		CreatedAt:   integration.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
