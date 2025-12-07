package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
)

// Module implements the router.ModuleRegisterer interface for the Insight module
type Module struct{}

// NewModule creates a new instance of the Insight module
func NewModule() router.ModuleRegisterer {
	return &Module{}
}

// ModuleName returns the name of the module
func (m *Module) ModuleName() string {
	return "Insight"
}

// RegisterHTTPRoutes registers all HTTP routes for the Insight module
func (m *Module) RegisterHTTPRoutes(router chi.Router) {
	router.Route("/insights", func(r chi.Router) {
		r.Post("/generate", m.handleGenerateInsight)
		r.Get("/", m.handleListInsights)
		r.Get("/{id}", m.handleGetInsight)
	})
}

// handleGenerateInsight generates a new insight
// @Summary Generate Insight
// @Description Generate a new insight
// @Tags insights
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Generate Insight Request"
// @Success 201 {object} map[string]interface{}
// @Router /insights/generate [post]
func (m *Module) handleGenerateInsight(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      "550e8400-e29b-41d4-a716-446655440000",
		"message": "generate insight",
	})
}

// handleListInsights lists all insights
// @Summary List Insights
// @Description List all insights
// @Tags insights
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /insights [get]
func (m *Module) handleListInsights(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"insights": []interface{}{},
		"message":  "list insights",
	})
}

// handleGetInsight gets an insight by ID
// @Summary Get Insight
// @Description Get an insight by ID
// @Tags insights
// @Security BearerAuth
// @Produce json
// @Param id path string true "Insight ID"
// @Success 200 {object} map[string]interface{}
// @Router /insights/{id} [get]
func (m *Module) handleGetInsight(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"id":      chi.URLParam(r, "id"),
		"message": "get insight",
	})
}
