package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	domerrors "github.com/jcsoftdev/pulzifi-back/modules/integration/domain/errors"
	bulkretrydeliveries "github.com/jcsoftdev/pulzifi-back/modules/integration/application/bulk_retry_deliveries"
	connectbyo "github.com/jcsoftdev/pulzifi-back/modules/integration/application/connect_byo"
	createdestination "github.com/jcsoftdev/pulzifi-back/modules/integration/application/create_destination"
	deletedestination "github.com/jcsoftdev/pulzifi-back/modules/integration/application/delete_destination"
	disconnectintegration "github.com/jcsoftdev/pulzifi-back/modules/integration/application/disconnect_integration"
	getdelivery "github.com/jcsoftdev/pulzifi-back/modules/integration/application/get_delivery"
	handleoauthcallback "github.com/jcsoftdev/pulzifi-back/modules/integration/application/handle_oauth_callback"
	listdeliveries "github.com/jcsoftdev/pulzifi-back/modules/integration/application/list_deliveries"
	listdestinations "github.com/jcsoftdev/pulzifi-back/modules/integration/application/list_destinations"
	listprovidertargets "github.com/jcsoftdev/pulzifi-back/modules/integration/application/list_provider_targets"
	retrydelivery "github.com/jcsoftdev/pulzifi-back/modules/integration/application/retry_delivery"
	startoauth "github.com/jcsoftdev/pulzifi-back/modules/integration/application/start_oauth"
	updatedestination "github.com/jcsoftdev/pulzifi-back/modules/integration/application/update_destination"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
	intoauth "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/oauth"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/shared/contextkeys"
	"github.com/jcsoftdev/pulzifi-back/shared/featureflags"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"go.uber.org/zap"
)

// Compile-time interface check.
var _ router.ModuleRegisterer = (*Module)(nil)

// Deps holds all dependencies for the Integration HTTP module.
type Deps struct {
	DB                *sql.DB
	IntRepo           repositories.IntegrationRepository // public-schema, encryption-aware
	Registry          services.ProviderRegistry
	StateSigner       *intoauth.StateSigner
	OAuthRedirectBase string // e.g. "http://localhost:3000"

	// Phase 2 additions:
	Flags           *featureflags.Reader
	Validator       connectbyo.Validator
	PlanLookup      connectbyo.PlanLookup
	TwilioPaidPlans []string
}

// Module is the Integration HTTP module.
type Module struct{ deps Deps }

// NewModule constructs the module with fully wired dependencies.
func NewModule(deps Deps) *Module { return &Module{deps: deps} }

// Deprecated: use NewModule with full Deps. Shim for the brief window before T24 wiring lands.
func NewModuleWithDB(db *sql.DB) router.ModuleRegisterer {
	return &Module{deps: Deps{DB: db}}
}

// ModuleName satisfies router.ModuleRegisterer.
func (m *Module) ModuleName() string { return "Integration" }

// RegisterHTTPRoutes registers tenant-scoped routes under the /api/v1 prefix.
// The OAuth callback route is registered separately at root via HandleOAuthCallback
// (not under tenant middleware) by the caller in cmd/server/main.go.
func (m *Module) RegisterHTTPRoutes(r chi.Router) {
	r.Route("/integrations", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware.Authenticate)
		r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)
		r.Get("/", m.handleListIntegrations)
		r.Post("/connect", m.handleConnectBYO)
		r.Delete("/{id}", m.handleDisconnect)
		r.Get("/oauth/{provider}/start", m.handleStartOAuth)
		r.Get("/{id}/targets", m.handleListTargets)
	})
	r.Route("/destinations", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware.Authenticate)
		r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)
		r.Get("/", m.handleListDestinations)
		r.Post("/", m.handleCreateDestination)
		r.Patch("/{id}", m.handleUpdateDestination)
		r.Delete("/{id}", m.handleDeleteDestination)
	})
	r.Route("/deliveries", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware.Authenticate)
		r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)
		r.Get("/", m.handleListDeliveries)
		r.Get("/{id}", m.handleGetDelivery)
		r.Post("/bulk-retry", m.handleBulkRetry)
		r.Post("/{id}/retry", m.handleRetryDelivery)
	})
}

