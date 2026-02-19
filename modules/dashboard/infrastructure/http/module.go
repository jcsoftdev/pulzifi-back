package http

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	getdashboardstats "github.com/jcsoftdev/pulzifi-back/modules/dashboard/application/get_dashboard_stats"
	"github.com/jcsoftdev/pulzifi-back/modules/dashboard/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
)

// Module implements the router.ModuleRegisterer interface for the Dashboard module
type Module struct {
	db *sql.DB
}

// NewModule creates a new instance of the Dashboard module without DB
func NewModule() router.ModuleRegisterer {
	return &Module{}
}

// NewModuleWithDB creates a new instance with database connection
func NewModuleWithDB(db *sql.DB) router.ModuleRegisterer {
	return &Module{db: db}
}

// ModuleName returns the name of the module
func (m *Module) ModuleName() string {
	return "Dashboard"
}

// RegisterHTTPRoutes registers all HTTP routes for the Dashboard module
func (m *Module) RegisterHTTPRoutes(r chi.Router) {
	r.Route("/dashboard", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware.Authenticate)
		r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)
		r.Use(middleware.RequireTenant)
		r.Get("/stats", m.handleGetStats)
	})
}

// handleGetStats returns aggregated dashboard statistics
func (m *Module) handleGetStats(w http.ResponseWriter, r *http.Request) {
	if m.db == nil {
		http.Error(w, "database not initialized", http.StatusInternalServerError)
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewDashboardPostgresRepository(m.db, tenant)
	handler := getdashboardstats.NewGetDashboardStatsHandler(repo)
	handler.HandleHTTP(w, r)
}
