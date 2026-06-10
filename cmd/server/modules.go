package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"time"

	authwiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/auth"
	billingwiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/billing"
	insightwiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/insight"
	intwiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/integration"
	monitoringwiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/monitoring"
	pagewiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/page"
	socialwiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/social"
	teamwiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/team"
	alert "github.com/jcsoftdev/pulzifi-back/modules/alert/infrastructure/http"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/bff"
	auth "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/http"
	authpersistence "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/persistence"
	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/services"
	billingcancel "github.com/jcsoftdev/pulzifi-back/modules/billing/application/cancel_subscription"
	createcheckoutsession "github.com/jcsoftdev/pulzifi-back/modules/billing/application/create_checkout_session"
	createportalsession "github.com/jcsoftdev/pulzifi-back/modules/billing/application/create_portal_session"
	getsubscription "github.com/jcsoftdev/pulzifi-back/modules/billing/application/get_subscription"
	billinggiftmonth "github.com/jcsoftdev/pulzifi-back/modules/billing/application/gift_month"
	handlewebhook "github.com/jcsoftdev/pulzifi-back/modules/billing/application/handle_webhook"
	listplans "github.com/jcsoftdev/pulzifi-back/modules/billing/application/list_plans"
	managecoupons "github.com/jcsoftdev/pulzifi-back/modules/billing/application/manage_coupons"
	reconcilesubscription "github.com/jcsoftdev/pulzifi-back/modules/billing/application/reconcile_subscription"
	updatesubscription "github.com/jcsoftdev/pulzifi-back/modules/billing/application/update_subscription"
	billing "github.com/jcsoftdev/pulzifi-back/modules/billing/infrastructure/http"
	billingpostgres "github.com/jcsoftdev/pulzifi-back/modules/billing/infrastructure/persistence/postgres"
	billingstripe "github.com/jcsoftdev/pulzifi-back/modules/billing/infrastructure/stripe"
	dashboard "github.com/jcsoftdev/pulzifi-back/modules/dashboard/infrastructure/http"
	emailservices "github.com/jcsoftdev/pulzifi-back/modules/email/domain/services"
	email "github.com/jcsoftdev/pulzifi-back/modules/email/infrastructure/http"
	emailproviders "github.com/jcsoftdev/pulzifi-back/modules/email/infrastructure/providers"
	insightservices "github.com/jcsoftdev/pulzifi-back/modules/insight/domain/services"
	insight "github.com/jcsoftdev/pulzifi-back/modules/insight/infrastructure/http"
	dispatchevent "github.com/jcsoftdev/pulzifi-back/modules/integration/application/dispatch_event"
	intdomainservices "github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
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
	monitoring "github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/http"
	orgservices "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
	organization "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/http"
	orgmessaging "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/messaging"
	orgpersistence "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/persistence"
	page "github.com/jcsoftdev/pulzifi-back/modules/page/infrastructure/http"
	report "github.com/jcsoftdev/pulzifi-back/modules/report/infrastructure/http"
	snapshotextractor "github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/extractor"
	socialhttp "github.com/jcsoftdev/pulzifi-back/modules/social/infrastructure/http"
	socialscheduler "github.com/jcsoftdev/pulzifi-back/modules/social/infrastructure/scheduler"
	team "github.com/jcsoftdev/pulzifi-back/modules/team/infrastructure/http"
	trialstatus "github.com/jcsoftdev/pulzifi-back/modules/usage/application/trial_status"
	usage "github.com/jcsoftdev/pulzifi-back/modules/usage/infrastructure/http"
	usagepersistence "github.com/jcsoftdev/pulzifi-back/modules/usage/infrastructure/persistence"
	workspace "github.com/jcsoftdev/pulzifi-back/modules/workspace/infrastructure/http"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/crypto"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/featureflags"
	"github.com/jcsoftdev/pulzifi-back/shared/integrationusage"
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

