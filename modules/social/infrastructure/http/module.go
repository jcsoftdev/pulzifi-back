// Package socialhttp provides the HTTP module registerer for the social module.
// Routes are registered only when SOCIAL_ENABLED=true (REQ-FLAG-01).
package socialhttp

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	createprofile "github.com/jcsoftdev/pulzifi-back/modules/social/application/create_profile"
	deleteprofile "github.com/jcsoftdev/pulzifi-back/modules/social/application/delete_profile"
	getchange "github.com/jcsoftdev/pulzifi-back/modules/social/application/get_change"
	getprofile "github.com/jcsoftdev/pulzifi-back/modules/social/application/get_profile"
	getquotastatus "github.com/jcsoftdev/pulzifi-back/modules/social/application/get_quota_status"
	listchanges "github.com/jcsoftdev/pulzifi-back/modules/social/application/list_changes"
	listprofiles "github.com/jcsoftdev/pulzifi-back/modules/social/application/list_profiles"
	runcheck "github.com/jcsoftdev/pulzifi-back/modules/social/application/run_check"
	updateprofile "github.com/jcsoftdev/pulzifi-back/modules/social/application/update_profile"
	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/social/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/social/infrastructure/persistence/postgres"
	sharedhttp "github.com/jcsoftdev/pulzifi-back/shared/http"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
)

// Compile-time interface check.
var _ router.ModuleRegisterer = (*Module)(nil)

// Deps contains all external dependencies for the Social module HTTP layer.
// Cross-module concerns (alert creator, plan limits) are provided via domain ports
// implemented by cmd/wiring/social adapters.
type Deps struct {
	DB          *sql.DB
	AlertCreator services.AlertCreator
	PlanLimits   services.PlanLimits
	Fetcher      services.SocialFetcher
	MediaStore   services.MediaStore
	Enabled      bool // SOCIAL_ENABLED

	// PostsPerCheck is the number of posts fetched per check (SOCIAL_POSTS_PER_CHECK).
	PostsPerCheck int

	// ChecksPerDay is the plan-level daily limit; 0 = unlimited.
	// For the HTTP manual-check trigger, pass 0 here — the run_check handler
	// resolves the real limit via PlanLimits.GetChecksPerDay internally.
	ChecksPerDay int
}

// Module implements router.ModuleRegisterer for the Social module.
type Module struct {
	db           *sql.DB
	alertCreator services.AlertCreator
	planLimits   services.PlanLimits
	fetcher      services.SocialFetcher
	mediaStore   services.MediaStore
	enabled      bool
	postsPerCheck int
	checksPerDay  int
}

// NewModule creates a new Social HTTP module.
func NewModule(deps Deps) *Module {
	return &Module{
		db:            deps.DB,
		alertCreator:  deps.AlertCreator,
		planLimits:    deps.PlanLimits,
		fetcher:       deps.Fetcher,
		mediaStore:    deps.MediaStore,
		enabled:       deps.Enabled,
		postsPerCheck: deps.PostsPerCheck,
		checksPerDay:  deps.ChecksPerDay,
	}
}

// ModuleName returns the name of the module.
func (m *Module) ModuleName() string {
	return "Social"
}

// RegisterHTTPRoutes registers all 9 social HTTP routes (REQ-HTTP-01 through REQ-HTTP-10).
// When SOCIAL_ENABLED=false, no routes are registered (REQ-FLAG-01, REQ-FLAG-02).
func (m *Module) RegisterHTTPRoutes(r chi.Router) {
	if !m.enabled {
		return
	}

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware.Authenticate)
		r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)
		r.Use(middleware.RequireTenant)

		// Workspace-scoped routes (REQ-HTTP-01, REQ-HTTP-02, REQ-HTTP-11)
		r.Post("/workspaces/{workspaceID}/social-profiles", m.handleCreateProfile)
		r.Get("/workspaces/{workspaceID}/social-profiles", m.handleListProfiles)
		r.Get("/workspaces/{workspaceID}/social-quota", m.handleGetQuotaStatus)
		r.Get("/workspaces/{workspaceID}/social-changes", m.handleListWorkspaceChanges)

		// Profile-scoped routes (REQ-HTTP-03 through REQ-HTTP-08)
		r.Get("/social-profiles/{id}", m.handleGetProfile)
		r.Patch("/social-profiles/{id}", m.handleUpdateProfile)
		r.Delete("/social-profiles/{id}", m.handleDeleteProfile)
		r.Post("/social-profiles/{id}/check", m.handleRunCheck)
		r.Get("/social-profiles/{id}/changes", m.handleListChanges)

		// Change detail (REQ-HTTP-09, REQ-HTTP-10)
		r.Get("/social-changes/{changeID}", m.handleGetChange)
	})
}

