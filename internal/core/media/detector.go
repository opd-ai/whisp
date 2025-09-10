package media

import (
	"bufio"
	"fmt"
	"image"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	// Import image format decoders
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	// Import additional image formats
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// DefaultMediaDetector implements MediaDetector using file analysis
type DefaultMediaDetector struct{}

// NewDefaultMediaDetector creates a new media detector
func NewDefaultMediaDetector() *DefaultMediaDetector {
	return &DefaultMediaDetector{}
}

// DetectMediaType determines the media type from file extension and content
func (d *DefaultMediaDetector) DetectMediaType(filePath string) (MediaType, error) {
	if filePath == "" {
		return MediaTypeUnknown, fmt.Errorf("file path is empty")
	}

	// First try by file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	if mediaType := d.getMediaTypeFromExtension(ext); mediaType != MediaTypeUnknown {
		return mediaType, nil
	}

	// For unknown extensions, check if file exists before trying to read content
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// If file doesn't exist and extension is unknown, assume document type
		return MediaTypeDocument, nil
	}

	// Then try by MIME type detection for existing files
	return d.detectByContent(filePath)
}

// GetMediaInfo extracts comprehensive information about a media file
func (d *DefaultMediaDetector) GetMediaInfo(filePath string) (*MediaInfo, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path is empty")
	}

	// Get file stats
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Detect media type
	mediaType, err := d.DetectMediaType(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect media type: %w", err)
	}

	mediaInfo := &MediaInfo{
		Type:     mediaType,
		Size:     info.Size(),
		MimeType: d.getMimeType(filePath),
	}

	// Get additional info for images
	if mediaType == MediaTypeImage {
		width, height, err := d.getImageDimensions(filePath)
		if err == nil {
			mediaInfo.Width = width
			mediaInfo.Height = height
		}
	}

	return mediaInfo, nil
}

// IsSupported checks if the file type is supported for preview
func (d *DefaultMediaDetector) IsSupported(filePath string) bool {
	mediaType, err := d.DetectMediaType(filePath)
	if err != nil {
		return false
	}

	// Currently support images and basic video thumbnails
	return mediaType == MediaTypeImage || mediaType == MediaTypeVideo
}

// getMediaTypeFromExtension returns media type based on file extension
func (d *DefaultMediaDetector) getMediaTypeFromExtension(ext string) MediaType {
	imageExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".bmp": true, ".tiff": true, ".tif": true, ".webp": true,
	}

	videoExts := map[string]bool{
		".mp4": true, ".avi": true, ".mov": true, ".wmv": true,
		".flv": true, ".webm": true, ".mkv": true, ".m4v": true,
	}

	audioExts := map[string]bool{
		".mp3": true, ".wav": true, ".flac": true, ".aac": true,
		".ogg": true, ".wma": true, ".m4a": true,
	}

	if imageExts[ext] {
		return MediaTypeImage
	}
	if videoExts[ext] {
		return MediaTypeVideo
	}
	if audioExts[ext] {
		return MediaTypeAudio
	}

	return MediaTypeUnknown
}

// detectByContent detects media type by analyzing file content
func (d *DefaultMediaDetector) detectByContent(filePath string) (MediaType, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return MediaTypeUnknown, err
	}
	defer file.Close()

	// Read first 512 bytes for content type detection
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && n == 0 {
		return MediaTypeUnknown, err
	}

	contentType := http.DetectContentType(buffer[:n])

	if strings.HasPrefix(contentType, "image/") {
		return MediaTypeImage, nil
	}
	if strings.HasPrefix(contentType, "video/") {
		return MediaTypeVideo, nil
	}
	if strings.HasPrefix(contentType, "audio/") {
		return MediaTypeAudio, nil
	}

	return MediaTypeDocument, nil
}

// getMimeType returns the MIME type for a file
func (d *DefaultMediaDetector) getMimeType(filePath string) string {
	ext := filepath.Ext(filePath)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		return "application/octet-stream"
	}
	return mimeType
}

// getImageDimensions returns the dimensions of an image file
func (d *DefaultMediaDetector) getImageDimensions(filePath string) (int, int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	// Decode just the config to get dimensions without loading full image
	config, _, err := image.DecodeConfig(bufio.NewReader(file))
	if err != nil {
		return 0, 0, err
	}

	return config.Width, config.Height, nil
}
