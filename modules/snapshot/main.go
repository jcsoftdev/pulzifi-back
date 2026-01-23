package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/application"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/minio"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/kafka"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	logger.Info("Starting Snapshot Service", zap.String("module", cfg.ModuleName))

	// Init MinIO
	minioClient, err := minio.NewClient(cfg)
	if err != nil {
		logger.Logger.Fatal("Failed to create MinIO client", zap.Error(err))
	}
	// Ensure bucket exists with retry
	ensureBucketCtx := context.Background()
	for i := 0; i < 30; i++ {
		if err := minioClient.EnsureBucket(ensureBucketCtx); err != nil {
			logger.Logger.Error("Failed to ensure bucket, retrying...", zap.Error(err), zap.Int("attempt", i+1))
			time.Sleep(2 * time.Second)
		} else {
			break
		}
		if i == 29 {
			logger.Logger.Fatal("Failed to ensure bucket after retries")
		}
	}

	// Init Snapshot Service
	snapshotService := application.NewSnapshotService(minioClient)

	// Init Kafka Producer (for results)
	producer, err := kafka.NewProducerClient(cfg)
	if err != nil {
		logger.Logger.Fatal("Failed to create Kafka producer", zap.Error(err))
	}
	defer producer.Close()

	// Init Kafka Consumer (for requests)
	consumer, err := kafka.NewConsumerClient(cfg, "snapshot-group", []string{"snapshot-requests"})
	if err != nil {
		logger.Logger.Fatal("Failed to create Kafka consumer", zap.Error(err))
	}
	defer consumer.Close()

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
			// Read with 100ms timeout
			topic, _, value, err := consumer.ReadMessage(100)
			if err != nil {
				// Check for timeout
				if kafkaErr, ok := err.(ckafka.Error); ok && kafkaErr.Code() == ckafka.ErrTimedOut {
					continue
				}
				// Log other errors but don't spam if it's just no message
				// logger.Error("Error reading message", zap.Error(err))
				continue
			}

			logger.Info("Received snapshot request", zap.String("topic", topic))

			var req entities.SnapshotRequest
			if err := json.Unmarshal(value, &req); err != nil {
				logger.Error("Failed to unmarshal request", zap.Error(err))
				continue
			}

			// Process with retry
			var result *entities.SnapshotResult
			var processErr error

			for attempt := 0; attempt < 3; attempt++ {
				if attempt > 0 {
					backoff := time.Duration(1<<attempt) * time.Second
					logger.Info("Retrying snapshot capture", zap.Int("attempt", attempt+1), zap.Duration("backoff", backoff))
					time.Sleep(backoff)
				}

				result, processErr = snapshotService.CaptureAndUpload(context.Background(), req)
				if processErr == nil {
					break
				}
			}

			if processErr != nil {
				logger.Error("Failed to capture snapshot after retries", zap.Error(processErr))

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
				if err := producer.Produce("snapshot-completed", req.PageID, resultBytes); err != nil {
					logger.Error("Failed to produce failure event", zap.Error(err))
				}
				continue
			}

			// Produce completion event
			resultBytes, _ := json.Marshal(result)
			if err := producer.Produce("snapshot-completed", req.PageID, resultBytes); err != nil {
				logger.Error("Failed to produce completion event", zap.Error(err))
			}
		}
	}
}
