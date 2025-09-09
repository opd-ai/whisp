package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2/app"

	"github.com/opd-ai/whisp/internal/core"
	"github.com/opd-ai/whisp/ui/adaptive"
)

func main() {
	log.Println("Starting Whisp Desktop UI Demo...")

	// Create Fyne application
	fyneApp := app.New()

	// Set up configuration path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}
	configPath := filepath.Join(homeDir, ".config", "whisp", "config.yaml")

	// Create core configuration
	config := &core.Config{
		ConfigPath: configPath,
		Debug:      true,
	}

	// Create core application
	coreApp, err := core.NewApp(config)
	if err != nil {
		log.Fatalf("Failed to create core app: %v", err)
	}

	// Detect platform
	platform := adaptive.DetectPlatform()
	log.Printf("Detected platform: %s", platform)

	// Create UI
	ui, err := adaptive.NewUI(fyneApp, coreApp, platform)
	if err != nil {
		log.Fatalf("Failed to create UI: %v", err)
	}

	// Initialize UI
	ctx := context.Background()
	if err := ui.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize UI: %v", err)
	}

	log.Println("Desktop UI Demo initialized successfully!")
	log.Println("Features available:")
	log.Println("  - Keyboard shortcuts:")
	log.Println("    * Ctrl+N: Add Friend")
	log.Println("    * Ctrl+Q: Quit")
	log.Println("    * Ctrl+,: Settings")
	log.Println("  - Window state persistence")
	log.Println("  - Enhanced About dialog")
	log.Println("  - Copy-to-clipboard in Tox ID dialog")

	// Show main window and run
	ui.ShowMainWindow()

	log.Println("Whisp Desktop UI Demo completed.")
}
