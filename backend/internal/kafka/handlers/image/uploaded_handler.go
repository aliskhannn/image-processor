package image

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"

	"github.com/aliskhannn/image-processor/internal/model"
)

type service interface {
	ProcessTask(ctx context.Context, task model.Task) error
}

type UploadedHandler struct {
	service service
}

func NewUploadedHandler(s service) *UploadedHandler {
	return &UploadedHandler{service: s}
}

func (h *UploadedHandler) Handle(ctx context.Context, msg kafka.Message) error {
	var task model.Task
	if err := json.Unmarshal(msg.Value, &task); err != nil {
		return fmt.Errorf("unmarshal task: %w", err)
	}

	err := h.service.ProcessTask(ctx, task)
	if err != nil {
		return fmt.Errorf("process task: %w", err)
	}

	return nil
}
