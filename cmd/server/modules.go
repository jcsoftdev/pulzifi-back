package main

import (
	"database/sql"

	alert "github.com/jcsoftdev/pulzifi-back/modules/alert/infrastructure/http"
	auth "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/http"
	authpersistence "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/persistence"
	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/services"
	insight "github.com/jcsoftdev/pulzifi-back/modules/insight/infrastructure/http"
	integration "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/http"
	monitoring "github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/http"
	organization "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/http"
	orgpersistence "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/persistence"
	page "github.com/jcsoftdev/pulzifi-back/modules/page/infrastructure/http"
	report "github.com/jcsoftdev/pulzifi-back/modules/report/infrastructure/http"
	usage "github.com/jcsoftdev/pulzifi-back/modules/usage/infrastructure/http"
	workspace "github.com/jcsoftdev/pulzifi-back/modules/workspace/infrastructure/http"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"go.uber.org/zap"
)

func registerAllModulesInternal(registry *router.Registry, db *sql.DB, eventBus *eventbus.EventBus, enableWorkers bool) {
	cfg := config.Load()

	userRepo := authpersistence.NewUserPostgresRepository(db)
	roleRepo := authpersistence.NewRolePostgresRepository(db)
	permRepo := authpersistence.NewPermissionPostgresRepository(db)
	sessionRepo := authpersistence.NewSessionPostgresRepository(db)
	orgRepo := orgpersistence.NewOrganizationPostgresRepository(db)

	authService := authservices.NewBcryptAuthService(userRepo, permRepo)
	cookieSecure := cfg.Environment == "production"

	// Create auth module and set global middleware
	authModule := auth.NewModule(auth.ModuleDeps{
		UserRepo:     userRepo,
		SessionRepo:  sessionRepo,
		RoleRepo:     roleRepo,
		PermRepo:     permRepo,
		AuthService:  authService,
		SessionTTL:   cfg.JWTExpiration,
		CookieDomain: cfg.CookieDomain,
		CookieSecure: cookieSecure,
	})
	authMiddleware := authModule.(*auth.Module).AuthMiddleware()

	// Set global middleware for all modules
	middleware.SetAuthMiddleware(authMiddleware)
	middleware.SetOrganizationMiddleware(middleware.NewOrganizationMiddleware(db))

	moduleInstances := []struct {
		name   string
		module router.ModuleRegisterer
	}{
		{"Auth", authModule},
		{"Organization", organization.NewModule(orgRepo)},
		{"Workspace", workspace.NewModuleWithDB(db)},
		{"Page", page.NewModuleWithDB(db)},
		{"Alert", alert.NewModuleWithDB(db)},
		{"Monitoring", monitoring.NewModuleWithDB(db, eventBus)},
		{"Integration", integration.NewModule()},
		{"Insight", insight.NewModuleWithDB(db)},
		{"Report", report.NewModule()},
		{"Usage", usage.NewModuleWithDB(db)},
	}

	logger.Info("Registering all modules", zap.Int("count", len(moduleInstances)))

	for _, m := range moduleInstances {
		registry.Register(m.module)
		logger.Info("Registered module", zap.String("module", m.name))

		// Special handling for Monitoring module to start background processes if enabled
		if m.name == "Monitoring" && enableWorkers {
			if monModule, ok := m.module.(*monitoring.Module); ok {
				monModule.StartBackgroundProcesses()
				logger.Info("Started background processes for Monitoring module")
			}
		}
	}

	logger.Info("All modules registered successfully", zap.Int("total", registry.Count()))
}
