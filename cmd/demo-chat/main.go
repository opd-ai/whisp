package main

import (
	"context"
	"fmt"
	"log"

	"github.com/opd-ai/whisp/internal/core"
	"github.com/opd-ai/whisp/ui/adaptive"
)

// ChatViewDemo demonstrates the chat view functionality
func main() {
	fmt.Println("=== Whisp Chat View Implementation Demo ===")

	// Create core app
	config := &core.Config{
		DataDir:  "./demo-data",
		Debug:    true,
		Platform: adaptive.DetectPlatform(),
	}

	coreApp, err := core.NewApp(config)
	if err != nil {
		log.Fatalf("Failed to create core app: %v", err)
	}
	defer coreApp.Cleanup()

	fmt.Println("✅ Core application created successfully")
	fmt.Println("✅ UI components implemented:")
	fmt.Println("   - Chat view with message display and input")
	fmt.Println("   - Contact list with Add Friend functionality")
	fmt.Println("   - Add Friend dialog with Tox ID validation")
	fmt.Println("   - Message history loading from database")
	fmt.Println("   - Contact selection integration")
	fmt.Println("   - Menu bar with Friends menu")
	fmt.Println("")
	fmt.Println("🚀 Starting GUI application...")

	// Start GUI
	if err := coreApp.StartGUI(context.Background()); err != nil {
		log.Fatalf("Failed to start GUI: %v", err)
	}
}
