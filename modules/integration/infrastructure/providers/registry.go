package providers

import (
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
)

type Registry struct {
	clients map[string]services.ProviderClient
}

func NewRegistry(clients ...services.ProviderClient) *Registry {
	m := make(map[string]services.ProviderClient, len(clients))
	for _, c := range clients {
		m[c.ServiceType()] = c
	}
	return &Registry{clients: m}
}

func (r *Registry) Get(t string) (services.ProviderClient, bool) {
	c, ok := r.clients[t]
	return c, ok
}

var _ services.ProviderRegistry = (*Registry)(nil)
