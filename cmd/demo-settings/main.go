package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/opd-ai/whisp/internal/core/config"
	"github.com/opd-ai/whisp/ui/shared"
)

// Demo application to test the settings panel functionality
// This demonstrates the complete settings system implementation
func main() {
	fmt.Println("Whisp Settings Panel Demo")
	fmt.Println("==========================")

	// Create a demo app
	demoApp := app.New()

	// Create temporary config directory for demo
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}

	configDir := filepath.Join(homeDir, ".config", "whisp-demo")
	configPath := filepath.Join(configDir, "config.yaml")

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Fatalf("Failed to create config directory: %v", err)
	}

	fmt.Printf("Demo config location: %s\n", configPath)

	// Initialize configuration manager
	configMgr, err := config.NewManager(configPath)
	if err != nil {
		log.Fatalf("Failed to initialize config manager: %v", err)
	}

	// Display current configuration
	fmt.Println("\nCurrent Configuration:")
	cfg := configMgr.GetConfig()
	fmt.Printf("- Theme: %s\n", cfg.UI.Theme)
	fmt.Printf("- Font Size: %s\n", cfg.UI.FontSize)
	fmt.Printf("- Language: %s\n", cfg.UI.Language)
	fmt.Printf("- Encryption Enabled: %t\n", cfg.Storage.EnableEncryption)
	fmt.Printf("- Save Message History: %t\n", cfg.Privacy.SaveMessageHistory)
	fmt.Printf("- Notifications Enabled: %t\n", cfg.Notifications.Enabled)
	fmt.Printf("- Log Level: %s\n", cfg.Advanced.LogLevel)

	// Create main window
	window := demoApp.NewWindow("Settings Panel Demo")
	window.Resize(fyne.NewSize(800, 600))

	// Create demo content
	label := widget.NewLabel("Click the button below to open the Settings Panel")

	configInfo := widget.NewEntry()
	configInfo.MultiLine = true
	configInfo.SetText(fmt.Sprintf(
		"Config file: %s\n\n"+
			"Current settings:\n"+
			"• Theme: %s\n"+
			"• Font Size: %s\n"+
			"• Language: %s\n"+
			"• Encryption: %t\n"+
			"• Save History: %t\n"+
			"• Notifications: %t\n"+
			"• Log Level: %s",
		configPath,
		cfg.UI.Theme,
		cfg.UI.FontSize,
		cfg.UI.Language,
		cfg.Storage.EnableEncryption,
		cfg.Privacy.SaveMessageHistory,
		cfg.Notifications.Enabled,
		cfg.Advanced.LogLevel,
	))

	refreshInfo := func() {
		newCfg := configMgr.GetConfig()
		configInfo.SetText(fmt.Sprintf(
			"Config file: %s\n\n"+
				"Current settings:\n"+
				"• Theme: %s\n"+
				"• Font Size: %s\n"+
				"• Language: %s\n"+
				"• Encryption: %t\n"+
				"• Save History: %t\n"+
				"• Notifications: %t\n"+
				"• Log Level: %s",
			configPath,
			newCfg.UI.Theme,
			newCfg.UI.FontSize,
			newCfg.UI.Language,
			newCfg.Storage.EnableEncryption,
			newCfg.Privacy.SaveMessageHistory,
			newCfg.Notifications.Enabled,
			newCfg.Advanced.LogLevel,
		))
	}

	settingsBtn := widget.NewButton("Open Settings Panel", func() {
		settingsDialog := shared.NewSettingsDialog(configMgr, window)
		settingsDialog.Show()

		// Refresh info after dialog closes (simple delay for demo)
		go func() {
			// In a real app, you'd have proper callbacks for when settings change
			time.Sleep(100 * time.Millisecond)
			refreshInfo()
		}()
	})
	settingsBtn.Importance = widget.HighImportance

	testBtn := widget.NewButton("Test Config Operations", func() {
		fmt.Println("\nTesting configuration operations...")

		// Test reading current config
		currentCfg := configMgr.GetConfig()
		fmt.Printf("Current theme: %s\n", currentCfg.UI.Theme)

		// Test modifying config
		if currentCfg.UI.Theme == "system" {
			currentCfg.UI.Theme = "dark"
		} else {
			currentCfg.UI.Theme = "system"
		}

		// Test saving config
		if err := configMgr.UpdateConfig(currentCfg); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
		} else {
			fmt.Printf("Successfully changed theme to: %s\n", currentCfg.UI.Theme)
			refreshInfo()
		}
	})

	// Layout
	content := container.NewVBox(
		widget.NewCard("Settings Panel Demo", "Test the complete settings system",
			container.NewVBox(
				label,
				settingsBtn,
				widget.NewSeparator(),
				testBtn,
			),
		),
		widget.NewCard("Configuration Information", "", configInfo),
	)

	window.SetContent(container.NewScroll(content))

	// Show initial status
	fmt.Println("\nDemo is ready! The GUI window should be open.")
	fmt.Println("Use the settings panel to modify configuration values.")
	fmt.Println("Changes will be saved to:", configPath)

	// Show window and run
	window.ShowAndRun()

	// Final status
	fmt.Println("\nDemo completed.")
	finalCfg := configMgr.GetConfig()
	fmt.Printf("Final theme: %s\n", finalCfg.UI.Theme)
	fmt.Printf("Final log level: %s\n", finalCfg.Advanced.LogLevel)
}
