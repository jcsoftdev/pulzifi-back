package messaging

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Subscriber subscribes to topics using MessageBus
type Subscriber struct {
	bus eventbus.MessageBus
	db  *sql.DB
}

// NewSubscriber creates a new subscriber
func NewSubscriber(bus eventbus.MessageBus, db *sql.DB) *Subscriber {
	return &Subscriber{
		bus: bus,
		db:  db,
	}
}

// ListenToEvents subscribes to topics
func (s *Subscriber) ListenToEvents(ctx context.Context) {
	logger.Info("Starting subscriber for organization module")

	// Topics this module subscribes to
	topics := []string{
		"user.deleted",
		"workspace.created",
	}

	for _, topic := range topics {
		// Define handler wrapper to pass context
		handler := func(t string, payload []byte) {
			logger.Info("Received event", zap.String("topic", t), zap.Int("payload_size", len(payload)))
			s.handleEvent(context.Background(), t, "", payload) // In-memory bus doesn't support key yet, passing empty
		}

		if err := s.bus.Subscribe(topic, handler); err != nil {
			logger.Error("Failed to subscribe to topic", zap.String("topic", topic), zap.Error(err))
		}
	}

	logger.Info("Organization module subscriber ready")

	// Keep alive until context cancelled
	<-ctx.Done()
	logger.Info("Stopping subscriber")
}

// handleEvent processes incoming Kafka messages
func (s *Subscriber) handleEvent(ctx context.Context, topic string, key string, payload []byte) {
	switch topic {
	case "user.deleted":
		s.handleUserDeleted(ctx, key, payload)
	case "workspace.created":
		s.handleWorkspaceCreated(ctx, key, payload)
	default:
		logger.Warn("Unknown event topic", zap.String("topic", topic))
	}
}

// userDeletedPayload represents the JSON payload for a user.deleted event
type userDeletedPayload struct {
	UserID string `json:"user_id"`
}

// handleUserDeleted handles user deletion events from other modules.
//
// Design D4: the destructive cascade (org soft-delete, hard-delete, DROP SCHEMA)
// is now owned entirely by the synchronous delete_organization use case. This
// subscriber's destructive branch has been neutralized to prevent double-cascade
// races. The handler is kept subscribed for future non-destructive listeners
// (analytics, logging, etc.).
func (s *Subscriber) handleUserDeleted(_ context.Context, _ string, payload []byte) {
	var event userDeletedPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		logger.Error("Failed to parse user.deleted payload", zap.Error(err))
		return
	}

	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		logger.Error("Invalid user_id in user.deleted payload",
			zap.String("raw_user_id", event.UserID),
			zap.Error(err),
		)
		return
	}

	// Org cascade is now handled synchronously by delete_organization use case
	// (invoked from DELETE /auth/me and DELETE /organizations/{id}). The subscriber
	// must NOT perform any destructive DB writes to avoid double-cascade races
	// (DAO-14, design §4).
	logger.Debug("user.deleted: org cascade handled synchronously, skipping destructive branch",
		zap.String("user_id", userID.String()),
	)
}

// handleWorkspaceCreated handles workspace creation events from other modules
// When a workspace is created, we may need to update organization metadata
func (s *Subscriber) handleWorkspaceCreated(ctx context.Context, workspaceID string, payload []byte) {
	logger.Info("Processing workspace.created event",
		zap.String("workspace_id", workspaceID),
	)

	// This is informational - just log that we received it
	// Organization module doesn't need to do anything when workspaces are created
	// This handler exists for future extensibility
}
