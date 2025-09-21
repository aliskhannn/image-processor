package processor

import (
	"bytes"
	"context"
	"fmt"
	"image/color"
	"io"
	"strconv"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"

	"github.com/aliskhannn/image-processor/internal/model"
)

// fileStorage defines the interface for file storage.
// It allows saving and loading files from a backend (e.g., local FS, S3, MinIO).
type fileStorage interface {
	Save(ctx context.Context, subdir, filename string, src io.Reader) (string, error)
	Load(ctx context.Context, subdir, filename string) (io.ReadCloser, error)
}

// Processor is responsible for executing image processing tasks
// such as resize, thumbnail generation, and watermarking.
type Processor struct {
	fileStorage fileStorage
}

// New creates a new Processor with the given file storage backend.
func New(fs fileStorage) *Processor {
	return &Processor{fileStorage: fs}
}

// Process iterates over all actions defined in the Task and
// calls the appropriate processing method.
func (p *Processor) Process(ctx context.Context, task model.Task) error {
	for _, action := range task.Actions {
		switch action.Name {
		case "resize":
			if err := p.resize(ctx, task, action.Params); err != nil {
				return fmt.Errorf("resize failed: %w", err)
			}
		case "thumbnail":
			if err := p.thumbnail(ctx, task, action.Params); err != nil {
				return fmt.Errorf("thumbnail failed: %w", err)
			}
		case "watermark":
			if err := p.watermark(ctx, task, action.Params); err != nil {
				return fmt.Errorf("watermark failed: %w", err)
			}
		default:
			// Unknown action, ignore.
			continue
		}
	}

	return nil
}

// resize resizes the image to the specified width and height.
func (p *Processor) resize(ctx context.Context, task model.Task, params map[string]string) error {
	width, err := strconv.Atoi(params["width"])
	if err != nil {
		return fmt.Errorf("invalid width: %v", err)
	}
	height, err := strconv.Atoi(params["height"])
	if err != nil {
		return fmt.Errorf("invalid height: %v", err)
	}

	// Load the original image from storage.
	srcReader, err := p.fileStorage.Load(ctx, "original", task.Filename)
	if err != nil {
		return fmt.Errorf("failed to load original image: %w", err)
	}
	defer srcReader.Close()

	// Decode into an image object.
	img, err := imaging.Decode(srcReader)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Perform resizing.
	resized := imaging.Resize(img, width, height, imaging.Lanczos)

	// Encode resized image into buffer for storage.
	buf := bytes.NewBuffer(nil)
	if err := imaging.Encode(buf, resized, imaging.JPEG); err != nil {
		return fmt.Errorf("failed to encode resized image: %w", err)
	}

	// Save resized version.
	_, err = p.fileStorage.Save(ctx, "resized", task.Filename, buf)
	if err != nil {
		return fmt.Errorf("failed to save resized image: %w", err)
	}

	return nil
}

// thumbnail generates a small thumbnail of the image.
func (p *Processor) thumbnail(ctx context.Context, task model.Task, params map[string]string) error {
	width, err := strconv.Atoi(params["width"])
	if err != nil {
		return fmt.Errorf("invalid width: %v", err)
	}
	height, err := strconv.Atoi(params["height"])
	if err != nil {
		return fmt.Errorf("invalid height: %v", err)
	}

	// Load the original image.
	srcReader, err := p.fileStorage.Load(ctx, "original", task.Filename)
	if err != nil {
		return fmt.Errorf("failed to load original image: %w", err)
	}
	defer srcReader.Close()

	// Decode into an image object.
	img, err := imaging.Decode(srcReader)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Generate thumbnail.
	thumb := imaging.Thumbnail(img, width, height, imaging.Lanczos)

	// Encode resized image into buffer for storage.
	buf := bytes.NewBuffer(nil)
	if err := imaging.Encode(buf, thumb, imaging.JPEG); err != nil {
		return fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	// Save thumbnail.
	_, err = p.fileStorage.Save(ctx, "thumbnails", task.Filename, buf)
	if err != nil {
		return fmt.Errorf("failed to save thumbnail: %w", err)
	}

	return nil
}

// watermark adds a watermark text to the image.
// For simplicity, the watermark will be placed in the bottom-right corner.
func (p *Processor) watermark(ctx context.Context, task model.Task, params map[string]string) error {
	text := params["text"]
	if text == "" {
		text = "Watermark"
	}

	// Load the original image.
	srcReader, err := p.fileStorage.Load(ctx, "original", task.Filename)
	if err != nil {
		return fmt.Errorf("failed to load original image: %w", err)
	}
	defer srcReader.Close()

	// Decode into an image object.
	img, err := imaging.Decode(srcReader)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Draw watermark text on top of the image.
	dc := gg.NewContextForImage(img)
	dc.SetColor(color.White)

	err = dc.LoadFontFace("sans-serif", 6)
	if err != nil {
		return fmt.Errorf("failed to load font: %w", err)
	}

	margin := 10.0
	x := float64(dc.Width()) - margin
	y := float64(dc.Height()) - margin

	dc.DrawStringAnchored(text, x, y, 1, 1) // bottom-right corner
	dc.Fill()

	// Encode modified image.
	buf := new(bytes.Buffer)
	if err := imaging.Encode(buf, dc.Image(), imaging.JPEG); err != nil {
		return fmt.Errorf("failed to encode watermarked image: %w", err)
	}

	// Save watermarked version.
	_, err = p.fileStorage.Save(context.Background(), "watermarked", task.Filename, buf)
	if err != nil {
		return fmt.Errorf("failed to save watermarked image: %w", err)
	}

	return nil
}
