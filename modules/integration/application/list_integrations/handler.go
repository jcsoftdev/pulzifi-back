package listintegrations

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
)

type Response struct {
	Integrations []IntegrationResponse `json:"integrations"`
}

type IntegrationResponse struct {
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

func (h *Handler) Handle(ctx context.Context) (*Response, error) {
	integrations, err := h.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	resp := &Response{Integrations: make([]IntegrationResponse, 0, len(integrations))}
	for _, i := range integrations {
		resp.Integrations = append(resp.Integrations, IntegrationResponse{
			ID:          i.ID.String(),
			ServiceType: i.ServiceType,
			Config:      i.Config,
			Enabled:     i.Enabled,
			CreatedAt:   i.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return resp, nil
}
