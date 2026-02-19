package http

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	authmw "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/middleware"
	add_workspace_member "github.com/jcsoftdev/pulzifi-back/modules/workspace/application/add_workspace_member"
	createworkspace "github.com/jcsoftdev/pulzifi-back/modules/workspace/application/create_workspace"
	delete_workspace "github.com/jcsoftdev/pulzifi-back/modules/workspace/application/delete_workspace"
	getworkspace "github.com/jcsoftdev/pulzifi-back/modules/workspace/application/get_workspace"
	list_workspace_members "github.com/jcsoftdev/pulzifi-back/modules/workspace/application/list_workspace_members"
	listworkspaces "github.com/jcsoftdev/pulzifi-back/modules/workspace/application/list_workspaces"
	remove_workspace_member "github.com/jcsoftdev/pulzifi-back/modules/workspace/application/remove_workspace_member"
	update_member_role "github.com/jcsoftdev/pulzifi-back/modules/workspace/application/update_member_role"
	updateworkspace "github.com/jcsoftdev/pulzifi-back/modules/workspace/application/update_workspace"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/value_objects"
	workspacemw "github.com/jcsoftdev/pulzifi-back/modules/workspace/infrastructure/middleware"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"go.uber.org/zap"
)

// HTTP constants
const (
	contentTypeHeader = "Content-Type"
	applicationJSON   = "application/json"
)

// Error messages
const (
	errInvalidWorkspaceID     = "invalid workspace ID"
	errInvalidUserID          = "invalid user ID"
	errWorkspaceNotFound      = "workspace not found"
	errDatabaseNotInitialized = "database not initialized"
	errNotWorkspaceMember     = "you are not a member of this workspace"
	errInternalServerError    = "internal server error"
)

// Module implements the router.ModuleRegisterer interface for the Workspace module
type Module struct {
	db *sql.DB
}

// NewModule creates a new instance of the Workspace module
func NewModule() router.ModuleRegisterer {
	return &Module{}
}

// NewModuleWithDB creates a new instance with database connection
func NewModuleWithDB(db *sql.DB) router.ModuleRegisterer {
	return &Module{
		db: db,
	}
}

// ModuleName returns the name of the module
func (m *Module) ModuleName() string {
	return "Workspace"
}

// RegisterHTTPRoutes registers all HTTP routes for the Workspace module
func (m *Module) RegisterHTTPRoutes(router chi.Router) {
	// Create workspace-specific authorization middleware
	workspaceAuth := workspacemw.NewWorkspaceAuthorizationMiddleware(m.db)

	router.Route("/workspaces", func(r chi.Router) {
		// ========================================
		// LEVEL 1: Global Authentication & Permissions
		// ========================================
		r.Use(middleware.AuthMiddleware.Authenticate)
		r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)
		r.Use(middleware.RequireTenant)

		// Public workspace endpoints (require global permission only)
		r.Group(func(r chi.Router) {
			// Level 1: Require global permission to use workspaces
			r.Use(middleware.AuthMiddleware.RequirePermission("workspaces", "write"))
			r.Post("/", m.handleCreateWorkspace)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware.RequirePermission("workspaces", "read"))
			r.Get("/", m.handleListWorkspaces)
		})

		// ========================================
		// LEVEL 2: Workspace-Specific Authorization
		// ========================================
		// These endpoints require membership in the specific workspace
		r.Group(func(r chi.Router) {
			// Level 1: Global permission check
			r.Use(middleware.AuthMiddleware.RequirePermission("workspaces", "read"))
			// Level 2: Workspace membership check (any role: viewer, editor, owner)
			r.Use(workspaceAuth.RequireWorkspaceMembership)

			r.Get("/{id}", m.handleGetWorkspace)
			r.Get("/{id}/members", m.handleListWorkspaceMembers)
		})

		r.Group(func(r chi.Router) {
			// Level 1: Global permission
			r.Use(middleware.AuthMiddleware.RequirePermission("workspaces", "write"))
			// Level 2: Workspace membership + minimum role = editor
			r.Use(workspaceAuth.RequireWorkspaceMembership)
			r.Use(workspaceAuth.RequireWorkspaceRole(value_objects.RoleEditor))

			r.Put("/{id}", m.handleUpdateWorkspace)
		})

		r.Group(func(r chi.Router) {
			// Level 1: Global permission
			r.Use(middleware.AuthMiddleware.RequirePermission("workspaces", "delete"))
			// Level 2: Workspace membership + minimum role = owner
			r.Use(workspaceAuth.RequireWorkspaceMembership)
			r.Use(workspaceAuth.RequireWorkspaceRole(value_objects.RoleOwner))

			r.Delete("/{id}", m.handleDeleteWorkspace)
		})

		// Member management (only owners)
		r.Group(func(r chi.Router) {
			// Level 1: Global permission
			r.Use(middleware.AuthMiddleware.RequirePermission("workspaces", "write"))
			// Level 2: Workspace membership + owner role
			r.Use(workspaceAuth.RequireWorkspaceMembership)
			r.Use(workspaceAuth.RequireWorkspaceRole(value_objects.RoleOwner))

			r.Post("/{id}/members", m.handleAddWorkspaceMember)
			r.Put("/{id}/members/{user_id}", m.handleUpdateMemberRole)
			r.Delete("/{id}/members/{user_id}", m.handleRemoveWorkspaceMember)
		})
	})
}

