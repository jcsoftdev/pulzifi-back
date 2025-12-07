package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
)

// Module implements the router.ModuleRegisterer interface for the Usage module
type Module struct{}

// NewModule creates a new instance of the Usage module
func NewModule() router.ModuleRegisterer {
	return &Module{}
}

// ModuleName returns the name of the module
func (m *Module) ModuleName() string {
	return "Usage"
}

// RegisterHTTPRoutes registers all HTTP routes for the Usage module
func (m *Module) RegisterHTTPRoutes(router chi.Router) {
	router.Route("/usage", func(r chi.Router) {
		r.Get("/metrics", m.handleGetMetrics)
		r.Get("/quotas", m.handleGetQuotas)
	})
}

// handleGetMetrics returns usage metrics
// @Summary Get Usage Metrics
// @Description Get usage metrics for the current organization
// @Tags usage
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /usage/metrics [get]
func (m *Module) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"metrics": map[string]interface{}{},
		"message": "get usage metrics",
	})
}

// handleGetQuotas returns usage quotas
// @Summary Get Usage Quotas
// @Description Get usage quotas for the current organization
// @Tags usage
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /usage/quotas [get]
func (m *Module) handleGetQuotas(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"quotas":  map[string]interface{}{},
		"message": "get usage quotas",
	})
}
