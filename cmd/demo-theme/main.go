package main

import (
	"fmt"
	"image/color"
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

	if err := os.MkdirAll(demoDir, 0755); err != nil {
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
			Primary:        color.NRGBA{R: 33, G: 150, B: 243, A: 255},  // Blue
			Secondary:      color.NRGBA{R: 0, G: 188, B: 212, A: 255},   // Cyan
			Background:     color.NRGBA{R: 250, G: 250, B: 250, A: 255}, // Light gray
			Surface:        color.NRGBA{R: 255, G: 255, B: 255, A: 255}, // White
			OnPrimary:      color.NRGBA{R: 255, G: 255, B: 255, A: 255}, // White
			OnSecondary:    color.NRGBA{R: 255, G: 255, B: 255, A: 255}, // White
			OnBackground:   color.NRGBA{R: 33, G: 33, B: 33, A: 255},    // Dark gray
			OnSurface:      color.NRGBA{R: 33, G: 33, B: 33, A: 255},    // Dark gray
			OnError:        color.NRGBA{R: 255, G: 255, B: 255, A: 255}, // White
			Success:        color.NRGBA{R: 76, G: 175, B: 80, A: 255},   // Green
			Warning:        color.NRGBA{R: 255, G: 152, B: 0, A: 255},   // Orange
			Error:          color.NRGBA{R: 244, G: 67, B: 54, A: 255},   // Red
			Info:           color.NRGBA{R: 33, G: 150, B: 243, A: 255},  // Blue
			Disabled:       color.NRGBA{R: 158, G: 158, B: 158, A: 255}, // Gray
			Highlight:      color.NRGBA{R: 227, G: 242, B: 253, A: 255}, // Light blue
			Border:         color.NRGBA{R: 224, G: 224, B: 224, A: 255}, // Light gray
			Divider:        color.NRGBA{R: 238, G: 238, B: 238, A: 255}, // Very light gray
			Shadow:         color.NRGBA{R: 0, G: 0, B: 0, A: 51},        // Transparent black
			SurfaceVariant: color.NRGBA{R: 245, G: 245, B: 245, A: 255}, // Very light gray
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
			Primary:        color.NRGBA{R: 156, G: 39, B: 176, A: 255},  // Purple
			Secondary:      color.NRGBA{R: 233, G: 30, B: 99, A: 255},   // Pink
			Background:     color.NRGBA{R: 18, G: 18, B: 18, A: 255},    // Very dark
			Surface:        color.NRGBA{R: 33, G: 33, B: 33, A: 255},    // Dark gray
			OnPrimary:      color.NRGBA{R: 255, G: 255, B: 255, A: 255}, // White
			OnSecondary:    color.NRGBA{R: 255, G: 255, B: 255, A: 255}, // White
			OnBackground:   color.NRGBA{R: 255, G: 255, B: 255, A: 255}, // White
			OnSurface:      color.NRGBA{R: 255, G: 255, B: 255, A: 255}, // White
			OnError:        color.NRGBA{R: 0, G: 0, B: 0, A: 255},       // Black
			Success:        color.NRGBA{R: 76, G: 175, B: 80, A: 255},   // Green
			Warning:        color.NRGBA{R: 255, G: 152, B: 0, A: 255},   // Orange
			Error:          color.NRGBA{R: 244, G: 67, B: 54, A: 255},   // Red
			Info:           color.NRGBA{R: 33, G: 150, B: 243, A: 255},  // Blue
			Disabled:       color.NRGBA{R: 97, G: 97, B: 97, A: 255},    // Gray
			Highlight:      color.NRGBA{R: 48, G: 48, B: 48, A: 255},    // Dark gray
			Border:         color.NRGBA{R: 66, G: 66, B: 66, A: 255},    // Medium gray
			Divider:        color.NRGBA{R: 48, G: 48, B: 48, A: 255},    // Dark gray
			Shadow:         color.NRGBA{R: 0, G: 0, B: 0, A: 102},       // Transparent black
			SurfaceVariant: color.NRGBA{R: 28, G: 28, B: 28, A: 255},    // Very dark gray
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
