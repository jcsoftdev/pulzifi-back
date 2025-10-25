package messaging

import (
	"context"
	"encoding/json"

	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/events"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Publisher publishes domain events to Kafka
type Publisher struct {
	kafkaClient *KafkaClient
}

// NewPublisher creates a new publisher
func NewPublisher(kafkaClient *KafkaClient) *Publisher {
	return &Publisher{
		kafkaClient: kafkaClient,
	}
}

// PublishOrganizationCreated publishes organization created event to Kafka
func (p *Publisher) PublishOrganizationCreated(ctx context.Context, event *events.OrganizationCreated) error {
	payload, err := json.Marshal(event)
	if err != nil {
		logger.Error("Failed to marshal organization created event", zap.Error(err))
		return err
	}

	err = p.kafkaClient.Producer.Produce(
		"organization.created",
		event.ID.String(),
		payload,
	)
	if err != nil {
		logger.Error("Failed to publish organization.created event", zap.Error(err))
		return err
	}

	logger.Info("Published organization.created event", zap.String("organization_id", event.ID.String()))
	return nil
}

// PublishOrganizationDeleted publishes organization deleted event to Kafka
func (p *Publisher) PublishOrganizationDeleted(ctx context.Context, event *events.OrganizationDeleted) error {
	payload, err := json.Marshal(event)
	if err != nil {
		logger.Error("Failed to marshal organization deleted event", zap.Error(err))
		return err
	}

	err = p.kafkaClient.Producer.Produce(
		"organization.deleted",
		event.ID.String(),
		payload,
	)
	if err != nil {
		logger.Error("Failed to publish organization.deleted event", zap.Error(err))
		return err
	}

	logger.Info("Published organization.deleted event", zap.String("organization_id", event.ID.String()))
	return nil
}

// PublishOrganizationUpdated publishes organization updated event to Kafka
func (p *Publisher) PublishOrganizationUpdated(ctx context.Context, event *events.OrganizationUpdated) error {
	payload, err := json.Marshal(event)
	if err != nil {
		logger.Error("Failed to marshal organization updated event", zap.Error(err))
		return err
	}

	err = p.kafkaClient.Producer.Produce(
		"organization.updated",
		event.ID.String(),
		payload,
	)
	if err != nil {
		logger.Error("Failed to publish organization.updated event", zap.Error(err))
		return err
	}

	logger.Info("Published organization.updated event", zap.String("organization_id", event.ID.String()))
	return nil
}