// handleCreateWorkspace creates a new workspace
// @Summary Create Workspace
// @Description Create a new workspace
// @Tags workspaces
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body createworkspace.CreateWorkspaceRequest true "Create Workspace Request"
// @Success 201 {object} createworkspace.CreateWorkspaceResponse
// @Router /workspaces [post]
func (m *Module) handleCreateWorkspace(w http.ResponseWriter, r *http.Request) {
	// If db is not available, return mock response
	if m.db == nil {
		w.Header().Set(contentTypeHeader, applicationJSON)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "550e8400-e29b-41d4-a716-446655440000",
			"message": "create workspace (mock - db not initialized)",
		})
		return
	}

	// Get tenant from context
	tenant := middleware.GetTenantFromContext(r.Context())

	// Create repository with dynamic tenant
	repo := persistence.NewWorkspacePostgresRepository(m.db, tenant)

	// Use real handler
	handler := createworkspace.NewCreateWorkspaceHandler(repo)
	handler.HandleHTTP(w, r)
}

// handleListWorkspaces lists all workspaces
// @Summary List Workspaces
// @Description List all workspaces
// @Tags workspaces
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit number of results (default: all)"
// @Success 200 {object} listworkspaces.ListWorkspacesResponse
// @Failure 401 {object} map[string]string
// @Router /workspaces [get]
func (m *Module) handleListWorkspaces(w http.ResponseWriter, r *http.Request) {
	// If db is not available, return empty response
	if m.db == nil {
		w.Header().Set(contentTypeHeader, applicationJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"workspaces": []interface{}{},
		})
		return
	}

	// Get tenant from context
	tenant := middleware.GetTenantFromContext(r.Context())

	// Create repository with dynamic tenant
	repo := persistence.NewWorkspacePostgresRepository(m.db, tenant)

	// Use handler
	handler := listworkspaces.NewListWorkspacesHandler(repo)
	response, err := handler.Handle(r.Context())
	if err != nil {
		w.Header().Set(contentTypeHeader, applicationJSON)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	// Get limit from query parameter
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			if len(response.Workspaces) > limit {
				response.Workspaces = response.Workspaces[:limit]
			}
		}
	}

	w.Header().Set(contentTypeHeader, applicationJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleGetWorkspace gets a workspace by ID
// @Summary Get Workspace
// @Description Get a workspace by ID
// @Tags workspaces
// @Security BearerAuth
// @Produce json
// @Param id path string true "Workspace ID"
// @Success 200 {object} getworkspace.GetWorkspaceResponse
// @Failure 404 {object} map[string]string
// @Router /workspaces/{id} [get]
func (m *Module) handleGetWorkspace(w http.ResponseWriter, r *http.Request) {
	// If db is not available, return mock response
	if m.db == nil {
		w.Header().Set(contentTypeHeader, applicationJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"id":      chi.URLParam(r, "id"),
			"message": "get workspace (mock - db not initialized)",
		})
		return
	}

	// Get workspace ID from URL
	workspaceIDStr := chi.URLParam(r, "id")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		w.Header().Set(contentTypeHeader, applicationJSON)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": errInvalidWorkspaceID,
		})
		return
	}

	// Get tenant from context
	tenant := middleware.GetTenantFromContext(r.Context())

	// Create repository with dynamic tenant
	repo := persistence.NewWorkspacePostgresRepository(m.db, tenant)

	// Use real handler
	handler := getworkspace.NewGetWorkspaceHandler(repo)
	response, err := handler.Handle(r.Context(), workspaceID)
	if err != nil {
		if err == getworkspace.ErrWorkspaceNotFound {
			w.Header().Set(contentTypeHeader, applicationJSON)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": errWorkspaceNotFound,
			})
			return
		}

		logger.Error("Failed to get workspace", zap.Error(err))
		w.Header().Set(contentTypeHeader, applicationJSON)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": errInternalServerError,
		})
		return
	}

	w.Header().Set(contentTypeHeader, applicationJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleUpdateWorkspace updates a workspace
// @Summary Update Workspace
// @Description Update a workspace
// @Tags workspaces
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Workspace ID"
// @Param request body map[string]string true "Update Workspace Request"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /workspaces/{id} [put]
func (m *Module) handleUpdateWorkspace(w http.ResponseWriter, r *http.Request) {
	// If db is not available, return mock response
	if m.db == nil {
		w.Header().Set(contentTypeHeader, applicationJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"id":      chi.URLParam(r, "id"),
			"message": "update workspace",
		})
		return
	}

	// Get tenant from context
	tenant := middleware.GetTenantFromContext(r.Context())

	// Create repository with dynamic tenant
	repo := persistence.NewWorkspacePostgresRepository(m.db, tenant)

	// Use real handler
	handler := updateworkspace.NewUpdateWorkspaceHandler(repo)
	handler.HandleHTTP(w, r)
}

