// Package http provides the Chi HTTP module for the billing feature.
// It implements router.ModuleRegisterer so cmd/server/modules.go can register
// it alongside the other 14 modules.
//
// # Raw-body investigation result
//
// No global JSON body-parsing middleware exists in this project. main.go applies
// only CORS, rate limiting, and tenant/logging middlewares. Body parsing happens
// per-handler via json.NewDecoder(r.Body). The Stripe webhook handler reads r.Body
// directly as raw bytes — no sub-router isolation is required. The route can live
// on the standard /billing sub-router and still receive an untouched body.
package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	cancelsubscription "github.com/jcsoftdev/pulzifi-back/modules/billing/application/cancel_subscription"
	createcheckoutsession "github.com/jcsoftdev/pulzifi-back/modules/billing/application/create_checkout_session"
	createportalsession "github.com/jcsoftdev/pulzifi-back/modules/billing/application/create_portal_session"
	getsubscription "github.com/jcsoftdev/pulzifi-back/modules/billing/application/get_subscription"
	giftmonth "github.com/jcsoftdev/pulzifi-back/modules/billing/application/gift_month"
	handlewebhook "github.com/jcsoftdev/pulzifi-back/modules/billing/application/handle_webhook"
	listplans "github.com/jcsoftdev/pulzifi-back/modules/billing/application/list_plans"
	managecoupons "github.com/jcsoftdev/pulzifi-back/modules/billing/application/manage_coupons"
	updatesubscription "github.com/jcsoftdev/pulzifi-back/modules/billing/application/update_subscription"
	"github.com/jcsoftdev/pulzifi-back/shared/contextkeys"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"go.uber.org/zap"
)

// Compile-time interface check.
var _ router.ModuleRegisterer = (*Module)(nil)

// billingSettingsPath is the canonical frontend route for the billing tab.
// Stripe checkout/portal return URLs are built dynamically from r.Host + this
// path so they respect the tenant subdomain the request came from.
const billingSettingsPath = "/settings/billing"

// buildTenantBillingURL constructs a fully qualified URL pointing to the
// billing tab on the SAME subdomain the request came from. Falls back to the
// raw r.Host when X-Forwarded-Host is absent. Query is appended as-is (must
// already start with "?" if non-empty).
func buildTenantBillingURL(r *http.Request, query string) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	return fmt.Sprintf("%s://%s%s%s", scheme, host, billingSettingsPath, query)
}

// Deps holds all external dependencies wired by cmd/server/modules.go.
type Deps struct {
	DB *sql.DB

	CheckoutHandler     *createcheckoutsession.Handler
	PortalHandler       *createportalsession.Handler
	SubscriptionHandler *getsubscription.Handler
	WebhookHandler      *handlewebhook.Handler
	UpdateSubHandler    *updatesubscription.Handler
	CouponHandler       *managecoupons.Handler
	GiftHandler         *giftmonth.Handler
	CancelHandler       *cancelsubscription.Handler
	ListPlansHandler    *listplans.Handler
}

// Module is the Billing HTTP module.
type Module struct{ deps Deps }

// NewModule constructs the module with fully wired dependencies.
func NewModule(deps Deps) *Module { return &Module{deps: deps} }

// ModuleName satisfies router.ModuleRegisterer.
func (m *Module) ModuleName() string { return "Billing" }

