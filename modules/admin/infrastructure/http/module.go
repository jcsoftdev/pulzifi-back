package http

import (
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
	adminservices "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/contextkeys"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"go.uber.org/zap"
)

type Module struct {
	listPendingHandler *listpendingusers.Handler
	approveHandler     *approveuser.Handler
	rejectHandler      *rejectuser.Handler
	authMiddleware     middleware.AuthVerifier
	notifier           adminservices.RegistrationNotifier
	userReader         adminservices.PendingUserReader
	frontendURL        string
}

type ModuleDeps struct {
	RegReqRepo          repositories.RegistrationRequestRepository
	UserReader          adminservices.PendingUserReader
	ApprovalProvisioner adminservices.ApprovalProvisioner
	RejectionProvisioner adminservices.RejectionProvisioner
	Notifier            adminservices.RegistrationNotifier
	AuthMiddleware      middleware.AuthVerifier
	FrontendURL         string
}

func NewModule(deps ModuleDeps) router.ModuleRegisterer {
	return &Module{
		listPendingHandler: listpendingusers.NewHandler(deps.RegReqRepo, deps.UserReader),
		approveHandler:     approveuser.NewHandler(deps.RegReqRepo, deps.ApprovalProvisioner),
		rejectHandler:      rejectuser.NewHandler(deps.RegReqRepo, deps.RejectionProvisioner),
		authMiddleware:     deps.AuthMiddleware,
		notifier:           deps.Notifier,
		userReader:         deps.UserReader,
		frontendURL:        deps.FrontendURL,
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

	reviewerIDStr, _ := r.Context().Value(contextkeys.UserIDKey).(string)
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

	// Send approval notification email (fire-and-forget)
	go func() {
		regReq, err := m.approveHandler.GetRegistrationRequest(r.Context(), requestID)
		if err != nil {
			logger.Error("Failed to get reg request for email", zap.Error(err))
			return
		}
		user, err := m.userReader.GetByID(r.Context(), regReq.UserID)
		if err != nil || user == nil {
			logger.Error("Failed to get user for approval email", zap.Error(err))
			return
		}
		if err := m.notifier.SendApproval(r.Context(), user.Email, user.FirstName, regReq.OrganizationSubdomain, m.frontendURL); err != nil {
			logger.Error("Failed to send approval email", zap.Error(err))
		}
	}()

	writeJSON(w, http.StatusOK, map[string]string{"message": "user approved successfully"})
}

func (m *Module) handleRejectUser(w http.ResponseWriter, r *http.Request) {
	requestID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request id"})
		return
	}

	reviewerIDStr, _ := r.Context().Value(contextkeys.UserIDKey).(string)
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

	// Send rejection notification email (fire-and-forget)
	go func() {
		regReq, err := m.rejectHandler.GetRegistrationRequest(r.Context(), requestID)
		if err != nil {
			logger.Error("Failed to get reg request for rejection email", zap.Error(err))
			return
		}
		user, err := m.userReader.GetByID(r.Context(), regReq.UserID)
		if err != nil || user == nil {
			logger.Error("Failed to get user for rejection email", zap.Error(err))
			return
		}
		if err := m.notifier.SendRejection(r.Context(), user.Email, user.FirstName); err != nil {
			logger.Error("Failed to send rejection email", zap.Error(err))
		}
	}()

	writeJSON(w, http.StatusOK, map[string]string{"message": "user rejected successfully"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