// --- Route handlers ---

// handleCreateProfile creates a new social profile.
// @Summary Create Social Profile
// @Description Add a new social profile to a workspace for monitoring
// @Tags social
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param workspaceID path string true "Workspace ID"
// @Param request body createprofile.Request true "Create Profile Request"
// @Success 201 {object} createprofile.Response
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /workspaces/{workspaceID}/social-profiles [post]
func (m *Module) handleCreateProfile(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := parsePathUUID(r, "workspaceID")
	if err != nil {
		sharedhttp.RespondError(w, http.StatusBadRequest, "invalid workspaceID")
		return
	}

	var req createprofile.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedhttp.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.WorkspaceID = workspaceID

	tenant := middleware.GetTenantFromContext(r.Context())
	profileRepo := postgres.NewProfilePostgresRepository(m.db, tenant)

	handler := createprofile.NewHandler(profileRepo, m.planLimits)
	resp, err := handler.Handle(r.Context(), &req)
	if err != nil {
		mapDomainError(w, err)
		return
	}

	sharedhttp.RespondJSON(w, http.StatusCreated, resp)
}

// handleListProfiles lists social profiles for a workspace.
// @Summary List Social Profiles
// @Description List all social profiles for a workspace
// @Tags social
// @Security BearerAuth
// @Produce json
// @Param workspaceID path string true "Workspace ID"
// @Success 200 {array} listprofiles.ProfileSummary
// @Failure 400 {object} map[string]string
// @Router /workspaces/{workspaceID}/social-profiles [get]
func (m *Module) handleListProfiles(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := parsePathUUID(r, "workspaceID")
	if err != nil {
		sharedhttp.RespondError(w, http.StatusBadRequest, "invalid workspaceID")
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	profileRepo := postgres.NewProfilePostgresRepository(m.db, tenant)

	handler := listprofiles.NewHandler(profileRepo)
	resp, err := handler.Handle(r.Context(), workspaceID)
	if err != nil {
		sharedhttp.RespondError(w, http.StatusInternalServerError, "failed to list profiles")
		return
	}

	sharedhttp.RespondJSON(w, http.StatusOK, resp)
}

// handleGetProfile retrieves a profile and its latest snapshot.
// @Summary Get Social Profile
// @Description Get a social profile by ID, including its latest snapshot
// @Tags social
// @Security BearerAuth
// @Produce json
// @Param id path string true "Profile ID"
// @Success 200 {object} getprofile.Response
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /social-profiles/{id} [get]
func (m *Module) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathUUID(r, "id")
	if err != nil {
		sharedhttp.RespondError(w, http.StatusBadRequest, "invalid profile id")
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	profileRepo := postgres.NewProfilePostgresRepository(m.db, tenant)
	snapshotRepo := postgres.NewSnapshotPostgresRepository(m.db, tenant)

	handler := getprofile.NewHandler(profileRepo, snapshotRepo)
	resp, err := handler.Handle(r.Context(), id)
	if err != nil {
		mapDomainError(w, err)
		return
	}

	sharedhttp.RespondJSON(w, http.StatusOK, resp)
}

// handleUpdateProfile updates a profile's interval or active status.
// @Summary Update Social Profile
// @Description Update a social profile's check interval or active flag
// @Tags social
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Profile ID"
// @Param request body updateprofile.Request true "Update Profile Request"
// @Success 200 {object} updateprofile.Response
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /social-profiles/{id} [patch]
func (m *Module) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathUUID(r, "id")
	if err != nil {
		sharedhttp.RespondError(w, http.StatusBadRequest, "invalid profile id")
		return
	}

	var req updateprofile.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedhttp.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	profileRepo := postgres.NewProfilePostgresRepository(m.db, tenant)

	handler := updateprofile.NewHandler(profileRepo, m.planLimits)
	resp, err := handler.Handle(r.Context(), id, &req)
	if err != nil {
		mapDomainError(w, err)
		return
	}

	sharedhttp.RespondJSON(w, http.StatusOK, resp)
}

