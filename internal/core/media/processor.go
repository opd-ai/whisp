package media

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
)

// DefaultImageProcessor implements ImageProcessor using standard libraries
type DefaultImageProcessor struct{}

// NewDefaultImageProcessor creates a new image processor
func NewDefaultImageProcessor() *DefaultImageProcessor {
	return &DefaultImageProcessor{}
}

// ResizeImage resizes an image to the specified dimensions while maintaining aspect ratio
func (p *DefaultImageProcessor) ResizeImage(src image.Image, width, height uint) image.Image {
	// Use the resize library to maintain aspect ratio
	return resize.Resize(width, height, src, resize.Lanczos3)
}

// DecodeImage decodes an image from a reader and returns the image and format
func (p *DefaultImageProcessor) DecodeImage(r io.Reader) (image.Image, string, error) {
	return image.Decode(r)
}

// EncodeImage encodes an image to the specified format
func (p *DefaultImageProcessor) EncodeImage(w io.Writer, img image.Image, format string) error {
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		return jpeg.Encode(w, img, &jpeg.Options{Quality: 85})
	case "png":
		return png.Encode(w, img)
	default:
		return fmt.Errorf("unsupported image format: %s", format)
	}
}

// CreateThumbnail creates a thumbnail from an image file
func (p *DefaultImageProcessor) CreateThumbnail(sourcePath, outputPath string, maxWidth, maxHeight int) error {
	// Open source image
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source image: %w", err)
	}
	defer sourceFile.Close()

	// Decode image
	img, format, err := p.DecodeImage(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Calculate thumbnail size maintaining aspect ratio
	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	thumbWidth, thumbHeight := calculateThumbnailSize(origWidth, origHeight, maxWidth, maxHeight)

	// Resize image
	thumbnail := p.ResizeImage(img, uint(thumbWidth), uint(thumbHeight))

	// Create output directory if needed
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Determine output format (prefer JPEG for thumbnails to save space)
	outputFormat := "jpeg"
	if format == "png" {
		// Keep PNG for transparency
		outputFormat = "png"
	}

	// Encode thumbnail
	if err := p.EncodeImage(outputFile, thumbnail, outputFormat); err != nil {
		return fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return nil
}

// calculateThumbnailSize calculates the optimal thumbnail dimensions
func calculateThumbnailSize(origWidth, origHeight, maxWidth, maxHeight int) (int, int) {
	if origWidth <= maxWidth && origHeight <= maxHeight {
		return origWidth, origHeight
	}

	// Calculate scaling factor to fit within bounds
	scaleX := float64(maxWidth) / float64(origWidth)
	scaleY := float64(maxHeight) / float64(origHeight)

	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	newWidth := int(float64(origWidth) * scale)
	newHeight := int(float64(origHeight) * scale)

	// Ensure minimum size
	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}

	return newWidth, newHeight
}
