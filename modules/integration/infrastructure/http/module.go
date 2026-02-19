package http

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	authmw "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/middleware"
	deleteintegration "github.com/jcsoftdev/pulzifi-back/modules/integration/application/delete_integration"
	listintegrations "github.com/jcsoftdev/pulzifi-back/modules/integration/application/list_integrations"
	upsertintegration "github.com/jcsoftdev/pulzifi-back/modules/integration/application/upsert_integration"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"go.uber.org/zap"
)

type Module struct {
	db *sql.DB
}

func NewModuleWithDB(db *sql.DB) router.ModuleRegisterer {
	return &Module{db: db}
}

func (m *Module) ModuleName() string {
	return "Integration"
}

func (m *Module) RegisterHTTPRoutes(r chi.Router) {
	r.Route("/integrations", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware.Authenticate)
		r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)

		r.Get("/", m.handleListIntegrations)
		r.Post("/", m.handleUpsertIntegration)
		r.Delete("/{id}", m.handleDeleteIntegration)
	})
}

func (m *Module) handleListIntegrations(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.GetSubdomainFromContext(r.Context())
	repo := persistence.NewIntegrationPostgresRepository(m.db, tenant)
	handler := listintegrations.NewHandler(repo)

	resp, err := handler.Handle(r.Context())
	if err != nil {
		logger.Error("Failed to list integrations", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (m *Module) handleUpsertIntegration(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.GetSubdomainFromContext(r.Context())

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

	var req upsertintegration.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	repo := persistence.NewIntegrationPostgresRepository(m.db, tenant)
	handler := upsertintegration.NewHandler(repo)

	resp, err := handler.Handle(r.Context(), &req, userID)
	if err != nil {
		if err == upsertintegration.ErrInvalidServiceType {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		logger.Error("Failed to upsert integration", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (m *Module) handleDeleteIntegration(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.GetSubdomainFromContext(r.Context())

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	repo := persistence.NewIntegrationPostgresRepository(m.db, tenant)
	handler := deleteintegration.NewHandler(repo)

	if err := handler.Handle(r.Context(), id); err != nil {
		logger.Error("Failed to delete integration", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