// handleDeleteWorkspace deletes a workspace
// @Summary Delete Workspace
// @Description Delete a workspace
// @Tags workspaces
// @Security BearerAuth
// @Produce json
// @Param id path string true "Workspace ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Router /workspaces/{id} [delete]
func (m *Module) handleDeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	// If db is not available, return mock response
	if m.db == nil {
		w.Header().Set(contentTypeHeader, applicationJSON)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Get workspace ID from URL
	workspaceIDStr := chi.URLParam(r, "id")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		w.Header().Set(contentTypeHeader, applicationJSON)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": errInvalidWorkspaceID,
		})
		return
	}

	// Get tenant and user ID from context
	tenant := middleware.GetTenantFromContext(r.Context())

	userIDStr, ok := r.Context().Value(authmw.UserIDKey).(string)
	if !ok {
		logger.Error("User ID not found in context")
		w.Header().Set(contentTypeHeader, applicationJSON)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "unauthorized",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Error("Invalid user ID", zap.Error(err))
		w.Header().Set(contentTypeHeader, applicationJSON)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": errInvalidUserID,
		})
		return
	}

	// Create repository with dynamic tenant
	repo := persistence.NewWorkspacePostgresRepository(m.db, tenant)

	// Create handler
	handler := delete_workspace.NewDeleteWorkspaceHandler(repo, m.db)

	// Execute delete
	if err := handler.Handle(r.Context(), workspaceID, userID); err != nil {
		if err == delete_workspace.ErrWorkspaceNotFound {
			w.Header().Set(contentTypeHeader, applicationJSON)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": errWorkspaceNotFound,
			})
			return
		}

		if err == delete_workspace.ErrWorkspaceNotOwned {
			w.Header().Set(contentTypeHeader, applicationJSON)
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "you do not own this workspace",
			})
			return
		}

		w.Header().Set(contentTypeHeader, applicationJSON)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "failed to delete workspace",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleAddWorkspaceMember adds a member to a workspace
