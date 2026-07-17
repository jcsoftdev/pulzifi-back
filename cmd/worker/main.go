package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"os"
	"os/signal"
	"syscall"
	"time"

	intwiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/integration"
	monitoringwiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/monitoring"
	socialwiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/social"
	workerjobs "github.com/jcsoftdev/pulzifi-back/cmd/worker/jobs"
	emailservices "github.com/jcsoftdev/pulzifi-back/modules/email/domain/services"
	emailproviders "github.com/jcsoftdev/pulzifi-back/modules/email/infrastructure/providers"
	dispatchevent "github.com/jcsoftdev/pulzifi-back/modules/integration/application/dispatch_event"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
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
	deliveryworker "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/worker"
	monitoring "github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/http"
	orgpersistence "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/persistence"
	socialscheduler "github.com/jcsoftdev/pulzifi-back/modules/social/infrastructure/scheduler"
	"github.com/jcsoftdev/pulzifi-back/shared/cache"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/crypto"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/integrationusage"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// createEmailProvider selects the email provider from EMAIL_PROVIDER.
// "log" is used for local/E2E testing so the pipeline never hits the real
// Resend API; anything else (including unset) falls back to Resend. Mirrors
// cmd/server/modules.go's createEmailProvider.
func createEmailProvider(cfg *config.Config) emailservices.EmailProvider {
	if cfg.EmailProvider == "log" {
		return emailproviders.NewLogProvider()
	}
	return emailproviders.NewResendProvider(cfg.ResendAPIKey, cfg.EmailFromAddress, cfg.EmailFromName)
}

func main() {
	cfg := config.Load()
	logger.Info("Starting Pulzifi Worker Service", zap.String("environment", cfg.Environment))

	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	// Initialize Redis.
	// In single mode optional; in HA mode required (fatal if unreachable).
	if err := cache.InitRedis(cfg); err != nil {
		if cfg.InstanceMode == "ha" {
			logger.Logger.Fatal("Worker: HA mode requires Redis but connection failed — refusing to start",
				zap.Error(err),
				zap.String("instance_mode", cfg.InstanceMode),
			)
		}
		logger.Warn("Worker: Failed to initialize Redis - operating without cache", zap.Error(err))
	} else {
		defer func() { _ = cache.CloseRedis() }()
		logger.Info("Worker: Redis initialized successfully")
	}

	// Build the event bus from config. When provider=outbox, the relay is
	// started below (after all subscribers are registered).
	bus, err := eventbus.NewFromConfig(cfg.EventBusProvider, cfg.OutboxBatchSize, cfg.OutboxMaxAttempts, db)
	if err != nil {
		logger.Error("Failed to initialize event bus", zap.Error(err))
		os.Exit(1) //nolint:gocritic // fatal startup error; OS reclaims db connection
	}
	defer bus.Close()

	emailProvider := createEmailProvider(cfg)

	// Wire the integration dispatcher onto the worker's bus BEFORE detection starts.
	// Change detection runs in THIS process and publishes change.detected / alert.created
	// events; without a subscriber here, the in-memory bus drops them and no integration
	// delivery (email destinations, Slack, etc.) is ever created. Mirrors cmd/server.
	subscribeIntegrationEvents(db, cfg, bus)

	// Monitoring background processes
	startMonitoringWithBus(db, cfg, emailProvider, bus)

	// Social scheduler (REQ-SCHED-05) — gated by SOCIAL_ENABLED.
	startSocialScheduler(db, cfg, bus, emailProvider)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start outbox relay when using the outbox provider (after all subscribers registered).
	if ob, ok := bus.(*eventbus.OutboxBus); ok {
		ob.StartRelay(ctx)
		logger.Info("Outbox relay started", zap.String("provider", cfg.EventBusProvider))
	}

	// Integration delivery worker
	delWorker := buildDeliveryWorker(db, cfg, emailProvider)
	go func() {
		if err := delWorker.Run(ctx); err != nil && err != context.Canceled {
			logger.Logger.Fatal("delivery worker stopped", zap.Error(err))
		}
	}()

	// ---------------------------------------------------------------------------
	// Trial expirer cron job
	// ---------------------------------------------------------------------------
	trialExpirer := workerjobs.NewTrialExpirer(db, emailProvider, cfg.FrontendURL, time.Hour)
	go func() {
		if err := trialExpirer.Run(ctx); err != nil && err != context.Canceled {
			logger.Warn("trial expirer stopped", zap.Error(err))
		}
	}()

	// ---------------------------------------------------------------------------
	// Snapshot retention job
	// ---------------------------------------------------------------------------
	retentionRunner, err := workerjobs.NewSnapshotRetentionRunner(db, cfg)
	if err != nil {
		logger.Warn("snapshot retention runner could not be initialized — skipping", zap.Error(err))
	} else {
		go func() {
			if runErr := retentionRunner.Run(ctx); runErr != nil && runErr != context.Canceled {
				logger.Warn("snapshot retention runner stopped", zap.Error(runErr))
			}
		}()
	}

	logger.Info("Worker Service is running...")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutdown signal received, shutting down worker...")
	cancel()
}

