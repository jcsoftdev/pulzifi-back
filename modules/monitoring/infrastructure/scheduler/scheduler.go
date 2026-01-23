package scheduler

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/shared/kafka"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type Scheduler struct {
	db       *sql.DB
	producer *kafka.ProducerClient
}

func NewScheduler(db *sql.DB, producer *kafka.ProducerClient) *Scheduler {
	return &Scheduler{
		db:       db,
		producer: producer,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				s.runCheck(ctx)
			}
		}
	}()
	logger.Info("Monitoring Scheduler started")
}

func (s *Scheduler) runCheck(ctx context.Context) {
	// Query all schema names
	rows, err := s.db.QueryContext(ctx, "SELECT schema_name FROM organizations WHERE deleted_at IS NULL")
	if err != nil {
		logger.Error("Scheduler failed to fetch organizations", zap.Error(err))
		return
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			continue
		}
		schemas = append(schemas, schema)
	}

	for _, schema := range schemas {
		s.processTenant(ctx, schema)
	}
}

func (s *Scheduler) processTenant(ctx context.Context, schema string) {
	repo := persistence.NewMonitoringConfigPostgresRepository(s.db, schema)
	tasks, err := repo.GetDueSnapshotTasks(ctx)
	if err != nil {
		logger.Error("Failed to get due tasks", zap.String("tenant", schema), zap.Error(err))
		return
	}

	for _, task := range tasks {
		// Send to Kafka
		payload := map[string]interface{}{
			"page_id":     task.PageID.String(),
			"url":         task.URL,
			"schema_name": schema,
		}
		bytes, _ := json.Marshal(payload)
		
		err := s.producer.Produce("snapshot-requests", task.PageID.String(), bytes)
		if err != nil {
			logger.Error("Failed to produce snapshot request", zap.String("page_id", task.PageID.String()), zap.Error(err))
			continue
		}
		
		s.updateLastCheckedAt(ctx, schema, task.PageID)
	}
}

func (s *Scheduler) updateLastCheckedAt(ctx context.Context, schema string, pageID uuid.UUID) {
    q := "UPDATE " + schema + ".pages SET last_checked_at = NOW() WHERE id = $1"
    s.db.ExecContext(ctx, q, pageID)
}
