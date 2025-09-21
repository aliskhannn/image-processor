package image

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"

	"github.com/aliskhannn/image-processor/internal/api/respond"
	"github.com/aliskhannn/image-processor/internal/model"
)

// service defines the interface for image-related operations.
type service interface {
	SaveImage(ctx context.Context, subdir, filename string, file io.Reader, actions []model.Action) (string, error)
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

// UploadFile handles the HTTP request for uploading an image.
// It reads the multipart form, saves the uploaded file via the service,
// enqueues background processing tasks, and responds with the saved file info.
func (h *Handler) UploadFile(c *ginext.Context) {
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		respond.Fail(c, http.StatusBadRequest, fmt.Errorf("parse multipart form failed: %v", err))
	}

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

	// Parse actions JSON from a form field "actions"
	// Example: [{"name":"resize","params":{"width":"800","height":"600"}},{"name":"watermark","params":{"text":"My watermark"}}]
	var actions []model.Action
	if err := c.ShouldBindJSON(&actions); err != nil {
		zlog.Logger.Err(err).Msg("failed to bind the actions")
		respond.Fail(c, http.StatusBadRequest, fmt.Errorf("failed to bind the actions"))
		return
	}

	dst, err := h.service.SaveImage(c.Request.Context(), "original", header.Filename, file, actions)
	if err != nil {
		zlog.Logger.Err(err).Msg("failed to save the image")
		respond.Fail(c, http.StatusInternalServerError, fmt.Errorf("failed to save the image: %v", err))
		return
	}

	zlog.Logger.Printf("saved file: %v", dst)

	// Respond with file info.
	respond.OK(c, map[string]string{
		"filename": header.Filename,
		"path":     dst,
	})
}
