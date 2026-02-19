package http

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	authmw "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"go.uber.org/zap"
)

// Module implements the router.ModuleRegisterer interface for the Usage module
type Module struct {
	db *sql.DB
}

// NewModule creates a new instance of the Usage module
func NewModule() router.ModuleRegisterer {
	return &Module{}
}

// NewModuleWithDB creates a new usage module with DB support
func NewModuleWithDB(db *sql.DB) router.ModuleRegisterer {
	return &Module{db: db}
}

// ModuleName returns the name of the module
func (m *Module) ModuleName() string {
	return "Usage"
}

// RegisterHTTPRoutes registers all HTTP routes for the Usage module
func (m *Module) RegisterHTTPRoutes(router chi.Router) {
	router.Route("/usage", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware.Authenticate)
		r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)
		r.Get("/metrics", m.handleGetMetrics)
		r.Get("/quotas", m.handleGetQuotas)
	})

	router.Route("/usage/admin", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware.Authenticate)
		r.Get("/plans", m.handleListPlans)
		r.Get("/organizations", m.handleListOrganizationsWithPlans)
		r.Put("/organizations/{id}/plan", m.handleAssignOrganizationPlan)
	})
}

func isSuperAdmin(r *http.Request) bool {
	roles, ok := r.Context().Value(authmw.UserRolesKey).([]string)
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]string{"error": "super admin required"})
}

var schemaNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func (m *Module) handleListPlans(w http.ResponseWriter, r *http.Request) {
	if !isSuperAdmin(r) {
		forbidden(w)
		return
	}

	if m.db == nil {
		http.Error(w, "database not initialized", http.StatusInternalServerError)
		return
	}

	rows, err := m.db.QueryContext(r.Context(), `
		SELECT id, code, name, description, checks_allowed_monthly, is_active, storage_period_days
		FROM public.plans
		WHERE is_active = TRUE
		ORDER BY checks_allowed_monthly ASC
	`)
	if err != nil {
		logger.Error("Failed to list plans", zap.Error(err))
		http.Error(w, "failed to list plans", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	plans := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id uuid.UUID
		var code, name string
		var description sql.NullString
		var checksAllowed int
		var isActive bool
		var storagePeriodDays int

		if err := rows.Scan(&id, &code, &name, &description, &checksAllowed, &isActive, &storagePeriodDays); err != nil {
			logger.Error("Failed to scan plan row", zap.Error(err))
			http.Error(w, "failed to list plans", http.StatusInternalServerError)
			return
		}

		plans = append(plans, map[string]interface{}{
			"id":                     id,
			"code":                   code,
			"name":                   name,
			"description":            description.String,
			"checks_allowed_monthly": checksAllowed,
			"is_active":              isActive,
			"storage_period_days":    storagePeriodDays,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"plans": plans})
}

func (m *Module) handleListOrganizationsWithPlans(w http.ResponseWriter, r *http.Request) {
	if !isSuperAdmin(r) {
		forbidden(w)
		return
	}

	if m.db == nil {
		http.Error(w, "database not initialized", http.StatusInternalServerError)
		return
	}

	rows, err := m.db.QueryContext(r.Context(), `
		SELECT
			o.id,
			o.name,
			o.subdomain,
			o.schema_name,
			COALESCE(p.code, ''),
			COALESCE(p.name, ''),
			COALESCE(p.checks_allowed_monthly, 0),
			COALESCE(p.storage_period_days, 7)
		FROM public.organizations o
		LEFT JOIN LATERAL (
			SELECT op.plan_id
			FROM public.organization_plans op
			WHERE op.organization_id = o.id
				AND op.deleted_at IS NULL
				AND op.status = 'active'
			ORDER BY op.started_at DESC
			LIMIT 1
		) active_plan ON TRUE
		LEFT JOIN public.plans p ON p.id = active_plan.plan_id
		WHERE o.deleted_at IS NULL
		ORDER BY o.created_at DESC
	`)
	if err != nil {
		logger.Error("Failed to list organizations with plans", zap.Error(err))
		http.Error(w, "failed to list organizations", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	organizations := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id uuid.UUID
		var name, subdomain, schemaName string
		var planCode, planName string
		var checksAllowed int
		var storagePeriodDays int

		if err := rows.Scan(&id, &name, &subdomain, &schemaName, &planCode, &planName, &checksAllowed, &storagePeriodDays); err != nil {
			logger.Error("Failed to scan organization plan row", zap.Error(err))
			http.Error(w, "failed to list organizations", http.StatusInternalServerError)
			return
		}

		organizations = append(organizations, map[string]interface{}{
			"id":                     id,
			"name":                   name,
			"subdomain":              subdomain,
			"schema_name":            schemaName,
			"plan_code":              planCode,
			"plan_name":              planName,
			"checks_allowed_monthly": checksAllowed,
			"storage_period_days":    storagePeriodDays,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"organizations": organizations})
}

func (m *Module) handleAssignOrganizationPlan(w http.ResponseWriter, r *http.Request) {
	if !isSuperAdmin(r) {
		forbidden(w)
		return
	}

	if m.db == nil {
		http.Error(w, "database not initialized", http.StatusInternalServerError)
		return
	}

	orgID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid organization id", http.StatusBadRequest)
		return
	}

	var req struct {
		PlanCode string `json:"plan_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PlanCode == "" {
		http.Error(w, "plan_code is required", http.StatusBadRequest)
		return
	}

	userID, _ := r.Context().Value(authmw.UserIDKey).(string)

	tx, err := m.db.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var planID uuid.UUID
	var checksAllowed int
	if err := tx.QueryRowContext(r.Context(), `
		SELECT id, checks_allowed_monthly
		FROM public.plans
		WHERE code = $1 AND is_active = TRUE
		LIMIT 1
	`, req.PlanCode).Scan(&planID, &checksAllowed); err != nil {
		http.Error(w, "plan not found", http.StatusBadRequest)
		return
	}

	_, err = tx.ExecContext(r.Context(), `
		UPDATE public.organization_plans
		SET status = 'inactive', ended_at = NOW(), updated_at = NOW()
		WHERE organization_id = $1
		  AND status = 'active'
		  AND deleted_at IS NULL
	`, orgID)
	if err != nil {
		http.Error(w, "failed to deactivate current plan", http.StatusInternalServerError)
		return
	}

	_, err = tx.ExecContext(r.Context(), `
		INSERT INTO public.organization_plans (organization_id, plan_id, status, started_at, created_by, created_at, updated_at)
		VALUES ($1, $2, 'active', NOW(), NULLIF($3, '')::uuid, NOW(), NOW())
	`, orgID, planID, userID)
	if err != nil {
		http.Error(w, "failed to assign plan", http.StatusInternalServerError)
		return
	}

	var schemaName string
	if err := tx.QueryRowContext(r.Context(), `SELECT schema_name FROM public.organizations WHERE id = $1`, orgID).Scan(&schemaName); err != nil {
		http.Error(w, "organization not found", http.StatusBadRequest)
		return
	}

	if !schemaNameRegex.MatchString(schemaName) {
		http.Error(w, "invalid organization schema", http.StatusBadRequest)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "failed to commit plan change", http.StatusInternalServerError)
		return
	}

	// Best-effort: sync checks_allowed on any active usage_tracking row.
	// The table may not exist yet for newly provisioned orgs, so failures are non-fatal.
	updateUsageSQL := fmt.Sprintf(`
		UPDATE %s.usage_tracking
		SET checks_allowed = $1, updated_at = NOW()
		WHERE period_start <= CURRENT_DATE AND period_end >= CURRENT_DATE
	`, schemaName)
	if _, err := m.db.ExecContext(r.Context(), updateUsageSQL, checksAllowed); err != nil {
		logger.Warn("Failed to sync tenant usage_tracking after plan change (non-fatal)",
			zap.String("schema", schemaName), zap.Error(err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"organization_id":        orgID,
		"plan_code":              req.PlanCode,
		"checks_allowed_monthly": checksAllowed,
	})
}

// handleGetMetrics returns usage metrics for the current tenant
// @Summary Get Usage Metrics
// @Description Get usage metrics for the current organization (checks, pages, workspaces)
// @Tags usage
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /usage/metrics [get]
func (m *Module) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if m.db == nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"metrics": map[string]interface{}{}})
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	if tenant == "" {
		http.Error(w, "tenant not found", http.StatusBadRequest)
		return
	}

	if _, err := m.db.ExecContext(r.Context(), middleware.GetSetSearchPathSQL(tenant)); err != nil {
		logger.Error("Failed to set tenant search_path for metrics", zap.Error(err))
		http.Error(w, "failed to load metrics", http.StatusInternalServerError)
		return
	}

	var totalChecks, successChecks, failedChecks int
	m.db.QueryRowContext(r.Context(), `SELECT COUNT(*), COUNT(*) FILTER (WHERE status = 'success'), COUNT(*) FILTER (WHERE status = 'error') FROM checks`).Scan(&totalChecks, &successChecks, &failedChecks) //nolint:errcheck

	var totalPages int
	m.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM pages WHERE deleted_at IS NULL`).Scan(&totalPages) //nolint:errcheck

	var totalWorkspaces int
	m.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM workspaces WHERE deleted_at IS NULL`).Scan(&totalWorkspaces) //nolint:errcheck

	var totalAlerts int
	m.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM alerts`).Scan(&totalAlerts) //nolint:errcheck

	var checksUsed, checksAllowed int
	m.db.QueryRowContext(r.Context(), `SELECT COALESCE(checks_used, 0), COALESCE(checks_allowed, 0) FROM usage_tracking WHERE period_start <= $1::date AND period_end >= $1::date ORDER BY period_end DESC LIMIT 1`, time.Now()).Scan(&checksUsed, &checksAllowed) //nolint:errcheck

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"metrics": map[string]interface{}{
			"checks": map[string]interface{}{
				"total":   totalChecks,
				"success": successChecks,
				"failed":  failedChecks,
			},
			"pages":              totalPages,
			"workspaces":         totalWorkspaces,
			"alerts":             totalAlerts,
			"checks_used":        checksUsed,
			"checks_allowed":     checksAllowed,
		},
	})
}

