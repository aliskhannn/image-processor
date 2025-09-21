package image

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"

	"github.com/aliskhannn/image-processor/internal/model"
)

type service interface {
	ProcessImage(ctx context.Context, img model.Image) (uuid.UUID, error)
}

type UploadedHandler struct {
	service service
}

func NewUploadedHandler(s service) *UploadedHandler {
	return &UploadedHandler{service: s}
}

func (h *UploadedHandler) Handle(ctx context.Context, msg kafka.Message) error {
	var img model.Image
	if err := json.Unmarshal(msg.Value, &img); err != nil {
		return fmt.Errorf("unmarshal task: %w", err)
	}

	_, err := h.service.ProcessImage(ctx, img)
	if err != nil {
		return fmt.Errorf("process task: %w", err)
	}

	return nil
}