// RegisterHTTPRoutes mounts all billing routes onto the provided Chi router.
//
// Route overview:
//
//	POST /billing/checkout     — ADMIN/SUPER_ADMIN only (NFR3), creates Checkout session
//	POST /billing/portal       — ADMIN/SUPER_ADMIN only (NFR3), creates Portal session
//	GET  /billing/subscription — any authenticated org member, returns subscription state
//	POST /billing/webhook      — NO auth, raw body, Stripe calls this endpoint
//
// Note on webhook raw body: no global JSON body parser exists in this project so
// r.Body arrives untouched for every route. The webhook handler reads io.ReadAll(r.Body)
// directly, which is correct for stripe-go's ConstructEvent call.
func (m *Module) RegisterHTTPRoutes(r chi.Router) {
	r.Route("/billing", func(r chi.Router) {
		// Public route — Stripe webhook must be unauthenticated and receive raw bytes.
		r.Post("/webhook", m.handleWebhook)
		r.Get("/plans", m.handleListPlans)

		// Authenticated routes — require valid JWT + org membership.
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware.Authenticate)
			r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)

			// FR7: any org member may read their subscription status.
			r.Get("/subscription", m.handleGetSubscription)

			// NFR3: checkout and portal creation require ADMIN or SUPER_ADMIN role.
			// We allow both via requireAdminOrSuperAdmin to match spec semantics
			// ("owner or admin" maps to ADMIN/SUPER_ADMIN in this system's JWT roles).
			r.Group(func(r chi.Router) {
				r.Use(m.requireAdminOrSuperAdmin())
				r.Post("/checkout", m.handleCheckout)
				r.Post("/portal", m.handlePortal)
				// In-place plan change (upgrade/downgrade with proration) —
				// mutates the existing Stripe Subscription without sending the
				// user to Stripe's hosted UI. Preview=true returns the prorated
				// charge without applying.
				r.Post("/subscription", m.handleUpdateSubscription)
				// Cancel at period end (user keeps access until the paid period
				// closes, then the plan deactivates) + resume a pending cancel.
				r.Post("/subscription/cancel", m.handleCancelSubscription)
				r.Post("/subscription/resume", m.handleResumeSubscription)
			})

			// Coupon management + gifts — SUPER_ADMIN only.
			r.Group(func(r chi.Router) {
				r.Use(m.requireSuperAdmin())
				r.Post("/admin/coupons", m.handleCreateCoupon)
				r.Get("/admin/coupons", m.handleListCoupons)
				r.Delete("/admin/coupons/{id}", m.handleRevokeCoupon)
				// Gift = 100%-off-once coupon attached to the org's active sub.
				r.Post("/admin/gift", m.handleGift)
			})
		})
	})
}

// requireSuperAdmin allows only SUPER_ADMIN JWT role. Coupon management is a
// platform-operator concern, not a per-org-admin one.
func (m *Module) requireSuperAdmin() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roles, _ := r.Context().Value(contextkeys.UserRolesKey).([]string)
			for _, role := range roles {
				if role == "SUPER_ADMIN" {
					next.ServeHTTP(w, r)
					return
				}
			}
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "super admin required"})
		})
	}
}

// requireAdminOrSuperAdmin returns a Chi middleware that allows requests only from
// users whose JWT carries role ADMIN or SUPER_ADMIN.
//
// This satisfies NFR3: "POST /billing/checkout and POST /billing/portal MUST require
// owner or admin role." In this system, org-level admins are represented by the ADMIN
// JWT role; SUPER_ADMIN always has elevated access.
//
// Role checking reads JWT claims from context (set by the global middleware.AuthMiddleware
// singleton). When auth middleware is not applied (e.g. unit tests that bypass auth),
// the roles slice will be empty and the middleware returns 403 — callers must inject
// roles into context manually.
func (m *Module) requireAdminOrSuperAdmin() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roles, _ := r.Context().Value(contextkeys.UserRolesKey).([]string)
			for _, role := range roles {
				if role == "ADMIN" || role == "SUPER_ADMIN" {
					next.ServeHTTP(w, r)
					return
				}
			}
			// Neither ADMIN nor SUPER_ADMIN — 403 Forbidden.
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "insufficient role"})
		})
	}
}

// ── Handlers ──────────────────────────────────────────────────────────────────

// handleWebhook handles POST /billing/webhook.
//
// It reads the raw body bytes — mandatory for stripe-go's webhook.ConstructEvent
// which validates the HMAC signature over the exact bytes Stripe sent.
// No authentication is applied; Stripe's signature serves as the auth mechanism.
func (m *Module) handleWebhook(w http.ResponseWriter, r *http.Request) {
	sigHeader := r.Header.Get("Stripe-Signature")
	if sigHeader == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing Stripe-Signature header"})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("billing: failed to read webhook body", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to read body"})
		return
	}

	if err := m.deps.WebhookHandler.Handle(r.Context(), body, sigHeader); err != nil {
		if errors.Is(err, handlewebhook.ErrInvalidSignature) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid webhook signature"})
			return
		}
		// Orphan customer is no longer surfaced here — the application layer
		// stores the event as 'deferred' with full payload and returns nil so
		// Stripe receives a clean 200. Reconciliation runs when the org is
		// later linked to the Stripe customer.
		logger.Error("billing: webhook handler error", zap.Error(err))
		// Return 500 so Stripe retries — only for non-signature errors.
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "webhook processing failed"})
		return
	}

	w.WriteHeader(http.StatusOK)
}

