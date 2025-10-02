package producer

import (
	"context"
	"encoding/json"
	"fmt"

	wbfkafka "github.com/wb-go/wbf/kafka"
	"github.com/wb-go/wbf/retry"

	"github.com/aliskhannn/image-processor/internal/config"
	"github.com/aliskhannn/image-processor/internal/model"
)

// Producer represents a Kafka producer.
type Producer struct {
	Client   *wbfkafka.Producer
	strategy retry.Strategy
	cfg      *config.Kafka
}

// New creates a new Producer.
// - cfg: Kafka configuration struct
// - s: retry strategy
// - uh: handler for processing uploaded image messages
func New(
	cfg *config.Kafka,
	s retry.Strategy,
) *Producer {
	producer := wbfkafka.NewProducer(cfg.Brokers, cfg.Topic)

	return &Producer{
		Client:   producer,
		cfg:      cfg,
		strategy: s,
	}
}

// Produce serializes the Task to JSON and sends it to Kafka using the producer.
// The Task ID is used as the message key for partitioning and ordering.
func (p *Producer) Produce(ctx context.Context, img model.Image) error {
	data, err := json.Marshal(img)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %v", err)
	}

	key := []byte(img.ID.String())

	if err = p.Client.SendWithRetry(ctx, p.strategy, key, data); err != nil {
		return fmt.Errorf("failed to send task: %v", err)
	}

	return nil
}
