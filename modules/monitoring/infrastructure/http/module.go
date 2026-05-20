package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	bulkupdatemonitoringconfig "github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/bulk_update_monitoring_config"
	createcheck "github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/create_check"
	createmonitoringconfig "github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/create_monitoring_config"
	createnotificationpreference "github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/create_notification_preference"
	getmonitoringconfig "github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/get_monitoring_config"
	listchecks "github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/list_checks"
	managesections "github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/manage_sections"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/orchestrator"
	updatemonitoringconfig "github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/update_monitoring_config"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/workers"
	monitoringservices "github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/scheduler"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/pubsub"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

// Deps contains all external dependencies for the Monitoring module.
// Cross-module dependencies (snapshot, email, insight) are provided via
// the SnapshotExecutor port, constructed by cmd/wiring/monitoring.
type Deps struct {
	DB               *sql.DB
	EventBus         *eventbus.EventBus
	SnapshotExecutor monitoringservices.SnapshotExecutor
}

// Module implements the router.ModuleRegisterer interface for the Monitoring module
type Module struct {
	db          *sql.DB
	eventBus    *eventbus.EventBus
	scheduler   *scheduler.Scheduler
	workerPool  *workers.WorkerPool
	checkBroker *pubsub.CheckBroker
}

// NewModule creates a new instance of the Monitoring module
func NewModule() router.ModuleRegisterer {
	return &Module{}
}

// NewModuleWithDeps creates a new instance with all dependencies injected.
// Cross-module concerns (snapshot, insight, email) are provided via deps.SnapshotExecutor.
func NewModuleWithDeps(deps Deps) *Module {
	m := &Module{
		db:       deps.DB,
		eventBus: deps.EventBus,
	}

	// Initialize the check broker for SSE push notifications.
	m.checkBroker = pubsub.NewCheckBroker()

	// Wire OnCheckDone so successful/error completions are pushed to SSE subscribers.
	if deps.SnapshotExecutor != nil {
		deps.SnapshotExecutor.SetOnCheckDone(func(pageID uuid.UUID, checkJSON []byte) {
			m.checkBroker.Publish(pageID.String(), checkJSON)
		})
	}

	// Create WorkerPool with a callback that marks failed checks in the DB.
	failCheck := func(ctx context.Context, checkID uuid.UUID, schemaName string, errMsg string) {
		repo := persistence.NewCheckPostgresRepository(m.db, schemaName)
		check, err := repo.GetByID(ctx, checkID)
		if err != nil || check == nil {
			return
		}
		if check.Status == "success" || check.Status == "error" {
			return // already in a terminal state
		}
		check.Status = "error"
		check.ErrorMessage = errMsg
		if updateErr := repo.Update(ctx, check); updateErr != nil {
			logger.Error("failCheck: failed to update check status", zap.Error(updateErr), zap.String("check_id", checkID.String()))
			return
		}
		// Publish the failed check to SSE subscribers.
		payload, _ := json.Marshal(listchecks.CheckResponse{
			ID:              check.ID,
			PageID:          check.PageID,
			Status:          check.Status,
			ScreenshotURL:   check.ScreenshotURL,
			HTMLSnapshotURL: check.HTMLSnapshotURL,
			ChangeDetected:  check.ChangeDetected,
			ChangeType:      check.ChangeType,
			ErrorMessage:    check.ErrorMessage,
			CheckedAt:       check.CheckedAt,
		})
		m.checkBroker.Publish(check.PageID.String(), payload)
	}
	m.workerPool = workers.NewWorkerPool(deps.SnapshotExecutor, 100, failCheck)

	// In API-only mode we still need immediate dispatch capability when user updates frequency.
	// Start a lightweight in-process worker to consume TriggerPageCheck jobs.
	if os.Getenv("ENABLE_WORKERS") == "false" {
		m.workerPool.Start(1)
		logger.Info("Monitoring inline worker started for immediate API dispatch", zap.Int("concurrency", 1))
	}

	repoFactory := persistence.NewPostgresRepositoryFactory(m.db)
	orch := orchestrator.NewOrchestrator(repoFactory, m.workerPool)

	// Push pending/error check events to SSE subscribers when the orchestrator
	// creates a check record or a dispatch fails.
	orch.SetOnCheckCreated(func(pageID uuid.UUID, checkJSON []byte) {
		m.checkBroker.Publish(pageID.String(), checkJSON)
	})

	// Create Scheduler instance
	m.scheduler = scheduler.NewScheduler(m.db, orch)

	return m
}

