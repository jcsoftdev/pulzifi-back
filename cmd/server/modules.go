package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"time"

	admin "github.com/jcsoftdev/pulzifi-back/modules/admin/infrastructure/http"
	adminpersistence "github.com/jcsoftdev/pulzifi-back/modules/admin/infrastructure/persistence"
	alert "github.com/jcsoftdev/pulzifi-back/modules/alert/infrastructure/http"
	billing "github.com/jcsoftdev/pulzifi-back/modules/billing/infrastructure/http"
	createcheckoutsession "github.com/jcsoftdev/pulzifi-back/modules/billing/application/create_checkout_session"
	createportalsession "github.com/jcsoftdev/pulzifi-back/modules/billing/application/create_portal_session"
	getsubscription "github.com/jcsoftdev/pulzifi-back/modules/billing/application/get_subscription"
	handlewebhook "github.com/jcsoftdev/pulzifi-back/modules/billing/application/handle_webhook"
	billingpostgres "github.com/jcsoftdev/pulzifi-back/modules/billing/infrastructure/persistence/postgres"
	billingstripe "github.com/jcsoftdev/pulzifi-back/modules/billing/infrastructure/stripe"
	billingwiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/billing"
	auth "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/http"
	authpersistence "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/persistence"
	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/services"
	dashboard "github.com/jcsoftdev/pulzifi-back/modules/dashboard/infrastructure/http"
	emailservices "github.com/jcsoftdev/pulzifi-back/modules/email/domain/services"
	email "github.com/jcsoftdev/pulzifi-back/modules/email/infrastructure/http"
	emailproviders "github.com/jcsoftdev/pulzifi-back/modules/email/infrastructure/providers"
	insight "github.com/jcsoftdev/pulzifi-back/modules/insight/infrastructure/http"
	insightwiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/insight"
	insightservices "github.com/jcsoftdev/pulzifi-back/modules/insight/domain/services"
	dispatchevent "github.com/jcsoftdev/pulzifi-back/modules/integration/application/dispatch_event"
	integration "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/http"
	intoauth "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/oauth"
	intpersistence "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/persistence"
	intproviders "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/providers"
	discordprovider "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/providers/discord"
	emailprovider "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/providers/email"
	gmailprovider "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/providers/gmail"
	outlookprovider "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/providers/outlook"
	sheetsprovider "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/providers/sheets"
	slackprovider "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/providers/slack"
	teamsprovider "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/providers/teams"
	twilioprovider "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/providers/twilio"
	intdomainservices "github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/featureflags"
	"github.com/jcsoftdev/pulzifi-back/shared/integrationusage"
	monitoring "github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/http"
	monitoringwiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/monitoring"
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
	"github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/bff"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/crypto"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/noncestore"
	"github.com/jcsoftdev/pulzifi-back/shared/pubsub"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	adminwiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/admin"
	authwiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/auth"
	intwiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/integration"
	pagewiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/page"
	teamwiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/team"
	"go.uber.org/zap"
)

// createEmailProvider creates the Resend email provider.
func createEmailProvider(cfg *config.Config) emailservices.EmailProvider {
	return emailproviders.NewResendProvider(cfg.ResendAPIKey, cfg.EmailFromAddress, cfg.EmailFromName)
}

