package messaging

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Subscriber subscribes to topics using MessageBus
type Subscriber struct {
	bus eventbus.MessageBus
}

// NewSubscriber creates a new subscriber
func NewSubscriber(bus eventbus.MessageBus) *Subscriber {
	return &Subscriber{
		bus: bus,
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

// handleUserDeleted handles user deletion events from other modules
// When a user is deleted, we should cascade delete their organizations
func (s *Subscriber) handleUserDeleted(ctx context.Context, userID string, payload []byte) {
	logger.Info("Processing user.deleted event",
		zap.String("user_id", userID),
	)

	// TODO: Implement cascade delete logic:
	// 1. Query organizations where owner_user_id = userID
	// 2. For each organization:
	//    a. Delete the tenant schema
	//    b. Mark organization as deleted in public.organizations
	//    c. Publish organization.deleted event
	logger.Warn("Cascade delete for user organizations not yet implemented",
		zap.String("user_id", userID),
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
