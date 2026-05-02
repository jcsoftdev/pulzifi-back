package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"time"

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
	dispatchevent "github.com/jcsoftdev/pulzifi-back/modules/integration/application/dispatch_event"
	integration "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/http"
	intoauth "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/oauth"
	intpersistence "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/persistence"
	intproviders "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/providers"
	emailprovider "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/providers/email"
	slackprovider "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/providers/slack"
	monitoring "github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/http"
	orgservices "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
	organization "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/http"
	orgmessaging "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/messaging"
	orgpersistence "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/persistence"
	page "github.com/jcsoftdev/pulzifi-back/modules/page/infrastructure/http"
	report "github.com/jcsoftdev/pulzifi-back/modules/report/infrastructure/http"
	snapshotextractor "github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/extractor"
	team "github.com/jcsoftdev/pulzifi-back/modules/team/infrastructure/http"
	usage "github.com/jcsoftdev/pulzifi-back/modules/usage/infrastructure/http"
	workspace "github.com/jcsoftdev/pulzifi-back/modules/workspace/infrastructure/http"
	"github.com/jcsoftdev/pulzifi-back/shared/bff"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/crypto"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/noncestore"
	"github.com/jcsoftdev/pulzifi-back/shared/pubsub"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	intwiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/integration"
	"go.uber.org/zap"
)

// createEmailProvider creates the Resend email provider.
func createEmailProvider(cfg *config.Config) emailservices.EmailProvider {
	return emailproviders.NewResendProvider(cfg.ResendAPIKey, cfg.EmailFromAddress, cfg.EmailFromName)
}

func registerAllModulesInternal(registry *router.Registry, db *sql.DB, eventBus *eventbus.EventBus, enableWorkers bool) (*bff.Handler, *integration.Module) {
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

	// ---------------------------------------------------------------------------
	// Integration module wiring
	// ---------------------------------------------------------------------------
	intKeyHex := cfg.IntegrationTokenKey
	if intKeyHex == "" {
		if cfg.Environment == "production" {
			logger.Logger.Fatal("INTEGRATION_TOKEN_KEY required in production")
		}
		logger.Warn("INTEGRATION_TOKEN_KEY not set — using insecure dev default")
		intKeyHex = "00000000000000000000000000000000000000000000000000000000000000ff"
	}
	intKey, err := hex.DecodeString(intKeyHex)
	if err != nil || len(intKey) != 32 {
		logger.Logger.Fatal("invalid INTEGRATION_TOKEN_KEY", zap.Error(err))
	}
	intEnc, err := crypto.NewAESGCM(intKey)
	if err != nil {
		logger.Logger.Fatal("crypto init failed", zap.Error(err))
	}

	intRepo := intpersistence.NewIntegrationPostgresRepository(db, intEnc)
	intStateSigner := intoauth.NewStateSigner(intKey, 10*time.Minute)

	// Provider registry — Slack + email (via adapter wrapping the existing email module).
	slackClient := slackprovider.New(cfg.SlackClientID, cfg.SlackClientSecret)
	intEmailClient := emailprovider.New(intwiring.NewEmailAdapter(emailProvider))
	intRegistry := intproviders.NewRegistry(slackClient, intEmailClient)

	intRepoFactory := intwiring.NewTenantRepoFactory(db)
	intOrgGuard := intwiring.NewOrgGuard(orgRepo)
	intDispatcher := dispatchevent.NewHandler(intRepoFactory, intOrgGuard)

	// Subscribe dispatcher to domain events.
	if err := eventbus.SubscribeDomainEvent(eventBus, eventbus.TopicChangeDetected, func(ev eventbus.DomainEvent) {
		if err := intDispatcher.Handle(context.Background(), ev); err != nil {
			logger.Error("integration dispatch failed", zap.Error(err), zap.String("event_type", ev.Type))
		}
	}); err != nil {
		logger.Error("subscribe TopicChangeDetected", zap.Error(err))
	}
	if err := eventbus.SubscribeDomainEvent(eventBus, eventbus.TopicAlertCreated, func(ev eventbus.DomainEvent) {
		if err := intDispatcher.Handle(context.Background(), ev); err != nil {
			logger.Error("integration dispatch failed", zap.Error(err), zap.String("event_type", ev.Type))
		}
	}); err != nil {
		logger.Error("subscribe TopicAlertCreated", zap.Error(err))
	}

	integrationMod := integration.NewModule(integration.Deps{
		DB:                db,
		IntRepo:           intRepo,
		Registry:          intRegistry,
		StateSigner:       intStateSigner,
		OAuthRedirectBase: cfg.IntegrationOAuthRedirectBase,
	})

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
		{"Alert", func() router.ModuleRegisterer {
			m := alert.NewModuleWithDB(db)
			if am, ok := m.(*alert.Module); ok {
				am.SetEventBus(eventBus)
			}
			return m
		}()},
		{"Monitoring", monitoring.NewModuleWithDB(db, eventBus, emailProvider, cfg.FrontendURL)},
		{"Integration", integrationMod},
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
		LogoutHandler:  authMod.LogoutHandler(),
		RefreshHandler: authMod.RefreshHandler(),
		TokenService:   authMod.TokenService(),
		NonceStore:     noncestore.New(),
		CookieDomain:   authMod.CookieDomain(),
		CookieSecure:   authMod.CookieSecure(),
		Logger:         logger.Logger,
	})

	return bffHandler, integrationMod
}
