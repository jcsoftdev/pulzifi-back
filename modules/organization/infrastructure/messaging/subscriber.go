package messaging

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Subscriber subscribes to Kafka topics for events from other modules
type Subscriber struct {
	kafkaClient *KafkaClient
}

// NewSubscriber creates a new subscriber
func NewSubscriber(kafkaClient *KafkaClient) *Subscriber {
	return &Subscriber{
		kafkaClient: kafkaClient,
	}
}

// ListenToEvents starts listening to Kafka topics for events from other modules
// This module subscribes to events from other services that affect organizations
func (s *Subscriber) ListenToEvents(ctx context.Context) {
	logger.Info("Starting Kafka subscriber for organization module")

	// Topics this module subscribes to (from other modules):
	// "user.deleted" - to cascade delete user's organizations
	// "workspace.created" - if needed for organization tracking

	logger.Info("Organization module subscriber ready to listen to Kafka topics")

	// Listen in a loop
	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping Kafka subscriber")
			return
		default:
			topic, key, payload, err := s.kafkaClient.Consumer.ReadMessage(5000) // 5 second timeout
			if err != nil {
				// Timeout is expected
				continue
			}

			logger.Info("Received event from Kafka",
				zap.String("topic", topic),
				zap.String("key", key),
				zap.Int("payload_size", len(payload)),
			)

			s.handleEvent(ctx, topic, key, payload)
		}
	}
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
