package listproviders

import (
	"context"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
)

// Request is the input for listing the provider catalog for an org.
type Request struct {
	OrgID uuid.UUID
}

// ProviderState is the per-provider capability + state returned to the UI.
type ProviderState struct {
	Key       string `json:"key"`
	AuthType  string `json:"authType"`
	Enabled   bool   `json:"enabled"`
	Connected bool   `json:"connected"`
}

// Response wraps the catalog.
type Response struct {
	Providers []ProviderState
}

// IntegrationLister is the narrow repo dependency.
type IntegrationLister interface {
	ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*entities.Integration, error)
}

// Handler builds the provider catalog from the registry, feature flags, and integrations.
type Handler struct {
	registry services.ProviderRegistry
	flags    services.FlagReader
	repo     IntegrationLister
}

// NewHandler wires the use case. Pass a true-nil flags interface when no reader exists.
func NewHandler(reg services.ProviderRegistry, flags services.FlagReader, repo IntegrationLister) *Handler {
	return &Handler{registry: reg, flags: flags, repo: repo}
}

// Handle returns the full known-provider set with computed enabled/connected/authType.
func (h *Handler) Handle(ctx context.Context, req Request) (*Response, error) {
	integrations, err := h.repo.ListByOrg(ctx, req.OrgID)
	if err != nil {
		return nil, err
	}
	active := make(map[string]bool, len(integrations))
	for _, i := range integrations {
		if i.Status == entities.IntegrationActive {
			active[i.ServiceType] = true
		}
	}

	states := make([]ProviderState, 0, len(services.KnownProviders))
	for _, d := range services.KnownProviders {
		_, registered := h.registry.Get(d.Key)
		enabled := registered && services.ProviderGateOpen(ctx, h.flags, req.OrgID, d.Key)

		connected := active[d.Key]
		if d.AuthType == services.AuthTypeNone {
			connected = true
		}

		states = append(states, ProviderState{
			Key:       d.Key,
			AuthType:  d.AuthType,
			Enabled:   enabled,
			Connected: connected,
		})
	}
	return &Response{Providers: states}, nil
}
