package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	approveuser "github.com/jcsoftdev/pulzifi-back/modules/admin/application/approve_user"
	listpendingusers "github.com/jcsoftdev/pulzifi-back/modules/admin/application/list_pending_users"
	rejectuser "github.com/jcsoftdev/pulzifi-back/modules/admin/application/reject_user"
	adminerrors "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories"
	authrepos "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	authmw "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/middleware"
	orgrepos "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/repositories"
	orgservices "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"go.uber.org/zap"
)

type Module struct {
	listPendingHandler *listpendingusers.Handler
	approveHandler     *approveuser.Handler
	rejectHandler      *rejectuser.Handler
	authMiddleware     *authmw.AuthMiddleware
}

type ModuleDeps struct {
	DB             *sql.DB
	RegReqRepo     repositories.RegistrationRequestRepository
	UserRepo       authrepos.UserRepository
	OrgRepo        orgrepos.OrganizationRepository
	OrgService     *orgservices.OrganizationService
	AuthMiddleware *authmw.AuthMiddleware
}

func NewModule(deps ModuleDeps) router.ModuleRegisterer {
	return &Module{
		listPendingHandler: listpendingusers.NewHandler(deps.RegReqRepo, deps.UserRepo),
		approveHandler:     approveuser.NewHandler(deps.DB, deps.RegReqRepo, deps.UserRepo, deps.OrgRepo, deps.OrgService),
		rejectHandler:      rejectuser.NewHandler(deps.DB, deps.RegReqRepo, deps.UserRepo),
		authMiddleware:     deps.AuthMiddleware,
	}
}

func (m *Module) ModuleName() string {
	return "Admin"
}

func (m *Module) RegisterHTTPRoutes(r chi.Router) {
	r.Route("/admin", func(r chi.Router) {
		r.Use(m.authMiddleware.Authenticate)
		r.Use(m.authMiddleware.RequireRole("SUPER_ADMIN"))

		r.Get("/users/pending", m.handleListPendingUsers)
		r.Put("/users/{id}/approve", m.handleApproveUser)
		r.Put("/users/{id}/reject", m.handleRejectUser)
	})
}

func (m *Module) handleListPendingUsers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	response, err := m.listPendingHandler.Handle(r.Context(), limit, offset)
	if err != nil {
		logger.Error("Failed to list pending users", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list pending users"})
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (m *Module) handleApproveUser(w http.ResponseWriter, r *http.Request) {
	requestID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request id"})
		return
	}

	reviewerIDStr, _ := r.Context().Value(authmw.UserIDKey).(string)
	reviewerID, err := uuid.Parse(reviewerIDStr)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	if err := m.approveHandler.Handle(r.Context(), requestID, reviewerID); err != nil {
		var adminErr adminerrors.AdminError
		if errors.As(err, &adminErr) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": adminErr.Message})
			return
		}
		logger.Error("Failed to approve user", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to approve user"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "user approved successfully"})
}

func (m *Module) handleRejectUser(w http.ResponseWriter, r *http.Request) {
	requestID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request id"})
		return
	}

	reviewerIDStr, _ := r.Context().Value(authmw.UserIDKey).(string)
	reviewerID, err := uuid.Parse(reviewerIDStr)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	if err := m.rejectHandler.Handle(r.Context(), requestID, reviewerID); err != nil {
		var adminErr adminerrors.AdminError
		if errors.As(err, &adminErr) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": adminErr.Message})
			return
		}
		logger.Error("Failed to reject user", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to reject user"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "user rejected successfully"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