// HandleOAuthCallback is exposed for root-mounting (no tenant middleware).
// It validates the HMAC-signed state internally and redirects the browser back
// to the tenant subdomain with the integration result.
func (m *Module) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	q := r.URL.Query()

	// Phase 3: Microsoft Teams consent_required → redirect to friendly modal.
	if oauthErr := q.Get("error"); oauthErr == "consent_required" || oauthErr == "AADSTS65001" {
		redirectURI := m.deps.OAuthRedirectBase + "/api/v1/integrations/oauth/" + provider + "/callback"

		// Try to extract a typed admin-consent URL via the provider client.
		// Discoverable via interface upcast — works only for Teams provider in Phase 3.
		if client, ok := m.deps.Registry.Get(provider); ok {
			if teamsLike, ok := client.(interface {
				BuildAdminConsentURL(redirect string) *domerrors.ErrConsentRequired
			}); ok {
				consentErr := teamsLike.BuildAdminConsentURL(redirectURI)
				target := fmt.Sprintf("%s/settings/integrations?integration_error=consent_required&admin_url=%s",
					m.deps.OAuthRedirectBase, url.QueryEscape(consentErr.AdminConsentURL))
				http.Redirect(w, r, target, http.StatusFound)
				return
			}
		}
		// Fallback: provider doesn't expose admin-consent URL synthesis.
		http.Redirect(w, r,
			m.deps.OAuthRedirectBase+"/settings/integrations?integration_error=consent_required",
			http.StatusFound)
		return
	}

	// Slack error param — try to get tenant redirect from state, fall back to root.
	if oauthErr := q.Get("error"); oauthErr != "" {
		rawState := q.Get("state")
		if st, err := m.deps.StateSigner.Verify(rawState); err == nil {
			http.Redirect(w, r,
				strings.TrimRight(st.ReturnHost, "/")+st.ReturnPath+
					"?integration="+provider+"&status=error",
				http.StatusFound)
		} else {
			http.Redirect(w, r, m.deps.OAuthRedirectBase+"?status=error", http.StatusFound)
		}
		return
	}

	code := q.Get("code")
	state := q.Get("state")

	handler := handleoauthcallback.NewHandler(m.deps.Registry, m.deps.StateSigner, m.deps.IntRepo)
	resp, err := handler.Handle(r.Context(), handleoauthcallback.Request{
		Provider: provider,
		Code:     code,
		State:    state,
	})
	if err != nil {
		logger.Error("OAuth callback failed", zap.Error(err))
		http.Redirect(w, r, m.deps.OAuthRedirectBase+"?status=error", http.StatusFound)
		return
	}

	redirectURL := strings.TrimRight(resp.ReturnHost, "/") + resp.ReturnPath +
		"?integration=" + resp.Provider + "&status=connected"
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// orgIDFromTenant resolves the org UUID from the tenant schema name.
func (m *Module) orgIDFromTenant(ctx context.Context, tenant string) (uuid.UUID, error) {
	var orgID uuid.UUID
	err := m.deps.DB.QueryRowContext(ctx,
		`SELECT id FROM public.organizations WHERE subdomain = $1 AND deleted_at IS NULL`,
		tenant,
	).Scan(&orgID)
	return orgID, err
}

// ---------------------------------------------------------------------------
// Integration handlers
// ---------------------------------------------------------------------------

// integrationResponse is the safe JSON representation of an integration (no tokens).
type integrationResponse struct {
	ID           uuid.UUID              `json:"id"`
	ServiceType  string                 `json:"service_type"`
	Status       entities.IntegrationStatus `json:"status"`
	ProviderMeta map[string]any         `json:"provider_meta"`
	CreatedAt    string                 `json:"created_at"`
	UpdatedAt    string                 `json:"updated_at"`
}

