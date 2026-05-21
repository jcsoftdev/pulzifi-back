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
	changepassword "github.com/jcsoftdev/pulzifi-back/modules/auth/application/change_password"
	checksubdomain "github.com/jcsoftdev/pulzifi-back/modules/auth/application/check_subdomain"
	deletecurrentuser "github.com/jcsoftdev/pulzifi-back/modules/auth/application/delete_current_user"
	forgotpassword "github.com/jcsoftdev/pulzifi-back/modules/auth/application/forgot_password"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/application/getcurrentuser"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/application/login"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/application/logout"
	refreshapp "github.com/jcsoftdev/pulzifi-back/modules/auth/application/refreshtoken"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/application/register"
	resetpassword "github.com/jcsoftdev/pulzifi-back/modules/auth/application/reset_password"
	updatecurrentuser "github.com/jcsoftdev/pulzifi-back/modules/auth/application/update_current_user"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	autherrors "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/cookies"
	authmw "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/middleware"
	oauthproviders "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/oauth"
	authpersistence "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"go.uber.org/zap"
)

type Module struct {
	registerHandler          *register.Handler
	checkSubdomainHandler    *checksubdomain.Handler
	loginHandler             *login.Handler
	logoutHandler            *logout.Handler
	refreshHandler           *refreshapp.Handler
	getCurrentUserHandler    *getcurrentuser.Handler
	forgotPasswordHandler    *forgotpassword.Handler
	resetPasswordHandler     *resetpassword.Handler
	updateCurrentUserHandler *updatecurrentuser.Handler
	changePasswordHandler    *changepassword.Handler
	deleteCurrentUserHandler *deletecurrentuser.Handler
	authMiddleware           *authmw.AuthMiddleware
	tokenService             services.TokenService
	userRepo                 repositories.UserRepository
	authService              services.AuthService
	notifier                 services.RegistrationNotifier
	oauthProviders           map[string]oauthproviders.Provider
	refreshTokenRepo         repositories.RefreshTokenRepository
	eventBus                 *eventbus.EventBus
	cookieDomain             string
	cookieSecure             bool
	frontendURL              string
	db                       *sql.DB
}

type ModuleDeps struct {
	UserRepo         repositories.UserRepository
	RefreshTokenRepo repositories.RefreshTokenRepository
	RoleRepo         repositories.RoleRepository
	PermRepo         repositories.PermissionRepository
	RegReqWriter     services.RegistrationRequestWriter
	OrgDirectory     services.OrganizationDirectory
	AuthService      services.AuthService
	TokenService     services.TokenService
	CookieDomain     string
	CookieSecure     bool
	FrontendURL      string
	Notifier         services.RegistrationNotifier
	EventBus         *eventbus.EventBus
	DB               *sql.DB
	OrgContextLookup services.OrgContextLookup // optional; nil → org omitted from /me
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

	getCurrentUserHandler := getcurrentuser.NewHandler(deps.UserRepo, deps.OrgContextLookup)

	var passwordResetRepo repositories.PasswordResetRepository
	if deps.DB != nil {
		passwordResetRepo = authpersistence.NewPasswordResetPostgresRepository(deps.DB)
	}

	return &Module{
		registerHandler:          register.NewHandler(deps.UserRepo, deps.RegReqWriter, deps.OrgDirectory),
		checkSubdomainHandler:    checksubdomain.NewHandler(deps.RegReqWriter, deps.OrgDirectory),
		loginHandler:             login.NewHandler(deps.AuthService, deps.UserRepo, deps.RefreshTokenRepo, deps.TokenService),
		logoutHandler:            logout.NewHandler(deps.RefreshTokenRepo),
		refreshHandler:           refreshapp.NewHandler(deps.RefreshTokenRepo, deps.UserRepo, deps.TokenService),
		getCurrentUserHandler:    getCurrentUserHandler,
		forgotPasswordHandler:    forgotpassword.NewHandler(deps.UserRepo, passwordResetRepo, deps.Notifier),
		resetPasswordHandler:     resetpassword.NewHandler(passwordResetRepo, deps.AuthService),
		updateCurrentUserHandler: updatecurrentuser.NewHandler(deps.UserRepo, getCurrentUserHandler),
		changePasswordHandler:    changepassword.NewHandler(deps.UserRepo, deps.AuthService),
		deleteCurrentUserHandler: deletecurrentuser.NewHandler(deps.UserRepo, deps.EventBus),
		authMiddleware:           authmw.NewAuthMiddleware(deps.TokenService),
		tokenService:             deps.TokenService,
		userRepo:                 deps.UserRepo,
		authService:              deps.AuthService,
		notifier:                 deps.Notifier,
		oauthProviders:           oauthProviderMap,
		refreshTokenRepo:         deps.RefreshTokenRepo,
		eventBus:                 deps.EventBus,
		cookieDomain:             deps.CookieDomain,
		cookieSecure:             deps.CookieSecure,
		frontendURL:              deps.FrontendURL,
		db:                       deps.DB,
	}
}

func (m *Module) AuthMiddleware() *authmw.AuthMiddleware {
	return m.authMiddleware
}

