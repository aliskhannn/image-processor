package image

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"

	"github.com/aliskhannn/image-processor/internal/api/respond"
	"github.com/aliskhannn/image-processor/internal/model"
	"github.com/aliskhannn/image-processor/internal/repository/image"
)

// service defines the interface for image-related operations.
type service interface {
	SaveImage(ctx context.Context, subdir, filename string, file io.Reader, action model.Action) (uuid.UUID, string, error)
	GetImage(ctx context.Context, id uuid.UUID) (model.Image, io.ReadCloser, error)
	DeleteImage(ctx context.Context, id uuid.UUID) error
}

// Handler provides HTTP handlers for image-related endpoints.
// It depends on a service interface to perform the business logic.
type Handler struct {
	service service
}

// NewHandler creates a new Handler with the given service.
func NewHandler(s service) *Handler {
	return &Handler{service: s}
}

// UploadRequest represents the action and its parameters sent by the client.
type UploadRequest struct {
	Action string            `json:"action"`
	Params map[string]string `json:"params"`
}

// Upload handles the HTTP request for uploading an image.
// It reads the multipart form, saves the uploaded file via the service,
// enqueues background processing tasks, and responds with the saved file info.
func (h *Handler) Upload(c *ginext.Context) {
	// Parse the multipart form with a 10MB max memory limit.
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		respond.Fail(c, http.StatusBadRequest, fmt.Errorf("parse multipart form failed: %v", err))
	}

	// Retrieve the uploaded file from the form.
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		zlog.Logger.Err(err).Msg("failed to upload the file")
		respond.Fail(c, http.StatusBadRequest, fmt.Errorf("failed to retrieve the file"))
		return
	}
	defer file.Close()

	zlog.Logger.Printf("uploaded file: %v", header.Filename)
	zlog.Logger.Printf("file size: %v", header.Size)
	zlog.Logger.Printf("MIME header: %v", header.Header)

	// Parse the "actions" JSON field from the form.
	actionsJSON := c.PostForm("actions")
	if actionsJSON == "" {
		zlog.Logger.Warn().Msg("no actions provided")
		respond.Fail(c, http.StatusBadRequest, fmt.Errorf("actions field is required"))
		return
	}

	var req UploadRequest
	if err := json.Unmarshal([]byte(actionsJSON), &req); err != nil {
		zlog.Logger.Err(err).Msg("failed to unmarshal the actions")
		respond.Fail(c, http.StatusBadRequest, fmt.Errorf("failed to unmarshal the actions"))
		return
	}

	// Convert the request to a model.Action.
	action := model.Action{
		Name:   req.Action,
		Params: req.Params,
	}

	// Save the uploaded image via the service.
	id, dst, err := h.service.SaveImage(c.Request.Context(), "original", header.Filename, file, action)
	if err != nil {
		zlog.Logger.Err(err).Msg("failed to save the image")
		respond.Fail(c, http.StatusInternalServerError, fmt.Errorf("failed to save the image: %v", err))
		return
	}

	zlog.Logger.Printf("saved file: %v", dst)

	// Respond with file info.
	respond.OK(c, map[string]interface{}{
		"id":       id,
		"filename": header.Filename,
		"path":     dst,
	})
}

// Get serves the actual image bytes for a given image ID.
func (h *Handler) Get(c *ginext.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		zlog.Logger.Warn().Msg("missing id")
		respond.Fail(c, http.StatusBadRequest, fmt.Errorf("missing id"))
		return
	}

	// Parse UUID from the path parameter.
	id, err := uuid.Parse(idStr)
	if err != nil {
		zlog.Logger.Err(err).Msg("failed to parse id")
		respond.Fail(c, http.StatusBadRequest, fmt.Errorf("invalid id: %v", err))
		return
	}

	// Retrieve the image from the service.
	_, reader, err := h.service.GetImage(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, image.ErrImageNotFound) {
			zlog.Logger.Warn().Msg("image not found")
			respond.Fail(c, http.StatusNotFound, fmt.Errorf("image not found"))
			return
		}

		zlog.Logger.Err(err).Msg("failed to get image")
		respond.Fail(c, http.StatusInternalServerError, fmt.Errorf("failed to get image: %v", err))
		return
	}
	defer reader.Close()

	// Disable browser caching to always fetch the latest image.
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	respond.JPEG(c, http.StatusOK, reader)
}

// GetMeta returns metadata about the image (filename, status, etc.) without serving the file itself..
func (h *Handler) GetMeta(c *ginext.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respond.Fail(c, http.StatusBadRequest, fmt.Errorf("invalid id"))
		return
	}

	img, _, err := h.service.GetImage(c.Request.Context(), id)
	if err != nil {
		respond.Fail(c, http.StatusNotFound, fmt.Errorf("image not found"))
		return
	}

	respond.OK(c, img)
}

// Delete removes an image by ID.
func (h *Handler) Delete(c *ginext.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		zlog.Logger.Warn().Msg("missing id")
		respond.Fail(c, http.StatusBadRequest, fmt.Errorf("missing id"))
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		zlog.Logger.Err(err).Msg("failed to parse id")
		respond.Fail(c, http.StatusBadRequest, fmt.Errorf("invalid id: %v", err))
		return
	}

	if err := h.service.DeleteImage(c.Request.Context(), id); err != nil {
		if errors.Is(err, image.ErrImageNotFound) {
			zlog.Logger.Warn().Msg("image not found")
			respond.Fail(c, http.StatusNotFound, fmt.Errorf("image not found"))
			return
		}

		zlog.Logger.Err(err).Msg("failed to delete the image")
		respond.Fail(c, http.StatusInternalServerError, fmt.Errorf("failed to delete image: %w", err))
		return
	}

	c.Status(http.StatusNoContent)
}
