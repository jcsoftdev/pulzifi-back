package main

import (
	"context"
	"database/sql"

	admin "github.com/jcsoftdev/pulzifi-back/modules/admin/infrastructure/http"
	adminpersistence "github.com/jcsoftdev/pulzifi-back/modules/admin/infrastructure/persistence"
	alert "github.com/jcsoftdev/pulzifi-back/modules/alert/infrastructure/http"
	auth "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/http"
	authpersistence "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/persistence"
	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/services"
	dashboard "github.com/jcsoftdev/pulzifi-back/modules/dashboard/infrastructure/http"
	emailservices "github.com/jcsoftdev/pulzifi-back/modules/email/domain/services"
	email "github.com/jcsoftdev/pulzifi-back/modules/email/infrastructure/http"
	emailproviders "github.com/jcsoftdev/pulzifi-back/modules/email/infrastructure/providers"
	insight "github.com/jcsoftdev/pulzifi-back/modules/insight/infrastructure/http"
	integration "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/http"
	monitoring "github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/http"
	organization "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/http"
	orgmessaging "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/messaging"
	orgpersistence "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/persistence"
	orgservices "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
	page "github.com/jcsoftdev/pulzifi-back/modules/page/infrastructure/http"
	snapshotextractor "github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/extractor"
	report "github.com/jcsoftdev/pulzifi-back/modules/report/infrastructure/http"
	team "github.com/jcsoftdev/pulzifi-back/modules/team/infrastructure/http"
	usage "github.com/jcsoftdev/pulzifi-back/modules/usage/infrastructure/http"
	workspace "github.com/jcsoftdev/pulzifi-back/modules/workspace/infrastructure/http"
	"github.com/jcsoftdev/pulzifi-back/shared/bff"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/noncestore"
	"github.com/jcsoftdev/pulzifi-back/shared/pubsub"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"go.uber.org/zap"
)

// createEmailProvider creates the Resend email provider.
func createEmailProvider(cfg *config.Config) emailservices.EmailProvider {
	return emailproviders.NewResendProvider(cfg.ResendAPIKey, cfg.EmailFromAddress, cfg.EmailFromName)
}

func registerAllModulesInternal(registry *router.Registry, db *sql.DB, eventBus *eventbus.EventBus, enableWorkers bool) *bff.Handler {
	cfg := config.Load()

	userRepo := authpersistence.NewUserPostgresRepository(db)
	roleRepo := authpersistence.NewRolePostgresRepository(db)
	permRepo := authpersistence.NewPermissionPostgresRepository(db)
	refreshTokenRepo := authpersistence.NewRefreshTokenPostgresRepository(db)
	orgRepo := orgpersistence.NewOrganizationPostgresRepository(db)

	regReqRepo := adminpersistence.NewRegistrationRequestPostgresRepository(db)
	orgService := orgservices.NewOrganizationService()

	authService := authservices.NewBcryptAuthService(userRepo, permRepo)
	jwtService := authservices.NewJWTService(cfg.JWTSecret, cfg.JWTExpiration, cfg.JWTRefreshExpiration, roleRepo, permRepo)
	cookieSecure := cfg.Environment == "production"

	// Create email provider (shared across modules)
	emailProvider := createEmailProvider(cfg)

	// Create auth module and set global middleware
	authModule := auth.NewModule(auth.ModuleDeps{
		UserRepo:         userRepo,
		RefreshTokenRepo: refreshTokenRepo,
		RoleRepo:         roleRepo,
		PermRepo:         permRepo,
		RegReqRepo:       regReqRepo,
		OrgRepo:          orgRepo,
		OrgService:       orgService,
		AuthService:      authService,
		TokenService:     jwtService,
		CookieDomain:     cfg.CookieDomain,
		CookieSecure:     cookieSecure,
		FrontendURL:      cfg.FrontendURL,
		EmailProvider:    emailProvider,
		EventBus:         eventBus,
		DB:               db,
	})
	authMod := authModule.(*auth.Module)
	authMiddleware := authMod.AuthMiddleware()

	// Set global middleware for all modules
	middleware.SetAuthMiddleware(authMiddleware)
	middleware.SetOrganizationMiddleware(middleware.NewOrganizationMiddleware(db))

	moduleInstances := []struct {
		name   string
		module router.ModuleRegisterer
	}{
		{"Auth", authModule},
		{"Admin", admin.NewModule(admin.ModuleDeps{
			DB:             db,
			RegReqRepo:     regReqRepo,
			UserRepo:       userRepo,
			OrgRepo:        orgRepo,
			OrgService:     orgService,
			AuthMiddleware: authMiddleware,
			EmailProvider:  emailProvider,
			FrontendURL:    cfg.FrontendURL,
		})},
		{"Email", email.NewModule(emailProvider)},
		{"Organization", organization.NewModule(orgRepo)},
		{"Workspace", workspace.NewModuleWithDB(db)},
		{"Page", page.NewModuleWithExtractor(db, snapshotextractor.NewHTTPClient(cfg.ExtractorURL))},
		{"Alert", alert.NewModuleWithDB(db)},
		{"Monitoring", monitoring.NewModuleWithDB(db, eventBus, emailProvider, cfg.FrontendURL)},
		{"Integration", integration.NewModuleWithDB(db)},
		{"Insight", insight.NewModuleWithDB(db, pubsub.NewInsightBroker())},
		{"Report", report.NewModuleWithDB(db)},
		{"Usage", usage.NewModuleWithDB(db)},
		{"Dashboard", dashboard.NewModuleWithDB(db)},
		{"Team", team.NewModuleWithDB(db, emailProvider, cfg.FrontendURL)},
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

	// Start organization event subscriber in background
	orgSubscriber := orgmessaging.NewSubscriber(eventBus, db)
	go func() {
		orgSubscriber.ListenToEvents(context.Background())
	}()
	logger.Info("Started organization event subscriber")

	logger.Info("All modules registered successfully", zap.Int("total", registry.Count()))

	// Create BFF auth handler using extracted auth module handlers
	bffHandler := bff.NewHandler(bff.HandlerDeps{
		LoginHandler:   authMod.LoginHandler(),
		RefreshHandler: authMod.RefreshHandler(),
		TokenService:   authMod.TokenService(),
		NonceStore:     noncestore.New(),
		CookieDomain:   authMod.CookieDomain(),
		CookieSecure:   authMod.CookieSecure(),
		Logger:         logger.Logger,
	})

	return bffHandler
}
