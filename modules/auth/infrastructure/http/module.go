package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
)

// Module implements the router.ModuleRegisterer interface for the Auth module
type Module struct{}

// NewModule creates a new instance of the Auth module
func NewModule() router.ModuleRegisterer {
	return &Module{}
}

// ModuleName returns the name of the module
func (m *Module) ModuleName() string {
	return "Auth"
}

// RegisterHTTPRoutes registers all HTTP routes for the Auth module
func (m *Module) RegisterHTTPRoutes(router chi.Router) {
	router.Route("/auth", func(r chi.Router) {
		r.Post("/register", m.handleRegister)
		r.Post("/login", m.handleLogin)
		r.Post("/refresh", m.handleRefresh)
		r.Post("/logout", m.handleLogout)
	})
}

// handleRegister registers a new user
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Register Request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /auth/register [post]
func (m *Module) handleRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      "550e8400-e29b-41d4-a716-446655440000",
		"message": "user registered successfully",
	})
}

// handleLogin authenticates a user
// @Summary Login user
// @Description Authenticate user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Login Request"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (m *Module) handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":       "550e8400-e29b-41d4-a716-446655440000",
		"access_token":  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		"refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		"message":       "login endpoint",
	})
}

// handleRefresh refreshes an access token
// @Summary Refresh token
// @Description Refresh an access token using a refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Refresh Request"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /auth/refresh [post]
func (m *Module) handleRefresh(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		"message":      "refresh endpoint",
	})
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "logout endpoint",
	})
}
