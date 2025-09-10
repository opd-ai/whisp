package media

import (
	"fmt"
	"path/filepath"
)

// NewManager creates a new media manager with all components
func NewManager(cacheDir string) *Manager {
	processor := NewDefaultImageProcessor()
	detector := NewDefaultMediaDetector()
	thumbnailGen := NewDefaultThumbnailGenerator(cacheDir, processor)

	return &Manager{
		thumbnailGen: thumbnailGen,
		detector:     detector,
		processor:    processor,
		cacheDir:     cacheDir,
	}
}

// GetMediaInfo returns information about a media file
func (m *Manager) GetMediaInfo(filePath string) (*MediaInfo, error) {
	return m.detector.GetMediaInfo(filePath)
}

// GenerateThumbnail creates a thumbnail for a media file
func (m *Manager) GenerateThumbnail(filePath string, maxWidth, maxHeight int) (string, error) {
	if !m.IsMediaFile(filePath) {
		return "", fmt.Errorf("file is not a supported media type: %s", filePath)
	}

	return m.thumbnailGen.GenerateThumbnail(filePath, maxWidth, maxHeight)
}

// IsMediaFile checks if the file is a supported media type
func (m *Manager) IsMediaFile(filePath string) bool {
	return m.detector.IsSupported(filePath)
}

// GetThumbnailPath returns the thumbnail path for a file
func (m *Manager) GetThumbnailPath(filePath string, maxWidth, maxHeight int) (string, bool) {
	return m.thumbnailGen.GetCachedThumbnail(filePath, maxWidth, maxHeight)
}

// Cleanup removes cached thumbnails
func (m *Manager) Cleanup() error {
	return m.thumbnailGen.ClearCache()
}

// GetCacheDir returns the cache directory path
func (m *Manager) GetCacheDir() string {
	return m.cacheDir
}

// SetCacheDir updates the cache directory (useful for configuration changes)
func (m *Manager) SetCacheDir(cacheDir string) {
	m.cacheDir = cacheDir
	// Update thumbnail generator cache dir if it supports it
	if gen, ok := m.thumbnailGen.(*DefaultThumbnailGenerator); ok {
		gen.cacheDir = cacheDir
	}
}

// GetSupportedImageFormats returns a list of supported image formats
func (m *Manager) GetSupportedImageFormats() []string {
	return []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".tif", ".webp"}
}

// GetSupportedVideoFormats returns a list of supported video formats
func (m *Manager) GetSupportedVideoFormats() []string {
	return []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm", ".mkv", ".m4v"}
}

// IsImageFile checks if the file is an image
func (m *Manager) IsImageFile(filePath string) bool {
	mediaType, err := m.detector.DetectMediaType(filePath)
	if err != nil {
		return false
	}
	return mediaType == MediaTypeImage
}

// IsVideoFile checks if the file is a video
func (m *Manager) IsVideoFile(filePath string) bool {
	mediaType, err := m.detector.DetectMediaType(filePath)
	if err != nil {
		return false
	}
	return mediaType == MediaTypeVideo
}

// GetMediaTypeString returns a human-readable media type string
func (m *Manager) GetMediaTypeString(filePath string) string {
	mediaType, err := m.detector.DetectMediaType(filePath)
	if err != nil {
		return "unknown"
	}
	return mediaType.String()
}

// ValidateFile checks if a file exists and is accessible
func (m *Manager) ValidateFile(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("file path is empty")
	}

	// Check if file exists and is readable
	info, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	_, err = filepath.EvalSymlinks(info)
	if err != nil {
		return fmt.Errorf("file not accessible: %w", err)
	}

	return nil
}
