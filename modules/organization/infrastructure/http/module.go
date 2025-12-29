package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/application/get_current_organization"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"go.uber.org/zap"
)

const (
	contentTypeJSON = "application/json"
)

// Module implements the router.ModuleRegisterer interface for the Organization module
type Module struct {
	getCurrentOrgHandler *get_current_organization.Handler
}

// NewModule creates a new instance of the Organization module
func NewModule(orgRepo repositories.OrganizationRepository) router.ModuleRegisterer {
	return &Module{
		getCurrentOrgHandler: get_current_organization.NewHandler(orgRepo),
	}
}

// ModuleName returns the name of the module
func (m *Module) ModuleName() string {
	return "Organization"
}

// RegisterHTTPRoutes registers all HTTP routes for the Organization module
func (m *Module) RegisterHTTPRoutes(router chi.Router) {
	router.Route("/organizations", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware.Authenticate)
		r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)
		r.Post("/", m.handleCreateOrganization)
		r.Get("/", m.handleListOrganizations)
		r.Get("/{id}", m.handleGetOrganization)
		r.Put("/{id}", m.handleUpdateOrganization)
		r.Delete("/{id}", m.handleDeleteOrganization)
	})

	router.Route("/organization", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware.Authenticate)
		r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)
		r.Get("/current", m.handleGetCurrentOrganization)
	})
}

// handleCreateOrganization creates a new organization
// @Summary Create Organization
// @Description Create a new organization with the provided name and subdomain
// @Tags organizations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Create Organization Request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /organizations [post]
func (m *Module) handleCreateOrganization(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      "550e8400-e29b-41d4-a716-446655440000",
		"message": "create organization",
	})
}

// handleListOrganizations lists all organizations
// @Summary List Organizations
// @Description List all organizations for the current user
// @Tags organizations
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /organizations [get]
func (m *Module) handleListOrganizations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"organizations": []interface{}{},
		"message":       "list organizations",
	})
}

// handleGetOrganization gets a specific organization
// @Summary Get Organization
// @Description Get a specific organization by ID
// @Tags organizations
// @Security BearerAuth
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /organizations/{id} [get]
func (m *Module) handleGetOrganization(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"id":      chi.URLParam(r, "id"),
		"message": "get organization",
	})
}

// handleUpdateOrganization updates an organization
// @Summary Update Organization
// @Description Update an organization
// @Tags organizations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param request body map[string]string true "Update Organization Request"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /organizations/{id} [put]
func (m *Module) handleUpdateOrganization(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"id":      chi.URLParam(r, "id"),
		"message": "update organization",
	})
}

// handleDeleteOrganization deletes an organization
// @Summary Delete Organization
// @Description Delete an organization
// @Tags organizations
// @Security BearerAuth
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /organizations/{id} [delete]
func (m *Module) handleDeleteOrganization(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(http.StatusNoContent)
}

// handleGetCurrentOrganization gets the current organization based on tenant
// @Summary Get Current Organization
// @Description Get the current organization based on the tenant from subdomain
// @Tags organizations
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /organization/current [get]
func (m *Module) handleGetCurrentOrganization(w http.ResponseWriter, r *http.Request) {
	// Extract subdomain from context (set by TenantMiddleware)
	subdomain := middleware.GetSubdomainFromContext(r.Context())
	if subdomain == "" {
		w.Header().Set("Content-Type", contentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Subdomain is required",
		})
		return
	}
	tenant := subdomain

	response, err := m.getCurrentOrgHandler.Handle(context.Background(), tenant)
	if err != nil {
		logger.Error("Failed to get current organization", zap.Error(err))
		w.Header().Set("Content-Type", contentTypeJSON)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get organization",
		})
		return
	}

	if response == nil {
		w.Header().Set("Content-Type", contentTypeJSON)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Organization not found",
		})
		return
	}

	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