// handleCheckout handles POST /billing/checkout.
// Expects JSON body: {"plan_id": "<plan code>", "billing_cycle": "monthly|yearly"}.
// Stripe price IDs are resolved server-side from public.plans by the use case.
func (m *Module) handleCheckout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body struct {
		PlanID        string `json:"plan_id"`
		BillingCycle  string `json:"billing_cycle"`
		PromotionCode string `json:"promotion_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	_ = ctx.Value(contextkeys.UserIDKey) // userID available for future audit logging
	orgID, orgEmail, orgName, err := m.resolveOrgContext(ctx, middleware.GetSubdomainFromContext(ctx))
	if err != nil {
		logger.Error("billing: failed to resolve org context", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "org lookup failed"})
		return
	}

	req := createcheckoutsession.Request{
		OrgID:         orgID,
		OrgEmail:      orgEmail,
		OrgName:       orgName,
		PlanID:        body.PlanID,
		BillingCycle:  body.BillingCycle,
		PromotionCode: body.PromotionCode,
		SuccessURL:    buildTenantBillingURL(r, "?success=true&session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:     buildTenantBillingURL(r, "?canceled=true"),
	}

	resp, err := m.deps.CheckoutHandler.Handle(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, createcheckoutsession.ErrInvalidBillingCycle),
			errors.Is(err, createcheckoutsession.ErrMissingPriceID):
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
			return
		case errors.Is(err, createcheckoutsession.ErrPlanNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		logger.Error("billing: checkout handler error", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "checkout failed"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"checkout_url": resp.CheckoutURL})
}

// handlePortal handles POST /billing/portal.
// Creates a Stripe Customer Portal session for the authenticated org.
func (m *Module) handlePortal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, _, _, err := m.resolveOrgContext(ctx, middleware.GetSubdomainFromContext(ctx))
	if err != nil {
		logger.Error("billing: failed to resolve org context", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "org lookup failed"})
		return
	}

	req := createportalsession.Request{
		OrgID:     orgID,
		ReturnURL: buildTenantBillingURL(r, ""),
	}

	resp, err := m.deps.PortalHandler.Handle(ctx, req)
	if err != nil {
		if errors.Is(err, createportalsession.ErrNoStripeCustomer) {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
			return
		}
		logger.Error("billing: portal handler error", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "portal session failed"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"portal_url": resp.PortalURL})
}

// handleUpdateSubscription handles POST /billing/subscription.
// Body: {"plan_id": "pro", "billing_cycle": "monthly", "preview": false}.
// Returns: {"subscription_id", "new_price_id", "billing_cycle", "prorated_amount_cents", "currency", "preview"}.
//
// Drives the in-place upgrade/downgrade flow: when preview=true the response
// carries the prorated amount Stripe would charge if applied; preview=false
// applies the change immediately (always_invoice proration) and the local
// org_plans row is updated synchronously.
func (m *Module) handleUpdateSubscription(w http.ResponseWriter, r *http.Request) {
	if m.deps.UpdateSubHandler == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "in-place upgrade not configured"})
		return
	}
	ctx := r.Context()
	var body struct {
		PlanID       string `json:"plan_id"`
		BillingCycle string `json:"billing_cycle"`
		Preview      bool   `json:"preview"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	orgID, _, _, err := m.resolveOrgContext(ctx, middleware.GetSubdomainFromContext(ctx))
	if err != nil {
		logger.Error("billing: failed to resolve org context", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "org lookup failed"})
		return
	}

	resp, err := m.deps.UpdateSubHandler.Handle(ctx, updatesubscription.Request{
		OrgID:        orgID,
		PlanID:       body.PlanID,
		BillingCycle: body.BillingCycle,
		Preview:      body.Preview,
	})
	if err != nil {
		switch {
		case errors.Is(err, updatesubscription.ErrInvalidBillingCycle),
			errors.Is(err, updatesubscription.ErrMissingPriceID):
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
			return
		case errors.Is(err, updatesubscription.ErrPlanNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		case errors.Is(err, updatesubscription.ErrNoActiveSubscription):
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		logger.Error("billing: update subscription handler error", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "update subscription failed"})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleCancelSubscription handles POST /billing/subscription/cancel.
// Schedules cancellation at the end of the current paid period.
func (m *Module) handleCancelSubscription(w http.ResponseWriter, r *http.Request) {
	if m.deps.CancelHandler == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "cancellation not configured"})
		return
	}
	ctx := r.Context()
	orgID, err := m.resolveOrgUUID(ctx)
	if err != nil {
		logger.Error("billing: failed to resolve org context", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "org lookup failed"})
		return
	}

	resp, err := m.deps.CancelHandler.Cancel(ctx, orgID)
	if err != nil {
		if errors.Is(err, cancelsubscription.ErrNoActiveSubscription) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		logger.Error("billing: cancel subscription failed", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "cancel subscription failed"})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleResumeSubscription handles POST /billing/subscription/resume.
