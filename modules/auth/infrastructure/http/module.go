package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/application/get_current_user"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/application/login"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/application/register"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/cookies"
	authmw "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"go.uber.org/zap"
)

type Module struct {
	registerHandler       *register.Handler
	loginHandler          *login.Handler
	getCurrentUserHandler *get_current_user.Handler
	authMiddleware        *authmw.AuthMiddleware
	sessionRepo           repositories.SessionRepository
	cookieDomain          string
	cookieSecure          bool
}

type ModuleDeps struct {
	UserRepo     repositories.UserRepository
	SessionRepo  repositories.SessionRepository
	RoleRepo     repositories.RoleRepository
	PermRepo     repositories.PermissionRepository
	AuthService  services.AuthService
	SessionTTL   time.Duration
	CookieDomain string
	CookieSecure bool
}

func NewModule(deps ModuleDeps) router.ModuleRegisterer {
	return &Module{
		registerHandler:       register.NewHandler(deps.UserRepo),
		loginHandler:          login.NewHandler(deps.AuthService, deps.UserRepo, deps.SessionRepo, deps.SessionTTL),
		getCurrentUserHandler: get_current_user.NewHandler(deps.UserRepo),
		authMiddleware:        authmw.NewAuthMiddleware(deps.SessionRepo, deps.UserRepo, deps.RoleRepo, deps.PermRepo),
		sessionRepo:           deps.SessionRepo,
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

func (m *Module) RegisterHTTPRoutes(router chi.Router) {
	router.Route("/auth", func(r chi.Router) {
		r.Post("/register", m.handleRegister)
		r.Post("/login", m.handleLogin)
		r.Post("/logout", m.handleLogout)

		r.Group(func(r chi.Router) {
			r.Use(m.authMiddleware.Authenticate)
			r.Get("/me", m.handleGetCurrentUser)
		})
	})
}

// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body register.Request true "Register Request"
// @Success 201 {object} register.Response
// @Failure 400 {object} map[string]string
// @Router /auth/register [post]
func (m *Module) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req register.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	response, err := m.registerHandler.Handle(context.Background(), &req)
	if err != nil {
		logger.Error("Failed to register user", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// @Summary Login user
// @Description Authenticate user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body login.Request true "Login Request"
// @Success 200 {object} login.Response
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (m *Module) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req login.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	response, err := m.loginHandler.Handle(r.Context(), &req)
	if err != nil {
		logger.Error("Login failed", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid credentials"})
		return
	}

	cookies.SetSessionCookie(
		w,
		response.SessionID,
		time.Now().Add(time.Duration(response.ExpiresIn)*time.Second),
		m.cookieDomain,
		m.cookieSecure,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleLogout logs out a user
// @Summary Logout user
// @Description Logout the current user
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /auth/logout [post]
func (m *Module) handleLogout(w http.ResponseWriter, r *http.Request) {
	if sessionID, err := cookies.GetSessionIDFromCookie(r); err == nil && sessionID != "" {
		_ = m.sessionRepo.DeleteByID(r.Context(), sessionID)
	}
	cookies.ClearSessionCookie(w, m.cookieDomain, m.cookieSecure)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "logged out successfully",
	})
}

// @Summary Get Current User
// @Description Get the current authenticated user's information
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} get_current_user.Response
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /auth/me [get]
func (m *Module) handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value(authmw.UserIDKey).(string)
	if !ok {
		logger.Error("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Error("Invalid user ID", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid user id"})
		return
	}

	response, err := m.getCurrentUserHandler.Handle(context.Background(), userID)
	if err != nil {
		logger.Error("Failed to get current user", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to get user"})
		return
	}

	if response == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "user not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