// handleDeleteProfile deletes a profile and its cascade data.
// @Summary Delete Social Profile
// @Description Delete a social profile and all its associated snapshots and changes
// @Tags social
// @Security BearerAuth
// @Param id path string true "Profile ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /social-profiles/{id} [delete]
func (m *Module) handleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathUUID(r, "id")
	if err != nil {
		sharedhttp.RespondError(w, http.StatusBadRequest, "invalid profile id")
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	profileRepo := postgres.NewProfilePostgresRepository(m.db, tenant)

	handler := deleteprofile.NewHandler(profileRepo)
	if err := handler.Handle(r.Context(), id); err != nil {
		mapDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleRunCheck triggers a manual check for a profile.
// @Summary Run Social Check
// @Description Trigger an immediate profile check (fetches latest data and detects changes)
// @Tags social
// @Security BearerAuth
// @Produce json
// @Param id path string true "Profile ID"
// @Success 200 {object} runcheck.Response
// @Failure 400 {object} map[string]string
// @Failure 402 {object} map[string]string "Daily quota exceeded"
// @Failure 404 {object} map[string]string
// @Router /social-profiles/{id}/check [post]
func (m *Module) handleRunCheck(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathUUID(r, "id")
	if err != nil {
		sharedhttp.RespondError(w, http.StatusBadRequest, "invalid profile id")
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	profileRepo := postgres.NewProfilePostgresRepository(m.db, tenant)
	snapshotRepo := postgres.NewSnapshotPostgresRepository(m.db, tenant)
	changeRepo := postgres.NewChangePostgresRepository(m.db, tenant)
	quotaRepo := postgres.NewCheckQuotaPostgresRepository(m.db, tenant)

	handler := runcheck.NewHandler(
		profileRepo,
		snapshotRepo,
		changeRepo,
		quotaRepo,
		m.fetcher,
		m.mediaStore,
		m.alertCreator,
		m.postsPerCheck,
		m.checksPerDay,
	)

	resp, err := handler.Handle(r.Context(), id)
	if err != nil {
		mapDomainError(w, err)
		return
	}

	sharedhttp.RespondJSON(w, http.StatusOK, resp)
}

// handleListChanges lists changes for a profile or workspace.
// @Summary List Profile Changes
// @Description List social changes for a specific profile, paginated
// @Tags social
// @Security BearerAuth
// @Produce json
// @Param id path string true "Profile ID"
// @Param limit query int false "Page size (default 20)"
// @Param offset query int false "Page offset (default 0)"
// @Success 200 {object} listchanges.Response
// @Failure 400 {object} map[string]string
// @Router /social-profiles/{id}/changes [get]
func (m *Module) handleListChanges(w http.ResponseWriter, r *http.Request) {
	profileID, err := parsePathUUID(r, "id")
	if err != nil {
		sharedhttp.RespondError(w, http.StatusBadRequest, "invalid profile id")
		return
	}

	q := r.URL.Query()
	limit := parseIntQuery(q.Get("limit"), 20)
	offset := parseIntQuery(q.Get("offset"), 0)

	req := &listchanges.Request{
		ProfileID: &profileID,
		Limit:     limit,
		Offset:    offset,
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	changeRepo := postgres.NewChangePostgresRepository(m.db, tenant)

	handler := listchanges.NewHandler(changeRepo)
	resp, err := handler.Handle(r.Context(), req)
	if err != nil {
		sharedhttp.RespondError(w, http.StatusInternalServerError, "failed to list changes")
		return
	}

	sharedhttp.RespondJSON(w, http.StatusOK, resp)
}

// handleListWorkspaceChanges lists all social changes for a workspace (across all profiles).
// GET /workspaces/{workspaceID}/social-changes
// Satisfies: REQ-FE-05 — feeds the unified changes-view feed on the frontend.
// @Summary List Workspace Social Changes
// @Description List all social changes across all profiles in a workspace, paginated
// @Tags social
// @Security BearerAuth
// @Produce json
// @Param workspaceID path string true "Workspace ID"
// @Param limit query int false "Page size (default 50)"
// @Param offset query int false "Page offset (default 0)"
// @Success 200 {object} listchanges.Response
// @Failure 400 {object} map[string]string
// @Router /workspaces/{workspaceID}/social-changes [get]
func (m *Module) handleListWorkspaceChanges(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := parsePathUUID(r, "workspaceID")
	if err != nil {
		sharedhttp.RespondError(w, http.StatusBadRequest, "invalid workspace id")
		return
	}

	q := r.URL.Query()
	limit := parseIntQuery(q.Get("limit"), 50)
	offset := parseIntQuery(q.Get("offset"), 0)

	req := &listchanges.Request{
		WorkspaceID: &workspaceID,
		Limit:       limit,
		Offset:      offset,
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	changeRepo := postgres.NewChangePostgresRepository(m.db, tenant)

	handler := listchanges.NewHandler(changeRepo)
	resp, err := handler.Handle(r.Context(), req)
	if err != nil {
		sharedhttp.RespondError(w, http.StatusInternalServerError, "failed to list workspace social changes")
		return
	}

	sharedhttp.RespondJSON(w, http.StatusOK, resp)
}

// handleGetChange retrieves a single change with full before/after payload.
// @Summary Get Social Change
// @Description Get a social change by ID with full before/after snapshot data
// @Tags social
// @Security BearerAuth
// @Produce json
// @Param changeID path string true "Change ID"
// @Success 200 {object} getchange.Response
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /social-changes/{changeID} [get]
func (m *Module) handleGetChange(w http.ResponseWriter, r *http.Request) {
	changeID, err := parsePathUUID(r, "changeID")
	if err != nil {
		sharedhttp.RespondError(w, http.StatusBadRequest, "invalid change id")
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	changeRepo := postgres.NewChangePostgresRepository(m.db, tenant)
	snapshotRepo := postgres.NewSnapshotPostgresRepository(m.db, tenant)

	handler := getchange.NewHandler(changeRepo, snapshotRepo)
	resp, err := handler.Handle(r.Context(), changeID)
	if err != nil {
		mapDomainError(w, err)
		return
	}
	if resp == nil {
		sharedhttp.RespondError(w, http.StatusNotFound, "change not found")
		return
	}

	sharedhttp.RespondJSON(w, http.StatusOK, resp)
}

// handleGetQuotaStatus returns quota usage for the workspace.
// @Summary Get Social Quota Status
// @Description Get the daily check quota usage for a workspace
// @Tags social
// @Security BearerAuth
// @Produce json
// @Param workspaceID path string true "Workspace ID"
// @Success 200 {object} getquotastatus.Response
// @Failure 400 {object} map[string]string
// @Router /workspaces/{workspaceID}/social-quota [get]
func (m *Module) handleGetQuotaStatus(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := parsePathUUID(r, "workspaceID")
	if err != nil {
		sharedhttp.RespondError(w, http.StatusBadRequest, "invalid workspaceID")
		return
	}

	handler := getquotastatus.NewHandler(m.planLimits)
	resp, err := handler.Handle(r.Context(), workspaceID)
	if err != nil {
		sharedhttp.RespondError(w, http.StatusInternalServerError, "failed to get quota status")
		return
	}

	sharedhttp.RespondJSON(w, http.StatusOK, resp)
}

// --- Helpers ---

// mapDomainError converts known domain errors to HTTP status codes.
func mapDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domainerrors.ErrProfileNotFound):
		sharedhttp.RespondError(w, http.StatusNotFound, "profile not found")
	case errors.Is(err, domainerrors.ErrProfileAlreadyExists):
		sharedhttp.RespondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domainerrors.ErrProfileLimitReached):
		sharedhttp.RespondError(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, domainerrors.ErrQuotaExceeded):
		sharedhttp.RespondError(w, http.StatusPaymentRequired, "daily check quota exceeded")
	case errors.Is(err, domainerrors.ErrPlatformNotSupported):
		sharedhttp.RespondError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domainerrors.ErrFetchFailed):
		sharedhttp.RespondError(w, http.StatusBadGateway, "provider fetch failed")
	default:
		sharedhttp.RespondError(w, http.StatusInternalServerError, err.Error())
	}
}

// parsePathUUID parses a URL path parameter as a UUID.
func parsePathUUID(r *http.Request, param string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, param))
}

// parseIntQuery parses an integer query parameter with a default value.
func parseIntQuery(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 0 {
		return defaultVal
	}
	return v
}

// ensure repositories interfaces are satisfied (checked at runtime wiring).
var (
	_ repositories.ProfileRepository  = (*postgres.ProfilePostgresRepository)(nil)
	_ repositories.SnapshotRepository = (*postgres.SnapshotPostgresRepository)(nil)
	_ repositories.ChangeRepository   = (*postgres.ChangePostgresRepository)(nil)
	_ services.CheckQuota             = (*postgres.CheckQuotaPostgresRepository)(nil)
)
