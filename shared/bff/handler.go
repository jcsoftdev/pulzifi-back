package bff

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/application/login"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/application/logout"
	refreshapp "github.com/jcsoftdev/pulzifi-back/modules/auth/application/refresh_token"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/cookies"
	"github.com/jcsoftdev/pulzifi-back/shared/noncestore"
	"go.uber.org/zap"
)

// Handler implements the BFF auth routes that manage cookies and nonces
// for cross-subdomain authentication. This lives in shared/ (not modules/auth)
// because it is HTTP cookie/session orchestration, not domain logic.
type Handler struct {
	loginHandler   *login.Handler
	logoutHandler  *logout.Handler
	refreshHandler *refreshapp.Handler
	tokenService   services.TokenService
	nonceStore     *noncestore.Store
	cookieDomain   string
	cookieSecure   bool
	logger         *zap.Logger
}

type HandlerDeps struct {
	LoginHandler   *login.Handler
	LogoutHandler  *logout.Handler
	RefreshHandler *refreshapp.Handler
	TokenService   services.TokenService
	NonceStore     *noncestore.Store
	CookieDomain   string
	CookieSecure   bool
	Logger         *zap.Logger
}

func NewHandler(deps HandlerDeps) *Handler {
	return &Handler{
		loginHandler:   deps.LoginHandler,
		logoutHandler:  deps.LogoutHandler,
		refreshHandler: deps.RefreshHandler,
		tokenService:   deps.TokenService,
		nonceStore:     deps.NonceStore,
		cookieDomain:   deps.CookieDomain,
		cookieSecure:   deps.CookieSecure,
		logger:         deps.Logger,
	}
}

// RegisterRoutes mounts the BFF auth routes on the given router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/login", h.handleLogin)
	r.Post("/refresh", h.handleRefresh)
	r.Post("/logout", h.handleLogoutPost)
	r.Get("/logout", h.handleLogoutGet)
	r.Get("/callback", h.handleCallback)
	r.Get("/set-base-session", h.handleSetBaseSession)
}

// POST /api/auth/login
// Calls loginHandler in-process, generates a nonce for cross-subdomain redirect,
// sets HttpOnly cookies on current origin, returns {nonce, tenant, expires_in}.
func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req login.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	response, err := h.loginHandler.Handle(r.Context(), &req)
	if err != nil {
		h.logger.Warn("BFF login failed", zap.Error(err))
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	// Generate nonce for cross-subdomain token exchange
	nonce := uuid.New().String()
	h.nonceStore.Save(nonce, noncestore.NonceEntry{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		ExpiresIn:    response.ExpiresIn,
	})

	// Set HttpOnly cookies on the current origin
	accessExpires := time.Now().Add(h.tokenService.GetTokenExpiration())
	cookies.SetAccessTokenCookie(w, r, response.AccessToken, accessExpires, h.cookieDomain, h.cookieSecure)

	refreshExpires := time.Now().Add(h.tokenService.GetRefreshTokenExpiration())
	cookies.SetRefreshTokenCookie(w, r, response.RefreshToken, refreshExpires, h.cookieDomain, h.cookieSecure)

	result := map[string]interface{}{
		"nonce":      nonce,
		"expires_in": response.ExpiresIn,
	}
	if response.Tenant != nil {
		result["tenant"] = *response.Tenant
	}

	writeJSON(w, http.StatusOK, result)
}

// POST /api/auth/refresh
// Reads refresh_token cookie, calls refreshHandler in-process,
// sets new HttpOnly cookies, returns {success: true}.
func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	refreshTokenStr, err := cookies.GetTokenFromCookie(r, cookies.RefreshTokenCookie)
	if err != nil || refreshTokenStr == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing refresh token"})
		return
	}

	req := &refreshapp.Request{RefreshToken: refreshTokenStr}
	response, err := h.refreshHandler.Handle(r.Context(), req)
	if err != nil {
		h.logger.Warn("BFF token refresh failed", zap.Error(err))
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired refresh token"})
		return
	}

	accessExpires := time.Now().Add(h.tokenService.GetTokenExpiration())
	cookies.SetAccessTokenCookie(w, r, response.AccessToken, accessExpires, h.cookieDomain, h.cookieSecure)

	refreshExpires := time.Now().Add(h.tokenService.GetRefreshTokenExpiration())
	cookies.SetRefreshTokenCookie(w, r, response.RefreshToken, refreshExpires, h.cookieDomain, h.cookieSecure)

	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

