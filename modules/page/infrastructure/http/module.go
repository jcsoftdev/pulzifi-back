package http

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	createpage "github.com/jcsoftdev/pulzifi-back/modules/page/application/create_page"
	bulkdeletepages "github.com/jcsoftdev/pulzifi-back/modules/page/application/bulk_delete_pages"
	deletepage "github.com/jcsoftdev/pulzifi-back/modules/page/application/delete_page"
	getpage "github.com/jcsoftdev/pulzifi-back/modules/page/application/get_page"
	listpages "github.com/jcsoftdev/pulzifi-back/modules/page/application/list_pages"
	updatepage "github.com/jcsoftdev/pulzifi-back/modules/page/application/update_page"
	"github.com/jcsoftdev/pulzifi-back/modules/page/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
)

// Module implements the router.ModuleRegisterer interface for the Page module
type Module struct {
	db *sql.DB
}

// NewModule creates a new instance of the Page module
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
	return "Page"
}

// RegisterHTTPRoutes registers all HTTP routes for the Page module
func (m *Module) RegisterHTTPRoutes(router chi.Router) {
	router.Route("/pages", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware.Authenticate)
		r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)
		r.Use(middleware.RequireTenant)
		r.Post("/", m.handleCreatePage)
		r.Post("/bulk-delete", m.handleBulkDeletePages)
		r.Get("/", m.handleListPages)
		r.Get("/{id}", m.handleGetPage)
		r.Put("/{id}", m.handleUpdatePage)
		r.Delete("/{id}", m.handleDeletePage)
	})
}

// handleCreatePage creates a new page
// @Summary Create Page
// @Description Create a new page
// @Tags pages
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body createpage.CreatePageRequest true "Create Page Request"
// @Success 201 {object} createpage.CreatePageResponse
// @Router /pages [post]
func (m *Module) handleCreatePage(w http.ResponseWriter, r *http.Request) {
	// If db is not available, return mock response
	if m.db == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "550e8400-e29b-41d4-a716-446655440000",
			"message": "create page (mock - db not initialized)",
		})
		return
	}

	// Get tenant from context
	tenant := middleware.GetTenantFromContext(r.Context())

	// Create repository with dynamic tenant
	repo := persistence.NewPagePostgresRepository(m.db, tenant)

	// Use real handler
	handler := createpage.NewCreatePageHandler(repo)
	handler.HandleHTTP(w, r)
}

// handleListPages lists all pages
// @Summary List Pages
// @Description List all pages for a workspace
// @Tags pages
// @Security BearerAuth
// @Produce json
// @Param workspace_id query string true "Workspace ID"
// @Success 200 {object} listpages.ListPagesResponse
// @Router /pages [get]
func (m *Module) handleListPages(w http.ResponseWriter, r *http.Request) {
	// If db is not available, return mock response
	if m.db == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"pages":   []interface{}{},
			"message": "list pages (mock - db not initialized)",
		})
		return
	}

	// Get tenant from context
	tenant := middleware.GetTenantFromContext(r.Context())

	// Create repository with dynamic tenant
	repo := persistence.NewPagePostgresRepository(m.db, tenant)

	// Use real handler
	handler := listpages.NewListPagesHandler(repo)
	handler.HandleHTTP(w, r)
}

// handleGetPage gets a page by ID
// @Summary Get Page
// @Description Get a page by ID
// @Tags pages
// @Security BearerAuth
// @Produce json
// @Param id path string true "Page ID"
// @Success 200 {object} getpage.GetPageResponse
// @Router /pages/{id} [get]
func (m *Module) handleGetPage(w http.ResponseWriter, r *http.Request) {
	if m.db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewPagePostgresRepository(m.db, tenant)
	handler := getpage.NewGetPageHandler(repo)
	handler.HandleHTTP(w, r)
}

// handleUpdatePage updates a page
// @Summary Update Page
// @Description Update a page
// @Tags pages
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Page ID"
// @Param request body updatepage.UpdatePageRequest true "Update Page Request"
// @Success 200 {object} updatepage.UpdatePageResponse
// @Router /pages/{id} [put]
func (m *Module) handleUpdatePage(w http.ResponseWriter, r *http.Request) {
	if m.db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewPagePostgresRepository(m.db, tenant)
	handler := updatepage.NewUpdatePageHandler(repo)
	handler.HandleHTTP(w, r)
}

// handleDeletePage deletes a page
// @Summary Delete Page
// @Description Delete a page
// @Tags pages
// @Security BearerAuth
// @Success 204
// @Router /pages/{id} [delete]
func (m *Module) handleDeletePage(w http.ResponseWriter, r *http.Request) {
	if m.db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewPagePostgresRepository(m.db, tenant)
	handler := deletepage.NewDeletePageHandler(repo)
	handler.HandleHTTP(w, r)
}

// handleBulkDeletePages deletes multiple pages at once
// @Summary Bulk Delete Pages
// @Description Delete multiple pages by ID
// @Tags pages
// @Security BearerAuth
// @Accept json
// @Param request body bulkdeletepages.BulkDeletePagesRequest true "Bulk Delete Request"
// @Success 204
// @Router /pages/bulk-delete [post]
func (m *Module) handleBulkDeletePages(w http.ResponseWriter, r *http.Request) {
	if m.db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewPagePostgresRepository(m.db, tenant)
	handler := bulkdeletepages.NewBulkDeletePagesHandler(repo)
	handler.HandleHTTP(w, r)
}
