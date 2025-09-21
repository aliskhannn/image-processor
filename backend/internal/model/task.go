package model

import "github.com/google/uuid"

// Task represents an image processing job that will be sent to the queue.
type Task struct {
	ID       uuid.UUID `json:"id"`
	Filename string    `json:"filename"`
	Path     string    `json:"file_path"`
	Actions  []Action  `json:"actions"` // list of actions to perform
}

// Action defines a single action and its optional parameters.
type Action struct {
	Name   string            `json:"name"`   // "resize", "thumbnail", "watermark"
	Params map[string]string `json:"params"` // e.g., width/height, watermark text, etc.
}
