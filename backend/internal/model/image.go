package model

import (
	"time"

	"github.com/google/uuid"
)

// Image represents an image processing job that will be sent to the queue.
type Image struct {
	ID         uuid.UUID  `json:"id"`
	OriginalID *uuid.UUID `json:"original_id"`
	Filename   string     `json:"filename"`
	Path       string     `json:"file_path"`
	Action     Action     `json:"actions"` // action to perform
	Status     string     `json:"status"`  // pending / processed / failed
	CreatedAt  time.Time  `json:"created_at"`
}

// Action defines a single action and its optional parameters.
type Action struct {
	Name   string            `json:"name"`   // "resize", "thumbnail", "watermark"
	Params map[string]string `json:"params"` // e.g., width/height, watermark text, etc.
}
