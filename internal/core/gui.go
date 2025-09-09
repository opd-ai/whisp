package core

import (
	"context"
	"fmt"
	"log"

	"fyne.io/fyne/v2/app"
	"github.com/opd-ai/whisp/ui/adaptive"
)

// StartGUI starts the graphical user interface
func (a *App) StartGUI(ctx context.Context) error {
	if a.config.Platform == "headless" {
		return fmt.Errorf("GUI not available in headless mode")
	}

	// Create Fyne application
	fyneApp := app.NewWithID("com.opd-ai.whisp")

	// Create adaptive UI
	ui, err := adaptive.NewUI(fyneApp, a, adaptive.DetectPlatform())
	if err != nil {
		return fmt.Errorf("failed to create UI: %w", err)
	}

	// Initialize UI with core app context
	if err := ui.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize UI: %w", err)
	}

	// Start UI in separate goroutine to avoid blocking
	go func() {
		log.Println("Starting GUI...")
		ui.ShowMainWindow()
	}()

	return nil
}
