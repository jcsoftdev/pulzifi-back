package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	generateinsights "github.com/jcsoftdev/pulzifi-back/modules/insight/application/generate_insights"
	insightAI "github.com/jcsoftdev/pulzifi-back/modules/insight/infrastructure/ai"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/application"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/extractor"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/storage"
	sharedAI "github.com/jcsoftdev/pulzifi-back/shared/ai"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	logger.Info("Starting Snapshot Service", zap.String("module", cfg.ModuleName))

	// Init DB
	db, err := database.Connect(cfg)
	if err != nil {
		logger.Logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Init Object Storage
	objectStorage, err := storage.NewObjectStorage(cfg)
	if err != nil {
		logger.Logger.Fatal("Failed to create object storage client", zap.Error(err))
	}
	// Ensure object storage is ready with retry
	ensureBucketCtx := context.Background()
	for i := 0; i < 30; i++ {
		if err := objectStorage.EnsureBucket(ensureBucketCtx); err != nil {
			logger.Logger.Error("Failed to initialize object storage, retrying...", zap.Error(err), zap.Int("attempt", i+1))
			time.Sleep(2 * time.Second)
		} else {
			break
		}
		if i == 29 {
			logger.Logger.Fatal("Failed to initialize object storage after retries")
		}
	}

	// Init Extractor Client
	extractorClient := extractor.NewHTTPClient(cfg.ExtractorURL)

	// Init Insight Handler (optional — requires OPENROUTER_API_KEY)
	var insightHandler *generateinsights.GenerateInsightsHandler
	if cfg.OpenRouterAPIKey != "" {
		openRouterClient := sharedAI.NewOpenRouterClient(cfg.OpenRouterAPIKey, cfg.OpenRouterModel)
		generator := insightAI.NewOpenRouterGenerator(openRouterClient)
		insightHandler = generateinsights.NewGenerateInsightsHandler(generator, db)
		logger.Info("Intelligent Insights enabled", zap.String("model", cfg.OpenRouterModel))
	} else {
		logger.Info("Intelligent Insights disabled (set OPENROUTER_API_KEY to enable)")
	}

	// Init Snapshot Service
	snapshotService := application.NewSnapshotService(objectStorage, extractorClient, db, insightHandler)

	// Init Messaging (In-Memory for MVP)
	messageBus := eventbus.GetInstance()

	// Use generic message bus (Kafka logic removed for MVP)
	// For now, we simulate subscription
	go func() {
		logger.Logger.Info("Snapshot service listening for requests (In-Memory)")
		messageBus.Subscribe("snapshot-requests", func(topic string, payload []byte) {
			logger.Logger.Info("Received snapshot request", zap.ByteString("payload", payload))
			var req entities.SnapshotRequest
			if err := json.Unmarshal(payload, &req); err != nil {
				logger.Logger.Error("Failed to unmarshal request", zap.Error(err))
				return
			}
			processRequest(messageBus, snapshotService, req)
		})
	}()

	// Signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("Snapshot Service is running...")

	runLoop := true
	for runLoop {
		select {
		case <-sigChan:
			runLoop = false
		default:
			// In In-Memory mode, the subscriber runs in its own goroutine
			// We just need to keep the main thread alive and handle graceful shutdown
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Helper to handle request processing (was inline before)
func processRequest(messageBus eventbus.MessageBus, snapshotService *application.SnapshotService, req entities.SnapshotRequest) {
	// Process with retry
	var result *entities.SnapshotResult
	var processErr error

	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<attempt) * time.Second
			logger.Logger.Info("Retrying snapshot capture", zap.Int("attempt", attempt+1), zap.Duration("backoff", backoff))
			time.Sleep(backoff)
		}

		result, processErr = snapshotService.CaptureAndUpload(context.Background(), req)
		if processErr == nil {
			break
		}
	}

	if processErr != nil {
		logger.Logger.Error("Failed to capture snapshot after retries", zap.Error(processErr))

		// Produce failure event
		failureResult := &entities.SnapshotResult{
			PageID:       req.PageID,
			URL:          req.URL,
			SchemaName:   req.SchemaName,
			Status:       "failed",
			ErrorMessage: processErr.Error(),
			CreatedAt:    time.Now(),
		}

		resultBytes, _ := json.Marshal(failureResult)
		if err := messageBus.Publish("snapshot-completed", req.PageID, resultBytes); err != nil {
			logger.Logger.Error("Failed to produce failure event", zap.Error(err))
		}
		return
	}

	// Produce completion event
	resultBytes, _ := json.Marshal(result)
	if err := messageBus.Publish("snapshot-completed", req.PageID, resultBytes); err != nil {
		logger.Logger.Error("Failed to produce completion event", zap.Error(err))
	}
}
