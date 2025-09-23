package consumer

import (
	"context"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	wbfkafka "github.com/wb-go/wbf/kafka"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"

	"github.com/aliskhannn/image-processor/internal/config"
)

// uploadedHandler defines the interface for handling uploaded image messages.
type uploadedHandler interface {
	Handle(ctx context.Context, msg kafka.Message) error
}

// Consumer represents a Kafka consumer along with its configuration
// and the handler that processes uploaded image messages.
type Consumer struct {
	Client          *wbfkafka.Consumer
	uploadedHandler uploadedHandler
	cfg             *config.Kafka
	strategy        retry.Strategy
}

// New creates a new Consumer.
// - cfg: Kafka configuration struct
// - s: retry strategy
// - uh: handler for processing uploaded image messages
func New(
	cfg *config.Kafka,
	s retry.Strategy,
	uh uploadedHandler,
) *Consumer {
	consumer := wbfkafka.NewConsumer(cfg.Brokers, cfg.Topic, cfg.GroupID)

	return &Consumer{
		Client:          consumer,
		uploadedHandler: uh,
		cfg:             cfg,
		strategy:        s,
	}
}

// Consume continuously fetches messages from Kafka, processes them using the handler,
// and commits offsets after successful processing. It stops gracefully on context cancellation.
func (c *Consumer) Consume(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	zlog.Logger.Info().
		Str("topic", c.cfg.Topic).
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
			msg, fetchErr = c.Client.Fetch(ctx)
			return fetchErr
		}, c.strategy)

		if err != nil {
			// Log error and retry after a short backoff.
			zlog.Logger.Err(err).Msg("failed to fetch message")
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// Process message using the uploadedHandler.
		if err := c.uploadedHandler.Handle(ctx, msg); err != nil {
			zlog.Logger.Err(err).
				Str("message", string(msg.Value)).
				Msg("failed to process image")
			continue
		}

		// Commit the message with retries.
		err = retry.Do(func() error {
			return c.Client.Commit(ctx, msg)
		}, c.strategy)
		if err != nil {
			zlog.Logger.Err(err).Msg("failed to commit message after retries")
		}

		zlog.Logger.Info().
			Int64("offset", msg.Offset).
			Str("message", string(msg.Value)).
			Msg("message handled successfully")
	}
}
