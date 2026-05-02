package main

import (
	"context"
	"encoding/hex"
	"os"
	"os/signal"
	"syscall"

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
	intwiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/integration"
	emailproviders "github.com/jcsoftdev/pulzifi-back/modules/email/infrastructure/providers"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
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
	defer db.Close()

	// ---------------------------------------------------------------------------
	// Monitoring background processes
	// ---------------------------------------------------------------------------
	mod := monitoring.NewModuleWithDB(db, eventbus.GetInstance(), nil, "")
	if monitoringModule, ok := mod.(*monitoring.Module); ok {
		monitoringModule.StartBackgroundProcesses()
	} else {
		logger.Logger.Fatal("Failed to cast monitoring module")
	}

	// ---------------------------------------------------------------------------
	// Integration delivery worker
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

	emailProvider := emailproviders.NewResendProvider(cfg.ResendAPIKey, cfg.EmailFromAddress, cfg.EmailFromName)
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
	if cfg.GmailIntegrationEnabled {
		workerProviders = append(workerProviders, gmailprovider.New(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.IntegrationOAuthRedirectBase))
	}
	if cfg.MicrosoftClientID != "" && cfg.MicrosoftClientSecret != "" {
		workerProviders = append(workerProviders, outlookprovider.New(cfg.MicrosoftClientID, cfg.MicrosoftClientSecret, cfg.IntegrationOAuthRedirectBase))
	}
	intRegistry := intproviders.NewRegistry(workerProviders...)

	intRepoFactory := intwiring.NewTenantRepoFactory(db)

	delWorker := deliveryworker.New(
		db,
		intRepoFactory,
		intRepo,
		intRegistry,
		services.NewPayloadBuilder(),
		deliveryworker.Config{
			PollInterval:   cfg.DeliveryPollInterval,
			PoolSize:       cfg.DeliveryWorkerPoolSize,
			MaxAttempts:    cfg.DeliveryMaxAttempts,
		},
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := delWorker.Run(ctx); err != nil && err != context.Canceled {
			logger.Logger.Fatal("delivery worker stopped", zap.Error(err))
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
