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
	invitetoplatform "github.com/jcsoftdev/pulzifi-back/modules/admin/application/invite_to_platform"
	listinvitations "github.com/jcsoftdev/pulzifi-back/modules/admin/application/list_invitations"
	listpendingusers "github.com/jcsoftdev/pulzifi-back/modules/admin/application/list_pending_users"
	rejectuser "github.com/jcsoftdev/pulzifi-back/modules/admin/application/reject_user"
	resendinvitation "github.com/jcsoftdev/pulzifi-back/modules/admin/application/resend_invitation"
	revokeinvitation "github.com/jcsoftdev/pulzifi-back/modules/admin/application/revoke_invitation"
	adminerrors "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories"
	authrepos "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	authmw "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/middleware"
	emailservices "github.com/jcsoftdev/pulzifi-back/modules/email/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/email/infrastructure/templates"
	orgrepos "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/repositories"
	orgservices "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"go.uber.org/zap"
)

type Module struct {
	listPendingHandler     *listpendingusers.Handler
	approveHandler         *approveuser.Handler
	rejectHandler          *rejectuser.Handler
	inviteHandler          *invitetoplatform.Handler
	listInvitationsHandler *listinvitations.Handler
	revokeHandler          *revokeinvitation.Handler
	resendHandler          *resendinvitation.Handler
	authMiddleware         *authmw.AuthMiddleware
	emailProvider          emailservices.EmailProvider
	userRepo               authrepos.UserRepository
	frontendURL            string
}

type ModuleDeps struct {
	DB                     *sql.DB
	RegReqRepo             repositories.RegistrationRequestRepository
	UserRepo               authrepos.UserRepository
	OrgRepo                orgrepos.OrganizationRepository
	OrgService             *orgservices.OrganizationService
	AuthMiddleware         *authmw.AuthMiddleware
	EmailProvider          emailservices.EmailProvider
	FrontendURL            string
	InviteHandler          *invitetoplatform.Handler
	ListInvitationsHandler *listinvitations.Handler
	RevokeHandler          *revokeinvitation.Handler
	ResendHandler          *resendinvitation.Handler
}

func NewModule(deps ModuleDeps) router.ModuleRegisterer {
	return &Module{
		listPendingHandler:     listpendingusers.NewHandler(deps.RegReqRepo, deps.UserRepo),
		approveHandler:         approveuser.NewHandler(deps.DB, deps.RegReqRepo, deps.UserRepo, deps.OrgRepo, deps.OrgService),
		rejectHandler:          rejectuser.NewHandler(deps.DB, deps.RegReqRepo, deps.UserRepo),
		inviteHandler:          deps.InviteHandler,
		listInvitationsHandler: deps.ListInvitationsHandler,
		revokeHandler:          deps.RevokeHandler,
		resendHandler:          deps.ResendHandler,
		authMiddleware:         deps.AuthMiddleware,
		emailProvider:          deps.EmailProvider,
		userRepo:               deps.UserRepo,
		frontendURL:            deps.FrontendURL,
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

		r.Route("/invitations", func(r chi.Router) {
			r.Post("/", m.handleCreateInvitation)
			r.Get("/", m.handleListInvitations)
			r.Delete("/{id}", m.handleRevokeInvitation)
			r.Post("/{id}/resend", m.handleResendInvitation)
		})
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

	// Send approval notification email (fire-and-forget)
	go func() {
		regReq, err := m.approveHandler.GetRegistrationRequest(r.Context(), requestID)
		if err != nil {
			logger.Error("Failed to get reg request for email", zap.Error(err))
			return
		}
		user, err := m.userRepo.GetByID(r.Context(), regReq.UserID)
		if err != nil || user == nil {
			logger.Error("Failed to get user for approval email", zap.Error(err))
			return
		}
		loginURL := m.frontendURL
		subject, html := templates.ApprovalNotification(user.FirstName, regReq.OrganizationSubdomain, loginURL)
		if err := m.emailProvider.Send(r.Context(), user.Email, subject, html); err != nil {
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

	// Send rejection notification email (fire-and-forget)
	go func() {
		regReq, err := m.rejectHandler.GetRegistrationRequest(r.Context(), requestID)
		if err != nil {
			logger.Error("Failed to get reg request for rejection email", zap.Error(err))
			return
		}
		user, err := m.userRepo.GetByID(r.Context(), regReq.UserID)
		if err != nil || user == nil {
			logger.Error("Failed to get user for rejection email", zap.Error(err))
			return
		}
		subject, html := templates.RejectionNotification(user.FirstName)
		if err := m.emailProvider.Send(r.Context(), user.Email, subject, html); err != nil {
			logger.Error("Failed to send rejection email", zap.Error(err))
		}
	}()

	writeJSON(w, http.StatusOK, map[string]string{"message": "user rejected successfully"})
}

// handleCreateInvitation creates a new platform invitation (SUPER_ADMIN only).
func (m *Module) handleCreateInvitation(w http.ResponseWriter, r *http.Request) {
	var req invitetoplatform.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email is required"})
		return
	}

	inviterIDStr, _ := r.Context().Value(authmw.UserIDKey).(string)
	inviterID, err := uuid.Parse(inviterIDStr)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	resp, err := m.inviteHandler.Handle(r.Context(), req, inviterID)
	if err != nil {
		switch {
		case errors.Is(err, adminerrors.ErrCannotInviteEmail):
			writeJSON(w, http.StatusConflict, map[string]string{"error": "cannot_invite_email"})
		case errors.Is(err, adminerrors.ErrDailyCapExceeded):
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "daily_cap_exceeded"})
		default:
			logger.Error("Failed to create invitation", zap.Error(err))
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create invitation"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// handleListInvitations lists platform invitations with optional status filter.
func (m *Module) handleListInvitations(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	resp, err := m.listInvitationsHandler.Handle(r.Context(), status, limit, offset)
	if err != nil {
		logger.Error("Failed to list invitations", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list invitations"})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleRevokeInvitation revokes a pending invitation by id.
func (m *Module) handleRevokeInvitation(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid invitation id"})
		return
	}

	revokerIDStr, _ := r.Context().Value(authmw.UserIDKey).(string)
	revokerID, err := uuid.Parse(revokerIDStr)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	if err := m.revokeHandler.Handle(r.Context(), id, revokerID); err != nil {
		if errors.Is(err, adminerrors.ErrInvitationAlreadyDecided) {
			writeJSON(w, http.StatusGone, map[string]string{"error": "invitation_already_decided"})
			return
		}
		logger.Error("Failed to revoke invitation", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to revoke invitation"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleResendInvitation re-issues a token and sends a fresh email for a pending invitation.
func (m *Module) handleResendInvitation(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid invitation id"})
		return
	}

	resp, err := m.resendHandler.Handle(r.Context(), id)
	if err != nil {
		if errors.Is(err, adminerrors.ErrInvitationAlreadyDecided) {
			writeJSON(w, http.StatusGone, map[string]string{"error": "invitation_already_decided"})
			return
		}
		logger.Error("Failed to resend invitation", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to resend invitation"})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
