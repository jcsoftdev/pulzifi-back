package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
)

// Module implements the router.ModuleRegisterer interface for the Monitoring module
type Module struct{}

// NewModule creates a new instance of the Monitoring module
func NewModule() router.ModuleRegisterer {
	return &Module{}
}

// ModuleName returns the name of the module
func (m *Module) ModuleName() string {
	return "Monitoring"
}

// RegisterHTTPRoutes registers all HTTP routes for the Monitoring module
func (m *Module) RegisterHTTPRoutes(router chi.Router) {
	router.Route("/monitoring", func(r chi.Router) {
		r.Route("/checks", func(cr chi.Router) {
			cr.Post("/", m.handleCreateCheck)
			cr.Get("/", m.handleListChecks)
			cr.Get("/{id}", m.handleGetCheck)
		})
	})
}

// handleCreateCheck creates a new monitoring check
// @Summary Create Monitoring Check
// @Description Create a new monitoring check
// @Tags monitoring
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Create Check Request"
// @Success 201 {object} map[string]interface{}
// @Router /monitoring/checks [post]
func (m *Module) handleCreateCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      "550e8400-e29b-41d4-a716-446655440000",
		"message": "create check",
	})
}

// handleListChecks lists all monitoring checks
// @Summary List Monitoring Checks
// @Description List all monitoring checks
// @Tags monitoring
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /monitoring/checks [get]
func (m *Module) handleListChecks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"checks":  []interface{}{},
		"message": "list checks",
	})
}

// handleGetCheck gets a monitoring check by ID
// @Summary Get Monitoring Check
// @Description Get a monitoring check by ID
// @Tags monitoring
// @Security BearerAuth
// @Produce json
// @Param id path string true "Check ID"
// @Success 200 {object} map[string]interface{}
// @Router /monitoring/checks/{id} [get]
func (m *Module) handleGetCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"id":      chi.URLParam(r, "id"),
		"message": "get check",
	})
}
