package http

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	createcheck "github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/create_check"
	createmonitoringconfig "github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/create_monitoring_config"
	createnotificationpreference "github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/create_notification_preference"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
)

// Module implements the router.ModuleRegisterer interface for the Monitoring module
type Module struct {
	db *sql.DB
}

// NewModule creates a new instance of the Monitoring module
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
	return "Monitoring"
}

// RegisterHTTPRoutes registers all HTTP routes for the Monitoring module
func (m *Module) RegisterHTTPRoutes(router chi.Router) {
	router.Route("/monitoring", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware.Authenticate)
		r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)
		r.Route("/checks", func(cr chi.Router) {
			cr.Post("/", m.handleCreateCheck)
			cr.Get("/", m.handleListChecks)
			cr.Get("/{id}", m.handleGetCheck)
		})
		r.Route("/configs", func(cr chi.Router) {
			cr.Post("/", m.handleCreateMonitoringConfig)
			cr.Get("/{pageId}", m.handleGetMonitoringConfig)
		})
		r.Route("/notification-preferences", func(cr chi.Router) {
			cr.Post("/", m.handleCreateNotificationPreference)
			cr.Get("/{id}", m.handleGetNotificationPreference)
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
// @Param request body createcheck.CreateCheckRequest true "Create Check Request"
// @Success 201 {object} createcheck.CreateCheckResponse
// @Router /monitoring/checks [post]
func (m *Module) handleCreateCheck(w http.ResponseWriter, r *http.Request) {
	// If db is not available, return mock response
	if m.db == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "550e8400-e29b-41d4-a716-446655440000",
			"message": "create check (mock - db not initialized)",
		})
		return
	}

	// Get tenant from context
	tenant := middleware.GetTenantFromContext(r.Context())

	// Create repository with dynamic tenant
	repo := persistence.NewCheckPostgresRepository(m.db, tenant)

	// Use real handler
	handler := createcheck.NewCreateCheckHandler(repo)
	handler.HandleHTTP(w, r)
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

// handleCreateMonitoringConfig creates a new monitoring config
// @Summary Create Monitoring Config
// @Description Create a new monitoring config
// @Tags monitoring
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body create_monitoring_config.CreateMonitoringConfigRequest true "Create Config Request"
// @Success 201 {object} create_monitoring_config.CreateMonitoringConfigResponse
// @Router /monitoring/configs [post]
func (m *Module) handleCreateMonitoringConfig(w http.ResponseWriter, r *http.Request) {
	// If db is not available, return mock response
	if m.db == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "550e8400-e29b-41d4-a716-446655440000",
			"message": "create monitoring config (mock - db not initialized)",
		})
		return
	}

	// Get tenant from context
	tenant := middleware.GetTenantFromContext(r.Context())

	// Create repository with dynamic tenant
	repo := persistence.NewMonitoringConfigPostgresRepository(m.db, tenant)

	// Use real handler
	handler := createmonitoringconfig.NewCreateMonitoringConfigHandler(repo)
	handler.HandleHTTP(w, r)
}

// handleGetMonitoringConfig gets a monitoring config by page ID
// @Summary Get Monitoring Config
// @Description Get a monitoring config by page ID
// @Tags monitoring
// @Security BearerAuth
// @Produce json
// @Param pageId path string true "Page ID"
// @Success 200 {object} map[string]interface{}
// @Router /monitoring/configs/{pageId} [get]
func (m *Module) handleGetMonitoringConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"page_id": chi.URLParam(r, "pageId"),
		"message": "get monitoring config",
	})
}

// handleCreateNotificationPreference creates a new notification preference
// @Summary Create Notification Preference
// @Description Create a new notification preference
// @Tags monitoring
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body create_notification_preference.CreateNotificationPreferenceRequest true "Create Preference Request"
// @Success 201 {object} create_notification_preference.CreateNotificationPreferenceResponse
// @Router /monitoring/notification-preferences [post]
func (m *Module) handleCreateNotificationPreference(w http.ResponseWriter, r *http.Request) {
	// If db is not available, return mock response
	if m.db == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "550e8400-e29b-41d4-a716-446655440000",
			"message": "create notification preference (mock - db not initialized)",
		})
		return
	}

	// Get tenant from context
	tenant := middleware.GetTenantFromContext(r.Context())

	// Create repository with dynamic tenant
	repo := persistence.NewNotificationPreferencePostgresRepository(m.db, tenant)

	// Use real handler
	handler := createnotificationpreference.NewCreateNotificationPreferenceHandler(repo)
	handler.HandleHTTP(w, r)
}

// handleGetNotificationPreference gets a notification preference by ID
// @Summary Get Notification Preference
// @Description Get a notification preference by ID
// @Tags monitoring
// @Security BearerAuth
// @Produce json
// @Param id path string true "Preference ID"
// @Success 200 {object} map[string]interface{}
// @Router /monitoring/notification-preferences/{id} [get]
func (m *Module) handleGetNotificationPreference(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"id":      chi.URLParam(r, "id"),
		"message": "get notification preference",
	})
}
