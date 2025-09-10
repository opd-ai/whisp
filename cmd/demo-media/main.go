package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/opd-ai/whisp/internal/core/media"
)

func main() {
	// Print header
	fmt.Println("🎬 Whisp Media Preview Demo")
	fmt.Println("===========================")
	fmt.Println()
	fmt.Println("This demo showcases the media preview functionality including:")
	fmt.Println("- Image thumbnail generation and caching")
	fmt.Println("- Media type detection (images, videos, audio, documents)")
	fmt.Println("- File preview capabilities")
	fmt.Println("- Cross-platform media support")
	fmt.Println()

	// Create temporary directory for demo
	tempDir, err := os.MkdirTemp("", "whisp_media_demo")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("📁 Demo directory: %s\n", tempDir)

	// Create media manager
	cacheDir := filepath.Join(tempDir, "cache")
	mediaManager := media.NewManager(cacheDir)

	// Create sample image files for testing
	testFiles := createSampleFiles(tempDir)

	// Test media detection and processing
	testMediaDetection(mediaManager, testFiles)

	// Test thumbnail generation
	testThumbnailGeneration(mediaManager, testFiles)

	// Create GUI demo
	createGUIDemo(mediaManager, testFiles)
}

// createSampleFiles creates sample files for testing
func createSampleFiles(tempDir string) []string {
	var files []string

	// Create sample PNG image
	pngPath := filepath.Join(tempDir, "sample.png")
	createSampleImage(pngPath, 300, 200, color.RGBA{255, 100, 100, 255})
	files = append(files, pngPath)

	// Create another sample image
	jpgPath := filepath.Join(tempDir, "photo.jpg") // We'll treat as PNG for simplicity
	createSampleImage(jpgPath, 400, 300, color.RGBA{100, 255, 100, 255})
	files = append(files, jpgPath)

	// Create text file
	txtPath := filepath.Join(tempDir, "document.txt")
	err := os.WriteFile(txtPath, []byte("This is a sample text document for testing."), 0o644)
	if err != nil {
		log.Printf("Failed to create text file: %v", err)
	} else {
		files = append(files, txtPath)
	}

	// Create fake video file (just for testing detection)
	videoPath := filepath.Join(tempDir, "video.mp4")
	err = os.WriteFile(videoPath, []byte("fake video content"), 0o644)
	if err != nil {
		log.Printf("Failed to create fake video file: %v", err)
	} else {
		files = append(files, videoPath)
	}

	return files
}

// createSampleImage creates a simple colored image
func createSampleImage(path string, width, height int, col color.RGBA) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Create a gradient pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Create gradient effect
			r := uint8((x * 255) / width)
			g := col.G
			b := uint8((y * 255) / height)
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	file, err := os.Create(path)
	if err != nil {
		log.Printf("Failed to create image file %s: %v", path, err)
		return
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		log.Printf("Failed to encode image %s: %v", path, err)
	}
}

// testMediaDetection tests media type detection
func testMediaDetection(manager *media.Manager, files []string) {
	fmt.Println("🔍 Testing Media Detection")
	fmt.Println("-------------------------")

	for _, file := range files {
		filename := filepath.Base(file)

		// Test if it's a media file
		isMedia := manager.IsMediaFile(file)
		fmt.Printf("📄 %s: Media file = %v", filename, isMedia)

		// Get media type string
		mediaType := manager.GetMediaTypeString(file)
		fmt.Printf(" (Type: %s)", mediaType)

		// Test specific type checks
		if manager.IsImageFile(file) {
			fmt.Printf(" [IMAGE]")
		}
		if manager.IsVideoFile(file) {
			fmt.Printf(" [VIDEO]")
		}

		fmt.Println()

		// Get detailed media info for media files
		if isMedia {
			info, err := manager.GetMediaInfo(file)
			if err != nil {
				fmt.Printf("   ❌ Error getting media info: %v\n", err)
			} else {
				fmt.Printf("   📊 Size: %d bytes", info.Size)
				if info.Width > 0 && info.Height > 0 {
					fmt.Printf(", Dimensions: %dx%d", info.Width, info.Height)
				}
				fmt.Printf(", MIME: %s\n", info.MimeType)
			}
		}
		fmt.Println()
	}
}

