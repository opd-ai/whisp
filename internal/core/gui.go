package core

import (
	"context"
	"fmt"
	"log"

	"fyne.io/fyne/v2/app"
	"github.com/opd-ai/whisp/internal/core/message"
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

// SendMessageFromUI handles message sending from UI
func (a *App) SendMessageFromUI(friendID uint32, content string) error {
	if a.messages == nil {
		return fmt.Errorf("message manager not initialized")
	}
	
	_, err := a.messages.SendMessage(friendID, content, message.MessageTypeNormal)
	return err
}

// AddContactFromUI handles contact addition from UI
func (a *App) AddContactFromUI(toxID, messageText string) error {
	if a.contacts == nil {
		return fmt.Errorf("contact manager not initialized")
	}
	
	_, err := a.contacts.AddContact(toxID, messageText)
	return err
}
