package media

import (
	"image"
	"io"
)

// MediaType represents the type of media file
type MediaType int

const (
	MediaTypeUnknown MediaType = iota
	MediaTypeImage
	MediaTypeVideo
	MediaTypeAudio
	MediaTypeDocument
)

// String returns the string representation of MediaType
func (mt MediaType) String() string {
	switch mt {
	case MediaTypeImage:
		return "image"
	case MediaTypeVideo:
		return "video"
	case MediaTypeAudio:
		return "audio"
	case MediaTypeDocument:
		return "document"
	default:
		return "unknown"
	}
}

// MediaInfo contains information about a media file
type MediaInfo struct {
	Type          MediaType `json:"type"`
	Width         int       `json:"width,omitempty"`
	Height        int       `json:"height,omitempty"`
	Duration      int       `json:"duration,omitempty"` // in seconds for video/audio
	Size          int64     `json:"size"`               // file size in bytes
	MimeType      string    `json:"mime_type"`
	ThumbnailPath string    `json:"thumbnail_path,omitempty"`
}

// ThumbnailGenerator generates thumbnails for media files
type ThumbnailGenerator interface {
	// GenerateThumbnail creates a thumbnail for the given file and returns the thumbnail path
	GenerateThumbnail(filePath string, maxWidth, maxHeight int) (string, error)

	// GenerateVideoThumbnail creates a thumbnail for video files
	GenerateVideoThumbnail(filePath string, maxWidth, maxHeight int) (string, error)

	// GetCachedThumbnail returns the cached thumbnail path if it exists
	GetCachedThumbnail(filePath string, maxWidth, maxHeight int) (string, bool)

	// ClearCache removes all cached thumbnails
	ClearCache() error
}

// MediaDetector detects media file types and properties
type MediaDetector interface {
	// DetectMediaType determines the media type from file extension and content
	DetectMediaType(filePath string) (MediaType, error)

	// GetMediaInfo extracts comprehensive information about a media file
	GetMediaInfo(filePath string) (*MediaInfo, error)

	// IsSupported checks if the file type is supported for preview
	IsSupported(filePath string) bool
}

// ImageProcessor processes images for thumbnails and previews
type ImageProcessor interface {
	// ResizeImage resizes an image to the specified dimensions
	ResizeImage(src image.Image, width, height uint) image.Image

	// DecodeImage decodes an image from a reader
	DecodeImage(r io.Reader) (image.Image, string, error)

	// EncodeImage encodes an image to the specified format
	EncodeImage(w io.Writer, img image.Image, format string) error

	// CreateThumbnail creates a thumbnail from an image file
	CreateThumbnail(sourcePath, outputPath string, maxWidth, maxHeight int) error
}

// Manager coordinates all media-related operations
type Manager struct {
	thumbnailGen ThumbnailGenerator
	detector     MediaDetector
	processor    ImageProcessor
	cacheDir     string
}

// ManagerInterface defines the public interface for the media manager
type ManagerInterface interface {
	// GetMediaInfo returns information about a media file
	GetMediaInfo(filePath string) (*MediaInfo, error)

	// GenerateThumbnail creates a thumbnail for a media file
	GenerateThumbnail(filePath string, maxWidth, maxHeight int) (string, error)

	// IsMediaFile checks if the file is a supported media type
	IsMediaFile(filePath string) bool

	// GetThumbnailPath returns the thumbnail path for a file
	GetThumbnailPath(filePath string, maxWidth, maxHeight int) (string, bool)

	// Cleanup removes cached thumbnails
	Cleanup() error
}