func (m *Module) LoginHandler() *login.Handler        { return m.loginHandler }
func (m *Module) LogoutHandler() *logout.Handler      { return m.logoutHandler }
func (m *Module) RefreshHandler() *refreshapp.Handler { return m.refreshHandler }
func (m *Module) TokenService() services.TokenService { return m.tokenService }
func (m *Module) CookieDomain() string                { return m.cookieDomain }
func (m *Module) CookieSecure() bool                  { return m.cookieSecure }

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

// handleRegister creates a new user registration request
// @Summary Register
// @Description Submit a new user registration request (requires admin approval)
// @Tags auth
// @Accept json
// @Produce json
// @Param request body register.Request true "Registration Request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /auth/register [post]
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
		if sendErr := m.notifier.SendRegistrationSubmitted(context.Background(), response.Email, response.FirstName, req.OrganizationName); sendErr != nil {
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

// handleLogin authenticates a user and returns JWT tokens
// @Summary Login
// @Description Authenticate with email and password. Returns access_token for use with the Authorize button above.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body login.Request true "Login credentials"
// @Success 200 {object} login.Response
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
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

	refreshExpires := time.Now().Add(m.tokenService.GetRefreshTokenExpiration())
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

// handleLogout revokes the refresh token and clears authentication cookies
// @Summary Logout
// @Description Revoke the refresh token and clear authentication cookies
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string
// @Router /auth/logout [post]
func (m *Module) handleLogout(w http.ResponseWriter, r *http.Request) {
	refreshTokenStr, _ := cookies.GetTokenFromCookie(r, cookies.RefreshTokenCookie)
	m.logoutHandler.Handle(r.Context(), refreshTokenStr)
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

	refreshExpires := time.Now().Add(m.tokenService.GetRefreshTokenExpiration())
	cookies.SetRefreshTokenCookie(w, r, response.RefreshToken, refreshExpires, m.cookieDomain, m.cookieSecure)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"expires_in": response.ExpiresIn,
		"tenant":     response.Tenant,
	})
}

// handleGetCurrentUser returns the authenticated user's profile
// @Summary Get current user
// @Description Get the profile of the currently authenticated user
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /auth/me [get]
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

	if err := m.deleteCurrentUserHandler.Handle(r.Context(), &deletecurrentuser.Request{UserID: userID}); err != nil {
		logger.Error("Failed to delete user", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete account"})
		return
	}

	cookies.ClearAuthCookies(w, r, m.cookieDomain, m.cookieSecure)
	writeJSON(w, http.StatusOK, map[string]string{"message": "account deleted successfully"})
}

func (m *Module) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email is required"})
		return
	}

	resp, _ := m.forgotPasswordHandler.Handle(r.Context(), &forgotpassword.Request{
		Email:       body.Email,
		FrontendURL: m.frontendURL,
	})
	writeJSON(w, http.StatusOK, resp)
}

func (m *Module) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if body.Token == "" || body.NewPassword == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "token and new_password are required"})
		return
	}

	resp, err := m.resetPasswordHandler.Handle(r.Context(), &resetpassword.Request{
		Token:       body.Token,
		NewPassword: body.NewPassword,
	})
	if err != nil {
		if err.Error() == "password must be at least 8 characters" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if err.Error() == "service unavailable" {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": err.Error()})
			return
		}
		if err.Error() == "invalid or expired token" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		logger.Error("Failed to reset password", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to reset password"})
		return
	}
	writeJSON(w, http.StatusOK, resp)
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

	// Persist refresh token so the rotation flow (POST /auth/refresh) can find it.
	refreshExpiry := time.Now().Add(m.tokenService.GetRefreshTokenExpiration())
	newRefreshToken := entities.NewRefreshToken(userID, refreshTokenStr, refreshExpiry)
	if err := m.refreshTokenRepo.Create(r.Context(), newRefreshToken); err != nil {
		logger.Error("Failed to store refresh token for OAuth user", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create session"})
		return
	}

	// Set cookies
	accessExpires := time.Now().Add(m.tokenService.GetTokenExpiration())
	cookies.SetAccessTokenCookie(w, r, accessTokenStr, accessExpires, m.cookieDomain, m.cookieSecure)
	refreshExpires := time.Now().Add(m.tokenService.GetRefreshTokenExpiration())
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

	var body struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	roles, _ := r.Context().Value(authmw.UserRolesKey).([]string)

	response, err := m.updateCurrentUserHandler.Handle(r.Context(), &updatecurrentuser.Request{
		UserID:    userID,
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Roles:     roles,
	})
	if err != nil {
		if err.Error() == "user not found" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		logger.Error("Failed to update user profile", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update profile"})
		return
	}
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

	var body struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	resp, err := m.changePasswordHandler.Handle(r.Context(), &changepassword.Request{
		UserID:          userID,
		CurrentPassword: body.CurrentPassword,
		NewPassword:     body.NewPassword,
	})
	if err != nil {
		if err.Error() == "password must be at least 8 characters" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if err.Error() == "user not found" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		if err.Error() == "current password is incorrect" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
			return
		}
		logger.Error("Failed to change password", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update password"})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