// handleGetQuotas returns usage quotas
// @Summary Get Usage Quotas
// @Description Get usage quotas for the current organization
// @Tags usage
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /usage/quotas [get]
func (m *Module) handleGetQuotas(w http.ResponseWriter, r *http.Request) {
	if m.db == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"quotas": map[string]interface{}{
				"checks_used":    300,
				"checks_allowed": 1000,
				"next_refill_at": "2025-10-20T00:00:00Z",
			},
			"message": "get usage quotas (mock - db not initialized)",
		})
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	if tenant == "" {
		http.Error(w, "tenant not found", http.StatusBadRequest)
		return
	}

	if _, err := m.db.ExecContext(r.Context(), middleware.GetSetSearchPathSQL(tenant)); err != nil {
		logger.Error("Failed to set tenant search_path for quotas", zap.Error(err), zap.String("tenant", tenant))
		http.Error(w, "failed to load quotas", http.StatusInternalServerError)
		return
	}

	q := `
		SELECT checks_used, checks_allowed, next_refill_at
		FROM usage_tracking
		WHERE period_start <= $1::date AND period_end >= $1::date
		ORDER BY period_end DESC
		LIMIT 1
	`

	var checksUsed int
	var checksAllowed int
	var nextRefillAt sql.NullTime

	err := m.db.QueryRowContext(r.Context(), q, time.Now()).Scan(&checksUsed, &checksAllowed, &nextRefillAt)
	if err != nil {
		if err == sql.ErrNoRows {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"quotas": map[string]interface{}{
					"checks_used":        0,
					"checks_allowed":     0,
					"next_refill_at":     nil,
					"storage_period_days": 7,
				},
				"message": "no active quota period",
			})
			return
		}

		logger.Error("Failed to query usage quotas", zap.Error(err), zap.String("tenant", tenant))
		http.Error(w, "failed to load quotas", http.StatusInternalServerError)
		return
	}

	// Fetch storage_period_days from the org's active plan
	storagePeriodDays := 7 // default
	spQuery := `
		SELECT COALESCE(p.storage_period_days, 7)
		FROM public.organizations o
		JOIN public.organization_plans op ON op.organization_id = o.id
			AND op.status = 'active' AND op.deleted_at IS NULL
		JOIN public.plans p ON p.id = op.plan_id
		WHERE o.schema_name = $1
		ORDER BY op.started_at DESC
		LIMIT 1
	`
	if err := m.db.QueryRowContext(r.Context(), spQuery, tenant).Scan(&storagePeriodDays); err != nil && err != sql.ErrNoRows {
		logger.Warn("Failed to fetch storage_period_days, using default", zap.Error(err), zap.String("tenant", tenant))
	}

	var refill interface{}
	if nextRefillAt.Valid {
		refill = nextRefillAt.Time.UTC().Format(time.RFC3339)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"quotas": map[string]interface{}{
			"checks_used":        checksUsed,
			"checks_allowed":     checksAllowed,
			"next_refill_at":     refill,
			"storage_period_days": storagePeriodDays,
		},
		"message": "get usage quotas",
	})
}