// NewModuleWithDB creates a new instance with database connection.
// Deprecated: use NewModuleWithDeps for clean dependency injection.
// Kept for the cmd/worker entry point which does not need snapshot/email/insight.
func NewModuleWithDB(db *sql.DB, eventBus *eventbus.EventBus) router.ModuleRegisterer {
	return NewModuleWithDeps(Deps{
		DB:               db,
		EventBus:         eventBus,
		SnapshotExecutor: nil,
	})
}

// StartBackgroundProcesses initializes and starts the Scheduler, Orchestrator, and Workers
func (m *Module) StartBackgroundProcesses() {
	if m.scheduler == nil || m.workerPool == nil {
		logger.Error("Scheduler or WorkerPool not initialized in Module")
		return
	}

	// Start Worker Pool
	m.workerPool.Start(5)

	// Start Scheduler
	m.scheduler.Start(context.Background())

	logger.Info("Monitoring Scheduler and Orchestrator initialized and started")
}

// ModuleName returns the name of the module
func (m *Module) ModuleName() string {
	return "Monitoring"
}

// RegisterHTTPRoutes registers all HTTP routes for the Monitoring module
func (m *Module) RegisterHTTPRoutes(router chi.Router) {
	router.Route("/monitoring", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware.Authenticate)

		// /checks sub-router: all routes require org+tenant authorization.
		r.Route("/checks", func(cr chi.Router) {
			cr.Use(middleware.OrgMiddleware.RequireOrganizationMembership)
			cr.Use(middleware.RequireTenant)
			cr.Get("/page/{pageId}/stream", m.handleCheckSSE)
			cr.Post("/", m.handleCreateCheck)
			cr.Get("/", m.handleListChecks)
			cr.Get("/{id}", m.handleGetCheck)
			cr.Get("/page/{pageId}", m.handleListChecksByPage)
			cr.Post("/page/{pageId}/run", m.handleRunNow)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)
			r.Use(middleware.RequireTenant)
			r.Route("/configs", func(cr chi.Router) {
				cr.Post("/", m.handleCreateMonitoringConfig)
				cr.Put("/bulk", m.handleBulkUpdateMonitoringConfig)
				cr.Get("/{pageId}", m.handleGetMonitoringConfig)
				cr.Put("/{pageId}", m.handleUpdateMonitoringConfig)
			})
			r.Route("/notification-preferences", func(cr chi.Router) {
				cr.Post("/", m.handleCreateNotificationPreference)
				cr.Get("/{id}", m.handleGetNotificationPreference)
			})
			r.Route("/sections/page/{pageId}", func(cr chi.Router) {
				cr.Get("/", m.handleListSections)
				cr.Post("/", m.handleSaveSections)
				cr.Delete("/{sectionId}", m.handleDeleteSection)
			})
		})
	})
}