func registerAllModulesInternal(
	registry *router.Registry,
	db *sql.DB,
	eventBus *eventbus.EventBus,
	enableWorkers bool,
	nonceStore noncestore.NonceStore,
	checkBroker pubsub.CheckBroker,
	insightBroker pubsub.InsightBroker,
) (*bff.Handler, *integration.Module) {
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

	// Composed org-context lookup for /me (flags + plan + identity in one query).
	orgContextLookup := intwiring.NewOrgContextLookup(db)

	// Auth wiring adapters — bridge auth module ports to concrete implementations
	// from admin, organization, and email modules without creating cross-module imports.
	authRegReqWriter := authwiring.NewRegistrationWriterAdapter(regReqRepo)
	authOrgDirectory := authwiring.NewOrganizationDirectoryAdapter(orgRepo, orgService)
	authNotifier := authwiring.NewNotifierAdapter(emailProvider)
	authTrialProvisioner := authwiring.NewTrialProvisioner(db, orgService)

	// Create auth module and set global middleware
	authModule := auth.NewModule(auth.ModuleDeps{
		UserRepo:         userRepo,
		RefreshTokenRepo: refreshTokenRepo,
		RoleRepo:         roleRepo,
		PermRepo:         permRepo,
		RegReqWriter:     authRegReqWriter,
		OrgDirectory:     authOrgDirectory,
		TrialProvisioner: authTrialProvisioner,
		TrialDays:        cfg.TrialDays,
		AuthService:      authService,
		TokenService:     jwtService,
		CookieDomain:     cfg.CookieDomain,
		CookieSecure:     cookieSecure,
		FrontendURL:      cfg.FrontendURL,
		Notifier:         authNotifier,
		EventBus:         eventBus,
		DB:               db,
		OrgContextLookup: orgContextLookup,
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
	intStateSigner := intoauth.NewStateSignerAdapter(intoauth.NewStateSigner(intKey, 10*time.Minute))

	// Provider registry — Slack + email (via adapter wrapping the existing email module).
	slackClient := slackprovider.New(cfg.SlackClientID, cfg.SlackClientSecret)
	intEmailClient := emailprovider.New(intwiring.NewEmailAdapter(emailProvider))

	// ---------------------------------------------------------------------------
	// Phase 2 integration providers
	// ---------------------------------------------------------------------------
	flagsReader := featureflags.NewReader(db)

	discordClient := discordprovider.New(cfg.DiscordClientID, cfg.DiscordClientSecret)

	twilioPlanLookup := intwiring.NewOrgPlanLookup(db)

	// Phase 3 additions
	sheetsClient := sheetsprovider.New(cfg.SheetsClientID, cfg.SheetsClientSecret)
	teamsClient := teamsprovider.New(cfg.MicrosoftClientID, cfg.MicrosoftClientSecret)

	quotaTracker := integrationusage.NewTracker(db)
	twilioAllowedFor := intwiring.TwilioAllowedFor(twilioPlanLookup, cfg.TwilioPaidPlans, cfg.TwilioQuotaPaidPerMonth)
	twilioQuotaAdapter := intwiring.NewTwilioQuotaAdapter(quotaTracker, twilioAllowedFor)

	twilioClient := twilioprovider.New(twilioprovider.Config{
		PaidPlans:          cfg.TwilioPaidPlans,
		PlatformAccountSID: cfg.TwilioAccountSID,
		PlatformAuthToken:  cfg.TwilioAuthToken,
		PlatformFromNumber: cfg.TwilioFromNumber,
	}, twilioPlanLookup, twilioQuotaAdapter)
	twilioValidator := twilioprovider.NewValidator()

	baseProviders := []intdomainservices.ProviderClient{slackClient, intEmailClient, discordClient, twilioClient, sheetsClient, teamsClient}
	if cfg.GmailIntegrationEnabled && cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		baseProviders = append(baseProviders, gmailprovider.New(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.IntegrationOAuthRedirectBase))
	}
	if cfg.MicrosoftClientID != "" && cfg.MicrosoftClientSecret != "" {
		baseProviders = append(baseProviders, outlookprovider.New(cfg.MicrosoftClientID, cfg.MicrosoftClientSecret, cfg.IntegrationOAuthRedirectBase))
	}
	intRegistry := intproviders.NewRegistry(baseProviders...)

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
		Flags:             flagsReader,
		Validator:         twilioValidator,
		PlanLookup:        twilioPlanLookup,
		TwilioPaidPlans:   cfg.TwilioPaidPlans,
	})

	moduleInstances := []struct {
		name   string
		module router.ModuleRegisterer
	}{
		{"Auth", authModule},
		{"Admin", admin.NewModule(admin.ModuleDeps{
			RegReqRepo:           regReqRepo,
			UserReader:           adminwiring.NewPendingUserAdapter(userRepo),
			ApprovalProvisioner:  adminwiring.NewApprovalProvisioner(db, orgService),
			RejectionProvisioner: adminwiring.NewRejectionProvisioner(db),
			Notifier:             adminwiring.NewNotifierAdapter(emailProvider),
			AuthMiddleware:       authMiddleware,
			FrontendURL:          cfg.FrontendURL,
		})},
		{"Email", email.NewModule(emailProvider)},
		{"Organization", organization.NewModule(orgRepo)},
		{"Workspace", workspace.NewModuleWithDB(db)},
		{"Page", page.NewModuleWithExtractor(db, pagewiring.NewExtractorPreviewStreamerAdapter(snapshotextractor.NewHTTPClient(cfg.ExtractorURL)))},
		{"Alert", func() router.ModuleRegisterer {
			m := alert.NewModuleWithDB(db)
			if am, ok := m.(*alert.Module); ok {
				am.SetEventBus(eventBus)
			}
			return m
		}()},
		{"Monitoring", func() router.ModuleRegisterer {
				snapshotWorker, err := monitoringwiring.NewSnapshotWorker(monitoringwiring.SnapshotWorkerDeps{
					DB:            db,
					EventBus:      eventBus,
					EmailProvider: emailProvider,
					FrontendURL:   cfg.FrontendURL,
					Cfg:           cfg,
				})
				if err != nil {
					logger.Error("Failed to build snapshot worker, monitoring will run without snapshot execution", zap.Error(err))
				}
				return monitoring.NewModuleWithDeps(monitoring.Deps{
					DB:               db,
					EventBus:         eventBus,
					SnapshotExecutor: snapshotWorker,
					CheckBroker:      checkBroker,
				})
			}()},
		{"Integration", integrationMod},
		{"Insight", insight.NewModuleWithDeps(
			db,
			insightBroker,
			func(tenant string) insightservices.CheckReader { return insightwiring.NewCheckReaderAdapter(db, tenant) },
			func(tenant string) insightservices.PageConfigReader { return insightwiring.NewPageConfigReaderAdapter(db, tenant) },
		)},
		{"Report", report.NewModuleWithDB(db)},
		{"Usage", usage.NewModuleWithDB(db)},
		{"Dashboard", dashboard.NewModuleWithDB(db)},
		{"Team", team.NewModuleWithDB(
			db,
			teamwiring.NewEmailerAdapter(emailProvider),
			teamwiring.NewInviteEmailBuilderAdapter(),
			teamwiring.NewTokenGeneratorAdapter(db),
			cfg.FrontendURL,
		)},
	}

	// ---------------------------------------------------------------------------
	// Billing module wiring (gated behind BILLING_ENABLED)
	// ---------------------------------------------------------------------------
	if cfg.BillingEnabled {
		stripeGateway := billingstripe.NewGateway(cfg.StripeSecretKey, cfg.StripeWebhookSecret)
		planAssigner := billingwiring.NewPlanAssigner(db)

		subscriptionRepo := billingpostgres.NewSubscriptionPostgresRepository(db)
		webhookRepo := billingpostgres.NewWebhookEventPostgresRepository(db)
		customerRepo := billingpostgres.NewCustomerPostgresRepository(db)

		checkoutHandler := createcheckoutsession.NewHandler(stripeGateway, customerRepo)
		portalHandler := createportalsession.NewHandler(stripeGateway, customerRepo)
		subscriptionHandler := getsubscription.NewHandler(subscriptionRepo)
		webhookHandler := handlewebhook.NewHandler(
			stripeGateway,
			cfg.StripeWebhookSecret,
			planAssigner,
			customerRepo,
			webhookRepo,
			subscriptionRepo,
		)

		billingMod := billing.NewModule(billing.Deps{
			DB:                       db,
			CheckoutHandler:          checkoutHandler,
			PortalHandler:            portalHandler,
			SubscriptionHandler:      subscriptionHandler,
			WebhookHandler:           webhookHandler,
			StripeCheckoutSuccessURL: cfg.StripeCheckoutSuccessURL,
			StripeCheckoutCancelURL:  cfg.StripeCheckoutCancelURL,
			StripePortalReturnURL:    cfg.StripePortalReturnURL,
		})

		moduleInstances = append(moduleInstances, struct {
			name   string
			module router.ModuleRegisterer
		}{"Billing", billingMod})

		logger.Info("Billing module enabled", zap.String("module", "Billing"))
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
		NonceStore:     nonceStore,
		CookieDomain:   authMod.CookieDomain(),
		CookieSecure:   authMod.CookieSecure(),
		Logger:         logger.Logger,
	})

	return bffHandler, integrationMod
}
