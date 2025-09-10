package media

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// DefaultThumbnailGenerator implements ThumbnailGenerator with file-based caching
type DefaultThumbnailGenerator struct {
	cacheDir  string
	processor ImageProcessor
	mu        sync.RWMutex
}

// NewDefaultThumbnailGenerator creates a new thumbnail generator
func NewDefaultThumbnailGenerator(cacheDir string, processor ImageProcessor) *DefaultThumbnailGenerator {
	return &DefaultThumbnailGenerator{
		cacheDir:  cacheDir,
		processor: processor,
	}
}

// GenerateThumbnail creates a thumbnail for the given file and returns the thumbnail path
func (g *DefaultThumbnailGenerator) GenerateThumbnail(filePath string, maxWidth, maxHeight int) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Check if cached thumbnail exists
	if thumbnailPath, exists := g.getCachedThumbnailPath(filePath, maxWidth, maxHeight); exists {
		return thumbnailPath, nil
	}

	// Determine media type
	detector := NewDefaultMediaDetector()
	mediaType, err := detector.DetectMediaType(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to detect media type: %w", err)
	}

	// Generate thumbnail based on media type
	switch mediaType {
	case MediaTypeImage:
		return g.generateImageThumbnail(filePath, maxWidth, maxHeight)
	case MediaTypeVideo:
		return g.GenerateVideoThumbnail(filePath, maxWidth, maxHeight)
	default:
		return "", fmt.Errorf("unsupported media type for thumbnail: %s", mediaType)
	}
}

// GenerateVideoThumbnail creates a thumbnail for video files
func (g *DefaultThumbnailGenerator) GenerateVideoThumbnail(filePath string, maxWidth, maxHeight int) (string, error) {
	// For now, return a placeholder for video thumbnails
	// In a full implementation, this would extract a frame from the video
	thumbnailPath := g.getThumbnailPath(filePath, maxWidth, maxHeight)

	// Create a simple placeholder image for video thumbnails
	return g.createVideoPlaceholder(thumbnailPath, maxWidth, maxHeight)
}

// GetCachedThumbnail returns the cached thumbnail path if it exists
func (g *DefaultThumbnailGenerator) GetCachedThumbnail(filePath string, maxWidth, maxHeight int) (string, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.getCachedThumbnailPath(filePath, maxWidth, maxHeight)
}

// ClearCache removes all cached thumbnails
func (g *DefaultThumbnailGenerator) ClearCache() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, err := os.Stat(g.cacheDir); os.IsNotExist(err) {
		return nil // Cache directory doesn't exist, nothing to clear
	}

	return os.RemoveAll(g.cacheDir)
}

// generateImageThumbnail creates a thumbnail for image files
func (g *DefaultThumbnailGenerator) generateImageThumbnail(filePath string, maxWidth, maxHeight int) (string, error) {
	thumbnailPath := g.getThumbnailPath(filePath, maxWidth, maxHeight)

	// Create thumbnail using image processor
	err := g.processor.CreateThumbnail(filePath, thumbnailPath, maxWidth, maxHeight)
	if err != nil {
		return "", fmt.Errorf("failed to create image thumbnail: %w", err)
	}

	return thumbnailPath, nil
}

// createVideoPlaceholder creates a placeholder thumbnail for video files
func (g *DefaultThumbnailGenerator) createVideoPlaceholder(thumbnailPath string, maxWidth, maxHeight int) (string, error) {
	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(thumbnailPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create thumbnail directory: %w", err)
	}

	// For now, create an empty file as placeholder
	// In a real implementation, this would create a video icon or extract a frame
	file, err := os.Create(thumbnailPath)
	if err != nil {
		return "", fmt.Errorf("failed to create video placeholder: %w", err)
	}
	defer file.Close()

	// Write minimal content to indicate this is a video placeholder
	_, err = file.WriteString("video_placeholder")
	if err != nil {
		return "", fmt.Errorf("failed to write video placeholder: %w", err)
	}

	return thumbnailPath, nil
}

// getThumbnailPath generates the path for a thumbnail file
func (g *DefaultThumbnailGenerator) getThumbnailPath(filePath string, maxWidth, maxHeight int) string {
	// Create a hash of the file path and dimensions for unique naming
	hash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s_%d_%d", filePath, maxWidth, maxHeight))))

	// Use JPEG extension for thumbnails
	fileName := fmt.Sprintf("%s.jpg", hash)

	return filepath.Join(g.cacheDir, fileName)
}

// getCachedThumbnailPath checks if a cached thumbnail exists and returns its path
func (g *DefaultThumbnailGenerator) getCachedThumbnailPath(filePath string, maxWidth, maxHeight int) (string, bool) {
	thumbnailPath := g.getThumbnailPath(filePath, maxWidth, maxHeight)

	// Check if thumbnail file exists
	if _, err := os.Stat(thumbnailPath); err == nil {
		return thumbnailPath, true
	}

	return "", false
}
