package http

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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
		r.Use(middleware.AuthMiddleware.Authenticate)
		r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)
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

// handleListAlerts lists all alerts for the current workspace
// @Summary List Alerts
// @Description List all alerts for the current workspace
// @Tags alerts
// @Security BearerAuth
// @Produce json
// @Param workspace_id query string false "Workspace ID (optional filter)"
// @Success 200 {object} map[string]interface{}
// @Router /alerts [get]
func (m *Module) handleListAlerts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// If db is not available, return mock response
	if m.db == nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":    []interface{}{},
			"message": "list alerts (mock - db not initialized)",
		})
		return
	}

	// Get tenant from context
	tenant := middleware.GetTenantFromContext(r.Context())

	// Get workspace_id from query params
	workspaceIDStr := r.URL.Query().Get("workspace_id")
	if workspaceIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "workspace_id query parameter is required",
		})
		return
	}

	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid workspace_id format",
		})
		return
	}

	// Create repository with dynamic tenant
	repo := persistence.NewAlertPostgresRepository(m.db, tenant)

	// List alerts
	alerts, err := repo.ListByWorkspace(r.Context(), workspaceID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "failed to list alerts",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  alerts,
		"count": len(alerts),
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
// @Failure 404 {object} map[string]string
// @Router /alerts/{id} [get]
func (m *Module) handleGetAlert(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid alert id"})
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewAlertPostgresRepository(m.db, tenant)

	alert, err := repo.GetByID(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to get alert"})
		return
	}
	if alert == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "alert not found"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(alert)
}

// handleUpdateAlert marks an alert as read
// @Summary Mark Alert as Read
// @Description Mark an alert as read
// @Tags alerts
// @Security BearerAuth
// @Produce json
// @Param id path string true "Alert ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /alerts/{id} [put]
func (m *Module) handleUpdateAlert(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid alert id"})
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewAlertPostgresRepository(m.db, tenant)

	if err := repo.MarkAsRead(r.Context(), id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to update alert"})
		return
	}

	alert, err := repo.GetByID(r.Context(), id)
	if err != nil || alert == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "alert not found"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(alert)
}

// handleDeleteAlert deletes an alert
// @Summary Delete Alert
// @Description Delete an alert
// @Tags alerts
// @Security BearerAuth
// @Success 204
// @Failure 400 {object} map[string]string
// @Router /alerts/{id} [delete]
func (m *Module) handleDeleteAlert(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid alert id"})
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewAlertPostgresRepository(m.db, tenant)

	if err := repo.Delete(r.Context(), id); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to delete alert"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
