package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	authmw "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/middleware"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/application/get_current_organization"
	orgentities "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/entities"
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
	orgRepo              repositories.OrganizationRepository
	getCurrentOrgHandler *get_current_organization.Handler
}

// NewModule creates a new instance of the Organization module
func NewModule(orgRepo repositories.OrganizationRepository) router.ModuleRegisterer {
	return &Module{
		orgRepo:              orgRepo,
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
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{"error": "use the registration/approval flow to create organizations"})
}

// handleListOrganizations lists all organizations for the current user
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

	userIDStr, ok := r.Context().Value(authmw.UserIDKey).(string)
	if !ok || userIDStr == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid user id"})
		return
	}

	orgs, err := m.orgRepo.List(r.Context(), userID, 100, 0)
	if err != nil {
		logger.Error("Failed to list organizations", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to list organizations"})
		return
	}
	if orgs == nil {
		orgs = []*orgentities.Organization{}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"organizations": orgs})
}

// handleGetOrganization gets a specific organization by ID
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

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid organization id"})
		return
	}

	org, err := m.orgRepo.GetByID(r.Context(), id)
	if err != nil {
		logger.Error("Failed to get organization", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to get organization"})
		return
	}
	if org == nil || org.IsDeleted() {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "organization not found"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(org)
}

// handleUpdateOrganization updates an organization's name
// @Summary Update Organization
// @Description Update an organization's name
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

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid organization id"})
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "name is required"})
		return
	}

	org, err := m.orgRepo.GetByID(r.Context(), id)
	if err != nil || org == nil || org.IsDeleted() {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "organization not found"})
		return
	}

	org.Name = req.Name
	org.UpdatedAt = time.Now()
	if err := m.orgRepo.Update(r.Context(), org); err != nil {
		logger.Error("Failed to update organization", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to update organization"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(org)
}

// handleDeleteOrganization soft-deletes an organization
// @Summary Delete Organization
// @Description Soft-delete an organization
// @Tags organizations
// @Security BearerAuth
// @Produce json
// @Param id path string true "Organization ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Router /organizations/{id} [delete]
func (m *Module) handleDeleteOrganization(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		w.Header().Set("Content-Type", contentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid organization id"})
		return
	}

	if err := m.orgRepo.Delete(r.Context(), id); err != nil {
		w.Header().Set("Content-Type", contentTypeJSON)
		logger.Error("Failed to delete organization", zap.Error(err))
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "organization not found"})
		return
	}

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
