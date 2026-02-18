package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	createcheck "github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/create_check"
	createmonitoringconfig "github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/create_monitoring_config"
	createnotificationpreference "github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/create_notification_preference"
	getmonitoringconfig "github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/get_monitoring_config"
	listchecks "github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/list_checks"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/orchestrator"
	updatemonitoringconfig "github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/update_monitoring_config"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/workers"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/scheduler"
	generateinsights "github.com/jcsoftdev/pulzifi-back/modules/insight/application/generate_insights"
	insightAI "github.com/jcsoftdev/pulzifi-back/modules/insight/infrastructure/ai"
	snapshotapp "github.com/jcsoftdev/pulzifi-back/modules/snapshot/application"
	snapshotextractor "github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/extractor"
	snapshotstorage "github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/storage"
	sharedAI "github.com/jcsoftdev/pulzifi-back/shared/ai"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"go.uber.org/zap"
)

// Module implements the router.ModuleRegisterer interface for the Monitoring module
type Module struct {
	db         *sql.DB
	eventBus   *eventbus.EventBus
	scheduler  *scheduler.Scheduler
	workerPool *workers.WorkerPool
}

// NewModule creates a new instance of the Monitoring module
func NewModule() router.ModuleRegisterer {
	return &Module{}
}

// NewModuleWithDB creates a new instance with database connection
func NewModuleWithDB(db *sql.DB, eventBus *eventbus.EventBus) router.ModuleRegisterer {
	m := &Module{
		db:       db,
		eventBus: eventBus,
	}

	// Initialize Configuration
	cfg := config.Load()

	// Initialize Snapshot Infrastructure
	objectStorage, err := snapshotstorage.NewObjectStorage(cfg)
	if err != nil {
		logger.Error("Failed to initialize object storage client", zap.Error(err))
	} else {
		if err := objectStorage.EnsureBucket(context.Background()); err != nil {
			logger.Error("Failed to initialize object storage", zap.Error(err))
		}
	}

	extractorClient := snapshotextractor.NewHTTPClient(cfg.ExtractorURL)

	var insightHandler *generateinsights.GenerateInsightsHandler
	if cfg.OpenRouterAPIKey != "" {
		openRouterClient := sharedAI.NewOpenRouterClient(cfg.OpenRouterAPIKey, cfg.OpenRouterModel)
		generator := insightAI.NewOpenRouterGenerator(openRouterClient)
		insightHandler = generateinsights.NewGenerateInsightsHandler(generator, m.db)
	}

	snapshotWorker := snapshotapp.NewSnapshotWorker(objectStorage, extractorClient, m.db, insightHandler)

	// Create WorkerPool
	m.workerPool = workers.NewWorkerPool(snapshotWorker, 100)

	// In API-only mode we still need immediate dispatch capability when user updates frequency.
	// Start a lightweight in-process worker to consume TriggerPageCheck jobs.
	if os.Getenv("ENABLE_WORKERS") == "false" {
		m.workerPool.Start(1)
		logger.Info("Monitoring inline worker started for immediate API dispatch", zap.Int("concurrency", 1))
	}

	repoFactory := persistence.NewPostgresRepositoryFactory(m.db)
	orch := orchestrator.NewOrchestrator(repoFactory, m.workerPool)

	// Create Scheduler instance
	m.scheduler = scheduler.NewScheduler(m.db, orch)

	return m
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
		r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)
		r.Use(middleware.RequireTenant)
		r.Route("/checks", func(cr chi.Router) {
			cr.Post("/", m.handleCreateCheck)
			cr.Get("/", m.handleListChecks)
			cr.Get("/{id}", m.handleGetCheck)
			cr.Get("/page/{pageId}", m.handleListChecksByPage)
		})
		r.Route("/configs", func(cr chi.Router) {
			cr.Post("/", m.handleCreateMonitoringConfig)
			cr.Get("/{pageId}", m.handleGetMonitoringConfig)
			cr.Put("/{pageId}", m.handleUpdateMonitoringConfig)
		})
		r.Route("/notification-preferences", func(cr chi.Router) {
			cr.Post("/", m.handleCreateNotificationPreference)
			cr.Get("/{id}", m.handleGetNotificationPreference)
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

	// Use real handler
	handler := createcheck.NewCreateCheckHandler(repo)
	handler.HandleHTTP(w, r)
}

// handleListChecks lists all monitoring checks
// @Summary List Monitoring Checks
// @Description List all monitoring checks
// @Tags monitoring
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /monitoring/checks [get]
func (m *Module) handleListChecks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"checks":  []interface{}{},
		"message": "list checks",
	})
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

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewCheckPostgresRepository(m.db, tenant)
	handler := listchecks.NewListChecksHandler(repo)
	handler.HandleHTTP(w, r)
}

// handleGetCheck gets a monitoring check by ID
// @Summary Get Monitoring Check
// @Description Get a monitoring check by ID
// @Tags monitoring
// @Security BearerAuth
// @Produce json
// @Param id path string true "Check ID"
// @Success 200 {object} map[string]interface{}
// @Router /monitoring/checks/{id} [get]
func (m *Module) handleGetCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"id":      chi.URLParam(r, "id"),
		"message": "get check",
	})
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

	// Get tenant from context
	tenant := middleware.GetTenantFromContext(r.Context())

	// Create repository with dynamic tenant
	repo := persistence.NewMonitoringConfigPostgresRepository(m.db, tenant)

	// Use real handler
	handler := createmonitoringconfig.NewCreateMonitoringConfigHandler(repo, m.scheduler)
	handler.HandleHTTP(w, r)
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

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewMonitoringConfigPostgresRepository(m.db, tenant)
	handler := getmonitoringconfig.NewGetMonitoringConfigHandler(repo)
	handler.HandleHTTP(w, r)
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

	// Get tenant from context
	tenant := middleware.GetTenantFromContext(r.Context())

	// Create repository with dynamic tenant
	repo := persistence.NewMonitoringConfigPostgresRepository(m.db, tenant)

	// Use real handler
	handler := updatemonitoringconfig.NewUpdateMonitoringConfigHandler(repo, m.eventBus, tenant, m.scheduler)
	handler.HandleHTTP(w, r)
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

	// Get tenant from context
	tenant := middleware.GetTenantFromContext(r.Context())

	// Create repository with dynamic tenant
	repo := persistence.NewNotificationPreferencePostgresRepository(m.db, tenant)

	// Use real handler
	handler := createnotificationpreference.NewCreateNotificationPreferenceHandler(repo)
	handler.HandleHTTP(w, r)
}

// handleGetNotificationPreference gets a notification preference by ID
// @Summary Get Notification Preference
// @Description Get a notification preference by ID
// @Tags monitoring
// @Security BearerAuth
// @Produce json
// @Param id path string true "Preference ID"
// @Success 200 {object} map[string]interface{}
// @Router /monitoring/notification-preferences/{id} [get]
func (m *Module) handleGetNotificationPreference(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"id":      chi.URLParam(r, "id"),
		"message": "get notification preference",
	})
}
