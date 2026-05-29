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
	workerjobs "github.com/jcsoftdev/pulzifi-back/cmd/worker/jobs"
	emailproviders "github.com/jcsoftdev/pulzifi-back/modules/email/infrastructure/providers"
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
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/crypto"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/integrationusage"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

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

	emailProvider := emailproviders.NewResendProvider(cfg.ResendAPIKey, cfg.EmailFromAddress, cfg.EmailFromName)

	// Monitoring background processes
	startMonitoring(db, cfg, emailProvider)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	logger.Info("Worker Service is running...")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutdown signal received, shutting down worker...")
	cancel()
}

// startMonitoring builds the snapshot worker and starts the monitoring module's
// background processes.
func startMonitoring(db *sql.DB, cfg *config.Config, emailProvider *emailproviders.ResendProvider) {
	snapshotWorker, err := monitoringwiring.NewSnapshotWorker(monitoringwiring.SnapshotWorkerDeps{
		DB:            db,
		EventBus:      eventbus.GetInstance(),
		EmailProvider: emailProvider,
		FrontendURL:   cfg.FrontendURL,
		Cfg:           cfg,
	})
	if err != nil {
		logger.Error("Failed to build snapshot worker for monitoring", zap.Error(err))
	}
	monitoringMod := monitoring.NewModuleWithDeps(monitoring.Deps{
		DB:               db,
		EventBus:         eventbus.GetInstance(),
		SnapshotExecutor: snapshotWorker,
	})
	monitoringMod.StartBackgroundProcesses()
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
func buildProviderRegistry(db *sql.DB, cfg *config.Config, emailProvider *emailproviders.ResendProvider) services.ProviderRegistry {
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

// buildDeliveryWorker wires the integration delivery worker from config.
func buildDeliveryWorker(db *sql.DB, cfg *config.Config, emailProvider *emailproviders.ResendProvider) *deliveryworker.Worker {
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
		services.NewPayloadBuilder(),
		deliveryworker.Config{
			PollInterval: cfg.DeliveryPollInterval,
			PoolSize:     cfg.DeliveryWorkerPoolSize,
			MaxAttempts:  cfg.DeliveryMaxAttempts,
		},
	)
}
