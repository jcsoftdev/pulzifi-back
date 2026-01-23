package consumer

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/kafka"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type SnapshotConsumer struct {
	db       *sql.DB
	consumer *kafka.ConsumerClient
}

func NewSnapshotConsumer(db *sql.DB) *SnapshotConsumer {
	cfg := config.Load()
	consumer, err := kafka.NewConsumerClient(cfg, "monitoring-snapshot-group", []string{"snapshot-completed"})
	if err != nil {
		logger.Error("Failed to create snapshot consumer", zap.Error(err))
		return nil
	}
	return &SnapshotConsumer{db: db, consumer: consumer}
}

func (c *SnapshotConsumer) Start(ctx context.Context) {
	if c.consumer == nil {
		return
	}
	
	go func() {
		defer c.consumer.Close()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				topic, _, value, err := c.consumer.ReadMessage(100)
				if err != nil {
					if kafkaErr, ok := err.(ckafka.Error); ok && kafkaErr.Code() == ckafka.ErrTimedOut {
						continue
					}
					// logger.Error("Error reading snapshot result", zap.Error(err))
					continue
				}

				c.processMessage(ctx, value, topic)
			}
		}
	}()
	logger.Info("Snapshot Consumer started")
}

func (c *SnapshotConsumer) processMessage(ctx context.Context, value []byte, topic string) {
	type SnapshotResult struct {
		PageID       string    `json:"page_id"`
		URL          string    `json:"url"`
		SchemaName   string    `json:"schema_name"`
		ImageURL     string    `json:"image_url"`
		Status       string    `json:"status"`
		ErrorMessage string    `json:"error_message,omitempty"`
		CreatedAt    time.Time `json:"created_at"`
	}

	var result SnapshotResult
	if err := json.Unmarshal(value, &result); err != nil {
		logger.Error("Failed to unmarshal snapshot result", zap.Error(err))
		return
	}
	
	logger.Info("Processing snapshot result", zap.String("page_id", result.PageID), zap.String("status", result.Status))

	if result.SchemaName == "" {
		logger.Error("Snapshot result missing schema name", zap.String("page_id", result.PageID))
		return
	}

	pageID, err := uuid.Parse(result.PageID)
	if err != nil {
		logger.Error("Invalid PageID in snapshot result", zap.String("page_id", result.PageID))
		return
	}

	// Create Check record
	checkRepo := persistence.NewCheckPostgresRepository(c.db, result.SchemaName)
	
	check := &entities.Check{
		ID:            uuid.New(),
		PageID:        pageID,
		Status:        result.Status,
		ScreenshotURL: result.ImageURL,
		CheckedAt:     result.CreatedAt,
		// Change detection logic would go here. For now default to false.
		ChangeDetected: false, 
	}
	if result.Status == "failed" {
		check.ErrorMessage = result.ErrorMessage
	} else {
		// Update Page Thumbnail
		// I'll execute raw SQL to update page
		q := "UPDATE " + result.SchemaName + ".pages SET thumbnail_url = $1, last_checked_at = $2 WHERE id = $3"
		if _, err := c.db.ExecContext(ctx, q, result.ImageURL, result.CreatedAt, pageID); err != nil {
			logger.Error("Failed to update page thumbnail", zap.Error(err))
		}
	}

	if err := checkRepo.Create(ctx, check); err != nil {
		logger.Error("Failed to create check record", zap.Error(err))
	}
}
