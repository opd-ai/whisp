package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/opd-ai/whisp/ui/theme"
)

func main() {
	fmt.Println("=== Whisp Theme System Demo ===")

	// Create demo directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	demoDir := filepath.Join(homeDir, ".local", "share", "whisp", "demo-theme")
	fmt.Printf("Demo directory: %s\n", demoDir)

	if err := os.MkdirAll(demoDir, 0o755); err != nil {
		log.Fatal(err)
	}

	// Create Fyne app
	fyneApp := app.NewWithID("ai.opd.whisp.theme-demo")

	// Initialize theme manager
	themeManager := theme.NewDefaultThemeManager(demoDir)
	if err := themeManager.Initialize(fyneApp); err != nil {
		log.Fatalf("Failed to initialize theme manager: %v", err)
	}

	fmt.Println("✓ Theme manager initialized")

	// Demo 1: Test basic theme switching
	fmt.Println("\n=== Testing Basic Theme Switching ===")

	// Test light theme
	if err := themeManager.SetTheme(theme.ThemeLight); err != nil {
		log.Printf("Failed to set light theme: %v", err)
	} else {
		fmt.Println("✓ Light theme applied")
	}

	// Test dark theme
	if err := themeManager.SetTheme(theme.ThemeDark); err != nil {
		log.Printf("Failed to set dark theme: %v", err)
	} else {
		fmt.Println("✓ Dark theme applied")
	}

	// Test system theme
	systemTheme := themeManager.DetectSystemTheme()
	fmt.Printf("✓ Detected system theme: %v\n", systemTheme)

	if err := themeManager.SetTheme(theme.ThemeSystem); err != nil {
		log.Printf("Failed to set system theme: %v", err)
	} else {
		fmt.Println("✓ System theme applied")
	}

	// Demo 2: Test custom themes
	fmt.Println("\n=== Testing Custom Themes ===")

	// Create a custom blue theme
	blueTheme := theme.CustomTheme{
		Name:        "Ocean Blue",
		Description: "A soothing blue theme",
		ColorScheme: theme.ColorScheme{
			Primary:        theme.NewSerializableColorFromRGBA(33, 150, 243, 255),  // Blue
			Secondary:      theme.NewSerializableColorFromRGBA(0, 188, 212, 255),   // Cyan
			Background:     theme.NewSerializableColorFromRGBA(250, 250, 250, 255), // Light gray
			Surface:        theme.NewSerializableColorFromRGBA(255, 255, 255, 255), // White
			OnPrimary:      theme.NewSerializableColorFromRGBA(255, 255, 255, 255), // White
			OnSecondary:    theme.NewSerializableColorFromRGBA(255, 255, 255, 255), // White
			OnBackground:   theme.NewSerializableColorFromRGBA(33, 33, 33, 255),    // Dark gray
			OnSurface:      theme.NewSerializableColorFromRGBA(33, 33, 33, 255),    // Dark gray
			OnError:        theme.NewSerializableColorFromRGBA(255, 255, 255, 255), // White
			Success:        theme.NewSerializableColorFromRGBA(76, 175, 80, 255),   // Green
			Warning:        theme.NewSerializableColorFromRGBA(255, 152, 0, 255),   // Orange
			Error:          theme.NewSerializableColorFromRGBA(244, 67, 54, 255),   // Red
			Info:           theme.NewSerializableColorFromRGBA(33, 150, 243, 255),  // Blue
			Disabled:       theme.NewSerializableColorFromRGBA(158, 158, 158, 255), // Gray
			Highlight:      theme.NewSerializableColorFromRGBA(227, 242, 253, 255), // Light blue
			Border:         theme.NewSerializableColorFromRGBA(224, 224, 224, 255), // Light gray
			Divider:        theme.NewSerializableColorFromRGBA(238, 238, 238, 255), // Very light gray
			Shadow:         theme.NewSerializableColorFromRGBA(0, 0, 0, 51),        // Transparent black
			SurfaceVariant: theme.NewSerializableColorFromRGBA(245, 245, 245, 255), // Very light gray
		},
	}

	if err := themeManager.CreateCustomTheme(blueTheme); err != nil {
		log.Printf("Failed to create blue theme: %v", err)
	} else {
		fmt.Println("✓ Custom 'Ocean Blue' theme created")
	}

	// Create a custom dark theme
	darkTheme := theme.CustomTheme{
		Name:        "Midnight Dark",
		Description: "A deep dark theme",
		ColorScheme: theme.ColorScheme{
			Primary:        theme.NewSerializableColorFromRGBA(156, 39, 176, 255),  // Purple
			Secondary:      theme.NewSerializableColorFromRGBA(233, 30, 99, 255),   // Pink
			Background:     theme.NewSerializableColorFromRGBA(18, 18, 18, 255),    // Very dark
			Surface:        theme.NewSerializableColorFromRGBA(33, 33, 33, 255),    // Dark gray
			OnPrimary:      theme.NewSerializableColorFromRGBA(255, 255, 255, 255), // White
			OnSecondary:    theme.NewSerializableColorFromRGBA(255, 255, 255, 255), // White
			OnBackground:   theme.NewSerializableColorFromRGBA(255, 255, 255, 255), // White
			OnSurface:      theme.NewSerializableColorFromRGBA(255, 255, 255, 255), // White
			OnError:        theme.NewSerializableColorFromRGBA(0, 0, 0, 255),       // Black
			Success:        theme.NewSerializableColorFromRGBA(76, 175, 80, 255),   // Green
			Warning:        theme.NewSerializableColorFromRGBA(255, 152, 0, 255),   // Orange
			Error:          theme.NewSerializableColorFromRGBA(244, 67, 54, 255),   // Red
			Info:           theme.NewSerializableColorFromRGBA(33, 150, 243, 255),  // Blue
			Disabled:       theme.NewSerializableColorFromRGBA(97, 97, 97, 255),    // Gray
			Highlight:      theme.NewSerializableColorFromRGBA(48, 48, 48, 255),    // Dark gray
			Border:         theme.NewSerializableColorFromRGBA(66, 66, 66, 255),    // Medium gray
			Divider:        theme.NewSerializableColorFromRGBA(48, 48, 48, 255),    // Dark gray
			Shadow:         theme.NewSerializableColorFromRGBA(0, 0, 0, 102),       // Transparent black
			SurfaceVariant: theme.NewSerializableColorFromRGBA(28, 28, 28, 255),    // Very dark gray
		},
	}

	if err := themeManager.CreateCustomTheme(darkTheme); err != nil {
		log.Printf("Failed to create dark theme: %v", err)
	} else {
		fmt.Println("✓ Custom 'Midnight Dark' theme created")
	}

	// List all custom themes
	customThemes := themeManager.ListCustomThemes()
	fmt.Printf("✓ Total custom themes created: %d\n", len(customThemes))
	for _, ct := range customThemes {
		fmt.Printf("  - %s: %s\n", ct.Name, ct.Description)
	}

	// Demo 3: Test theme preferences
	fmt.Println("\n=== Testing Theme Preferences ===")

	prefs := themeManager.GetPreferences()
	fmt.Printf("✓ Current preferences: Theme=%v, FollowSystem=%v\n", prefs.ThemeType, prefs.FollowSystemTheme)

	// Demo 4: Create a simple GUI window to show theme in action
	fmt.Println("\n=== Creating Theme Demo Window ===")

	window := fyneApp.NewWindow("Whisp Theme Demo")
	window.Resize(fyne.NewSize(600, 400))

	// Create theme selection widgets
	lightBtn := widget.NewButton("Light Theme", func() {
		themeManager.SetTheme(theme.ThemeLight)
		fmt.Println("Switched to light theme")
	})

	darkBtn := widget.NewButton("Dark Theme", func() {
		themeManager.SetTheme(theme.ThemeDark)
		fmt.Println("Switched to dark theme")
	})

	systemBtn := widget.NewButton("System Theme", func() {
		themeManager.SetTheme(theme.ThemeSystem)
		fmt.Println("Switched to system theme")
	})

	// Theme selection buttons
	themeSelector := container.NewHBox(
		lightBtn,
		darkBtn,
		systemBtn,
	)

	// Sample UI elements to demonstrate theming
	sampleText := widget.NewRichTextFromMarkdown(`
# Theme Demo

This window demonstrates the **Whisp Theme System** in action.

## Features:
- Light and dark themes
- System theme detection
- Custom color schemes
- Persistent preferences

Try switching between themes using the buttons above!
`)

	sampleEntry := widget.NewEntry()
	sampleEntry.SetText("Type here to see text input styling...")

	sampleProgress := widget.NewProgressBar()
	sampleProgress.SetValue(0.7)

	sampleCheck := widget.NewCheck("Enable theme persistence", func(checked bool) {
		fmt.Printf("Theme persistence: %v\n", checked)
	})
	sampleCheck.SetChecked(true)

	// Layout the demo window
	content := container.NewBorder(
		widget.NewCard("Theme Selection", "", themeSelector),
		container.NewVBox(
			widget.NewSeparator(),
			widget.NewLabel("Progress Demo:"),
			sampleProgress,
			sampleCheck,
		),
		nil,
		nil,
		container.NewBorder(
			nil,
			container.NewVBox(
				widget.NewSeparator(),
				widget.NewLabel("Text Input Demo:"),
				sampleEntry,
			),
			nil,
			nil,
			sampleText,
		),
	)

	window.SetContent(content)

	// Set up theme change callback
	themeManager.OnThemeChanged(func(newTheme theme.ThemeType) {
		fmt.Printf("Theme changed to: %v\n", newTheme)
	})

	// Apply initial theme
	themeManager.ApplyTheme(fyneApp)

	fmt.Println("✓ Demo window created")
	fmt.Println("\n=== Theme System Demo Complete ===")
	fmt.Println("The demo window is now open. Try switching themes!")
	fmt.Println("Close the window to exit the demo.")

	// Show window and run app
	window.ShowAndRun()
}
