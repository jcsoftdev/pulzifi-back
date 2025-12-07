package http

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	createalert "github.com/jcsoftdev/pulzifi-back/modules/alert/application/create_alert"
	"github.com/jcsoftdev/pulzifi-back/modules/alert/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
)

// Module implements the router.ModuleRegisterer interface for the Alert module
type Module struct {
	db *sql.DB
}

// NewModule creates a new instance of the Alert module
func NewModule() router.ModuleRegisterer {
	return &Module{}
}

// NewModuleWithDB creates a new instance with database connection
func NewModuleWithDB(db *sql.DB) router.ModuleRegisterer {
	return &Module{
		db: db,
	}
}

// ModuleName returns the name of the module
func (m *Module) ModuleName() string {
	return "Alert"
}

// RegisterHTTPRoutes registers all HTTP routes for the Alert module
func (m *Module) RegisterHTTPRoutes(router chi.Router) {
	router.Route("/alerts", func(r chi.Router) {
		r.Post("/", m.handleCreateAlert)
		r.Get("/", m.handleListAlerts)
		r.Get("/{id}", m.handleGetAlert)
		r.Put("/{id}", m.handleUpdateAlert)
		r.Delete("/{id}", m.handleDeleteAlert)
	})
}

// handleCreateAlert creates a new alert
// @Summary Create Alert
// @Description Create a new alert
// @Tags alerts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body createalert.CreateAlertRequest true "Create Alert Request"
// @Success 201 {object} createalert.CreateAlertResponse
// @Router /alerts [post]
func (m *Module) handleCreateAlert(w http.ResponseWriter, r *http.Request) {
	// If db is not available, return mock response for now
	if m.db == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "550e8400-e29b-41d4-a716-446655440000",
			"message": "create alert (mock - db not initialized)",
		})
		return
	}

	// Get tenant from context
	tenant := middleware.GetTenantFromContext(r.Context())

	// Create repository with dynamic tenant
	repo := persistence.NewAlertPostgresRepository(m.db, tenant)

	// Use real handler
	handler := createalert.NewCreateAlertHandler(repo)
	handler.HandleHTTP(w, r)
}

// handleListAlerts lists all alerts
// @Summary List Alerts
// @Description List all alerts
// @Tags alerts
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /alerts [get]
func (m *Module) handleListAlerts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"alerts":  []interface{}{},
		"message": "list alerts",
	})
}

// handleGetAlert gets an alert by ID
// @Summary Get Alert
// @Description Get an alert by ID
// @Tags alerts
// @Security BearerAuth
// @Produce json
// @Param id path string true "Alert ID"
// @Success 200 {object} map[string]interface{}
// @Router /alerts/{id} [get]
func (m *Module) handleGetAlert(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"id":      chi.URLParam(r, "id"),
		"message": "get alert",
	})
}

// handleUpdateAlert updates an alert
// @Summary Update Alert
// @Description Update an alert
// @Tags alerts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Alert ID"
// @Param request body map[string]string true "Update Alert Request"
// @Success 200 {object} map[string]interface{}
// @Router /alerts/{id} [put]
func (m *Module) handleUpdateAlert(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"id":      chi.URLParam(r, "id"),
		"message": "update alert",
	})
}

// handleDeleteAlert deletes an alert
// @Summary Delete Alert
// @Description Delete an alert
// @Tags alerts
// @Security BearerAuth
// @Success 204
// @Router /alerts/{id} [delete]
func (m *Module) handleDeleteAlert(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
