package main

import (
	"context"
	"log"
	"time"

	"fyne.io/fyne/v2/app"

	"github.com/opd-ai/whisp/internal/core"
	"github.com/opd-ai/whisp/ui/adaptive"
)

func main() {
	log.Println("Starting Whisp Mobile UI Demo...")

	// Create Fyne application
	fyneApp := app.New()

	// Create core application with minimal config
	config := &core.Config{
		DataDir: "./demo-data",
	}
	coreApp, err := core.NewApp(config)
	if err != nil {
		log.Fatalf("Failed to create core app: %v", err)
	}

	// Force mobile platform for demo purposes
	platform := adaptive.PlatformAndroid
	log.Printf("Using platform: %s (Mobile: %t)", platform, platform.IsMobile())

	// Create adaptive UI with mobile platform
	ui, err := adaptive.NewUI(fyneApp, coreApp, platform)
	if err != nil {
		log.Fatalf("Failed to create UI: %v", err)
	}

	// Initialize UI
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := ui.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize UI: %v", err)
	}

	log.Println("Mobile UI Demo Features:")
	log.Println("- Bottom tab navigation (Contacts, Chat, Settings)")
	log.Println("- Touch-optimized buttons and layouts")
	log.Println("- Mobile window sizing (360x640)")
	log.Println("- Pull-to-refresh in contacts")
	log.Println("- Automatic navigation on contact selection")
	log.Println("- Mobile-specific settings view")

	// Show main window with mobile layout
	ui.ShowMainWindow()
}
