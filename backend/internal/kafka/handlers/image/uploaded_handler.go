package image

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/wb-go/wbf/zlog"

	"github.com/aliskhannn/image-processor/internal/model"
	"github.com/aliskhannn/image-processor/internal/repository/image"
)

// service defines the interface for processing uploaded images.
type service interface {
	ProcessImage(ctx context.Context, img model.Image) (uuid.UUID, error)
}

// UploadedHandler handles Kafka messages for newly uploaded images.
// It relies on a service that implements image processing logic.
type UploadedHandler struct {
	service service
}

// NewUploadedHandler creates a new handler with the given service.
func NewUploadedHandler(s service) *UploadedHandler {
	return &UploadedHandler{service: s}
}

// Handle processes a Kafka message containing an uploaded image.
// It unmarshals the message, calls the service to process the image,
// and logs the result.
func (h *UploadedHandler) Handle(ctx context.Context, msg kafka.Message) error {
	var img model.Image
	if err := json.Unmarshal(msg.Value, &img); err != nil {
		return fmt.Errorf("unmarshal task: %w", err)
	}

	id, err := h.service.ProcessImage(ctx, img)
	if err != nil {
		if errors.Is(err, image.ErrImageNotFound) {
			return fmt.Errorf("process task: %w", image.ErrImageNotFound)
		}

		return fmt.Errorf("process task: %w", err)
	}

	zlog.Logger.Printf("image processed: %s", id)

	return nil
}
