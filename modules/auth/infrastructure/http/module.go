package http

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	adminrepos "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories"
	checksubdomain "github.com/jcsoftdev/pulzifi-back/modules/auth/application/check_subdomain"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/application/get_current_user"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/application/login"
	refreshapp "github.com/jcsoftdev/pulzifi-back/modules/auth/application/refresh_token"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/application/register"
	autherrors "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/cookies"
	authmw "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/middleware"
	emailservices "github.com/jcsoftdev/pulzifi-back/modules/email/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/email/infrastructure/templates"
	oauthproviders "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/oauth"
	orgrepos "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/repositories"
	orgservices "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"go.uber.org/zap"
)

type Module struct {
	registerHandler        *register.Handler
	checkSubdomainHandler  *checksubdomain.Handler
	loginHandler           *login.Handler
	refreshHandler         *refreshapp.Handler
	getCurrentUserHandler  *get_current_user.Handler
	authMiddleware         *authmw.AuthMiddleware
	tokenService           services.TokenService
	authService            services.AuthService
	userRepo               repositories.UserRepository
	emailProvider          emailservices.EmailProvider
	oauthProviders         map[string]oauthproviders.Provider
	refreshTokenRepo       repositories.RefreshTokenRepository
	eventBus               *eventbus.EventBus
	cookieDomain           string
	cookieSecure           bool
	frontendURL            string
	db                     *sql.DB
}

type ModuleDeps struct {
	UserRepo         repositories.UserRepository
	RefreshTokenRepo repositories.RefreshTokenRepository
	RoleRepo         repositories.RoleRepository
	PermRepo         repositories.PermissionRepository
	RegReqRepo       adminrepos.RegistrationRequestRepository
	OrgRepo          orgrepos.OrganizationRepository
	OrgService       *orgservices.OrganizationService
	AuthService      services.AuthService
	TokenService     services.TokenService
	CookieDomain     string
	CookieSecure     bool
	FrontendURL      string
	EmailProvider    emailservices.EmailProvider
	EventBus         *eventbus.EventBus
	DB               *sql.DB
}

func NewModule(deps ModuleDeps) router.ModuleRegisterer {
	cfg := config.Load()
	oauthProviderMap := make(map[string]oauthproviders.Provider)
	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		oauthProviderMap["google"] = oauthproviders.NewGoogleProvider(
			cfg.GoogleClientID, cfg.GoogleClientSecret,
			cfg.OAuthRedirectBaseURL+"/api/v1/auth/oauth/google/callback",
		)
	}
	if cfg.GitHubClientID != "" && cfg.GitHubClientSecret != "" {
		oauthProviderMap["github"] = oauthproviders.NewGitHubProvider(
			cfg.GitHubClientID, cfg.GitHubClientSecret,
			cfg.OAuthRedirectBaseURL+"/api/v1/auth/oauth/github/callback",
		)
	}

	return &Module{
		registerHandler:       register.NewHandler(deps.UserRepo, deps.RegReqRepo, deps.OrgRepo, deps.OrgService),
		checkSubdomainHandler: checksubdomain.NewHandler(deps.RegReqRepo, deps.OrgRepo, deps.OrgService),
		loginHandler:          login.NewHandler(deps.AuthService, deps.UserRepo, deps.RefreshTokenRepo, deps.TokenService),
		refreshHandler:        refreshapp.NewHandler(deps.RefreshTokenRepo, deps.UserRepo, deps.TokenService),
		getCurrentUserHandler: get_current_user.NewHandler(deps.UserRepo),
		authMiddleware:        authmw.NewAuthMiddleware(deps.TokenService),
		tokenService:          deps.TokenService,
		authService:           deps.AuthService,
		userRepo:              deps.UserRepo,
		emailProvider:         deps.EmailProvider,
		eventBus:              deps.EventBus,
		oauthProviders:        oauthProviderMap,
		refreshTokenRepo:      deps.RefreshTokenRepo,
		cookieDomain:          deps.CookieDomain,
		cookieSecure:          deps.CookieSecure,
		frontendURL:           deps.FrontendURL,
		db:                    deps.DB,
	}
}

func (m *Module) AuthMiddleware() *authmw.AuthMiddleware {
	return m.authMiddleware
}

func (m *Module) ModuleName() string {
	return "Auth"
}

func (m *Module) RegisterHTTPRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", m.handleRegister)
		r.Post("/check-subdomain", m.handleCheckSubdomain)
		r.Post("/login", m.handleLogin)
		r.Post("/logout", m.handleLogout)
		r.Post("/refresh", m.handleRefresh)
		r.Post("/forgot-password", m.handleForgotPassword)
		r.Post("/reset-password", m.handleResetPassword)

		// OAuth routes
		r.Get("/oauth/{provider}", m.handleOAuthRedirect)
		r.Get("/oauth/{provider}/callback", m.handleOAuthCallback)

		r.Group(func(r chi.Router) {
			r.Use(m.authMiddleware.Authenticate)
			r.Get("/me", m.handleGetCurrentUser)
			r.Put("/me", m.handleUpdateCurrentUser)
			r.Put("/me/password", m.handleChangePassword)
			r.Delete("/me", m.handleDeleteCurrentUser)
		})
	})
}