// testThumbnailGeneration tests thumbnail generation and caching
func testThumbnailGeneration(manager *media.Manager, files []string) {
	fmt.Println("🖼️  Testing Thumbnail Generation")
	fmt.Println("-------------------------------")

	for _, file := range files {
		filename := filepath.Base(file)

		if !manager.IsMediaFile(file) {
			fmt.Printf("⏭️  Skipping %s (not a media file)\n", filename)
			continue
		}

		fmt.Printf("🎨 Generating thumbnail for %s...\n", filename)

		// Test thumbnail generation
		thumbnailPath, err := manager.GenerateThumbnail(file, 150, 100)
		if err != nil {
			fmt.Printf("   ❌ Error: %v\n", err)
		} else {
			fmt.Printf("   ✅ Thumbnail created: %s\n", filepath.Base(thumbnailPath))

			// Test cache retrieval
			cachedPath, cached := manager.GetThumbnailPath(file, 150, 100)
			if cached && cachedPath == thumbnailPath {
				fmt.Printf("   💾 Cache working correctly\n")
			} else {
				fmt.Printf("   ⚠️  Cache issue: cached=%v, paths match=%v\n",
					cached, cachedPath == thumbnailPath)
			}
		}
		fmt.Println()
	}

	// Test cache cleanup
	fmt.Printf("🧹 Testing cache cleanup...\n")
	err := manager.Cleanup()
	if err != nil {
		fmt.Printf("   ❌ Error cleaning cache: %v\n", err)
	} else {
		fmt.Printf("   ✅ Cache cleaned successfully\n")
	}
	fmt.Println()
}

// createGUIDemo creates a simple GUI to demonstrate media preview
func createGUIDemo(manager *media.Manager, files []string) {
	fmt.Println("🖥️  Launching GUI Demo...")
	fmt.Println("Press Ctrl+C to exit after exploring the GUI")
	fmt.Println()

	// Create Fyne app
	myApp := app.NewWithID("com.whisp.media-demo")

	window := myApp.NewWindow("Whisp Media Preview Demo")
	window.Resize(fyne.NewSize(800, 600))

	// Create content
	content := container.NewVBox(
		widget.NewCard("Media Preview Demo",
			"This demonstrates the media preview functionality implemented for Whisp.",
			widget.NewLabel("Select a file below to see its preview:")),
	)

	// Add file list with preview
	for _, file := range files {
		fileCard := createFileCard(manager, file)
		content.Add(fileCard)
	}

	// Add supported formats info
	infoCard := widget.NewCard("Supported Formats", "",
		container.NewVBox(
			widget.NewLabel("📷 Images: "+fmt.Sprintf("%v", manager.GetSupportedImageFormats())),
			widget.NewLabel("🎬 Videos: "+fmt.Sprintf("%v", manager.GetSupportedVideoFormats())),
			widget.NewLabel("🔧 Features: Thumbnail generation, caching, media detection"),
		))
	content.Add(infoCard)

	// Add scroll container
	scroll := container.NewScroll(content)
	window.SetContent(scroll)

	// Show and run
	window.ShowAndRun()
}

// createFileCard creates a card widget for each file
func createFileCard(manager *media.Manager, filePath string) *widget.Card {
	filename := filepath.Base(filePath)

	// File info
	info := widget.NewLabel(fmt.Sprintf("File: %s", filename))

	// Media type
	mediaType := manager.GetMediaTypeString(filePath)
	typeLabel := widget.NewLabel(fmt.Sprintf("Type: %s", mediaType))

	// Media file status
	isMedia := manager.IsMediaFile(filePath)
	statusLabel := widget.NewLabel(fmt.Sprintf("Media file: %v", isMedia))

	content := container.NewVBox(info, typeLabel, statusLabel)

	// Add thumbnail info for media files
	if isMedia {
		thumbnailBtn := widget.NewButton("Generate Thumbnail", func() {
			_, err := manager.GenerateThumbnail(filePath, 150, 100)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("Error: %v", err))
			} else {
				statusLabel.SetText("✅ Thumbnail generated successfully!")
			}
		})
		content.Add(thumbnailBtn)
	}

	return widget.NewCard("", "", content)
}
