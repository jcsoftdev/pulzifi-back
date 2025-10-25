package kafka

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// ProducerClient wraps Kafka producer
type ProducerClient struct {
	producer *kafka.Producer
}

// ConsumerClient wraps Kafka consumer
type ConsumerClient struct {
	consumer *kafka.Consumer
}

// NewProducerClient creates a new Kafka producer
func NewProducerClient(cfg *config.Config) (*ProducerClient, error) {
	brokers := cfg.KafkaBrokers
	if brokers == "" {
		brokers = "localhost:9092"
	}

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"client.id":         cfg.ModuleName,
	})

	if err != nil {
		logger.Error("Failed to create Kafka producer", zap.Error(err))
		return nil, err
	}

	logger.Info("Kafka producer created", zap.String("brokers", brokers))
	return &ProducerClient{producer: p}, nil
}

// NewConsumerClient creates a new Kafka consumer
func NewConsumerClient(cfg *config.Config, groupID string, topics []string) (*ConsumerClient, error) {
	brokers := cfg.KafkaBrokers
	if brokers == "" {
		brokers = "localhost:9092"
	}

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		logger.Error("Failed to create Kafka consumer", zap.Error(err))
		return nil, err
	}

	if err := c.SubscribeTopics(topics, nil); err != nil {
		logger.Error("Failed to subscribe to topics", zap.Error(err))
		return nil, err
	}

	logger.Info("Kafka consumer created",
		zap.String("brokers", brokers),
		zap.String("group_id", groupID),
		zap.Strings("topics", topics))

	return &ConsumerClient{consumer: c}, nil
}

// Produce sends a message to Kafka
func (p *ProducerClient) Produce(topic string, key string, value []byte) error {
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(key),
		Value: value,
	}

	deliveryChan := make(chan kafka.Event, 1)
	err := p.producer.Produce(msg, deliveryChan)
	if err != nil {
		logger.Error("Failed to produce message", zap.Error(err))
		return err
	}

	// Wait for delivery
	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		logger.Error("Failed to deliver message", zap.Error(m.TopicPartition.Error))
		return m.TopicPartition.Error
	}

	logger.Info("Message produced",
		zap.String("topic", topic),
		zap.String("key", key),
		zap.Int32("partition", m.TopicPartition.Partition),
		zap.Int64("offset", int64(m.TopicPartition.Offset)),
	)

	return nil
}

// Flush waits for all messages to be delivered
func (p *ProducerClient) Flush(timeoutMs int) int {
	remaining := p.producer.Flush(timeoutMs)
	if remaining > 0 {
		logger.Warn("Messages still in queue after flush", zap.Int("remaining", remaining))
	}
	return remaining
}

// Close closes the producer
func (p *ProducerClient) Close() {
	p.producer.Close()
	logger.Info("Kafka producer closed")
}

// ReadMessage reads a message from Kafka
func (c *ConsumerClient) ReadMessage(timeoutMs int) (topic string, key string, value []byte, err error) {
	msg, err := c.consumer.ReadMessage(time.Duration(timeoutMs) * time.Millisecond)
	if err != nil {
		return "", "", nil, err
	}

	return *msg.TopicPartition.Topic, string(msg.Key), msg.Value, nil
}

// Close closes the consumer
func (c *ConsumerClient) Close() {
	c.consumer.Close()
	logger.Info("Kafka consumer closed")
}
