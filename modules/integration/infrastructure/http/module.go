package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
)

// Module implements the router.ModuleRegisterer interface for the Integration module
type Module struct{}

// NewModule creates a new instance of the Integration module
func NewModule() router.ModuleRegisterer {
	return &Module{}
}

// ModuleName returns the name of the module
func (m *Module) ModuleName() string {
	return "Integration"
}

// RegisterHTTPRoutes registers all HTTP routes for the Integration module
func (m *Module) RegisterHTTPRoutes(router chi.Router) {
	router.Route("/integrations", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware.Authenticate)
		r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)
		r.Route("/webhooks", func(wr chi.Router) {
			wr.Post("/", m.handleCreateWebhook)
			wr.Get("/", m.handleListWebhooks)
			wr.Get("/{id}", m.handleGetWebhook)
		})
	})
}

// handleCreateWebhook creates a new webhook integration
// @Summary Create Webhook Integration
// @Description Create a new webhook integration
// @Tags integrations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Create Webhook Request"
// @Success 201 {object} map[string]interface{}
// @Router /integrations/webhooks [post]
func (m *Module) handleCreateWebhook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      "550e8400-e29b-41d4-a716-446655440000",
		"message": "create webhook",
	})
}

// handleListWebhooks lists all webhook integrations
// @Summary List Webhook Integrations
// @Description List all webhook integrations
// @Tags integrations
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /integrations/webhooks [get]
func (m *Module) handleListWebhooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"webhooks": []interface{}{},
		"message":  "list webhooks",
	})
}

// handleGetWebhook gets a webhook integration by ID
// @Summary Get Webhook Integration
// @Description Get a webhook integration by ID
// @Tags integrations
// @Security BearerAuth
// @Produce json
// @Param id path string true "Webhook ID"
// @Success 200 {object} map[string]interface{}
// @Router /integrations/webhooks/{id} [get]
func (m *Module) handleGetWebhook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"id":      chi.URLParam(r, "id"),
		"message": "get webhook",
	})
}
