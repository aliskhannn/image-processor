package image

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"

	"github.com/aliskhannn/image-processor/internal/model"
)

// fileStorage defines the interface for storing files (e.g., local filesystem or S3).
type fileStorage interface {
	Save(ctx context.Context, subdir, filename string, src io.Reader) (string, error)
}

// producer defines the interface for enqueueing tasks into a message broker (e.g., Kafka).
type producer interface {
	Enqueue(ctx context.Context, task model.Task) error
}

type imgProcessor interface {
	Process(ctx context.Context, task model.Task) error
}

// Service provides business logic for image operations.
// It saves uploaded images to storage and publishes processing tasks to a queue.
type Service struct {
	fileStorage  fileStorage
	producer     producer
	imgProcessor imgProcessor
}

// NewService creates a new Service with the given storage and producer.
func NewService(fs fileStorage, p producer, imgP imgProcessor) *Service {
	return &Service{fileStorage: fs, producer: p, imgProcessor: imgP}
}

// SaveImage saves the uploaded file to storage and enqueues background processing tasks.
// Each action is converted into a separate Task and sent to the queue for asynchronous processing.
// Returns the path to the saved file or an error.
func (s *Service) SaveImage(ctx context.Context, subdir, filename string, file io.Reader, actions []model.Action) (string, error) {
	// Save the original file to storage.
	dst, err := s.fileStorage.Save(ctx, subdir, filename, file)
	if err != nil {
		return "", fmt.Errorf("upload: failed to save file: %w", err)
	}

	task := model.Task{
		ID:       uuid.New(),
		Filename: filename,
		Path:     dst,
		Actions:  actions,
	}

	// Enqueue the task for asynchronous processing.
	if err := s.producer.Enqueue(ctx, task); err != nil {
		return "", fmt.Errorf("upload: failed to enqueue task: %w", err)
	}

	return dst, nil
}

func (s *Service) ProcessTask(ctx context.Context, task model.Task) error {
	err := s.imgProcessor.Process(ctx, task)
	if err != nil {
		return fmt.Errorf("process: failed to process task: %w", err)
	}

	return nil
}
