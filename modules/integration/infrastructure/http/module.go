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
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
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

		r.Route("/webhooks", func(r chi.Router) {
			r.Post("/", m.handleCreateWebhook)
			r.Get("/", m.handleListWebhooks)
			r.Get("/{id}", m.handleGetWebhook)
		})
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

func (m *Module) handleCreateWebhook(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		URL    string   `json:"url"`
		Secret string   `json:"secret"`
		Events []string `json:"events"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.URL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "url is required"})
		return
	}

	config := map[string]interface{}{
		"url":    req.URL,
		"secret": req.Secret,
		"events": req.Events,
	}

	repo := persistence.NewIntegrationPostgresRepository(m.db, tenant)
	integration := entities.NewIntegration("webhook", config, userID)

	if err := repo.Create(r.Context(), integration); err != nil {
		logger.Error("Failed to create webhook integration", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":           integration.ID,
		"service_type": integration.ServiceType,
		"config":       integration.Config,
		"enabled":      integration.Enabled,
		"created_at":   integration.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (m *Module) handleListWebhooks(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.GetSubdomainFromContext(r.Context())
	repo := persistence.NewIntegrationPostgresRepository(m.db, tenant)

	webhooks, err := repo.ListByServiceType(r.Context(), "webhook")
	if err != nil {
		logger.Error("Failed to list webhook integrations", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	if webhooks == nil {
		webhooks = []*entities.Integration{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  webhooks,
		"count": len(webhooks),
	})
}

func (m *Module) handleGetWebhook(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.GetSubdomainFromContext(r.Context())

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	repo := persistence.NewIntegrationPostgresRepository(m.db, tenant)
	integration, err := repo.GetByID(r.Context(), id)
	if err != nil {
		logger.Error("Failed to get webhook integration", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	if integration == nil || integration.ServiceType != "webhook" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "webhook not found"})
		return
	}

	writeJSON(w, http.StatusOK, integration)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