// handleCreateCheck creates a new monitoring check
// @Summary Create Monitoring Check
// @Description Create a new monitoring check
// @Tags monitoring
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body createcheck.CreateCheckRequest true "Create Check Request"
// @Success 201 {object} createcheck.CreateCheckResponse
// @Router /monitoring/checks [post]
func (m *Module) handleCreateCheck(w http.ResponseWriter, r *http.Request) {
	// If db is not available, return mock response
	if m.db == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "550e8400-e29b-41d4-a716-446655440000",
			"message": "create check (mock - db not initialized)",
		})
		return
	}

	// Get tenant from context
	tenant := middleware.GetTenantFromContext(r.Context())

	// Create repository with dynamic tenant
	repo := persistence.NewCheckPostgresRepository(m.db, tenant)

	// Use handler with parsed inputs
	var req createcheck.CreateCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.PageID == [16]byte{} || req.Status == "" {
		http.Error(w, "page_id and status are required", http.StatusBadRequest)
		return
	}

	handler := createcheck.NewCreateCheckHandler(repo)
	resp, err := handler.Handle(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// handleListChecks lists monitoring checks for a page
// @Summary List Monitoring Checks
// @Description List monitoring checks filtered by page_id
// @Tags monitoring
// @Security BearerAuth
// @Produce json
// @Param page_id query string true "Page ID"
// @Success 200 {object} listchecks.ListChecksResponse
// @Failure 400 {object} map[string]string
// @Router /monitoring/checks [get]
func (m *Module) handleListChecks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pageIDStr := r.URL.Query().Get("page_id")
	if pageIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "page_id query parameter is required"})
		return
	}
	pageID, err := uuid.Parse(pageIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid page_id format"})
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewCheckPostgresRepository(m.db, tenant)
	handler := listchecks.NewListChecksHandler(repo)

	resp, err := handler.Handle(r.Context(), pageID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to list checks"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// handleListChecksByPage lists monitoring checks for a specific page
// @Summary List Monitoring Checks By Page
// @Description List monitoring checks for a specific page
// @Tags monitoring
// @Security BearerAuth
// @Produce json
// @Param pageId path string true "Page ID"
// @Success 200 {object} listchecks.ListChecksResponse
// @Router /monitoring/checks/page/{pageId} [get]
func (m *Module) handleListChecksByPage(w http.ResponseWriter, r *http.Request) {
	if m.db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	pageIDStr := chi.URLParam(r, "pageId")
	pageID, err := uuid.Parse(pageIDStr)
	if err != nil {
		http.Error(w, "Invalid page ID", http.StatusBadRequest)
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewCheckPostgresRepository(m.db, tenant)
	handler := listchecks.NewListChecksHandler(repo)

	// Optional section_id filter
	var resp *listchecks.ListChecksResponse
	sectionIDStr := r.URL.Query().Get("section_id")
	if sectionIDStr != "" {
		sectionID, err := uuid.Parse(sectionIDStr)
		if err != nil {
			http.Error(w, "Invalid section_id", http.StatusBadRequest)
			return
		}
		resp, err = handler.HandleBySection(r.Context(), pageID, &sectionID)
		if err != nil {
			logger.Error("Failed to list checks by section", zap.Error(err))
			http.Error(w, "Failed to list checks", http.StatusInternalServerError)
			return
		}
	} else {
		resp, err = handler.Handle(r.Context(), pageID)
		if err != nil {
			logger.Error("Failed to list checks", zap.Error(err))
			http.Error(w, "Failed to list checks", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// handleRunNow triggers an immediate monitoring check for a page
// @Summary Run Check Now
// @Description Trigger an immediate monitoring check for a page
// @Tags monitoring
// @Security BearerAuth
// @Param pageId path string true "Page ID"
// @Success 202
// @Failure 400 {object} map[string]string
// @Failure 402 {object} map[string]string
// @Router /monitoring/checks/page/{pageId}/run [post]
func (m *Module) handleRunNow(w http.ResponseWriter, r *http.Request) {
	if m.db == nil || m.scheduler == nil {
		http.Error(w, "Service not initialized", http.StatusInternalServerError)
		return
	}

	pageID, err := uuid.Parse(chi.URLParam(r, "pageId"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid pageId"})
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())

	if err := m.scheduler.TriggerPageCheck(r.Context(), tenant, pageID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, orchestrator.ErrPageNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "page not found"})
			return
		}
		if errors.Is(err, orchestrator.ErrQuotaExceeded) {
			w.WriteHeader(http.StatusPaymentRequired)
			json.NewEncoder(w).Encode(map[string]string{"error": "monthly check quota exceeded"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to trigger check"})
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// handleGetCheck gets a monitoring check by ID
// @Summary Get Monitoring Check
// @Description Get a monitoring check by ID
// @Tags monitoring
// @Security BearerAuth
// @Produce json
// @Param id path string true "Check ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /monitoring/checks/{id} [get]
func (m *Module) handleGetCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid check id"})
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewCheckPostgresRepository(m.db, tenant)

	check, err := repo.GetByID(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to get check"})
		return
	}
	if check == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "check not found"})
		return
	}

	resp := listchecks.CheckResponse{
		ID:              check.ID,
		PageID:          check.PageID,
		SectionID:       check.SectionID,
		ParentCheckID:   check.ParentCheckID,
		Status:          check.Status,
		ScreenshotURL:   check.ScreenshotURL,
		HTMLSnapshotURL: check.HTMLSnapshotURL,
		ChangeDetected:  check.ChangeDetected,
		ChangeType:      check.ChangeType,
		ErrorMessage:    check.ErrorMessage,
		CheckedAt:       check.CheckedAt,
	}

	// If this is a parent check, include its section checks.
	if check.SectionID == nil {
		sectionChecks, err := repo.ListByParentCheckID(r.Context(), check.ID)
		if err == nil && len(sectionChecks) > 0 {
			resp.Sections = make([]*listchecks.CheckResponse, len(sectionChecks))
			for i, sc := range sectionChecks {
				resp.Sections[i] = &listchecks.CheckResponse{
					ID:              sc.ID,
					PageID:          sc.PageID,
					SectionID:       sc.SectionID,
					ParentCheckID:   sc.ParentCheckID,
					Status:          sc.Status,
					ScreenshotURL:   sc.ScreenshotURL,
					HTMLSnapshotURL: sc.HTMLSnapshotURL,
					ChangeDetected:  sc.ChangeDetected,
					ChangeType:      sc.ChangeType,
					ErrorMessage:    sc.ErrorMessage,
					CheckedAt:       sc.CheckedAt,
				}
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// handleCreateMonitoringConfig creates a new monitoring config
// @Summary Create Monitoring Config
// @Description Create a new monitoring config
// @Tags monitoring
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body create_monitoring_config.CreateMonitoringConfigRequest true "Create Config Request"
// @Success 201 {object} create_monitoring_config.CreateMonitoringConfigResponse
// @Router /monitoring/configs [post]
func (m *Module) handleCreateMonitoringConfig(w http.ResponseWriter, r *http.Request) {
	// If db is not available, return mock response
	if m.db == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "550e8400-e29b-41d4-a716-446655440000",
			"message": "create monitoring config (mock - db not initialized)",
		})
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewMonitoringConfigPostgresRepository(m.db, tenant)

	var req createmonitoringconfig.CreateMonitoringConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.PageID == [16]byte{} || req.CheckFrequency == "" {
		http.Error(w, "page_id and check_frequency are required", http.StatusBadRequest)
		return
	}

	handler := createmonitoringconfig.NewCreateMonitoringConfigHandler(repo, m.scheduler)
	resp, err := handler.Handle(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// handleGetMonitoringConfig gets a monitoring config by page ID
// @Summary Get Monitoring Config
// @Description Get a monitoring config by page ID
// @Tags monitoring
// @Security BearerAuth
// @Produce json
// @Param pageId path string true "Page ID"
// @Success 200 {object} map[string]interface{}
// @Router /monitoring/configs/{pageId} [get]
func (m *Module) handleGetMonitoringConfig(w http.ResponseWriter, r *http.Request) {
	if m.db == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"page_id": chi.URLParam(r, "pageId"),
			"message": "get monitoring config (mock - db not initialized)",
		})
		return
	}

	pageIDStr := chi.URLParam(r, "pageId")
	pageID, err := uuid.Parse(pageIDStr)
	if err != nil {
		http.Error(w, "Invalid page ID", http.StatusBadRequest)
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewMonitoringConfigPostgresRepository(m.db, tenant)
	handler := getmonitoringconfig.NewGetMonitoringConfigHandler(repo)

	resp, err := handler.Handle(r.Context(), pageID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if resp == nil {
		http.Error(w, "Monitoring config not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// handleUpdateMonitoringConfig updates or creates a monitoring config by page ID
// @Summary Update or Create Monitoring Config (Upsert)
// @Description Update an existing monitoring config or create a new one if it doesn't exist
// @Tags monitoring
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param pageId path string true "Page ID"
// @Param request body updatemonitoringconfig.UpdateMonitoringConfigRequest true "Update Config Request"
// @Success 200 {object} updatemonitoringconfig.UpdateMonitoringConfigResponse
// @Router /monitoring/configs/{pageId} [put]
func (m *Module) handleUpdateMonitoringConfig(w http.ResponseWriter, r *http.Request) {
	// If db is not available, return mock response
	if m.db == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "550e8400-e29b-41d4-a716-446655440000",
			"message": "update monitoring config (mock - db not initialized)",
		})
		return
	}

	pageIDStr := chi.URLParam(r, "pageId")
	pageID, err := uuid.Parse(pageIDStr)
	if err != nil {
		logger.Error("Invalid page ID", zap.Error(err))
		http.Error(w, "invalid page_id", http.StatusBadRequest)
		return
	}

	var req updatemonitoringconfig.UpdateMonitoringConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewMonitoringConfigPostgresRepository(m.db, tenant)

	handler := updatemonitoringconfig.NewUpdateMonitoringConfigHandler(repo, m.eventBus, tenant, m.scheduler)
	response, err := handler.Handle(r.Context(), pageID, &req)
	if err != nil {
		logger.Error("Failed to update monitoring config", zap.Error(err))
		http.Error(w, "failed to update monitoring config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleBulkUpdateMonitoringConfig updates check frequency for multiple pages at once
// @Summary Bulk Update Monitoring Config
// @Description Update check frequency for multiple pages
// @Tags monitoring
// @Security BearerAuth
// @Accept json
// @Param request body bulkupdatemonitoringconfig.BulkUpdateMonitoringConfigRequest true "Bulk Update Request"
// @Success 204
// @Router /monitoring/configs/bulk [put]
func (m *Module) handleBulkUpdateMonitoringConfig(w http.ResponseWriter, r *http.Request) {
	if m.db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	var req bulkupdatemonitoringconfig.BulkUpdateMonitoringConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.PageIDs) == 0 {
		http.Error(w, "page_ids must not be empty", http.StatusBadRequest)
		return
	}
	if req.CheckFrequency == "" {
		http.Error(w, "check_frequency is required", http.StatusBadRequest)
		return
	}

	pageIDs := make([]uuid.UUID, 0, len(req.PageIDs))
	for _, idStr := range req.PageIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "Invalid page ID: "+idStr, http.StatusBadRequest)
			return
		}
		pageIDs = append(pageIDs, id)
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewMonitoringConfigPostgresRepository(m.db, tenant)
	handler := bulkupdatemonitoringconfig.NewBulkUpdateMonitoringConfigHandler(repo)

	if err := handler.Handle(r.Context(), pageIDs, req.CheckFrequency); err != nil {
		logger.Error("Failed to bulk update monitoring config", zap.Error(err))
		http.Error(w, "Failed to update check frequency", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleCreateNotificationPreference creates a new notification preference
// @Summary Create Notification Preference
// @Description Create a new notification preference
// @Tags monitoring
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body create_notification_preference.CreateNotificationPreferenceRequest true "Create Preference Request"
// @Success 201 {object} create_notification_preference.CreateNotificationPreferenceResponse
// @Router /monitoring/notification-preferences [post]
func (m *Module) handleCreateNotificationPreference(w http.ResponseWriter, r *http.Request) {
	// If db is not available, return mock response
	if m.db == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "550e8400-e29b-41d4-a716-446655440000",
			"message": "create notification preference (mock - db not initialized)",
		})
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewNotificationPreferencePostgresRepository(m.db, tenant)

	var req createnotificationpreference.CreateNotificationPreferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if (req.WorkspaceID == nil && req.PageID == nil) || (req.WorkspaceID != nil && req.PageID != nil) {
		http.Error(w, "Either workspace_id or page_id must be provided, not both", http.StatusBadRequest)
		return
	}
	if req.UserID == [16]byte{} {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	handler := createnotificationpreference.NewCreateNotificationPreferenceHandler(repo)
	resp, err := handler.Handle(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// handleGetNotificationPreference gets a notification preference by ID
// @Summary Get Notification Preference
// @Description Get a notification preference by ID
// @Tags monitoring
// @Security BearerAuth
// @Produce json
// @Param id path string true "Preference ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /monitoring/notification-preferences/{id} [get]
func (m *Module) handleGetNotificationPreference(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid preference id"})
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewNotificationPreferencePostgresRepository(m.db, tenant)

	pref, err := repo.GetByID(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to get notification preference"})
		return
	}
	if pref == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "notification preference not found"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pref)
}

// handleListSections returns all monitored sections for a page
// @Summary List Monitored Sections
// @Description List all monitored sections for a page
// @Tags monitoring
// @Security BearerAuth
// @Produce json
// @Param pageId path string true "Page ID"
// @Success 200 {object} managesections.ListSectionsResponse
// @Router /monitoring/sections/page/{pageId} [get]
func (m *Module) handleListSections(w http.ResponseWriter, r *http.Request) {
	if m.db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	pageID, err := uuid.Parse(chi.URLParam(r, "pageId"))
	if err != nil {
		http.Error(w, "invalid page_id", http.StatusBadRequest)
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	sectionRepo := persistence.NewMonitoredSectionPostgresRepository(m.db, tenant)
	configRepo := persistence.NewMonitoringConfigPostgresRepository(m.db, tenant)
	handler := managesections.NewManageSectionsHandler(sectionRepo, configRepo)

	resp, err := handler.List(r.Context(), pageID)
	if err != nil {
		logger.Error("Failed to list sections", zap.Error(err))
		http.Error(w, "failed to list sections", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleSaveSections replaces all monitored sections for a page
// @Summary Save Monitored Sections
// @Description Replace all monitored sections for a page
// @Tags monitoring
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param pageId path string true "Page ID"
// @Param request body managesections.SaveSectionsRequest true "Save Sections Request"
// @Success 200 {object} managesections.ListSectionsResponse
// @Router /monitoring/sections/page/{pageId} [post]
func (m *Module) handleSaveSections(w http.ResponseWriter, r *http.Request) {
	if m.db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	pageID, err := uuid.Parse(chi.URLParam(r, "pageId"))
	if err != nil {
		http.Error(w, "invalid page_id", http.StatusBadRequest)
		return
	}

	var req managesections.SaveSectionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	sectionRepo := persistence.NewMonitoredSectionPostgresRepository(m.db, tenant)
	configRepo := persistence.NewMonitoringConfigPostgresRepository(m.db, tenant)
	handler := managesections.NewManageSectionsHandler(sectionRepo, configRepo)

	resp, err := handler.SaveAll(r.Context(), pageID, &req)
	if err != nil {
		logger.Error("Failed to save sections", zap.Error(err))
		http.Error(w, "failed to save sections", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleDeleteSection deletes a single monitored section
// @Summary Delete Monitored Section
// @Description Delete a monitored section by ID
// @Tags monitoring
// @Security BearerAuth
// @Param pageId path string true "Page ID"
// @Param sectionId path string true "Section ID"
// @Success 204
// @Router /monitoring/sections/page/{pageId}/{sectionId} [delete]
func (m *Module) handleDeleteSection(w http.ResponseWriter, r *http.Request) {
	if m.db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	sectionID, err := uuid.Parse(chi.URLParam(r, "sectionId"))
	if err != nil {
		http.Error(w, "invalid section_id", http.StatusBadRequest)
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	sectionRepo := persistence.NewMonitoredSectionPostgresRepository(m.db, tenant)
	configRepo := persistence.NewMonitoringConfigPostgresRepository(m.db, tenant)
	handler := managesections.NewManageSectionsHandler(sectionRepo, configRepo)

	if err := handler.DeleteSection(r.Context(), sectionID); err != nil {
		logger.Error("Failed to delete section", zap.Error(err))
		http.Error(w, "failed to delete section", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleCheckSSE streams check-updated events to the client using SSE.
// The client connects with /checks/page/{pageId}/stream and receives events
// whenever a check for that page completes (success or error).
func (m *Module) handleCheckSSE(w http.ResponseWriter, r *http.Request) {
	pageIDStr := chi.URLParam(r, "pageId")
	if _, err := uuid.Parse(pageIDStr); err != nil {
		http.Error(w, "invalid pageId", http.StatusBadRequest)
		return
	}

	// Verify the page exists in the current tenant's schema.
	tenant := middleware.GetTenantFromContext(r.Context())
	if tenant != "" {
		q := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s.pages WHERE id = $1 AND deleted_at IS NULL)`, pq.QuoteIdentifier(tenant))
		var exists bool
		if err := m.db.QueryRowContext(r.Context(), q, pageIDStr).Scan(&exists); err != nil || !exists {
			http.Error(w, "page not found", http.StatusNotFound)
			return
		}
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	rc := http.NewResponseController(w)
	if err := rc.Flush(); err != nil {
		return
	}

	ch, unsubscribe := m.checkBroker.Subscribe(pageIDStr)
	defer unsubscribe()

	// 5-minute timeout — browser's EventSource auto-reconnects after close.
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	// Send heartbeat comments every 30s to keep the connection alive through proxies.
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case payload, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "event: check:updated\ndata: %s\n\n", payload)
			if err := rc.Flush(); err != nil {
				return
			}
		case <-heartbeat.C:
			if _, err := fmt.Fprintf(w, ": heartbeat\n\n"); err != nil {
				return
			}
			if err := rc.Flush(); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