// POST /api/auth/logout
// Revokes the refresh token, clears auth cookies, and returns {success: true}.
func (h *Handler) handleLogoutPost(w http.ResponseWriter, r *http.Request) {
	refreshTokenStr, _ := cookies.GetTokenFromCookie(r, cookies.RefreshTokenCookie)
	h.logoutHandler.Handle(r.Context(), refreshTokenStr)
	cookies.ClearAuthCookies(w, r, h.cookieDomain, h.cookieSecure)
	clearTenantHintCookie(w, r, h.cookieDomain, h.cookieSecure)
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

// GET /api/auth/logout?redirectTo=/login
// Revokes the refresh token, clears auth cookies, and redirects.
// Used for cross-subdomain cleanup.
func (h *Handler) handleLogoutGet(w http.ResponseWriter, r *http.Request) {
	refreshTokenStr, _ := cookies.GetTokenFromCookie(r, cookies.RefreshTokenCookie)
	h.logoutHandler.Handle(r.Context(), refreshTokenStr)
	cookies.ClearAuthCookies(w, r, h.cookieDomain, h.cookieSecure)
	clearTenantHintCookie(w, r, h.cookieDomain, h.cookieSecure)

	redirectTo := r.URL.Query().Get("redirectTo")
	if redirectTo == "" {
		redirectTo = "/login"
	}
	http.Redirect(w, r, redirectTo, http.StatusFound)
}

// GET /api/auth/callback?nonce=<uuid>&redirectTo=/
// Consumes the nonce, sets HttpOnly cookies on the tenant subdomain, and
// redirects to the app root.
func (h *Handler) handleCallback(w http.ResponseWriter, r *http.Request) {
	nonce := r.URL.Query().Get("nonce")
	if nonce == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	entry := h.nonceStore.Consume(nonce)
	if entry == nil {
		http.Redirect(w, r, "/login?error=SessionExpired", http.StatusFound)
		return
	}

	accessExpires := time.Now().Add(h.tokenService.GetTokenExpiration())
	cookies.SetAccessTokenCookie(w, r, entry.AccessToken, accessExpires, h.cookieDomain, h.cookieSecure)

	refreshExpires := time.Now().Add(h.tokenService.GetRefreshTokenExpiration())
	cookies.SetRefreshTokenCookie(w, r, entry.RefreshToken, refreshExpires, h.cookieDomain, h.cookieSecure)

	redirectTo := r.URL.Query().Get("redirectTo")
	if redirectTo == "" {
		redirectTo = "/"
	}
	http.Redirect(w, r, redirectTo, http.StatusFound)
}

// GET /api/auth/set-base-session?nonce=<uuid>&tenant=<name>&returnTo=<url>
// Peeks the nonce (does NOT consume), sets HttpOnly cookies + tenant_hint on
// the base domain, then redirects to returnTo.
func (h *Handler) handleSetBaseSession(w http.ResponseWriter, r *http.Request) {
	nonce := r.URL.Query().Get("nonce")
	returnTo := r.URL.Query().Get("returnTo")
	tenant := r.URL.Query().Get("tenant")

	if nonce == "" || returnTo == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	entry := h.nonceStore.Peek(nonce)
	if entry == nil {
		http.Redirect(w, r, "/login?error=SessionExpired", http.StatusFound)
		return
	}

	accessExpires := time.Now().Add(h.tokenService.GetTokenExpiration())
	cookies.SetAccessTokenCookie(w, r, entry.AccessToken, accessExpires, h.cookieDomain, h.cookieSecure)

	refreshExpires := time.Now().Add(h.tokenService.GetRefreshTokenExpiration())
	cookies.SetRefreshTokenCookie(w, r, entry.RefreshToken, refreshExpires, h.cookieDomain, h.cookieSecure)

	if tenant != "" {
		setTenantHintCookie(w, r, tenant, h.cookieDomain, h.cookieSecure)
	}

	http.Redirect(w, r, returnTo, http.StatusFound)
}

// --- helpers ---

func setTenantHintCookie(w http.ResponseWriter, r *http.Request, tenant, staticDomain string, secure bool) {
	expires := time.Now().Add(7 * 24 * time.Hour)
	cookie := &http.Cookie{
		Name:     "tenant_hint",
		Value:    tenant,
		Path:     "/",
		Expires:  expires,
		MaxAge:   7 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
	// Use the same domain resolution as auth cookies
	if staticDomain != "" {
		cookie.Domain = staticDomain
	}
	http.SetCookie(w, cookie)
}

func clearTenantHintCookie(w http.ResponseWriter, r *http.Request, staticDomain string, secure bool) {
	cookie := &http.Cookie{
		Name:     "tenant_hint",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
	if staticDomain != "" {
		cookie.Domain = staticDomain
	}
	http.SetCookie(w, cookie)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
