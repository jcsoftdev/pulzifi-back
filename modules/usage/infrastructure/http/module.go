package http

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	assignplan "github.com/jcsoftdev/pulzifi-back/modules/usage/application/assign_plan"
	getmetrics "github.com/jcsoftdev/pulzifi-back/modules/usage/application/get_metrics"
	getquotas "github.com/jcsoftdev/pulzifi-back/modules/usage/application/get_quotas"
	giftmonth "github.com/jcsoftdev/pulzifi-back/modules/usage/application/gift_month"
	listorgs "github.com/jcsoftdev/pulzifi-back/modules/usage/application/list_organizations_with_plans"
	listplans "github.com/jcsoftdev/pulzifi-back/modules/usage/application/list_plans"
	trialstatus "github.com/jcsoftdev/pulzifi-back/modules/usage/application/trial_status"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/services"
	usagepersistence "github.com/jcsoftdev/pulzifi-back/modules/usage/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/shared/contextkeys"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
)

// ModuleDeps carries all dependencies for the Usage module.
type ModuleDeps struct {
	DB                 *sql.DB
	GetMetricsHandler  *getmetrics.Handler
	GetQuotasHandler   *getquotas.Handler
	ListPlansHandler   *listplans.Handler
	ListOrgsHandler    *listorgs.Handler
	AssignPlanHandler  *assignplan.Handler
	GiftMonthHandler   *giftmonth.Handler
	TrialStatusHandler *trialstatus.Handler
}

// Module implements the router.ModuleRegisterer interface for the Usage module.
type Module struct {
	deps ModuleDeps
}

// NewModule creates a new Usage module.
func NewModule(deps ModuleDeps) router.ModuleRegisterer {
	return &Module{deps: deps}
}

// NewModuleWithDB creates a new Usage module (backward-compat shim, sets DB only).
// Callers should prefer NewModule with full deps for production; this path
// keeps cmd/server/modules.go compilable during the transition.
func NewModuleWithDB(db *sql.DB) router.ModuleRegisterer {
	return &Module{deps: ModuleDeps{DB: db}}
}

// NewModuleWithTrial creates a Usage module with the trial-status handler
// wired in. The other use cases stay nil-defaulted (handlers fall back to
// the existing mock responses) so this constructor is safe to drop in.
func NewModuleWithTrial(db *sql.DB, trialStatusHandler *trialstatus.Handler) router.ModuleRegisterer {
	return &Module{deps: ModuleDeps{DB: db, TrialStatusHandler: trialStatusHandler}}
}

// ModuleName returns the name of the module.
func (m *Module) ModuleName() string {
	return "Usage"
}

// RegisterHTTPRoutes registers all HTTP routes for the Usage module.
func (m *Module) RegisterHTTPRoutes(r chi.Router) {
	r.Route("/usage", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware.Authenticate)
		r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)
		r.Get("/metrics", m.handleGetMetrics)
		r.Get("/quotas", m.handleGetQuotas)
		r.Get("/trial-status", m.handleTrialStatus)
	})

	r.Route("/usage/admin", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware.Authenticate)
		r.Get("/plans", m.handleListPlans)
		r.Get("/organizations", m.handleListOrganizationsWithPlans)
		r.Put("/organizations/{id}/plan", m.handleAssignOrganizationPlan)
		r.Post("/organizations/{id}/gift-month", m.handleGiftMonth)
	})
}

func isSuperAdmin(r *http.Request) bool {
	roles, ok := r.Context().Value(contextkeys.UserRolesKey).([]string)
	if !ok {
		return false
	}
	for _, role := range roles {
		if role == "SUPER_ADMIN" {
			return true
		}
	}
	return false
}

func forbidden(w http.ResponseWriter) {
	writeJSON(w, http.StatusForbidden, map[string]string{"error": "super admin required"})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func (m *Module) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	if m.deps.GetMetricsHandler == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"metrics": map[string]interface{}{}})
		return
	}
	tenant := middleware.GetTenantFromContext(r.Context())
	if tenant == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "tenant not found"})
		return
	}
	resp, err := m.deps.GetMetricsHandler.Handle(r.Context(), &getmetrics.Request{Tenant: tenant})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load metrics"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"metrics": resp})
}

