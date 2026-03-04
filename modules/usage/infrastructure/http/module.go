package http

import (
	"context"
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
		r.Post("/organizations/{id}/gift-month", m.handleGiftMonth)
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

// quotaPeriod holds the billing period boundaries and quota info.
type quotaPeriod struct {
	periodStart      time.Time
	periodEnd        time.Time
	checksAllowed    int
	checksUsed       int
	nextRefillAt     sql.NullTime
	storagePeriodDays int
}

// billingPeriodForDate returns the period_start and period_end for the given date,
// anchored to the day-of-month when the plan started.
// E.g. if anchorDay=15 and today is Mar 20 → period is Mar 15 – Apr 14.
// If anchorDay=15 and today is Mar 10 → period is Feb 15 – Mar 14.
// Handles months with fewer days by clamping to end-of-month.
func billingPeriodForDate(today time.Time, anchorDay int) (start, end time.Time) {
	year, month, day := today.Date()

	// Clamp anchor to last day of current month
	lastDay := daysInMonth(year, month)
	clampedAnchor := anchorDay
	if clampedAnchor > lastDay {
		clampedAnchor = lastDay
	}

	if day >= clampedAnchor {
		// We're in the period that started this month
		start = time.Date(year, month, clampedAnchor, 0, 0, 0, 0, time.UTC)
	} else {
		// We're in the period that started last month
		prevMonth := month - 1
		prevYear := year
		if prevMonth < 1 {
			prevMonth = 12
			prevYear--
		}
		prevLastDay := daysInMonth(prevYear, prevMonth)
		prevAnchor := anchorDay
		if prevAnchor > prevLastDay {
			prevAnchor = prevLastDay
		}
		start = time.Date(prevYear, prevMonth, prevAnchor, 0, 0, 0, 0, time.UTC)
	}

	// End is one day before the next period starts
	nextMonth := start.Month() + 1
	nextYear := start.Year()
	if nextMonth > 12 {
		nextMonth = 1
		nextYear++
	}
	nextLastDay := daysInMonth(nextYear, nextMonth)
	nextAnchor := anchorDay
	if nextAnchor > nextLastDay {
		nextAnchor = nextLastDay
	}
	end = time.Date(nextYear, nextMonth, nextAnchor, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)

	return start, end
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// ensureCurrentPeriod returns the active billing period for the tenant, creating one if it doesn't exist.
// The billing cycle is anchored to the day-of-month from organization_plans.started_at.
func (m *Module) ensureCurrentPeriod(ctx context.Context, tenant string) (*quotaPeriod, error) {
	now := time.Now()

	// Try to find an existing period covering today
	q := `
		SELECT period_start, period_end, checks_allowed, checks_used, next_refill_at
		FROM usage_tracking
		WHERE period_start <= $1::date AND period_end >= $1::date
		ORDER BY period_end DESC
		LIMIT 1
	`
	var qp quotaPeriod
	err := m.db.QueryRowContext(ctx, q, now).Scan(
		&qp.periodStart, &qp.periodEnd, &qp.checksAllowed, &qp.checksUsed, &qp.nextRefillAt,
	)
	if err == nil {
		// Fetch storage_period_days
		qp.storagePeriodDays = m.fetchStoragePeriodDays(ctx, tenant)
		return &qp, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("query usage_tracking: %w", err)
	}

	// No active period — look up the org's plan to create one
	planQuery := `
		SELECT p.checks_allowed_monthly, COALESCE(p.storage_period_days, 7), op.started_at
		FROM public.organizations o
		JOIN public.organization_plans op ON op.organization_id = o.id
			AND op.status = 'active' AND op.deleted_at IS NULL
		JOIN public.plans p ON p.id = op.plan_id
		WHERE o.schema_name = $1
		ORDER BY op.started_at DESC
		LIMIT 1
	`
	var checksAllowed, storageDays int
	var startedAt time.Time
	if err := m.db.QueryRowContext(ctx, planQuery, tenant).Scan(&checksAllowed, &storageDays, &startedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active plan found for tenant %s", tenant)
		}
		return nil, fmt.Errorf("query plan: %w", err)
	}

	anchorDay := startedAt.Day()
	periodStart, periodEnd := billingPeriodForDate(now, anchorDay)
	nextRefill := periodEnd.AddDate(0, 0, 1)

	insertQ := `
		INSERT INTO usage_tracking (period_start, period_end, checks_allowed, checks_used, last_refill_at, next_refill_at, created_at, updated_at)
		VALUES ($1, $2, $3, 0, NOW(), $4, NOW(), NOW())
		ON CONFLICT DO NOTHING
	`
	if _, err := m.db.ExecContext(ctx, insertQ, periodStart, periodEnd, checksAllowed, nextRefill); err != nil {
		return nil, fmt.Errorf("insert usage_tracking: %w", err)
	}

	return &quotaPeriod{
		periodStart:       periodStart,
		periodEnd:         periodEnd,
		checksAllowed:     checksAllowed,
		checksUsed:        0,
		nextRefillAt:      sql.NullTime{Time: nextRefill, Valid: true},
		storagePeriodDays: storageDays,
	}, nil
}

func (m *Module) fetchStoragePeriodDays(ctx context.Context, tenant string) int {
	storagePeriodDays := 7
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
	if err := m.db.QueryRowContext(ctx, spQuery, tenant).Scan(&storagePeriodDays); err != nil && err != sql.ErrNoRows {
		logger.Warn("Failed to fetch storage_period_days, using default", zap.Error(err), zap.String("tenant", tenant))
	}
	return storagePeriodDays
}

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

// handleGiftMonth grants an organization a free month of their current plan.
// It creates a new usage_tracking row for the next billing period with checks_used=0.
// @Summary Gift Free Month
// @Description Grant an organization a free month of their current plan (SUPER_ADMIN only)
// @Tags usage
// @Security BearerAuth
// @Param id path string true "Organization ID"
// @Success 200 {object} map[string]interface{}
// @Router /usage/admin/organizations/{id}/gift-month [post]
func (m *Module) handleGiftMonth(w http.ResponseWriter, r *http.Request) {
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

	// Look up org schema and active plan
	planQuery := `
		SELECT o.schema_name, p.checks_allowed_monthly, op.started_at
		FROM public.organizations o
		JOIN public.organization_plans op ON op.organization_id = o.id
			AND op.status = 'active' AND op.deleted_at IS NULL
		JOIN public.plans p ON p.id = op.plan_id
		WHERE o.id = $1
		ORDER BY op.started_at DESC
		LIMIT 1
	`
	var schemaName string
	var checksAllowed int
	var startedAt time.Time
	if err := m.db.QueryRowContext(r.Context(), planQuery, orgID).Scan(&schemaName, &checksAllowed, &startedAt); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "organization has no active plan", http.StatusBadRequest)
			return
		}
		logger.Error("Failed to query org plan for gift month", zap.Error(err))
		http.Error(w, "failed to look up plan", http.StatusInternalServerError)
		return
	}

	if !schemaNameRegex.MatchString(schemaName) {
		http.Error(w, "invalid organization schema", http.StatusBadRequest)
		return
	}

	// Set search path to tenant schema
	if _, err := m.db.ExecContext(r.Context(), middleware.GetSetSearchPathSQL(schemaName)); err != nil {
		logger.Error("Failed to set search_path for gift month", zap.Error(err))
		http.Error(w, "failed to access organization data", http.StatusInternalServerError)
		return
	}

	anchorDay := startedAt.Day()
	now := time.Now()

	// Ensure current period exists first
	currentStart, currentEnd := billingPeriodForDate(now, anchorDay)
	_, err = m.db.ExecContext(r.Context(), `
		INSERT INTO usage_tracking (period_start, period_end, checks_allowed, checks_used, last_refill_at, next_refill_at, created_at, updated_at)
		VALUES ($1, $2, $3, 0, NOW(), $4, NOW(), NOW())
		ON CONFLICT DO NOTHING
	`, currentStart, currentEnd, checksAllowed, currentEnd.AddDate(0, 0, 1))
	if err != nil {
		logger.Warn("Failed to ensure current period for gift month", zap.Error(err))
	}

	// Calculate the NEXT billing period after the latest existing one
	// Find the latest period end for this tenant
	var latestEnd time.Time
	err = m.db.QueryRowContext(r.Context(), `
		SELECT period_end FROM usage_tracking ORDER BY period_end DESC LIMIT 1
	`).Scan(&latestEnd)
	if err != nil {
		logger.Error("Failed to find latest period", zap.Error(err))
		http.Error(w, "failed to determine billing period", http.StatusInternalServerError)
		return
	}

	// Next period starts the day after the latest period ends
	nextPeriodAnchor := latestEnd.AddDate(0, 0, 1)
	giftStart, giftEnd := billingPeriodForDate(nextPeriodAnchor, anchorDay)
	giftNextRefill := giftEnd.AddDate(0, 0, 1)

	// Insert the gifted period
	insertQ := `
		INSERT INTO usage_tracking (period_start, period_end, checks_allowed, checks_used, last_refill_at, next_refill_at, created_at, updated_at)
		VALUES ($1, $2, $3, 0, NOW(), $4, NOW(), NOW())
	`
	result, err := m.db.ExecContext(r.Context(), insertQ, giftStart, giftEnd, checksAllowed, giftNextRefill)
	if err != nil {
		logger.Error("Failed to insert gifted month", zap.Error(err), zap.String("schema", schemaName))
		http.Error(w, "failed to gift month", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "gifted period already exists", http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"organization_id": orgID,
		"gifted_period": map[string]interface{}{
			"period_start":    giftStart.Format("2006-01-02"),
			"period_end":      giftEnd.Format("2006-01-02"),
			"checks_allowed":  checksAllowed,
			"next_refill_at":  giftNextRefill.Format(time.RFC3339),
		},
		"message": "free month gifted successfully",
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
	if qp, err := m.ensureCurrentPeriod(r.Context(), tenant); err == nil {
		checksUsed = qp.checksUsed
		checksAllowed = qp.checksAllowed
	}

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

	qp, err := m.ensureCurrentPeriod(r.Context(), tenant)
	if err != nil {
		logger.Error("Failed to ensure current billing period", zap.Error(err), zap.String("tenant", tenant))
		http.Error(w, "failed to load quotas", http.StatusInternalServerError)
		return
	}

	var refill interface{}
	if qp.nextRefillAt.Valid {
		refill = qp.nextRefillAt.Time.UTC().Format(time.RFC3339)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"quotas": map[string]interface{}{
			"checks_used":        qp.checksUsed,
			"checks_allowed":     qp.checksAllowed,
			"next_refill_at":     refill,
			"storage_period_days": qp.storagePeriodDays,
		},
		"message": "get usage quotas",
	})
}