func registerAllModulesInternal(
	registry *router.Registry,
	db *sql.DB,
	eventBus eventbus.MessageBus,
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

	orgService := orgservices.NewOrganizationService()

	authService := authservices.NewBcryptAuthService(userRepo, permRepo)
	jwtService := authservices.NewJWTService(cfg.JWTSecret, cfg.JWTExpiration, cfg.JWTRefreshExpiration, roleRepo, permRepo)
	cookieSecure := cfg.Environment == "production"

	// Create email provider (shared across modules)
	emailProvider := createEmailProvider(cfg)

	// Composed org-context lookup for /me (flags + plan + identity in one query).
	orgContextLookup := intwiring.NewOrgContextLookup(db)

	// Auth wiring adapters — bridge auth module ports to concrete implementations
	// from organization and email modules without creating cross-module imports.
	authOrgDirectory := authwiring.NewOrganizationDirectoryAdapter(orgRepo, orgService)
	authOnboardingAdapter := authwiring.NewOnboardingProfileAdapter(db)
	authNotifier := authwiring.NewNotifierAdapter(emailProvider)
	authTrialProvisioner := authwiring.NewTrialProvisioner(db, orgService)
	authMembershipChecker := authwiring.NewMembershipChecker(db)

	// Create auth module and set global middleware
	authModule := auth.NewModule(auth.ModuleDeps{
		UserRepo:            userRepo,
		RefreshTokenRepo:    refreshTokenRepo,
		RoleRepo:            roleRepo,
		PermRepo:            permRepo,
		OrgDirectory:        authOrgDirectory,
		OnboardingWriter:    authOnboardingAdapter,
		OnboardingOrgFinder: authOnboardingAdapter,
		TrialProvisioner:  authTrialProvisioner,
		MembershipChecker: authMembershipChecker,
		TrialDays:         cfg.TrialDays,
		AuthService:       authService,
		TokenService:      jwtService,
		CookieDomain:      cfg.CookieDomain,
		CookieSecure:      cookieSecure,
		FrontendURL:       cfg.FrontendURL,
		Notifier:          authNotifier,
		EventBus:          eventBus,
		DB:                db,
		OrgContextLookup:  orgContextLookup,
	})
	authMod := authModule.(*auth.Module)
	authMiddleware := authMod.AuthMiddleware()

	// Set global middleware for all modules
	middleware.SetAuthMiddleware(authMiddleware)
	middleware.SetOrganizationMiddleware(middleware.NewOrganizationMiddleware(db))

	// Integration module wiring (providers, crypto, event subscriptions).
	integrationMod := buildIntegrationModule(db, cfg, eventBus, emailProvider, orgRepo)

	moduleInstances := []struct {
		name   string
		module router.ModuleRegisterer
	}{
		{"Auth", authModule},
		{"Email", email.NewModule(emailProvider)},
		{"Organization", organization.NewModule(orgRepo)},
		{"Workspace", workspace.NewModuleWithDB(db)},
		{"Page", page.NewModuleWithExtractor(db, pagewiring.NewExtractorPreviewStreamerAdapter(snapshotextractor.NewHTTPClientWithKey(cfg.ExtractorURL, cfg.ExtractorAPIKey)))},
		{"Alert", buildAlertModule(db, eventBus)},
		{"Monitoring", buildMonitoringModule(db, eventBus, emailProvider, checkBroker, cfg)},
		{"Integration", integrationMod},
		{"Insight", buildInsightModule(db, insightBroker)},
		{"Report", report.NewModuleWithDB(db)},
		{"Usage", buildUsageModule(db)},
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
		moduleInstances = append(moduleInstances, struct {
			name   string
			module router.ModuleRegisterer
		}{"Billing", buildBillingModule(db, cfg)})

		logger.Info("Billing module enabled", zap.String("module", "Billing"))
	}

	// ---------------------------------------------------------------------------
	// Social module wiring (gated behind SOCIAL_ENABLED — REQ-FLAG-01, REQ-FLAG-02)
	// ---------------------------------------------------------------------------
	socialHandlerFactory := socialwiring.NewTenantHandlerFactory(db, eventBus, cfg)
	socialMod := buildSocialModule(db, eventBus, cfg, socialHandlerFactory)
	moduleInstances = append(moduleInstances, struct {
		name   string
		module router.ModuleRegisterer
	}{"Social", socialMod})
	if cfg.SocialEnabled {
		logger.Info("Social module enabled", zap.String("module", "Social"))
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

		// Special handling for Social module to start its scheduler when workers enabled
		if m.name == "Social" && enableWorkers {
			sched := socialscheduler.NewScheduler(db, socialHandlerFactory, cfg.SocialEnabled)
			go sched.Start(context.Background())
			logger.Info("Started social scheduler",
				zap.Bool("social_enabled", cfg.SocialEnabled))
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

// devKeyFile is the gitignored file that persists the generated dev integration key.
// Keeping it stable across restarts lets existing dev DB rows remain decryptable.
const devKeyFile = ".dev-integration-key"

// loadOrGenerateDevIntegrationKey returns a stable 32-byte hex key for dev.
// On first call it generates a random key, persists it to devKeyFile, and logs it.
// On subsequent calls it loads the file. If the file is unreadable for any reason
// a fresh random key is generated for the current process (not persisted).
//
// IMPORTANT: devKeyFile is gitignored — never commit it.
func loadOrGenerateDevIntegrationKey() string {
	// Try to load from the persistent dev file first.
	if data, err := os.ReadFile(devKeyFile); err == nil {
		candidate := strings.TrimSpace(string(data))
		if len(candidate) == 64 {
			logger.Info("Loaded dev INTEGRATION_TOKEN_KEY from " + devKeyFile + " (gitignored)")
			return candidate
		}
	}

	// Generate a fresh random 32-byte key.
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		logger.Logger.Fatal("failed to generate random dev integration key", zap.Error(err))
	}
	keyHex := hex.EncodeToString(keyBytes)

	// Persist to the dev file so future restarts stay consistent.
	abs, _ := filepath.Abs(devKeyFile)
	if err := os.WriteFile(devKeyFile, []byte(keyHex+"\n"), 0600); err != nil {
		logger.Warn("Could not persist dev integration key to "+devKeyFile+" — key will change on restart", zap.Error(err))
	} else {
		logger.Warn("Generated new dev INTEGRATION_TOKEN_KEY and saved to " + abs +
			" (gitignored). Set INTEGRATION_TOKEN_KEY env var to a stable value if you need cross-restart consistency.")
	}

	return keyHex
}

// buildIntegrationModule wires the integration module: token crypto, the
// provider registry (Slack, email, Discord, Twilio, Sheets, Teams, optional
// Gmail/Outlook), and the domain-event subscriptions that drive dispatch.
func buildIntegrationModule(
	db *sql.DB,
	cfg *config.Config,
	eventBus eventbus.MessageBus,
	emailProvider emailservices.EmailProvider,
	orgRepo *orgpersistence.OrganizationPostgresRepository,
) *integration.Module {
	intKeyHex := cfg.IntegrationTokenKey
	if intKeyHex == "" {
		if cfg.Environment == "production" {
			logger.Logger.Fatal("INTEGRATION_TOKEN_KEY required in production")
		}
		intKeyHex = loadOrGenerateDevIntegrationKey()
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

	flagsReader := featureflags.NewReader(db)
	twilioPlanLookup := intwiring.NewOrgPlanLookup(db)
	twilioValidator := twilioprovider.NewValidator()
	intRegistry := buildIntegrationProviders(db, cfg, emailProvider, twilioPlanLookup)

	intRepoFactory := intwiring.NewTenantRepoFactory(db)
	intOrgGuard := intwiring.NewOrgGuard(orgRepo)
	intChannelEntitlement := intwiring.NewChannelEntitlementAdapter(twilioPlanLookup, cfg.IntegrationPaidPlans)
	intDispatcher := dispatchevent.NewHandlerWithEntitlement(intRepoFactory, intOrgGuard, intChannelEntitlement)

	subscribeIntegrationEvents(eventBus, intDispatcher)

	return integration.NewModule(integration.Deps{
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
}

// buildIntegrationProviders assembles the integration provider registry,
// appending Gmail/Outlook when their OAuth credentials are configured.
func buildIntegrationProviders(
	db *sql.DB,
	cfg *config.Config,
	emailProvider emailservices.EmailProvider,
	twilioPlanLookup *intwiring.OrgPlanLookup,
) intdomainservices.ProviderRegistry {
	slackClient := slackprovider.New(cfg.SlackClientID, cfg.SlackClientSecret)
	intEmailClient := emailprovider.New(intwiring.NewEmailAdapter(emailProvider))
	discordClient := discordprovider.New(cfg.DiscordClientID, cfg.DiscordClientSecret)
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

	baseProviders := []intdomainservices.ProviderClient{slackClient, intEmailClient, discordClient, twilioClient, sheetsClient, teamsClient}
	if cfg.GmailIntegrationEnabled && cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		baseProviders = append(baseProviders, gmailprovider.New(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.IntegrationOAuthRedirectBase))
	}
	if cfg.MicrosoftClientID != "" && cfg.MicrosoftClientSecret != "" {
		baseProviders = append(baseProviders, outlookprovider.New(cfg.MicrosoftClientID, cfg.MicrosoftClientSecret, cfg.IntegrationOAuthRedirectBase))
	}
	return intproviders.NewRegistry(baseProviders...)
}

// subscribeIntegrationEvents wires the integration dispatcher to the
// change-detected and alert-created domain events.
func subscribeIntegrationEvents(eventBus eventbus.MessageBus, intDispatcher *dispatchevent.Handler) {
	handle := func(ev eventbus.DomainEvent) {
		if err := intDispatcher.Handle(context.Background(), ev); err != nil {
			logger.Error("integration dispatch failed", zap.Error(err), zap.String("event_type", ev.Type))
		}
	}
	if err := eventbus.SubscribeDomainEvent(eventBus, eventbus.TopicChangeDetected, handle); err != nil {
		logger.Error("subscribe TopicChangeDetected", zap.Error(err))
	}
	if err := eventbus.SubscribeDomainEvent(eventBus, eventbus.TopicAlertCreated, handle); err != nil {
		logger.Error("subscribe TopicAlertCreated", zap.Error(err))
	}
}

// buildAlertModule constructs the alert module and wires its event bus.
func buildAlertModule(db *sql.DB, eventBus eventbus.MessageBus) router.ModuleRegisterer {
	m := alert.NewModuleWithDB(db)
	if am, ok := m.(*alert.Module); ok {
		am.SetEventBus(eventBus)
	}
	return m
}

// buildMonitoringModule constructs the monitoring module, building its snapshot
// worker. On snapshot-worker failure the module still runs, just without
// snapshot execution.
func buildMonitoringModule(db *sql.DB, eventBus eventbus.MessageBus, emailProvider emailservices.EmailProvider, checkBroker pubsub.CheckBroker, cfg *config.Config) router.ModuleRegisterer {
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

	// Build a presigner for browser-facing URL mediation.
	// Uses the same object storage provider as the snapshot worker.
	// On failure, mediation runs in pass-through mode (safe public default).
	presigner := monitoringwiring.NewSnapshotPresigner(cfg)

	return monitoring.NewModuleWithDeps(monitoring.Deps{
		DB:                    db,
		EventBus:              eventBus,
		SnapshotExecutor:      snapshotWorker,
		CheckBroker:           checkBroker,
		SnapshotPresigner:     presigner,
		SnapshotBucketPrivate: cfg.SnapshotBucketPrivate,
		SnapshotPresignTTL:    cfg.SnapshotPresignTTL,
		SchedulerMode:         cfg.SchedulerMode,
	})
}

// buildInsightModule constructs the insight module with its per-tenant reader
// adapters.
func buildInsightModule(db *sql.DB, insightBroker pubsub.InsightBroker) router.ModuleRegisterer {
	return insight.NewModuleWithDeps(
		db,
		insightBroker,
		func(tenant string) insightservices.CheckReader {
			return insightwiring.NewCheckReaderAdapter(db, tenant)
		},
		func(tenant string) insightservices.PageConfigReader {
			return insightwiring.NewPageConfigReaderAdapter(db, tenant)
		},
	)
}

// buildUsageModule constructs the usage module with its trial-status handler.
func buildUsageModule(db *sql.DB) router.ModuleRegisterer {
	trialReader := usagepersistence.NewTrialPlanPostgresReader(db)
	return usage.NewModuleWithTrial(db, trialstatus.NewHandler(trialReader))
}

// buildBillingModule wires the Stripe-backed billing module: gateway, repos,
// plan assigner, and all use-case handlers. Gated behind BILLING_ENABLED by the
// caller.
func buildBillingModule(db *sql.DB, cfg *config.Config) router.ModuleRegisterer {
	stripeGateway := billingstripe.NewGateway(cfg.StripeSecretKey, cfg.StripeWebhookSecret)
	planAssigner := billingwiring.NewPlanAssigner(db)

	subscriptionRepo := billingpostgres.NewSubscriptionPostgresRepository(db)
	webhookRepo := billingpostgres.NewWebhookEventPostgresRepository(db)
	customerRepo := billingpostgres.NewCustomerPostgresRepository(db)
	planRepo := billingpostgres.NewPlanPostgresRepository(db)
	listPlansHandler := listplans.NewHandler(planRepo)

	reconcileHandler := reconcilesubscription.NewHandler(stripeGateway, planAssigner, webhookRepo)

	checkoutHandler := createcheckoutsession.
		NewHandler(stripeGateway, customerRepo, planRepo).
		WithReconciler(reconcileHandler)
	portalHandler := createportalsession.NewHandler(stripeGateway, customerRepo)
	subscriptionHandler := getsubscription.NewHandler(subscriptionRepo).WithStripeGateway(stripeGateway)
	usageReader := billingwiring.NewUsageReader(db)
	updateSubHandler := updatesubscription.NewHandler(stripeGateway, planRepo, subscriptionRepo, usageReader, planAssigner)
	couponHandler := managecoupons.NewHandler(stripeGateway, planRepo, cfg.FrontendURL)
	giftHandler := billinggiftmonth.NewHandler(stripeGateway, subscriptionRepo, planRepo, planAssigner)
	cancelHandler := billingcancel.NewHandler(stripeGateway, subscriptionRepo)
	trialConverter := billingwiring.NewTrialConverter(db)
	webhookHandler := handlewebhook.NewHandler(
		stripeGateway,
		cfg.StripeWebhookSecret,
		planAssigner,
		customerRepo,
		webhookRepo,
		subscriptionRepo,
	).WithTrialConverter(trialConverter).
		WithPlanRepository(planRepo)

	return billing.NewModule(billing.Deps{
		DB:                  db,
		CheckoutHandler:     checkoutHandler,
		PortalHandler:       portalHandler,
		SubscriptionHandler: subscriptionHandler,
		WebhookHandler:      webhookHandler,
		UpdateSubHandler:    updateSubHandler,
		CouponHandler:       couponHandler,
		GiftHandler:         giftHandler,
		CancelHandler:       cancelHandler,
		ListPlansHandler:    listPlansHandler,
	})
}

// buildSocialModule wires the social module: Apify fetcher, media store, plan
// limits, alert creator, and HTTP routes. Gated behind SOCIAL_ENABLED by the
// module's RegisterHTTPRoutes (REQ-FLAG-01, REQ-FLAG-02, REQ-FLAG-03).
func buildSocialModule(
	db *sql.DB,
	bus eventbus.MessageBus,
	cfg *config.Config,
	handlerFactory *socialwiring.TenantHandlerFactory,
) router.ModuleRegisterer {
	_ = handlerFactory // used by the scheduler; not needed for HTTP module construction

	planLimits := socialwiring.NewPlanLimits(db)
	orgLookup := socialwiring.NewOrgLookup(db)

	// The HTTP manual-check trigger uses a context-aware alert creator because
	// the tenant is in the request context, not fixed at construction time.
	alertCreator := socialwiring.NewHTTPAlertCreator(bus, orgLookup)

	// TenantHandlerFactory builds the fetcher and media store internally; reuse
	// them for the HTTP module so we share the same singletons.
	fetcher := handlerFactory.Fetcher()
	mediaStore := handlerFactory.MediaStore()

	// The HTTP module's manual-check trigger passes checksPerDay=0 (unlimited at
	// construction); the quota repo enforces the real limit at consume time via the
	// atomic ON CONFLICT WHERE guard in the postgres CheckQuota implementation.
	return socialhttp.NewModule(socialhttp.Deps{
		DB:            db,
		AlertCreator:  alertCreator,
		PlanLimits:    planLimits,
		Fetcher:       fetcher,
		MediaStore:    mediaStore,
		Enabled:       cfg.SocialEnabled,
		PostsPerCheck: cfg.SocialPostsPerCheck,
		ChecksPerDay:  0,
	})
}
