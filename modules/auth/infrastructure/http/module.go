package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	adminrepos "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/application/get_current_user"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/application/login"
	refreshapp "github.com/jcsoftdev/pulzifi-back/modules/auth/application/refresh_token"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/application/register"
	autherrors "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/cookies"
	authmw "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/middleware"
	orgrepos "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/repositories"
	orgservices "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"go.uber.org/zap"
)

type Module struct {
	registerHandler       *register.Handler
	loginHandler          *login.Handler
	refreshHandler        *refreshapp.Handler
	getCurrentUserHandler *get_current_user.Handler
	authMiddleware        *authmw.AuthMiddleware
	tokenService          services.TokenService
	cookieDomain          string
	cookieSecure          bool
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
}

func NewModule(deps ModuleDeps) router.ModuleRegisterer {
	return &Module{
		registerHandler:       register.NewHandler(deps.UserRepo, deps.RegReqRepo, deps.OrgRepo, deps.OrgService),
		loginHandler:          login.NewHandler(deps.AuthService, deps.UserRepo, deps.RefreshTokenRepo, deps.TokenService),
		refreshHandler:        refreshapp.NewHandler(deps.RefreshTokenRepo, deps.UserRepo, deps.TokenService),
		getCurrentUserHandler: get_current_user.NewHandler(deps.UserRepo),
		authMiddleware:        authmw.NewAuthMiddleware(deps.TokenService),
		tokenService:          deps.TokenService,
		cookieDomain:          deps.CookieDomain,
		cookieSecure:          deps.CookieSecure,
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
		r.Post("/login", m.handleLogin)
		r.Post("/logout", m.handleLogout)
		r.Post("/refresh", m.handleRefresh)

		r.Group(func(r chi.Router) {
			r.Use(m.authMiddleware.Authenticate)
			r.Get("/me", m.handleGetCurrentUser)
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

	writeJSON(w, http.StatusCreated, response)
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

	response, err := m.getCurrentUserHandler.Handle(context.Background(), userID)
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

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
