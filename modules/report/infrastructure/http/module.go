package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
)

// Module implements the router.ModuleRegisterer interface for the Report module
type Module struct{}

// NewModule creates a new instance of the Report module
func NewModule() router.ModuleRegisterer {
	return &Module{}
}

// ModuleName returns the name of the module
func (m *Module) ModuleName() string {
	return "Report"
}

// RegisterHTTPRoutes registers all HTTP routes for the Report module
func (m *Module) RegisterHTTPRoutes(router chi.Router) {
	router.Route("/reports", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware.Authenticate)
		r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)
		r.Post("/", m.handleCreateReport)
		r.Get("/", m.handleListReports)
		r.Get("/{id}", m.handleGetReport)
	})
}

// handleCreateReport creates a new report
// @Summary Create Report
// @Description Create a new report
// @Tags reports
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Create Report Request"
// @Success 201 {object} map[string]interface{}
// @Router /reports [post]
func (m *Module) handleCreateReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      "550e8400-e29b-41d4-a716-446655440000",
		"message": "create report",
	})
}

// handleListReports lists all reports
// @Summary List Reports
// @Description List all reports
// @Tags reports
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /reports [get]
func (m *Module) handleListReports(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"reports": []interface{}{},
		"message": "list reports",
	})
}

// handleGetReport gets a report by ID
// @Summary Get Report
// @Description Get a report by ID
// @Tags reports
// @Security BearerAuth
// @Produce json
// @Param id path string true "Report ID"
// @Success 200 {object} map[string]interface{}
// @Router /reports/{id} [get]
func (m *Module) handleGetReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"id":      chi.URLParam(r, "id"),
		"message": "get report",
	})
}