func toIntegrationResponse(i *entities.Integration) integrationResponse {
	return integrationResponse{
		ID:           i.ID,
		ServiceType:  i.ServiceType,
		Status:       i.Status,
		ProviderMeta: i.ProviderMeta,
		CreatedAt:    i.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    i.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// GET /integrations
func (m *Module) handleListIntegrations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := middleware.GetSubdomainFromContext(ctx)
	orgID, err := m.orgIDFromTenant(ctx, tenant)
	if err != nil {
		logger.Error("Failed to resolve org from tenant", zap.Error(err), zap.String("tenant", tenant))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	integrations, err := m.deps.IntRepo.ListByOrg(ctx, orgID)
	if err != nil {
		logger.Error("Failed to list integrations", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	alwaysVisible := map[string]bool{"slack": true, "email": true}
	result := make([]integrationResponse, 0, len(integrations))
	for _, i := range integrations {
		if !alwaysVisible[i.ServiceType] && m.deps.Flags != nil {
			on, _ := m.deps.Flags.IsOn(ctx, orgID, "integrations."+i.ServiceType)
			if !on {
				continue
			}
		}
		result = append(result, toIntegrationResponse(i))
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

// DELETE /integrations/{id}
func (m *Module) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.GetSubdomainFromContext(r.Context())

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	destRepo := persistence.NewDestinationPostgresRepository(m.deps.DB, tenant)
	handler := disconnectintegration.NewHandler(m.deps.IntRepo, destRepo)
	if _, err := handler.Handle(r.Context(), disconnectintegration.Request{IntegrationID: id}); err != nil {
		logger.Error("Failed to disconnect integration", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /integrations/oauth/{provider}/start
func (m *Module) handleStartOAuth(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	tenant := middleware.GetSubdomainFromContext(r.Context())

	returnPath := r.URL.Query().Get("return_path")
	if returnPath == "" {
		returnPath = "/settings/integrations"
	}
	// Sanitize: must start with "/" and not "//" (open redirect prevention)
	if !strings.HasPrefix(returnPath, "/") || strings.HasPrefix(returnPath, "//") {
		returnPath = "/settings/integrations"
	}

	// Capture the tenant's origin for the post-OAuth redirect.
	scheme := "http"
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = strings.TrimSpace(strings.SplitN(proto, ",", 2)[0])
	}
	returnHost := scheme + "://" + r.Host

	orgID, err := m.orgIDFromTenant(r.Context(), tenant)
	if err != nil {
		logger.Error("Failed to resolve org from tenant", zap.Error(err), zap.String("tenant", tenant))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	if provider != "slack" && provider != "email" && m.deps.Flags != nil {
		on, _ := m.deps.Flags.IsOn(r.Context(), orgID, "integrations."+provider)
		if !on {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "provider not enabled for this organization"})
			return
		}
	}

	userIDStr, _ := r.Context().Value(contextkeys.UserIDKey).(string)
	userID, _ := uuid.Parse(userIDStr)

	redirectURI := strings.TrimRight(m.deps.OAuthRedirectBase, "/") +
		"/api/v1/integrations/oauth/" + provider + "/callback"

	handler := startoauth.NewHandler(m.deps.Registry, m.deps.StateSigner)
	resp, err := handler.Handle(r.Context(), startoauth.Request{
		Provider:    provider,
		Tenant:      tenant,
		ReturnPath:  returnPath,
		RedirectURI: redirectURI,
		ReturnHost:  returnHost,
		OrgID:       orgID,
		UserID:      userID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "unknown provider") {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "unknown provider: " + provider})
			return
		}
		logger.Error("Failed to start OAuth", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	http.Redirect(w, r, resp.AuthorizeURL, http.StatusFound)
}

// GET /integrations/{id}/targets
func (m *Module) handleListTargets(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.GetSubdomainFromContext(r.Context())

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	orgID, err := m.orgIDFromTenant(r.Context(), tenant)
	if err != nil {
		logger.Error("Failed to resolve org from tenant", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	// Verify the integration belongs to this org.
	integ, err := m.deps.IntRepo.GetByID(r.Context(), id)
	if err != nil {
		logger.Error("Failed to get integration", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	if integ == nil || integ.OrgID != orgID {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "integration not found"})
		return
	}

	handler := listprovidertargets.NewHandler(m.deps.IntRepo, m.deps.Registry)
	resp, err := handler.Handle(r.Context(), listprovidertargets.Request{
		OrgID:       orgID,
		ServiceType: integ.ServiceType,
	})
	if err != nil {
		logger.Error("Failed to list targets", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"targets": resp.Targets})
}

// ---------------------------------------------------------------------------
// Destination handlers
// ---------------------------------------------------------------------------

// GET /destinations
func (m *Module) handleListDestinations(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.GetSubdomainFromContext(r.Context())
	q := r.URL.Query()

	scopeTypeStr := q.Get("scope_type")
	if scopeTypeStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "scope_type required"})
		return
	}
	scopeIDStr := q.Get("scope_id")
	if scopeIDStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "scope_id required"})
		return
	}
	scopeID, err := uuid.Parse(scopeIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid scope_id"})
		return
	}

	destRepo := persistence.NewDestinationPostgresRepository(m.deps.DB, tenant)
	handler := listdestinations.NewHandler(destRepo)
	resp, err := handler.Handle(r.Context(), listdestinations.Request{
		ScopeType: entities.ScopeType(scopeTypeStr),
		ScopeID:   scopeID,
	})
	if err != nil {
		logger.Error("Failed to list destinations", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": resp.Destinations})
}

// POST /destinations
func (m *Module) handleCreateDestination(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.GetSubdomainFromContext(r.Context())

	var req createdestination.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	destRepo := persistence.NewDestinationPostgresRepository(m.deps.DB, tenant)
	handler := createdestination.NewHandler(destRepo)
	resp, err := handler.Handle(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, resp.Destination)
}

// PATCH /destinations/{id}
func (m *Module) handleUpdateDestination(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.GetSubdomainFromContext(r.Context())

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	var req updatedestination.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	req.ID = id // override from URL param

	destRepo := persistence.NewDestinationPostgresRepository(m.deps.DB, tenant)
	handler := updatedestination.NewHandler(destRepo)
	resp, err := handler.Handle(r.Context(), req)
	if err != nil {
		if err.Error() == "destination not found" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "destination not found"})
			return
		}
		logger.Error("Failed to update destination", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, resp.Destination)
}

// DELETE /destinations/{id}
func (m *Module) handleDeleteDestination(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.GetSubdomainFromContext(r.Context())

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	destRepo := persistence.NewDestinationPostgresRepository(m.deps.DB, tenant)
	handler := deletedestination.NewHandler(destRepo)
	if _, err := handler.Handle(r.Context(), deletedestination.Request{ID: id}); err != nil {
		logger.Error("Failed to delete destination", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ---------------------------------------------------------------------------
// Delivery handlers
// ---------------------------------------------------------------------------

// GET /deliveries
func (m *Module) handleListDeliveries(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.GetSubdomainFromContext(r.Context())
	q := r.URL.Query()

	destIDStr := q.Get("destination_id")
	if destIDStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "destination_id required"})
		return
	}
	destID, err := uuid.Parse(destIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid destination_id"})
		return
	}

	limit := 50
	offset := 0
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			offset = n
		}
	}

	deliveryRepo := persistence.NewDeliveryPostgresRepository(m.deps.DB, tenant)
	handler := listdeliveries.NewHandler(deliveryRepo)
	resp, err := handler.Handle(r.Context(), listdeliveries.Request{
		DestinationID: destID,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		logger.Error("Failed to list deliveries", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": resp.Deliveries})
}

// POST /deliveries/{id}/retry
// Phase 2: gate behind RequireRole('OWNER').
func (m *Module) handleRetryDelivery(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.GetSubdomainFromContext(r.Context())

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	deliveryRepo := persistence.NewDeliveryPostgresRepository(m.deps.DB, tenant)
	handler := retrydelivery.NewHandler(deliveryRepo)
	if _, err := handler.Handle(r.Context(), retrydelivery.Request{ID: id}); err != nil {
		logger.Error("Failed to retry delivery", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// POST /integrations/connect
func (m *Module) handleConnectBYO(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body struct {
		Provider    string            `json:"provider"`
		Credentials map[string]string `json:"credentials"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	tenant := middleware.GetSubdomainFromContext(ctx)
	orgID, err := m.orgIDFromTenant(ctx, tenant)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "org lookup failed"})
		return
	}

	userIDStr, ok := ctx.Value(contextkeys.UserIDKey).(string)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	h := connectbyo.NewHandler(m.deps.IntRepo, m.deps.Validator, m.deps.Flags, m.deps.PlanLookup, m.deps.TwilioPaidPlans)
	resp, err := h.Handle(ctx, connectbyo.Request{
		Provider:    body.Provider,
		OrgID:       orgID,
		UserID:      userID,
		Credentials: body.Credentials,
	})
	if err != nil {
		if strings.Contains(err.Error(), "plan lookup") {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "plan lookup unavailable"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, toIntegrationResponse(resp.Integration))
}

// GET /deliveries/{id}
func (m *Module) handleGetDelivery(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	tenant := middleware.GetSubdomainFromContext(r.Context())
	repo := persistence.NewDeliveryPostgresRepository(m.deps.DB, tenant)

	h := getdelivery.NewHandler(repo)
	resp, err := h.Handle(r.Context(), getdelivery.Request{ID: id})
	if err != nil {
		logger.Error("Failed to get delivery", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if resp.Delivery == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, resp.Delivery)
}

// POST /deliveries/bulk-retry
func (m *Module) handleBulkRetry(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs []uuid.UUID `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	tenant := middleware.GetSubdomainFromContext(r.Context())
	repo := persistence.NewDeliveryPostgresRepository(m.deps.DB, tenant)

	h := bulkretrydeliveries.NewHandler(repo)
	resp, err := h.Handle(r.Context(), bulkretrydeliveries.Request{IDs: body.IDs})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}
