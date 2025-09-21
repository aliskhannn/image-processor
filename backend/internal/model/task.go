package model

import "github.com/google/uuid"

// Task represents an image processing job that will be sent to the queue.
type Task struct {
	ID       uuid.UUID `json:"id"`
	Filename string    `json:"filename"`
	Path     string    `json:"file_path"`
	Actions  string    `json:"actions"` // e.g., "resize", "thumbnail", "watermark"
}
