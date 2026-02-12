package messaging

import (
	"context"
	"encoding/json"

	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/events"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Publisher publishes domain events to the MessageBus (Kafka or In-Memory)
type Publisher struct {
	bus eventbus.MessageBus
}

// NewPublisher creates a new publisher
func NewPublisher(bus eventbus.MessageBus) *Publisher {
	return &Publisher{
		bus: bus,
	}
}

// PublishOrganizationCreated publishes organization created event
func (p *Publisher) PublishOrganizationCreated(ctx context.Context, event *events.OrganizationCreated) error {
	payload, err := json.Marshal(event)
	if err != nil {
		logger.Error("Failed to marshal organization created event", zap.Error(err))
		return err
	}

	err = p.bus.Publish(
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

// PublishOrganizationDeleted publishes organization deleted event
func (p *Publisher) PublishOrganizationDeleted(ctx context.Context, event *events.OrganizationDeleted) error {
	payload, err := json.Marshal(event)
	if err != nil {
		logger.Error("Failed to marshal organization deleted event", zap.Error(err))
		return err
	}

	err = p.bus.Publish(
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

// PublishOrganizationUpdated publishes organization updated event
func (p *Publisher) PublishOrganizationUpdated(ctx context.Context, event *events.OrganizationUpdated) error {
	payload, err := json.Marshal(event)
	if err != nil {
		logger.Error("Failed to marshal organization updated event", zap.Error(err))
		return err
	}

	err = p.bus.Publish(
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