// startMonitoringWithBus builds the snapshot worker and starts the monitoring
// module's background processes using the provided event bus.
func startMonitoringWithBus(db *sql.DB, cfg *config.Config, emailProvider emailservices.EmailProvider, bus eventbus.MessageBus) {
	snapshotWorker, err := monitoringwiring.NewSnapshotWorker(monitoringwiring.SnapshotWorkerDeps{
		DB:            db,
		EventBus:      bus,
		EmailProvider: emailProvider,
		FrontendURL:   cfg.FrontendURL,
		Cfg:           cfg,
	})
	if err != nil {
		logger.Error("Failed to build snapshot worker for monitoring", zap.Error(err))
	}
	monitoringMod := monitoring.NewModuleWithDeps(monitoring.Deps{
		DB:               db,
		EventBus:         bus,
		SnapshotExecutor: snapshotWorker,
		SchedulerMode:    cfg.SchedulerMode,
	})
	monitoringMod.StartBackgroundProcesses()
}

// subscribeIntegrationEvents builds the integration dispatch handler and subscribes
// it to change-detected and alert-created domain events on the worker's bus. Detection
// runs in the worker process, so the subscriber must live here too — otherwise events
// published in this process reach no consumer and every delivery is dropped. Mirrors
// cmd/server/modules.go's subscribeIntegrationEvents.
func subscribeIntegrationEvents(db *sql.DB, cfg *config.Config, bus eventbus.MessageBus) {
	orgRepo := orgpersistence.NewOrganizationPostgresRepository(db)
	intRepoFactory := intwiring.NewTenantRepoFactory(db)
	intOrgGuard := intwiring.NewOrgGuard(orgRepo)
	twilioPlanLookup := intwiring.NewOrgPlanLookup(db)
	intChannelEntitlement := intwiring.NewChannelEntitlementAdapter(twilioPlanLookup, cfg.IntegrationPaidPlans)
	intDispatcher := dispatchevent.NewHandlerWithEntitlement(intRepoFactory, intOrgGuard, intChannelEntitlement)
	intDispatcher.SetFallbackResolver(intwiring.NewMemberEmailResolver(db))

	handle := func(ev eventbus.DomainEvent) {
		if err := intDispatcher.Handle(context.Background(), ev); err != nil {
			logger.Error("integration dispatch failed", zap.Error(err), zap.String("event_type", ev.Type))
		}
	}
	if err := eventbus.SubscribeDomainEvent(bus, eventbus.TopicChangeDetected, handle); err != nil {
		logger.Error("subscribe TopicChangeDetected", zap.Error(err))
	}
	if err := eventbus.SubscribeDomainEvent(bus, eventbus.TopicAlertCreated, handle); err != nil {
		logger.Error("subscribe TopicAlertCreated", zap.Error(err))
	}
	logger.Info("Worker: integration event dispatcher subscribed (change.detected, alert.created)")
}

// buildIntegrationKey resolves and validates the integration token key,
// falling back to an insecure dev default outside production.
func buildIntegrationKey(cfg *config.Config) []byte {
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
	return intKey
}

// buildProviderRegistry assembles the provider clients enabled by config.
func buildProviderRegistry(db *sql.DB, cfg *config.Config, emailProvider emailservices.EmailProvider) services.ProviderRegistry {
	slackClient := slackprovider.New(cfg.SlackClientID, cfg.SlackClientSecret)
	intEmailClient := emailprovider.New(intwiring.NewEmailAdapter(emailProvider))
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
	workerProviders := []services.ProviderClient{slackClient, intEmailClient, discordClient, twilioClient, sheetsClient, teamsClient}
	if cfg.GmailIntegrationEnabled && cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		workerProviders = append(workerProviders, gmailprovider.New(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.IntegrationOAuthRedirectBase))
	}
	if cfg.MicrosoftClientID != "" && cfg.MicrosoftClientSecret != "" {
		workerProviders = append(workerProviders, outlookprovider.New(cfg.MicrosoftClientID, cfg.MicrosoftClientSecret, cfg.IntegrationOAuthRedirectBase))
	}
	return intproviders.NewRegistry(workerProviders...)
}

// startSocialScheduler constructs the social scheduler and starts it in a
// background goroutine. It is a no-op when SOCIAL_ENABLED=false (REQ-SCHED-05).
func startSocialScheduler(db *sql.DB, cfg *config.Config, bus eventbus.MessageBus, emailProvider emailservices.EmailProvider) {
	handlerFactory := socialwiring.NewTenantHandlerFactory(db, bus, cfg, emailProvider)
	sched := socialscheduler.NewScheduler(db, handlerFactory, cfg.SocialEnabled)
	go sched.Start(context.Background())
	logger.Info("Social scheduler started",
		zap.Bool("social_enabled", cfg.SocialEnabled))
}

// buildDeliveryWorker wires the integration delivery worker from config.
func buildDeliveryWorker(db *sql.DB, cfg *config.Config, emailProvider emailservices.EmailProvider) *deliveryworker.Worker {
	intKey := buildIntegrationKey(cfg)
	intEnc, err := crypto.NewAESGCM(intKey)
	if err != nil {
		logger.Logger.Fatal("crypto init failed", zap.Error(err))
	}

	intRepo := intpersistence.NewIntegrationPostgresRepository(db, intEnc)
	intRegistry := buildProviderRegistry(db, cfg, emailProvider)
	intRepoFactory := intwiring.NewTenantRepoFactory(db)

	return deliveryworker.New(
		db,
		intRepoFactory,
		intRepo,
		intRegistry,
		services.NewPayloadBuilder(cfg.AppDomain),
		deliveryworker.Config{
			PollInterval: cfg.DeliveryPollInterval,
			PoolSize:     cfg.DeliveryWorkerPoolSize,
			MaxAttempts:  cfg.DeliveryMaxAttempts,
		},
	)
}