// @Summary Add Workspace Member
// @Description Add a user to a workspace
// @Tags workspaces
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Workspace ID"
// @Param request body add_workspace_member.AddWorkspaceMemberRequest true "Add Member Request"
// @Success 201 {object} add_workspace_member.AddWorkspaceMemberResponse
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /workspaces/{id}/members [post]
func (m *Module) handleAddWorkspaceMember(w http.ResponseWriter, r *http.Request) {
	if m.db == nil {
		http.Error(w, errDatabaseNotInitialized, http.StatusInternalServerError)
		return
	}

	// Get workspace ID from URL
	workspaceIDStr := chi.URLParam(r, "id")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		http.Error(w, errInvalidWorkspaceID, http.StatusBadRequest)
		return
	}

	// Get inviter ID from context
	inviterIDStr, ok := r.Context().Value(authmw.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	inviterID, err := uuid.Parse(inviterIDStr)
	if err != nil {
		http.Error(w, errInvalidUserID, http.StatusBadRequest)
		return
	}

	// Parse request
	var req add_workspace_member.AddWorkspaceMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Get tenant and create repository
	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewWorkspacePostgresRepository(m.db, tenant)

	// Execute handler
	handler := add_workspace_member.NewAddWorkspaceMemberHandler(repo)
	resp, err := handler.Handle(r.Context(), workspaceID, inviterID, &req)
	if err != nil {
		switch err {
		case add_workspace_member.ErrWorkspaceNotFound:
			http.Error(w, errWorkspaceNotFound, http.StatusNotFound)
		case add_workspace_member.ErrNotWorkspaceMember:
			http.Error(w, errNotWorkspaceMember, http.StatusForbidden)
		case add_workspace_member.ErrInsufficientPermissions:
			http.Error(w, "insufficient permissions", http.StatusForbidden)
		case add_workspace_member.ErrMemberAlreadyExists:
			http.Error(w, "user is already a member", http.StatusBadRequest)
		default:
			logger.Error("Failed to add workspace member", zap.Error(err))
			http.Error(w, errInternalServerError, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set(contentTypeHeader, applicationJSON)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// handleListWorkspaceMembers lists all members of a workspace
// @Summary List Workspace Members
// @Description List all members of a workspace
// @Tags workspaces
// @Security BearerAuth
// @Produce json
// @Param id path string true "Workspace ID"
// @Success 200 {object} list_workspace_members.ListWorkspaceMembersResponse
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /workspaces/{id}/members [get]
func (m *Module) handleListWorkspaceMembers(w http.ResponseWriter, r *http.Request) {
	if m.db == nil {
		http.Error(w, errDatabaseNotInitialized, http.StatusInternalServerError)
		return
	}

	// Get workspace ID from URL
	workspaceIDStr := chi.URLParam(r, "id")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		http.Error(w, errInvalidWorkspaceID, http.StatusBadRequest)
		return
	}

	// Get requester ID from context
	requesterIDStr, ok := r.Context().Value(authmw.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	requesterID, err := uuid.Parse(requesterIDStr)
	if err != nil {
		http.Error(w, errInvalidUserID, http.StatusBadRequest)
		return
	}

	// Get tenant and create repository
	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewWorkspacePostgresRepository(m.db, tenant)

	// Execute handler
	handler := list_workspace_members.NewListWorkspaceMembersHandler(repo)
	resp, err := handler.Handle(r.Context(), workspaceID, requesterID)
	if err != nil {
		switch err {
		case list_workspace_members.ErrWorkspaceNotFound:
			http.Error(w, errWorkspaceNotFound, http.StatusNotFound)
		case list_workspace_members.ErrNotWorkspaceMember:
			http.Error(w, errNotWorkspaceMember, http.StatusForbidden)
		default:
			logger.Error("Failed to list workspace members", zap.Error(err))
			http.Error(w, errInternalServerError, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set(contentTypeHeader, applicationJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// handleRemoveWorkspaceMember removes a member from a workspace
// @Summary Remove Workspace Member
// @Description Remove a user from a workspace
// @Tags workspaces
// @Security BearerAuth
// @Produce json
// @Param id path string true "Workspace ID"
// @Param user_id path string true "User ID to remove"
// @Success 204
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /workspaces/{id}/members/{user_id} [delete]
func (m *Module) handleRemoveWorkspaceMember(w http.ResponseWriter, r *http.Request) {
	if m.db == nil {
		http.Error(w, errDatabaseNotInitialized, http.StatusInternalServerError)
		return
	}

	// Get workspace ID from URL
	workspaceIDStr := chi.URLParam(r, "id")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		http.Error(w, errInvalidWorkspaceID, http.StatusBadRequest)
		return
	}

	// Get member to remove ID from URL
	memberToRemoveIDStr := chi.URLParam(r, "user_id")
	memberToRemoveID, err := uuid.Parse(memberToRemoveIDStr)
	if err != nil {
		http.Error(w, errInvalidUserID, http.StatusBadRequest)
		return
	}

	// Get requester ID from context
	requesterIDStr, ok := r.Context().Value(authmw.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	requesterID, err := uuid.Parse(requesterIDStr)
	if err != nil {
		http.Error(w, errInvalidUserID, http.StatusBadRequest)
		return
	}

	// Get tenant and create repository
	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewWorkspacePostgresRepository(m.db, tenant)

	// Execute handler
	handler := remove_workspace_member.NewRemoveWorkspaceMemberHandler(repo)
	err = handler.Handle(r.Context(), workspaceID, requesterID, memberToRemoveID)
	if err != nil {
		switch err {
		case remove_workspace_member.ErrWorkspaceNotFound:
			http.Error(w, errWorkspaceNotFound, http.StatusNotFound)
		case remove_workspace_member.ErrNotWorkspaceMember:
			http.Error(w, errNotWorkspaceMember, http.StatusForbidden)
		case remove_workspace_member.ErrInsufficientPermissions:
			http.Error(w, "insufficient permissions", http.StatusForbidden)
		case remove_workspace_member.ErrCannotRemoveOwner:
			http.Error(w, "cannot remove the owner from workspace", http.StatusBadRequest)
		case remove_workspace_member.ErrCannotRemoveSelf:
			http.Error(w, "cannot remove yourself from workspace", http.StatusBadRequest)
		case remove_workspace_member.ErrMemberNotFound:
			http.Error(w, "member not found in workspace", http.StatusNotFound)
		default:
			logger.Error("Failed to remove workspace member", zap.Error(err))
			http.Error(w, errInternalServerError, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleUpdateMemberRole updates a member's role in a workspace
// @Summary Update Workspace Member Role
// @Description Update a member's role in a workspace
// @Tags workspaces
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Workspace ID"
// @Param user_id path string true "User ID to update"
// @Param request body update_member_role.UpdateMemberRoleRequest true "Update Member Role Request"
// @Success 200
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /workspaces/{id}/members/{user_id} [put]
func (m *Module) handleUpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	if m.db == nil {
		http.Error(w, errDatabaseNotInitialized, http.StatusInternalServerError)
		return
	}

	// Get workspace ID from URL
	workspaceIDStr := chi.URLParam(r, "id")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		http.Error(w, errInvalidWorkspaceID, http.StatusBadRequest)
		return
	}

	// Get target user ID from URL
	targetUserIDStr := chi.URLParam(r, "user_id")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		http.Error(w, errInvalidUserID, http.StatusBadRequest)
		return
	}

	// Get requester ID from context
	requesterIDStr, ok := r.Context().Value(authmw.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	requesterID, err := uuid.Parse(requesterIDStr)
	if err != nil {
		http.Error(w, errInvalidUserID, http.StatusBadRequest)
		return
	}

	// Parse request
	var req update_member_role.UpdateMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Get tenant and create repository
	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewWorkspacePostgresRepository(m.db, tenant)

	// Execute handler
	handler := update_member_role.NewUpdateMemberRoleHandler(repo)
	err = handler.Handle(r.Context(), workspaceID, requesterID, targetUserID, &req)
	if err != nil {
		switch err {
		case update_member_role.ErrWorkspaceNotFound:
			http.Error(w, errWorkspaceNotFound, http.StatusNotFound)
		case update_member_role.ErrNotWorkspaceMember:
			http.Error(w, errNotWorkspaceMember, http.StatusForbidden)
		case update_member_role.ErrInsufficientPermissions:
			http.Error(w, "insufficient permissions", http.StatusForbidden)
		case update_member_role.ErrInvalidRole:
			http.Error(w, "invalid role", http.StatusBadRequest)
		case update_member_role.ErrCannotChangeOwnRole:
			http.Error(w, "cannot change your own role", http.StatusBadRequest)
		default:
			logger.Error("Failed to update workspace member role", zap.Error(err))
			http.Error(w, errInternalServerError, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
