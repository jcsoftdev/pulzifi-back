package messaging

import (
	kafkaclient "github.com/jcsoftdev/pulzifi-back/shared/kafka"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// KafkaClient wraps both producer and consumer clients
type KafkaClient struct {
	Producer *kafkaclient.ProducerClient
	Consumer *kafkaclient.ConsumerClient
}

// NewKafkaClient initializes both Kafka producer and consumer
func NewKafkaClient(cfg *config.Config) (*KafkaClient, error) {
	// Create producer
	producer, err := kafkaclient.NewProducerClient(cfg)
	if err != nil {
		logger.Error("Failed to create Kafka producer", zap.Error(err))
		return nil, err
	}

	// Create consumer for organization events from other modules
	// Topics: user.deleted, workspace.created, etc.
	topics := []string{
		"user.deleted",
		"workspace.created",
	}
	
	consumer, err := kafkaclient.NewConsumerClient(cfg, "organization-service", topics)
	if err != nil {
		logger.Error("Failed to create Kafka consumer", zap.Error(err))
		producer.Close()
		return nil, err
	}

	logger.Info("Kafka client initialized successfully")

	return &KafkaClient{
		Producer: producer,
		Consumer: consumer,
	}, nil
}

// Close closes both producer and consumer
func (k *KafkaClient) Close() {
	k.Producer.Close()
	k.Consumer.Close()
	logger.Info("Kafka client closed")
}
