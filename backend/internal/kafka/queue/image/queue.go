package image

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	wbfkafka "github.com/wb-go/wbf/kafka"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"

	"github.com/aliskhannn/image-processor/internal/config"
	"github.com/aliskhannn/image-processor/internal/model"
)

// uploadedHandler defines the interface for handling uploaded image messages.
type uploadedHandler interface {
	Handle(ctx context.Context, msg kafka.Message) error
}

// Queue wraps Kafka producer and consumer for sending and receiving tasks.
// It also holds configuration and retry strategy for sending messages.
type Queue struct {
	Producer *wbfkafka.Producer
	Consumer Consumer
	strategy retry.Strategy
}

// Consumer represents a Kafka consumer along with its configuration
// and the handler that processes uploaded image messages.
type Consumer struct {
	client          *wbfkafka.Consumer
	uploadedHandler uploadedHandler
	cfg             *config.Kafka
}

// NewQueue creates a new Queue with Kafka producer and consumer.
// - brokers: list of Kafka brokers
// - topic: Kafka topic name
// - groupID: consumer group ID
// - cfg: Kafka configuration struct
// - s: retry strategy for producer
// - uh: handler for processing uploaded image messages
func NewQueue(
	brokers []string,
	topic, groupID string,
	cfg *config.Kafka,
	s retry.Strategy,
	uh uploadedHandler,
) *Queue {
	producer := wbfkafka.NewProducer(brokers, topic)
	consumer := wbfkafka.NewConsumer(brokers, topic, groupID)

	return &Queue{
		Producer: producer,
		Consumer: Consumer{
			client:          consumer,
			uploadedHandler: uh,
			cfg:             cfg,
		},
		strategy: s,
	}
}

// Enqueue serializes the Task to JSON and sends it to Kafka using the producer.
// The Task ID is used as the message key for partitioning and ordering.
func (q *Queue) Enqueue(ctx context.Context, img model.Image) error {
	data, err := json.Marshal(img)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %v", err)
	}

	key := []byte(img.ID.String())

	if err = q.Producer.SendWithRetry(ctx, q.strategy, key, data); err != nil {
		return fmt.Errorf("failed to send task: %v", err)
	}

	return nil
}

// Consume continuously fetches messages from Kafka, processes them using the handler,
// and commits offsets after successful processing. It stops gracefully on context cancellation.
func (q *Queue) Consume(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		// Close the consumer when exiting.
		if err := q.Consumer.client.Close(); err != nil {
			zlog.Logger.Err(err).Msg("failed to close consumer")
			return
		}
		zlog.Logger.Info().Msg("consumer closed")
	}()

	zlog.Logger.Info().
		Str("topic", q.Consumer.cfg.Topic).
		Msg("starting consumer")

	for {
		// Exit if context is canceled (graceful shutdown).
		if ctx.Err() != nil {
			zlog.Logger.Info().Msg("shutdown signal received, stopping consumer")
			return
		}

		// Fetch a message from Kafka with retries.
		var msg kafka.Message
		err := retry.Do(func() error {
			var fetchErr error
			msg, fetchErr = q.Consumer.client.Fetch(ctx)
			return fetchErr
		}, q.strategy)

		if err != nil {
			// Log error and retry after a short backoff.
			zlog.Logger.Err(err).Msg("failed to fetch message")
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// Process message using the uploadedHandler.
		if err := q.Consumer.uploadedHandler.Handle(ctx, msg); err != nil {
			zlog.Logger.Err(err).
				Str("message", string(msg.Value)).
				Msg("failed to process image")
			continue
		}

		// Commit the message with retries.
		err = retry.Do(func() error {
			return q.Consumer.client.Commit(ctx, msg)
		}, q.strategy)
		if err != nil {
			zlog.Logger.Err(err).Msg("failed to commit message after retries")
		}

		zlog.Logger.Info().
			Int64("offset", msg.Offset).
			Str("message", string(msg.Value)).
			Msg("message handled successfully")
	}
}
