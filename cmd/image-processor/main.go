package main

import (
	"context"
	"errors"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"

	"github.com/aliskhannn/image-processor/internal/api/handlers/image"
	"github.com/aliskhannn/image-processor/internal/api/router"
	"github.com/aliskhannn/image-processor/internal/api/server"
	"github.com/aliskhannn/image-processor/internal/config"
	"github.com/aliskhannn/image-processor/internal/infra/kafka/consumer"
	"github.com/aliskhannn/image-processor/internal/infra/kafka/producer"
	imagemsg "github.com/aliskhannn/image-processor/internal/kafka/handlers/image"
	"github.com/aliskhannn/image-processor/internal/processor"
	imagerepo "github.com/aliskhannn/image-processor/internal/repository/image"
	imagesvc "github.com/aliskhannn/image-processor/internal/service/image"
	"github.com/aliskhannn/image-processor/internal/storage/file"
)

func main() {
	// Context & signals: used for graceful shutdown on system interrupts.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize logger and load application configuration.
	zlog.Init()
	cfg := config.MustLoad("./config/config.yml")

	// Connect to PostgreSQL (master and slaves).
	opts := &dbpg.Options{
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	}

	// Collect slave DSNs for replica connections.
	slaveDNSs := make([]string, 0, len(cfg.Database.Slaves))

	for _, s := range cfg.Database.Slaves {
		slaveDNSs = append(slaveDNSs, s.DSN())
	}
	zlog.Logger.Info().Msgf("db url: %s", cfg.Database.Master.DSN())
	db, err := dbpg.New(cfg.Database.Master.DSN(), slaveDNSs, opts)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to connect to database")
	}

	// Retry strategy for Kafka and other external calls.
	strategy := retry.Strategy{
		Attempts: cfg.Retry.Attempts,
		Delay:    cfg.Retry.Delay,
		Backoff:  cfg.Retry.Backoff,
	}

	// Initialize file storage (MinIO).
	storage, err := file.NewStorage(ctx, cfg.Storage.Endpoint, cfg.Storage.AccessKey, cfg.Storage.SecretKey, cfg.Storage.BucketName, cfg.Storage.UseSSL)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to connect to storage")
	}

	// Initialize repository, producer, processor, and service layer.
	repo := imagerepo.NewRepository(db)
	p := producer.New(&cfg.Kafka, strategy)
	imageProcessor := processor.New(storage)
	service := imagesvc.NewService(storage, p, imageProcessor, repo)

	// Kafka message handler for uploaded images.
	uploadedHandler := imagemsg.NewUploadedHandler(service)

	// HTTP handler for image routes.
	imgHandler := image.NewHandler(service)

	// Kafka consumer for processing uploaded image events.
	c := consumer.New(&cfg.Kafka, strategy, uploadedHandler)

	// Start Kafka consumer in a separate goroutine.
	var wg sync.WaitGroup
	wg.Add(1)
	go c.Consume(ctx, &wg)

	// Start HTTP server in a separate goroutine.
	r := router.Setup(imgHandler)
	s := server.New(cfg.Server.HTTPPort, r)
	go func() {
		if err := s.ListenAndServe(); err != nil {
			zlog.Logger.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	// Block until context is canceled (SIGINT/SIGTERM).
	<-ctx.Done()
	zlog.Logger.Info().Msg("context done")

	// Wait for Kafka consumer goroutine to finish.
	wg.Wait()

	// Graceful shutdown with timeout for HTTP server.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	zlog.Logger.Info().Msg("shutting down server")
	if err := s.Shutdown(shutdownCtx); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to shutdown server")
	}
	if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
		zlog.Logger.Info().Msg("timeout exceeded, forcing shutdown")
	}

	// Close master and slave databases.
	if err := db.Master.Close(); err != nil {
		zlog.Logger.Printf("failed to close master DB: %v", err)
	}
	for i, s := range db.Slaves {
		if err := s.Close(); err != nil {
			zlog.Logger.Printf("failed to close slave DB %d: %v", i, err)
		}
	}

	// Close Kafka producer and consumer clients.
	if err = p.Client.Close(); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to close kafka producer client")
	}
	if err = c.Client.Close(); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to close kafka consumer client")
	}
}