// Clears a pending period-end cancellation.
func (m *Module) handleResumeSubscription(w http.ResponseWriter, r *http.Request) {
	if m.deps.CancelHandler == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "cancellation not configured"})
		return
	}
	ctx := r.Context()
	orgID, err := m.resolveOrgUUID(ctx)
	if err != nil {
		logger.Error("billing: failed to resolve org context", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "org lookup failed"})
		return
	}

	resp, err := m.deps.CancelHandler.Resume(ctx, orgID)
	if err != nil {
		if errors.Is(err, cancelsubscription.ErrNoActiveSubscription) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		logger.Error("billing: resume subscription failed", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "resume subscription failed"})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleCreateCoupon handles POST /billing/admin/coupons (SUPER_ADMIN).
// Body: {plan_code, billing_cycle, code?, max_redemptions?, expires_at?}.
func (m *Module) handleCreateCoupon(w http.ResponseWriter, r *http.Request) {
	if m.deps.CouponHandler == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "coupons not configured"})
		return
	}
	var body struct {
		PlanCode       string `json:"plan_code"`
		BillingCycle   string `json:"billing_cycle"`
		Code           string `json:"code"`
		MaxRedemptions int64  `json:"max_redemptions"`
		ExpiresAt      int64  `json:"expires_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	resp, err := m.deps.CouponHandler.Create(r.Context(), managecoupons.CreateRequest{
		PlanCode:       body.PlanCode,
		BillingCycle:   body.BillingCycle,
		Code:           body.Code,
		MaxRedemptions: body.MaxRedemptions,
		ExpiresAt:      body.ExpiresAt,
	})
	if err != nil {
		switch {
		case errors.Is(err, managecoupons.ErrInvalidBillingCycle),
			errors.Is(err, managecoupons.ErrMissingPriceID):
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
			return
		case errors.Is(err, managecoupons.ErrPlanNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		logger.Error("billing: create coupon failed", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "create coupon failed"})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleListCoupons handles GET /billing/admin/coupons (SUPER_ADMIN).
func (m *Module) handleListCoupons(w http.ResponseWriter, r *http.Request) {
	if m.deps.CouponHandler == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "coupons not configured"})
		return
	}
	list, err := m.deps.CouponHandler.List(r.Context())
	if err != nil {
		logger.Error("billing: list coupons failed", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list coupons failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"coupons": list})
}

// handleRevokeCoupon handles DELETE /billing/admin/coupons/{id} (SUPER_ADMIN).
func (m *Module) handleRevokeCoupon(w http.ResponseWriter, r *http.Request) {
	if m.deps.CouponHandler == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "coupons not configured"})
		return
	}
	id := chi.URLParam(r, "id")
	if err := m.deps.CouponHandler.Revoke(r.Context(), id); err != nil {
		logger.Error("billing: revoke coupon failed", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "revoke coupon failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}

// handleGift handles POST /billing/admin/gift (SUPER_ADMIN).
// Body: {"org_id": "<uuid>"}. Applies a 100%-off-once coupon to the org's
// active subscription → next invoice $0. Quota refills via Stripe webhook.
func (m *Module) handleGift(w http.ResponseWriter, r *http.Request) {
	if m.deps.GiftHandler == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "gift not configured"})
		return
	}
	var body struct {
		OrgID    string `json:"org_id"`
		PlanCode string `json:"plan_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	orgID, err := uuid.Parse(body.OrgID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid org_id"})
		return
	}
	if body.PlanCode == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "plan_code required"})
		return
	}
	// Resolve owner email + org name in case the org has no Stripe customer yet
	// (gift to a new / trial org creates the customer + free trial sub).
	orgEmail, orgName := m.resolveOrgIdentity(r.Context(), orgID)
	resp, err := m.deps.GiftHandler.Handle(r.Context(), orgID, body.PlanCode, orgEmail, orgName)
	if err != nil {
		switch {
		case errors.Is(err, giftmonth.ErrNoActiveSubscription):
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		case errors.Is(err, giftmonth.ErrPlanNotFound), errors.Is(err, giftmonth.ErrMissingPriceID):
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
			return
		}
		logger.Error("billing: gift failed", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "gift failed"})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleGetSubscription handles GET /billing/subscription.
