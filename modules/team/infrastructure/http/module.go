package http

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	authmw "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/middleware"
	invitemember "github.com/jcsoftdev/pulzifi-back/modules/team/application/invite_member"
	listmembers "github.com/jcsoftdev/pulzifi-back/modules/team/application/list_members"
	removemember "github.com/jcsoftdev/pulzifi-back/modules/team/application/remove_member"
	updatemember "github.com/jcsoftdev/pulzifi-back/modules/team/application/update_member"
	"github.com/jcsoftdev/pulzifi-back/modules/team/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"go.uber.org/zap"
)

const (
	contentTypeHeader = "Content-Type"
	applicationJSON   = "application/json"
)

type Module struct {
	db *sql.DB
}

func NewModuleWithDB(db *sql.DB) router.ModuleRegisterer {
	return &Module{db: db}
}

func (m *Module) ModuleName() string {
	return "Team"
}

func (m *Module) RegisterHTTPRoutes(r chi.Router) {
	r.Route("/team", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware.Authenticate)
		r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)

		r.Get("/members", m.handleListMembers)
		r.Post("/members", m.handleInviteMember)
		r.Put("/members/{member_id}", m.handleUpdateMember)
		r.Delete("/members/{member_id}", m.handleRemoveMember)
	})
}

func (m *Module) handleListMembers(w http.ResponseWriter, r *http.Request) {
	subdomain := middleware.GetSubdomainFromContext(r.Context())

	repo := persistence.NewTeamMemberPostgresRepository(m.db)
	handler := listmembers.NewListMembersHandler(repo)

	resp, err := handler.Handle(r.Context(), subdomain)
	if err != nil {
		logger.Error("Failed to list team members", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (m *Module) handleInviteMember(w http.ResponseWriter, r *http.Request) {
	subdomain := middleware.GetSubdomainFromContext(r.Context())

	inviterIDStr, ok := r.Context().Value(authmw.UserIDKey).(string)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	inviterID, err := uuid.Parse(inviterIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	var req invitemember.InviteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	repo := persistence.NewTeamMemberPostgresRepository(m.db)
	handler := invitemember.NewInviteMemberHandler(repo)

	resp, err := handler.Handle(r.Context(), subdomain, inviterID, &req)
	if err != nil {
		switch err {
		case invitemember.ErrUserNotFound:
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case invitemember.ErrAlreadyMember:
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		case invitemember.ErrOrganizationNotFound:
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		default:
			logger.Error("Failed to invite team member", zap.Error(err))
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (m *Module) handleUpdateMember(w http.ResponseWriter, r *http.Request) {
	memberIDStr := chi.URLParam(r, "member_id")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid member id"})
		return
	}

	var req updatemember.UpdateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	repo := persistence.NewTeamMemberPostgresRepository(m.db)
	handler := updatemember.NewUpdateMemberHandler(repo)

	if err := handler.Handle(r.Context(), memberID, &req); err != nil {
		switch err {
		case updatemember.ErrMemberNotFound:
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case updatemember.ErrCannotUpdateOwnerRole:
			writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		default:
			logger.Error("Failed to update team member", zap.Error(err))
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "member updated"})
}

func (m *Module) handleRemoveMember(w http.ResponseWriter, r *http.Request) {
	memberIDStr := chi.URLParam(r, "member_id")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid member id"})
		return
	}

	requesterIDStr, ok := r.Context().Value(authmw.UserIDKey).(string)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	requesterID, err := uuid.Parse(requesterIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid requester id"})
		return
	}

	repo := persistence.NewTeamMemberPostgresRepository(m.db)
	handler := removemember.NewRemoveMemberHandler(repo)

	if err := handler.Handle(r.Context(), memberID, requesterID); err != nil {
		switch err {
		case removemember.ErrMemberNotFound:
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case removemember.ErrCannotRemoveOwner:
			writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		case removemember.ErrCannotRemoveSelf:
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			logger.Error("Failed to remove team member", zap.Error(err))
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set(contentTypeHeader, applicationJSON)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