func (m *Module) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req register.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request", zap.Error(err))
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	response, err := m.registerHandler.Handle(context.Background(), &req)
	if err != nil {
		logger.Error("Failed to register user", zap.Error(err))
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	go func() {
		subject, html := templates.RegistrationSubmitted(response.FirstName, req.OrganizationName)
		if sendErr := m.emailProvider.Send(context.Background(), response.Email, subject, html); sendErr != nil {
			logger.Error("Failed to send registration confirmation email", zap.Error(sendErr))
		}
	}()

	writeJSON(w, http.StatusCreated, response)
}

func (m *Module) handleCheckSubdomain(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Subdomain string `json:"subdomain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Subdomain == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "subdomain is required"})
		return
	}

	response, err := m.checkSubdomainHandler.Handle(r.Context(), req.Subdomain)
	if err != nil {
		logger.Error("Failed to check subdomain", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to check subdomain"})
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (m *Module) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req login.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	response, err := m.loginHandler.Handle(r.Context(), &req)
	if err != nil {
		var userErr autherrors.UserError
		if errors.As(err, &userErr) {
			if userErr.Code == autherrors.ErrUserNotApproved.Code || userErr.Code == autherrors.ErrUserRejected.Code {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": userErr.Message, "code": userErr.Code})
				return
			}
		}
		logger.Error("Login failed", zap.Error(err))
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	accessExpires := time.Now().Add(m.tokenService.GetTokenExpiration())
	cookies.SetAccessTokenCookie(w, r, response.AccessToken, accessExpires, m.cookieDomain, m.cookieSecure)

	refreshExpires := m.tokenService.GetRefreshTokenExpiration()
	cookies.SetRefreshTokenCookie(w, r, response.RefreshToken, refreshExpires, m.cookieDomain, m.cookieSecure)

	logger.Info("Login successful, JWT cookies set",
		zap.String("host", r.Host),
		zap.Bool("secure", m.cookieSecure),
	)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  response.AccessToken,
		"refresh_token": response.RefreshToken,
		"expires_in":    response.ExpiresIn,
		"tenant":        response.Tenant,
	})
}

func (m *Module) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookies.ClearAuthCookies(w, r, m.cookieDomain, m.cookieSecure)
	writeJSON(w, http.StatusOK, map[string]interface{}{"message": "logged out successfully"})
}

func (m *Module) handleRefresh(w http.ResponseWriter, r *http.Request) {
	refreshTokenStr, err := cookies.GetTokenFromCookie(r, cookies.RefreshTokenCookie)
	if err != nil || refreshTokenStr == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing refresh token"})
		return
	}

	req := &refreshapp.Request{RefreshToken: refreshTokenStr}
	response, err := m.refreshHandler.Handle(r.Context(), req)
	if err != nil {
		logger.Warn("Token refresh failed", zap.Error(err))
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired refresh token"})
		return
	}

	accessExpires := time.Now().Add(m.tokenService.GetTokenExpiration())
	cookies.SetAccessTokenCookie(w, r, response.AccessToken, accessExpires, m.cookieDomain, m.cookieSecure)

	refreshExpires := m.tokenService.GetRefreshTokenExpiration()
	cookies.SetRefreshTokenCookie(w, r, response.RefreshToken, refreshExpires, m.cookieDomain, m.cookieSecure)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"expires_in": response.ExpiresIn,
		"tenant":     response.Tenant,
	})
}

func (m *Module) handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value(authmw.UserIDKey).(string)
	if !ok {
		logger.Error("User ID not found in context")
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Error("Invalid user ID", zap.Error(err))
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	roles, _ := r.Context().Value(authmw.UserRolesKey).([]string)

	response, err := m.getCurrentUserHandler.Handle(r.Context(), userID, roles)
	if err != nil {
		logger.Error("Failed to get current user", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get user"})
		return
	}

	if response == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (m *Module) handleDeleteCurrentUser(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value(authmw.UserIDKey).(string)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	if err := m.userRepo.Delete(r.Context(), userID); err != nil {
		logger.Error("Failed to delete user", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete account"})
		return
	}

	// Publish user.deleted event to trigger cascade cleanup (org memberships, etc.)
	if m.eventBus != nil {
		payload, _ := json.Marshal(map[string]string{"user_id": userID.String()})
		if err := m.eventBus.Publish("user.deleted", userID.String(), payload); err != nil {
			logger.Error("Failed to publish user.deleted event", zap.Error(err))
		}
	}

	// Clear auth cookies
	cookies.ClearAuthCookies(w, r, m.cookieDomain, m.cookieSecure)

	writeJSON(w, http.StatusOK, map[string]string{"message": "account deleted successfully"})
}

func (m *Module) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email is required"})
		return
	}

	// Always return 200 to prevent email enumeration
	defer func() {
		writeJSON(w, http.StatusOK, map[string]string{"message": "if an account exists with that email, a password reset link has been sent"})
	}()

	if m.db == nil {
		return
	}

	user, err := m.userRepo.GetByEmail(r.Context(), req.Email)
	if err != nil || user == nil {
		return
	}

	// Generate secure token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		logger.Error("Failed to generate reset token", zap.Error(err))
		return
	}
	token := hex.EncodeToString(tokenBytes)
	expiresAt := time.Now().Add(1 * time.Hour)

	// Store token in password_resets table
	_, err = m.db.ExecContext(r.Context(),
		`INSERT INTO public.password_resets (id, user_id, token, expires_at, created_at) VALUES ($1, $2, $3, $4, NOW())`,
		uuid.New(), user.ID, token, expiresAt,
	)
	if err != nil {
		logger.Error("Failed to store password reset token", zap.Error(err))
		return
	}

	// Send password reset email
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", m.frontendURL, token)
	subject, html := templates.PasswordReset(user.FirstName, resetURL)
	go func() {
		if err := m.emailProvider.Send(context.Background(), user.Email, subject, html); err != nil {
			logger.Error("Failed to send password reset email", zap.Error(err))
		}
	}()
}

func (m *Module) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Token == "" || req.NewPassword == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "token and new_password are required"})
		return
	}
	if len(req.NewPassword) < 8 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
		return
	}

	if m.db == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "service unavailable"})
		return
	}

	// Find and validate token
	var userID uuid.UUID
	var expiresAt time.Time
	var used bool
	err := m.db.QueryRowContext(r.Context(),
		`SELECT user_id, expires_at, used FROM public.password_resets WHERE token = $1`,
		req.Token,
	).Scan(&userID, &expiresAt, &used)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or expired token"})
		return
	}
	if used || time.Now().After(expiresAt) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or expired token"})
		return
	}

	// Hash new password
	hashedPassword, err := m.authService.HashPassword(req.NewPassword)
	if err != nil {
		logger.Error("Failed to hash password", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to reset password"})
		return
	}

	// Update password and invalidate token in a transaction
	tx, err := m.db.BeginTx(r.Context(), nil)
	if err != nil {
		logger.Error("Failed to begin transaction", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to reset password"})
		return
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(r.Context(),
		`UPDATE public.users SET password_hash = $1, updated_at = NOW() WHERE id = $2`,
		hashedPassword, userID,
	)
	if err != nil {
		logger.Error("Failed to update password", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to reset password"})
		return
	}

	_, err = tx.ExecContext(r.Context(),
		`UPDATE public.password_resets SET used = true WHERE token = $1`,
		req.Token,
	)
	if err != nil {
		logger.Error("Failed to invalidate token", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to reset password"})
		return
	}

	// Activate any pending organization memberships — the user has accepted their invitation
	_, err = tx.ExecContext(r.Context(),
		`UPDATE public.organization_members SET invitation_status = 'active' WHERE user_id = $1 AND invitation_status = 'pending'`,
		userID,
	)
	if err != nil {
		logger.Error("Failed to activate pending organization memberships", zap.Error(err))
		// Non-fatal: proceed with password reset even if this update fails
	}

	if err := tx.Commit(); err != nil {
		logger.Error("Failed to commit password reset", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to reset password"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "password reset successfully"})
}

func (m *Module) handleOAuthRedirect(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")
	provider, ok := m.oauthProviders[providerName]
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("unsupported OAuth provider: %s", providerName)})
		return
	}

	// Generate state parameter for CSRF protection
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate state"})
		return
	}
	state := hex.EncodeToString(stateBytes)

	// Store state in a short-lived cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600, // 10 minutes
		HttpOnly: true,
		Secure:   m.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, provider.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func (m *Module) handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")
	provider, ok := m.oauthProviders[providerName]
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("unsupported OAuth provider: %s", providerName)})
		return
	}

	// Validate state parameter
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value == "" || stateCookie.Value != r.URL.Query().Get("state") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid state parameter"})
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing authorization code"})
		return
	}

	// Exchange code for user info
	oauthUser, token, err := provider.Exchange(r.Context(), code)
	if err != nil {
		logger.Error("OAuth exchange failed", zap.String("provider", providerName), zap.Error(err))
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed to authenticate with provider"})
		return
	}

	if m.db == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "service unavailable"})
		return
	}

	// Check if OAuth link exists
	var existingUserID uuid.UUID
	err = m.db.QueryRowContext(r.Context(),
		`SELECT user_id FROM public.user_oauth_providers WHERE provider = $1 AND provider_user_id = $2`,
		providerName, oauthUser.ProviderID,
	).Scan(&existingUserID)

	var userID uuid.UUID
	if err == nil {
		// Existing OAuth link — log in
		userID = existingUserID
	} else {
		// Check if user exists by email
		existingUser, _ := m.userRepo.GetByEmail(r.Context(), oauthUser.Email)
		if existingUser != nil {
			userID = existingUser.ID
		} else {
			// Create new user (auto-approved for OAuth)
			randomPassword := make([]byte, 32)
			rand.Read(randomPassword)
			hashedPw, _ := m.authService.HashPassword(hex.EncodeToString(randomPassword))

			newUserID := uuid.New()
			avatarPtr := &oauthUser.AvatarURL
			if oauthUser.AvatarURL == "" {
				avatarPtr = nil
			}

			_, err = m.db.ExecContext(r.Context(),
				`INSERT INTO public.users (id, email, password_hash, first_name, last_name, avatar_url, status, email_verified, created_at, updated_at)
				 VALUES ($1, $2, $3, $4, $5, $6, 'approved', true, NOW(), NOW())`,
				newUserID, oauthUser.Email, hashedPw, oauthUser.FirstName, oauthUser.LastName, avatarPtr,
			)
			if err != nil {
				logger.Error("Failed to create OAuth user", zap.Error(err))
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create account"})
				return
			}
			userID = newUserID
		}

		// Create OAuth provider link
		accessToken := ""
		refreshToken := ""
		if token != nil {
			accessToken = token.AccessToken
			refreshToken = token.RefreshToken
		}
		_, err = m.db.ExecContext(r.Context(),
			`INSERT INTO public.user_oauth_providers (user_id, provider, provider_user_id, email, access_token, refresh_token)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 ON CONFLICT (provider, provider_user_id) DO UPDATE SET access_token = $5, refresh_token = $6, updated_at = NOW()`,
			userID, providerName, oauthUser.ProviderID, oauthUser.Email, accessToken, refreshToken,
		)
		if err != nil {
			logger.Error("Failed to save OAuth provider link", zap.Error(err))
			// Non-fatal — user can still log in
		}
	}

	// Generate JWT tokens
	accessTokenStr, err := m.tokenService.GenerateAccessToken(r.Context(), userID, oauthUser.Email)
	if err != nil {
		logger.Error("Failed to generate access token for OAuth user", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		return
	}

	refreshTokenStr, err := m.tokenService.GenerateRefreshToken(r.Context(), userID)
	if err != nil {
		logger.Error("Failed to generate refresh token for OAuth user", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		return
	}

	// Set cookies
	accessExpires := time.Now().Add(m.tokenService.GetTokenExpiration())
	cookies.SetAccessTokenCookie(w, r, accessTokenStr, accessExpires, m.cookieDomain, m.cookieSecure)
	refreshExpires := m.tokenService.GetRefreshTokenExpiration()
	cookies.SetRefreshTokenCookie(w, r, refreshTokenStr, refreshExpires, m.cookieDomain, m.cookieSecure)

	// Redirect to frontend
	redirectURL := m.frontendURL
	if redirectURL == "" {
		redirectURL = "/"
	}
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func (m *Module) handleUpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value(authmw.UserIDKey).(string)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	var req struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	user, err := m.userRepo.GetByID(r.Context(), userID)
	if err != nil || user == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}

	if err := m.userRepo.Update(r.Context(), user); err != nil {
		logger.Error("Failed to update user profile", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update profile"})
		return
	}

	roles, _ := r.Context().Value(authmw.UserRolesKey).([]string)
	response, _ := m.getCurrentUserHandler.Handle(r.Context(), userID, roles)
	writeJSON(w, http.StatusOK, response)
}

func (m *Module) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value(authmw.UserIDKey).(string)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if len(req.NewPassword) < 8 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
		return
	}

	user, err := m.userRepo.GetByID(r.Context(), userID)
	if err != nil || user == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	if err := m.authService.ValidateCredentials(r.Context(), user, req.CurrentPassword); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "current password is incorrect"})
		return
	}

	hashedPassword, err := m.authService.HashPassword(req.NewPassword)
	if err != nil {
		logger.Error("Failed to hash new password", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update password"})
		return
	}

	user.PasswordHash = hashedPassword
	if err := m.userRepo.Update(r.Context(), user); err != nil {
		logger.Error("Failed to update user password", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update password"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "password updated successfully"})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
