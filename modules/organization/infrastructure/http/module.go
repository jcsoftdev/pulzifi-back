package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	deleteorganization "github.com/jcsoftdev/pulzifi-back/modules/organization/application/delete_organization"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/application/get_current_organization"
	orgentities "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/repositories"
	orgservices "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/contextkeys"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"go.uber.org/zap"
)

const (
	contentTypeJSON = "application/json"
)

// DeleteOrgHandler is the port the HTTP layer calls to delete an organization.
// Using an interface allows test doubles to stand in for *deleteorganization.Handler.
type DeleteOrgHandler interface {
	Handle(ctx context.Context, req *deleteorganization.Request) (*deleteorganization.Response, error)
}

// Module implements the router.ModuleRegisterer interface for the Organization module
type Module struct {
	orgRepo              repositories.OrganizationRepository
	getCurrentOrgHandler *get_current_organization.Handler
	deleteOrgHandler     DeleteOrgHandler // nil when use case not wired (legacy mode)
}

// NewModule creates a new instance of the Organization module (legacy, without cascade use case).
func NewModule(orgRepo repositories.OrganizationRepository) router.ModuleRegisterer {
	return &Module{
		orgRepo:              orgRepo,
		getCurrentOrgHandler: get_current_organization.NewHandler(orgRepo),
	}
}

// newModuleWithDeleteHandler creates a Module with only the delete handler injected.
// orgRepo is intentionally nil; callers must not invoke routes that require it.
// Used only by package-internal tests — not for production wiring (use NewModuleWithAll).
func newModuleWithDeleteHandler(deleteHandler DeleteOrgHandler) *Module {
	return &Module{
		deleteOrgHandler: deleteHandler,
	}
}

// NewModuleWithAll creates a fully wired Module (orgRepo + cascade delete handler).
func NewModuleWithAll(orgRepo repositories.OrganizationRepository, deleteHandler DeleteOrgHandler) router.ModuleRegisterer {
	return &Module{
		orgRepo:              orgRepo,
		getCurrentOrgHandler: get_current_organization.NewHandler(orgRepo),
		deleteOrgHandler:     deleteHandler,
	}
}

// ServeDeleteOrganization is the public handler entry point exposed for testing and routing.
func (m *Module) ServeDeleteOrganization(w http.ResponseWriter, r *http.Request) {
	m.handleDeleteOrganization(w, r)
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
	})

	// DELETE /organizations/{id} is intentionally outside the membership group.
	// Its only legitimate caller is SUPER_ADMIN acting from the apex domain (no
	// tenant subdomain), so RequireOrganizationMembership would 403 before the
	// in-handler SUPER_ADMIN guard could run (DAO-12). Authentication is still
	// enforced; the SUPER_ADMIN role check inside handleDeleteOrganization is the
	// sole authz gate.
	router.With(middleware.AuthMiddleware.Authenticate).Delete("/organizations/{id}", m.handleDeleteOrganization)

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
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "use the registration/approval flow to create organizations"})
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

	userIDStr, ok := r.Context().Value(contextkeys.UserIDKey).(string)
	if !ok || userIDStr == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid user id"})
		return
	}

	orgs, err := m.orgRepo.List(r.Context(), userID, 100, 0)
	if err != nil {
		logger.Error("Failed to list organizations", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to list organizations"})
		return
	}
	if orgs == nil {
		orgs = []*orgentities.Organization{}
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"organizations": orgs})
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
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid organization id"})
		return
	}

	org, err := m.orgRepo.GetByID(r.Context(), id)
	if err != nil {
		logger.Error("Failed to get organization", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to get organization"})
		return
	}
	if org == nil || org.IsDeleted() {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "organization not found"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(org)
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
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid organization id"})
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "name is required"})
		return
	}

	org, err := m.orgRepo.GetByID(r.Context(), id)
	if err != nil || org == nil || org.IsDeleted() {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "organization not found"})
		return
	}

	org.Name = req.Name
	org.UpdatedAt = time.Now()
	if err := m.orgRepo.Update(r.Context(), org); err != nil {
		logger.Error("Failed to update organization", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to update organization"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(org)
}

// handleDeleteOrganization cascades the full organization deletion as SUPER_ADMIN.
// @Summary Delete Organization (SUPER_ADMIN)
// @Description Full cascade deletion of an organization by SUPER_ADMIN. Requires SUPER_ADMIN role.
// @Tags organizations
// @Security BearerAuth
// @Produce json
// @Param id path string true "Organization ID"
// @Success 204
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /organizations/{id} [delete]
func (m *Module) handleDeleteOrganization(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", contentTypeJSON)

	// ── SUPER_ADMIN authz guard ─────────────────────────────────────────────
	// Must come BEFORE any other work (DAO-12).
	roles, _ := r.Context().Value(contextkeys.UserRolesKey).([]string)
	isSuperAdmin := false
	for _, role := range roles {
		if role == "SUPER_ADMIN" {
			isSuperAdmin = true
			break
		}
	}
	if !isSuperAdmin {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"code":    "forbidden",
			"message": "super_admin role required",
		})
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid organization id"})
		return
	}

	// ── Delegate to the cascade use case when wired ─────────────────────────
	if m.deleteOrgHandler != nil {
		actorIDStr, _ := r.Context().Value(contextkeys.UserIDKey).(string)
		actorID, _ := uuid.Parse(actorIDStr)

		_, usecaseErr := m.deleteOrgHandler.Handle(r.Context(), &deleteorganization.Request{
			OrgID:     id,
			ActorID:   actorID,
			ActorType: "super_admin",
		})
		if usecaseErr != nil {
			switch {
			case errors.Is(usecaseErr, orgservices.ErrOrgNotFound):
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"code":    "org_not_found",
					"message": "organization not found",
				})
			case errors.Is(usecaseErr, orgservices.ErrBillingActive):
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"code":    "billing_active",
					"message": "organization has an active paid subscription that could not be cancelled",
				})
			default:
				logger.Error("delete_organization: unexpected error",
					zap.String("org_id", id.String()),
					zap.Error(usecaseErr),
				)
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"code":    "internal_error",
					"message": "failed to delete organization",
				})
			}
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// ── Legacy soft-delete fallback (before cascade use case is wired) ─────
	if err := m.orgRepo.Delete(r.Context(), id); err != nil {
		logger.Error("Failed to delete organization", zap.Error(err))
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "organization not found"})
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
		_ = json.NewEncoder(w).Encode(map[string]string{
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
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get organization",
		})
		return
	}

	if response == nil {
		w.Header().Set("Content-Type", contentTypeJSON)
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Organization not found",
		})
		return
	}

	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
