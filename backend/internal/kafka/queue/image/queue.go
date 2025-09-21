package image

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wb-go/wbf/kafka"
	"github.com/wb-go/wbf/retry"

	"github.com/aliskhannn/image-processor/internal/config"
	"github.com/aliskhannn/image-processor/internal/model"
)

// Queue wraps Kafka producer and consumer for sending and receiving tasks.
// It also holds configuration and retry strategy for sending messages.
type Queue struct {
	Producer *kafka.Producer
	Consumer *kafka.Consumer
	cfg      *config.Config
	strategy retry.Strategy
}

// NewQueue creates a new Queue with the given Kafka brokers, topic, group ID,
// configuration, and retry strategy.
func NewQueue(brokers []string, topic, groupID string, cfg *config.Config, s retry.Strategy) *Queue {
	producer := kafka.NewProducer(brokers, topic)
	consumer := kafka.NewConsumer(brokers, topic, groupID)

	return &Queue{
		Producer: producer,
		Consumer: consumer,
		cfg:      cfg,
		strategy: s,
	}
}

// Enqueue serializes the Task to JSON and sends it to Kafka using the producer.
// The Task ID is used as the message key for partitioning and ordering.
func (q *Queue) Enqueue(ctx context.Context, task model.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %v", err)
	}

	key := []byte(task.ID.String())

	if err = q.Producer.SendWithRetry(ctx, q.strategy, key, data); err != nil {
		return fmt.Errorf("failed to send task: %v", err)
	}

	return nil
}
