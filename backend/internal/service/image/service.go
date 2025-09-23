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
	Load(ctx context.Context, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
}

// producer defines the interface for enqueueing tasks into a message broker (e.g., Kafka).
type producer interface {
	Enqueue(ctx context.Context, task model.Image) error
}

type imgProcessor interface {
	Process(ctx context.Context, img model.Image) (model.Image, error)
}

type repository interface {
	SaveImage(ctx context.Context, img model.Image) (uuid.UUID, error)
	GetImage(ctx context.Context, id uuid.UUID) (model.Image, error)
	DeleteImage(ctx context.Context, id uuid.UUID) error
}

// Service provides business logic for image operations.
// It saves uploaded images to storage and publishes processing tasks to a queue.
type Service struct {
	fileStorage  fileStorage
	producer     producer
	imgProcessor imgProcessor
	repository   repository
}

// NewService creates a new Service with the given storage and producer.
func NewService(
	fs fileStorage,
	p producer,
	imgP imgProcessor,
	r repository,
) *Service {
	return &Service{
		fileStorage:  fs,
		producer:     p,
		imgProcessor: imgP,
		repository:   r,
	}
}

// SaveImage saves the uploaded file to storage, records it in the database,
// and enqueues a background processing task for the specified action.
// Returns the generated image ID, the path to the saved file, or an error.
func (s *Service) SaveImage(ctx context.Context, subdir, filename string, file io.Reader, action model.Action) (uuid.UUID, string, error) {
	// Save the original file to storage.
	dst, err := s.fileStorage.Save(ctx, subdir, filename, file)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("save image: failed to save image in storage: %w", err)
	}

	img := model.Image{
		Filename: filename,
		Path:     dst,
		Action:   action,
		Status:   "pending",
	}

	id, err := s.repository.SaveImage(ctx, img)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("save image: failed to save image to db: %w", err)
	}

	img.ID = id

	// Enqueue the task for asynchronous processing.
	if err := s.producer.Enqueue(ctx, img); err != nil {
		return uuid.Nil, "", fmt.Errorf("save image: failed to enqueue task: %w", err)
	}

	return id, dst, nil
}

func (s *Service) GetImage(ctx context.Context, id uuid.UUID) (model.Image, io.ReadCloser, error) {
	img, err := s.repository.GetImage(ctx, id)
	if err != nil {
		return model.Image{}, nil, fmt.Errorf("get image: failed to get image: %w", err)
	}

	srcReader, err := s.fileStorage.Load(ctx, img.Path)
	if err != nil {
		return model.Image{}, nil, fmt.Errorf("get image: failed to load file: %w", err)
	}

	return img, srcReader, nil
}

func (s *Service) DeleteImage(ctx context.Context, id uuid.UUID) error {
	img, err := s.repository.GetImage(ctx, id)
	if err != nil {
		return fmt.Errorf("get image: failed to get image: %w", err)
	}

	err = s.repository.DeleteImage(ctx, id)
	if err != nil {
		return fmt.Errorf("delete image: failed to delete image from db: %w", err)
	}

	err = s.fileStorage.Delete(ctx, img.Path)
	if err != nil {
		return fmt.Errorf("delete image: failed to delete image from storage: %w", err)
	}

	return nil
}

func (s *Service) ProcessImage(ctx context.Context, img model.Image) (uuid.UUID, error) {
	img, err := s.imgProcessor.Process(ctx, img)
	if err != nil {
		return uuid.Nil, fmt.Errorf("process image: failed to process task: %w", err)
	}

	id, err := s.repository.SaveImage(ctx, img)
	if err != nil {
		return uuid.Nil, fmt.Errorf("process image: failed to save image to db: %w", err)
	}

	return id, nil
}