// Returns the current subscription state for the authenticated org.
func (m *Module) handleGetSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, _, _, err := m.resolveOrgContext(ctx, middleware.GetSubdomainFromContext(ctx))
	if err != nil {
		logger.Error("billing: failed to resolve org context", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "org lookup failed"})
		return
	}
	m.handleGetSubscriptionWithOrgID(w, r, orgID)
}

// handleGetSubscriptionWithOrgID is the core of GET /billing/subscription.
// It is separated from handleGetSubscription so unit tests can call it without
// a real database (org context resolution is skipped).
//
// FR7: when the org has no Stripe subscription, returns HTTP 200 with null
// stripe fields rather than 404 — callers should use null-check, not 404 handling.
func (m *Module) handleGetSubscriptionWithOrgID(w http.ResponseWriter, r *http.Request, orgID string) {
	ctx := r.Context()

	resp, err := m.deps.SubscriptionHandler.Handle(ctx, orgID)
	if err != nil {
		if errors.Is(err, getsubscription.ErrSubscriptionNotFound) {
			// FR7: return 200 with null stripe fields.
			// The frontend must treat stripe_status: null as "no Stripe subscription".
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"stripe_status":      nil,
				"payment_status":     nil,
				"billing_status":     nil,
				"current_period_end": nil,
			})
			return
		}
		logger.Error("billing: get subscription handler error", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "subscription lookup failed"})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// resolveOrgContext queries the org's ID, email (from owner user), and name from the subdomain.
// Returns (orgID string, email string, name string, err).
// All queries are on the public schema — no tenant schema involved.
func (m *Module) resolveOrgContext(ctx context.Context, subdomain string) (string, string, string, error) {
	var orgID, orgName string
	var ownerEmail sql.NullString

	err := m.deps.DB.QueryRowContext(ctx, `
		SELECT o.id::text, o.name, u.email
		FROM public.organizations o
		LEFT JOIN public.users u ON u.id = o.owner_user_id
		WHERE o.subdomain = $1 AND o.deleted_at IS NULL
		LIMIT 1
	`, subdomain).Scan(&orgID, &orgName, &ownerEmail)
	if err != nil {
		return "", "", "", err
	}

	return orgID, ownerEmail.String, orgName, nil
}

// resolveOrgUUID resolves the authenticated org's UUID from the request
// subdomain. Wraps resolveOrgContext for handlers that need a uuid.UUID.
func (m *Module) resolveOrgUUID(ctx context.Context) (uuid.UUID, error) {
	orgIDStr, _, _, err := m.resolveOrgContext(ctx, middleware.GetSubdomainFromContext(ctx))
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(orgIDStr)
}

// resolveOrgIdentity returns the org's owner email + name by org UUID. Best
// effort: returns empty strings on error (EnsureCustomer tolerates empties).
func (m *Module) resolveOrgIdentity(ctx context.Context, orgID uuid.UUID) (email, name string) {
	var orgName string
	var ownerEmail sql.NullString
	err := m.deps.DB.QueryRowContext(ctx, `
		SELECT o.name, u.email
		FROM public.organizations o
		LEFT JOIN public.users u ON u.id = o.owner_user_id
		WHERE o.id = $1 AND o.deleted_at IS NULL
		LIMIT 1
	`, orgID).Scan(&orgName, &ownerEmail)
	if err != nil {
		return "", ""
	}
	return ownerEmail.String, orgName
}

// handleListPlans handles GET /billing/plans.
// Public — no authentication required. Returns the active plan catalog with
// prices sourced from public.plans (synced from Stripe via webhook).
func (m *Module) handleListPlans(w http.ResponseWriter, r *http.Request) {
	if m.deps.ListPlansHandler == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "plan catalog unavailable"})
		return
	}
	resp, err := m.deps.ListPlansHandler.Handle(r.Context())
	if err != nil {
		logger.Error("billing: list plans error", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load plans"})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}
