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

const defaultFontPath = "internal/assets/fonts/DejaVuSans.ttf"

// fileStorage defines the interface for file storage.
// It allows saving and loading files from a backend (e.g., local FS, S3, MinIO).
type fileStorage interface {
	Save(ctx context.Context, subdir, filename string, src io.Reader) (string, error)
	Load(ctx context.Context, path string) (io.ReadCloser, error)
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
func (p *Processor) Process(ctx context.Context, img model.Image) (model.Image, error) {
	switch img.Action.Name {
	case "resize":
		return p.resize(ctx, img)
	case "thumbnail":
		return p.thumbnail(ctx, img)
	case "watermark":
		return p.watermark(ctx, img)
	default:
		return model.Image{}, fmt.Errorf("unknown task action: %s", img.Action.Name)
	}
}

// resize resizes the image to the specified width and height.
func (p *Processor) resize(ctx context.Context, img model.Image) (model.Image, error) {
	params := img.Action.Params

	width, err := strconv.Atoi(params["width"])
	if err != nil {
		return model.Image{}, fmt.Errorf("invalid width: %v", err)
	}
	height, err := strconv.Atoi(params["height"])
	if err != nil {
		return model.Image{}, fmt.Errorf("invalid height: %v", err)
	}

	// Load the original image from storage.
	srcReader, err := p.fileStorage.Load(ctx, img.Path)
	if err != nil {
		return model.Image{}, fmt.Errorf("failed to load original image: %w", err)
	}
	defer srcReader.Close()

	// Decode into an image object.
	image, err := imaging.Decode(srcReader)
	if err != nil {
		return model.Image{}, fmt.Errorf("failed to decode image: %w", err)
	}

	// Perform resizing.
	resized := imaging.Resize(image, width, height, imaging.Lanczos)

	// Encode resized image into buffer for storage.
	buf := bytes.NewBuffer(nil)
	if err := imaging.Encode(buf, resized, imaging.JPEG); err != nil {
		return model.Image{}, fmt.Errorf("failed to encode resized image: %w", err)
	}

	// Save resized version.
	dst, err := p.fileStorage.Save(ctx, "resized", img.Filename, buf)
	if err != nil {
		return model.Image{}, fmt.Errorf("failed to save resized image: %w", err)
	}

	img.Path = dst
	img.Status = "processed"

	return img, nil
}

// thumbnail generates a small thumbnail of the image.
func (p *Processor) thumbnail(ctx context.Context, img model.Image) (model.Image, error) {
	params := img.Action.Params

	width, err := strconv.Atoi(params["width"])
	if err != nil {
		return model.Image{}, fmt.Errorf("invalid width: %v", err)
	}
	height, err := strconv.Atoi(params["height"])
	if err != nil {
		return model.Image{}, fmt.Errorf("invalid height: %v", err)
	}

	// Load the original image.
	srcReader, err := p.fileStorage.Load(ctx, img.Path)
	if err != nil {
		return model.Image{}, fmt.Errorf("failed to load original image: %w", err)
	}
	defer srcReader.Close()

	// Decode into an image object.
	image, err := imaging.Decode(srcReader)
	if err != nil {
		return model.Image{}, fmt.Errorf("failed to decode image: %w", err)
	}

	// Generate thumbnail.
	thumb := imaging.Thumbnail(image, width, height, imaging.Lanczos)

	// Encode resized image into buffer for storage.
	buf := bytes.NewBuffer(nil)
	if err := imaging.Encode(buf, thumb, imaging.JPEG); err != nil {
		return model.Image{}, fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	// Save thumbnail.
	dst, err := p.fileStorage.Save(ctx, "thumbnails", img.Filename, buf)
	if err != nil {
		return model.Image{}, fmt.Errorf("failed to save thumbnail: %w", err)
	}

	img.Path = dst
	img.Status = "processed"

	return img, nil
}

// watermark adds a watermark text to the image.
// For simplicity, the watermark will be placed in the bottom-right corner.
func (p *Processor) watermark(ctx context.Context, img model.Image) (model.Image, error) {
	params := img.Action.Params

	text := params["text"]
	if text == "" {
		text = "Watermark"
	}

	// Load the original image.
	srcReader, err := p.fileStorage.Load(ctx, img.Path)
	if err != nil {
		return model.Image{}, fmt.Errorf("failed to load original image: %w", err)
	}
	defer srcReader.Close()

	// Decode into an image object.
	image, err := imaging.Decode(srcReader)
	if err != nil {
		return model.Image{}, fmt.Errorf("failed to decode image: %w", err)
	}

	// Draw watermark text on top of the image.
	dc := gg.NewContextForImage(image)
	dc.SetColor(color.White)

	fontSize := float64(dc.Width()) * 0.05 // 5% of the image width

	err = dc.LoadFontFace(defaultFontPath, fontSize)
	if err != nil {
		return model.Image{}, fmt.Errorf("failed to load font: %w", err)
	}

	tw, th := dc.MeasureString(text) // calculate font size

	margin := 10.0
	x := float64(dc.Width()) - tw - margin
	y := float64(dc.Height()) - th - margin

	dc.DrawStringAnchored(text, x, y, 1, 1) // bottom-right corner
	dc.Fill()

	// Encode modified image.
	buf := new(bytes.Buffer)
	if err := imaging.Encode(buf, dc.Image(), imaging.JPEG); err != nil {
		return model.Image{}, fmt.Errorf("failed to encode watermarked image: %w", err)
	}

	// Save watermarked version.
	dst, err := p.fileStorage.Save(ctx, "watermarked", img.Filename, buf)
	if err != nil {
		return model.Image{}, fmt.Errorf("failed to save watermarked image: %w", err)
	}

	img.Path = dst
	img.Status = "processed"

	return img, nil
}