func (m *Module) handleGetQuotas(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.GetTenantFromContext(r.Context())
	if tenant == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "tenant not found"})
		return
	}

	// Build a tenant-scoped GetQuotas handler on demand. The repositories
	// touch tenant schemas, so they cannot be wired once at startup the way
	// a single-tenant module would; constructing per-request keeps the
	// dependency graph simple until the usage module is fully hexagonalised.
	handler := m.deps.GetQuotasHandler
	if handler == nil {
		if m.deps.DB == nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "usage module not configured"})
			return
		}
		usageRepo := usagepersistence.NewUsageTrackingPostgresRepository(m.deps.DB, tenant)
		orgPlanRepo := usagepersistence.NewOrganizationPlanPostgresRepository(m.deps.DB)
		handler = getquotas.NewHandler(usageRepo, orgPlanRepo, services.New())
	}

	resp, err := handler.Handle(r.Context(), &getquotas.Request{Tenant: tenant})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load quotas"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"quotas": resp, "message": resp.Message})
}

func (m *Module) handleListPlans(w http.ResponseWriter, r *http.Request) {
	if !isSuperAdmin(r) {
		forbidden(w)
		return
	}
	if m.deps.ListPlansHandler == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "handler not initialized"})
		return
	}
	resp, err := m.deps.ListPlansHandler.Handle(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list plans"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"plans": resp.Plans})
}

func (m *Module) handleListOrganizationsWithPlans(w http.ResponseWriter, r *http.Request) {
	if !isSuperAdmin(r) {
		forbidden(w)
		return
	}
	if m.deps.ListOrgsHandler == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "handler not initialized"})
		return
	}
	resp, err := m.deps.ListOrgsHandler.Handle(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list organizations"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"organizations": resp.Organizations})
}

func (m *Module) handleAssignOrganizationPlan(w http.ResponseWriter, r *http.Request) {
	if !isSuperAdmin(r) {
		forbidden(w)
		return
	}
	if m.deps.AssignPlanHandler == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "handler not initialized"})
		return
	}
	orgID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid organization id"})
		return
	}
	var body struct {
		PlanCode string `json:"plan_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.PlanCode == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "plan_code is required"})
		return
	}
	userID, _ := r.Context().Value(contextkeys.UserIDKey).(string)
	resp, err := m.deps.AssignPlanHandler.Handle(r.Context(), &assignplan.Request{
		OrgID: orgID, PlanCode: body.PlanCode, ActorUserID: userID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"organization_id":        resp.OrganizationID,
		"plan_code":              resp.PlanCode,
		"checks_allowed_monthly": resp.ChecksAllowedMonthly,
	})
}

func (m *Module) handleTrialStatus(w http.ResponseWriter, r *http.Request) {
	if m.deps.TrialStatusHandler == nil {
		// Backwards-compat path: trial isn't wired in yet, surface a benign default.
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"is_trial":       false,
			"is_expired":     false,
			"needs_upgrade":  false,
			"converted":      false,
			"days_remaining": 0,
		})
		return
	}
	subdomain := middleware.GetSubdomainFromContext(r.Context())
	if subdomain == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "tenant required"})
		return
	}
	resp, err := m.deps.TrialStatusHandler.Handle(r.Context(), &trialstatus.Request{Subdomain: subdomain})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load trial status"})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (m *Module) handleGiftMonth(w http.ResponseWriter, r *http.Request) {
	if !isSuperAdmin(r) {
		forbidden(w)
		return
	}
	if m.deps.GiftMonthHandler == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "handler not initialized"})
		return
	}
	orgID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid organization id"})
		return
	}
	resp, err := m.deps.GiftMonthHandler.Handle(r.Context(), &giftmonth.Request{OrgID: orgID})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"organization_id": resp.OrganizationID,
		"gifted_period":   resp.GiftedPeriod,
		"message":         resp.Message,
	})
}
