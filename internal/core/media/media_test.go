package media

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// TestMediaTypes tests the MediaType enum and String method
func TestMediaTypes(t *testing.T) {
	tests := []struct {
		name      string
		mediaType MediaType
		expected  string
	}{
		{"Unknown type", MediaTypeUnknown, "unknown"},
		{"Image type", MediaTypeImage, "image"},
		{"Video type", MediaTypeVideo, "video"},
		{"Audio type", MediaTypeAudio, "audio"},
		{"Document type", MediaTypeDocument, "document"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mediaType.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestDefaultMediaDetector tests the media detection functionality
func TestDefaultMediaDetector(t *testing.T) {
	detector := NewDefaultMediaDetector()

	t.Run("Detect by extension", func(t *testing.T) {
		tests := []struct {
			filename string
			expected MediaType
		}{
			{"test.jpg", MediaTypeImage},
			{"test.png", MediaTypeImage},
			{"test.gif", MediaTypeImage},
			{"test.mp4", MediaTypeVideo},
			{"test.avi", MediaTypeVideo},
			{"test.mp3", MediaTypeAudio},
			{"test.wav", MediaTypeAudio},
			{"test.txt", MediaTypeDocument}, // Unknown extensions for non-existent files return document type
		}

		for _, tt := range tests {
			mediaType, err := detector.DetectMediaType(tt.filename)
			if err != nil {
				t.Errorf("Error detecting media type for %s: %v", tt.filename, err)
				continue
			}
			if mediaType != tt.expected {
				t.Errorf("Expected %v for %s, got %v", tt.expected, tt.filename, mediaType)
			}
		}
	})

	t.Run("Empty file path", func(t *testing.T) {
		_, err := detector.DetectMediaType("")
		if err == nil {
			t.Error("Expected error for empty file path")
		}
	})

	t.Run("IsSupported", func(t *testing.T) {
		// Create a temporary image file for testing
		tempDir := t.TempDir()
		testImagePath := filepath.Join(tempDir, "test.png")

		// Create a simple test image
		img := image.NewRGBA(image.Rect(0, 0, 10, 10))
		for y := 0; y < 10; y++ {
			for x := 0; x < 10; x++ {
				img.Set(x, y, color.RGBA{255, 0, 0, 255})
			}
		}

		file, err := os.Create(testImagePath)
		if err != nil {
			t.Fatalf("Failed to create test image: %v", err)
		}
		defer file.Close()

		err = png.Encode(file, img)
		if err != nil {
			t.Fatalf("Failed to encode test image: %v", err)
		}

		if !detector.IsSupported(testImagePath) {
			t.Error("Expected test image to be supported")
		}

		// Test unsupported file
		unsupportedPath := filepath.Join(tempDir, "test.txt")
		err = os.WriteFile(unsupportedPath, []byte("test content"), 0o644)
		if err != nil {
			t.Fatalf("Failed to create test text file: %v", err)
		}

		if detector.IsSupported(unsupportedPath) {
			t.Error("Expected text file to be unsupported")
		}
	})
}

// TestDefaultImageProcessor tests image processing functionality
func TestDefaultImageProcessor(t *testing.T) {
	processor := NewDefaultImageProcessor()

	t.Run("ResizeImage", func(t *testing.T) {
		// Create a test image
		img := image.NewRGBA(image.Rect(0, 0, 100, 100))

		// Resize to smaller dimensions
		resized := processor.ResizeImage(img, 50, 50)
		bounds := resized.Bounds()

		if bounds.Dx() != 50 || bounds.Dy() != 50 {
			t.Errorf("Expected resized image to be 50x50, got %dx%d", bounds.Dx(), bounds.Dy())
		}
	})

	t.Run("CreateThumbnail", func(t *testing.T) {
		// Create a temporary directory and test image
		tempDir := t.TempDir()
		sourcePath := filepath.Join(tempDir, "source.png")
		thumbnailPath := filepath.Join(tempDir, "thumbnail.jpg")

		// Create a test image
		img := image.NewRGBA(image.Rect(0, 0, 200, 200))
		for y := 0; y < 200; y++ {
			for x := 0; x < 200; x++ {
				img.Set(x, y, color.RGBA{uint8(x), uint8(y), 100, 255})
			}
		}

		// Save the test image
		file, err := os.Create(sourcePath)
		if err != nil {
			t.Fatalf("Failed to create source image: %v", err)
		}
		defer file.Close()

		err = png.Encode(file, img)
		if err != nil {
			t.Fatalf("Failed to encode source image: %v", err)
		}

		// Create thumbnail
		err = processor.CreateThumbnail(sourcePath, thumbnailPath, 100, 100)
		if err != nil {
			t.Fatalf("Failed to create thumbnail: %v", err)
		}

		// Verify thumbnail exists
		if _, err := os.Stat(thumbnailPath); os.IsNotExist(err) {
			t.Error("Thumbnail file was not created")
		}
	})
}

// TestDefaultThumbnailGenerator tests thumbnail generation and caching
func TestDefaultThumbnailGenerator(t *testing.T) {
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")
	processor := NewDefaultImageProcessor()
	generator := NewDefaultThumbnailGenerator(cacheDir, processor)

	t.Run("GenerateThumbnail", func(t *testing.T) {
		// Create a test image
		sourcePath := filepath.Join(tempDir, "test.png")
		img := image.NewRGBA(image.Rect(0, 0, 100, 100))

		file, err := os.Create(sourcePath)
		if err != nil {
			t.Fatalf("Failed to create test image: %v", err)
		}
		defer file.Close()

		err = png.Encode(file, img)
		if err != nil {
			t.Fatalf("Failed to encode test image: %v", err)
		}

		// Generate thumbnail
		thumbnailPath, err := generator.GenerateThumbnail(sourcePath, 50, 50)
		if err != nil {
			t.Fatalf("Failed to generate thumbnail: %v", err)
		}

		// Verify thumbnail exists
		if _, err := os.Stat(thumbnailPath); os.IsNotExist(err) {
			t.Error("Thumbnail file was not created")
		}

		// Test caching - second call should return cached version
		cachedPath, err := generator.GenerateThumbnail(sourcePath, 50, 50)
		if err != nil {
			t.Fatalf("Failed to get cached thumbnail: %v", err)
		}

		if cachedPath != thumbnailPath {
			t.Error("Expected cached thumbnail path to match original")
		}
	})

	t.Run("GetCachedThumbnail", func(t *testing.T) {
		// Test with non-existent file
		_, exists := generator.GetCachedThumbnail("nonexistent.jpg", 50, 50)
		if exists {
			t.Error("Expected no cached thumbnail for non-existent file")
		}
	})

	t.Run("ClearCache", func(t *testing.T) {
		err := generator.ClearCache()
		if err != nil {
			t.Errorf("Failed to clear cache: %v", err)
		}

		// Verify cache directory is removed
		if _, err := os.Stat(cacheDir); !os.IsNotExist(err) {
			t.Error("Expected cache directory to be removed")
		}
	})
}

// TestMediaManager tests the main manager functionality
func TestMediaManager(t *testing.T) {
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "media_cache")
	manager := NewManager(cacheDir)

	t.Run("Manager creation", func(t *testing.T) {
		if manager == nil {
			t.Fatal("Expected manager to be created")
		}

		if manager.GetCacheDir() != cacheDir {
			t.Errorf("Expected cache dir %s, got %s", cacheDir, manager.GetCacheDir())
		}
	})

	t.Run("IsMediaFile", func(t *testing.T) {
		// Create a test image file
		testImagePath := filepath.Join(tempDir, "test.jpg")
		file, err := os.Create(testImagePath)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		file.Close()

		if !manager.IsMediaFile(testImagePath) {
			t.Error("Expected test image to be recognized as media file")
		}

		// Test with text file
		testTextPath := filepath.Join(tempDir, "test.txt")
		err = os.WriteFile(testTextPath, []byte("test"), 0o644)
		if err != nil {
			t.Fatalf("Failed to create test text file: %v", err)
		}

		if manager.IsMediaFile(testTextPath) {
			t.Error("Expected text file to not be recognized as media file")
		}
	})

	t.Run("GetSupportedFormats", func(t *testing.T) {
		imageFormats := manager.GetSupportedImageFormats()
		if len(imageFormats) == 0 {
			t.Error("Expected at least one supported image format")
		}

		videoFormats := manager.GetSupportedVideoFormats()
		if len(videoFormats) == 0 {
			t.Error("Expected at least one supported video format")
		}
	})

	t.Run("ValidateFile", func(t *testing.T) {
		// Test empty path
		err := manager.ValidateFile("")
		if err == nil {
			t.Error("Expected error for empty file path")
		}

		// Test valid file path
		validPath := filepath.Join(tempDir, "valid.txt")
		err = os.WriteFile(validPath, []byte("test"), 0o644)
		if err != nil {
			t.Fatalf("Failed to create valid test file: %v", err)
		}

		err = manager.ValidateFile(validPath)
		if err != nil {
			t.Errorf("Expected no error for valid file, got: %v", err)
		}
	})

	t.Run("SetCacheDir", func(t *testing.T) {
		newCacheDir := filepath.Join(tempDir, "new_cache")
		manager.SetCacheDir(newCacheDir)

		if manager.GetCacheDir() != newCacheDir {
			t.Errorf("Expected cache dir %s, got %s", newCacheDir, manager.GetCacheDir())
		}
	})
}
