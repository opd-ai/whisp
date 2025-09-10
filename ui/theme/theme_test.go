package theme

import (
	"image/color"
	"os"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

func TestThemeTypes(t *testing.T) {
	tests := []struct {
		name      string
		themeType ThemeType
		expected  string
	}{
		{"Light theme", ThemeLight, "light"},
		{"Dark theme", ThemeDark, "dark"},
		{"System theme", ThemeSystem, "system"},
		{"Custom theme", ThemeCustom, "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: ThemeType is an iota, so we can't directly convert to string
			// This test just verifies the constants exist
			_ = tt.themeType
		})
	}
}

func TestColorSchemes(t *testing.T) {
	t.Run("LightColorScheme", func(t *testing.T) {
		scheme := LightColorScheme

		// Verify primary colors are set (check that they're not zero values)
		if scheme.Primary == (SerializableColor{}) {
			t.Error("Primary color should not be zero value")
		}
		if scheme.Background == (SerializableColor{}) {
			t.Error("Background color should not be zero value")
		}
		if scheme.OnBackground == (SerializableColor{}) {
			t.Error("OnBackground color should not be zero value")
		}
	})

	t.Run("DarkColorScheme", func(t *testing.T) {
		scheme := DarkColorScheme

		// Verify primary colors are set (check that they're not zero values)
		if scheme.Primary == (SerializableColor{}) {
			t.Error("Primary color should not be zero value")
		}
		if scheme.Background == (SerializableColor{}) {
			t.Error("Background color should not be zero value")
		}
		if scheme.OnBackground == (SerializableColor{}) {
			t.Error("OnBackground color should not be zero value")
		}
	})
}

func TestWhispTheme(t *testing.T) {
	t.Run("NewLightTheme", func(t *testing.T) {
		whispTheme := NewLightTheme()

		if whispTheme == nil {
			t.Fatal("Theme should not be nil")
		}
		if whispTheme.IsDark() {
			t.Error("Light theme should not be dark")
		}
		if whispTheme.GetVariant() != theme.VariantLight {
			t.Error("Light theme should have light variant")
		}
	})

	t.Run("NewDarkTheme", func(t *testing.T) {
		whispTheme := NewDarkTheme()

		if whispTheme == nil {
			t.Fatal("Theme should not be nil")
		}
		if !whispTheme.IsDark() {
			t.Error("Dark theme should be dark")
		}
		if whispTheme.GetVariant() != theme.VariantDark {
			t.Error("Dark theme should have dark variant")
		}
	})

	t.Run("ThemeColors", func(t *testing.T) {
		whispTheme := NewLightTheme()

		// Test basic color lookups
		primaryColor := whispTheme.Color(theme.ColorNamePrimary, theme.VariantLight)
		if primaryColor == nil {
			t.Error("Primary color should not be nil")
		}

		backgroundColor := whispTheme.Color(theme.ColorNameBackground, theme.VariantLight)
		if backgroundColor == nil {
			t.Error("Background color should not be nil")
		}
	})

	t.Run("ThemeFonts", func(t *testing.T) {
		whispTheme := NewLightTheme()

		// Test font lookup
		regularFont := whispTheme.Font(fyne.TextStyle{})
		if regularFont == nil {
			t.Error("Regular font should not be nil")
		}

		boldFont := whispTheme.Font(fyne.TextStyle{Bold: true})
		if boldFont == nil {
			t.Error("Bold font should not be nil")
		}
	})

	t.Run("ThemeIcons", func(t *testing.T) {
		whispTheme := NewLightTheme()

		// Test icon lookup using a basic icon
		documentIcon := whispTheme.Icon(theme.IconNameDocument)
		if documentIcon == nil {
			t.Error("Document icon should not be nil")
		}
	})

	t.Run("ThemeSizes", func(t *testing.T) {
		whispTheme := NewLightTheme()

		// Test size lookup
		textSize := whispTheme.Size(theme.SizeNameText)
		if textSize <= 0 {
			t.Error("Text size should be positive")
		}
	})
}

func TestSystemThemeDetector(t *testing.T) {
	t.Run("NewSystemThemeDetector", func(t *testing.T) {
		detector := NewSystemThemeDetector()
		if detector == nil {
			t.Fatal("Detector should not be nil")
		}
	})

	t.Run("DetectSystemTheme", func(t *testing.T) {
		detector := NewSystemThemeDetector()
		themeType := detector.DetectSystemTheme()

		// Should return a valid theme type
		validTypes := map[ThemeType]bool{
			ThemeLight: true,
			ThemeDark:  true,
		}

		if !validTypes[themeType] {
			t.Errorf("Invalid theme type detected: %v", themeType)
		}
	})
}

func TestThemeManager(t *testing.T) {
	// Create temporary directory for test configuration
	tempDir, err := os.MkdirTemp("", "whisp_theme_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("NewDefaultThemeManager", func(t *testing.T) {
		manager := NewDefaultThemeManager(tempDir)
		if manager == nil {
			t.Fatal("Manager should not be nil")
		}

		// Check default values
		if manager.GetThemeType() != ThemeSystem {
			t.Error("Default theme should be system")
		}

		if manager.GetCurrentTheme() == nil {
			t.Error("Current theme should not be nil")
		}
	})

	t.Run("Initialize", func(t *testing.T) {
		manager := NewDefaultThemeManager(tempDir)
		app := test.NewApp()

		err := manager.Initialize(app)
		if err != nil {
			t.Fatalf("Failed to initialize theme manager: %v", err)
		}
	})

	t.Run("SetTheme", func(t *testing.T) {
		manager := NewDefaultThemeManager(tempDir)
		app := test.NewApp()
		manager.Initialize(app)

		// Test setting light theme
		err := manager.SetTheme(ThemeLight)
		if err != nil {
			t.Fatalf("Failed to set light theme: %v", err)
		}
		if manager.GetThemeType() != ThemeLight {
			t.Error("Theme type should be light")
		}

		// Test setting dark theme
		err = manager.SetTheme(ThemeDark)
		if err != nil {
			t.Fatalf("Failed to set dark theme: %v", err)
		}
		if manager.GetThemeType() != ThemeDark {
			t.Error("Theme type should be dark")
		}
	})

	t.Run("CustomThemes", func(t *testing.T) {
		manager := NewDefaultThemeManager(tempDir)
		app := test.NewApp()
		manager.Initialize(app)

		// Create custom theme
		customTheme := CustomTheme{
			Name:        "Test Theme",
			Description: "A test theme",
			ColorScheme: ColorScheme{
				Primary:      NewSerializableColorFromRGBA(255, 0, 0, 255),
				Secondary:    NewSerializableColorFromRGBA(0, 255, 0, 255),
				Background:   NewSerializableColorFromRGBA(255, 255, 255, 255),
				Surface:      NewSerializableColorFromRGBA(248, 248, 248, 255),
				OnPrimary:    NewSerializableColorFromRGBA(255, 255, 255, 255),
				OnSecondary:  NewSerializableColorFromRGBA(255, 255, 255, 255),
				OnBackground: NewSerializableColorFromRGBA(0, 0, 0, 255),
				OnSurface:    NewSerializableColorFromRGBA(0, 0, 0, 255),
			},
		}

		// Test creating custom theme
		err := manager.CreateCustomTheme(customTheme)
		if err != nil {
			t.Fatalf("Failed to create custom theme: %v", err)
		}

		// Test retrieving custom theme
		retrievedTheme, err := manager.GetCustomTheme("Test Theme")
		if err != nil {
			t.Fatalf("Failed to get custom theme: %v", err)
		}
		if retrievedTheme.Name != "Test Theme" {
			t.Error("Retrieved theme name mismatch")
		}

		// Test listing custom themes
		themes := manager.ListCustomThemes()
		if len(themes) != 1 {
			t.Error("Should have one custom theme")
		}

		// Test updating custom theme
		customTheme.Description = "Updated description"
		err = manager.UpdateCustomTheme(customTheme)
		if err != nil {
			t.Fatalf("Failed to update custom theme: %v", err)
		}

		// Test deleting custom theme
		err = manager.DeleteCustomTheme("Test Theme")
		if err != nil {
			t.Fatalf("Failed to delete custom theme: %v", err)
		}

		themes = manager.ListCustomThemes()
		if len(themes) != 0 {
			t.Error("Should have no custom themes after deletion")
		}
	})

	t.Run("Preferences", func(t *testing.T) {
		// Create isolated temporary directory for this test
		testTempDir, err := os.MkdirTemp("", "whisp_theme_test_preferences_*")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(testTempDir)

		manager := NewDefaultThemeManager(testTempDir)
		app := test.NewApp()
		manager.Initialize(app)

		// Test getting preferences (before initialization)
		prefs := manager.GetPreferences()
		if prefs.ThemeType != ThemeSystem {
			t.Error("Default theme type should be system")
		}

		// After initialization, the theme may be resolved to the actual system theme

		// Test setting preferences
		newPrefs := ThemePreferences{
			ThemeType:         ThemeDark,
			FollowSystemTheme: false,
			AutoSwitchEnabled: true,
			LightThemeStart:   time.Date(0, 1, 1, 6, 0, 0, 0, time.UTC),
			DarkThemeStart:    time.Date(0, 1, 1, 18, 0, 0, 0, time.UTC),
		}

		err = manager.SetPreferences(newPrefs)
		if err != nil {
			t.Fatalf("Failed to set preferences: %v", err)
		}

		updatedPrefs := manager.GetPreferences()
		if updatedPrefs.ThemeType != ThemeDark {
			t.Error("Theme type should be updated")
		}
		if updatedPrefs.AutoSwitchEnabled != true {
			t.Error("Auto switch should be enabled")
		}
	})

	t.Run("AutoSwitch", func(t *testing.T) {
		manager := NewDefaultThemeManager(tempDir)
		app := test.NewApp()
		manager.Initialize(app)

		// Enable auto switch
		lightStart := time.Date(0, 1, 1, 6, 0, 0, 0, time.UTC)
		darkStart := time.Date(0, 1, 1, 18, 0, 0, 0, time.UTC)

		manager.EnableAutoSwitch(true, lightStart, darkStart)

		prefs := manager.GetPreferences()
		if !prefs.AutoSwitchEnabled {
			t.Error("Auto switch should be enabled")
		}

		// Test check auto switch (won't actually switch in test)
		manager.CheckAutoSwitch()

		// Disable auto switch
		manager.EnableAutoSwitch(false, lightStart, darkStart)

		prefs = manager.GetPreferences()
		if prefs.AutoSwitchEnabled {
			t.Error("Auto switch should be disabled")
		}
	})

	t.Run("ThemeCallbacks", func(t *testing.T) {
		manager := NewDefaultThemeManager(tempDir)
		app := test.NewApp()
		manager.Initialize(app)

		callbackCalled := false
		var callbackTheme ThemeType

		manager.OnThemeChanged(func(themeType ThemeType) {
			callbackCalled = true
			callbackTheme = themeType
		})

		// Change theme to trigger callback
		err := manager.SetTheme(ThemeDark)
		if err != nil {
			t.Fatalf("Failed to set theme: %v", err)
		}

		// Give some time for the callback goroutine
		time.Sleep(10 * time.Millisecond)

		if !callbackCalled {
			t.Error("Theme change callback should have been called")
		}
		if callbackTheme != ThemeDark {
			t.Error("Callback should have received dark theme")
		}
	})

	t.Run("PersistenceIntegration", func(t *testing.T) {
		// Create isolated temporary directory for this test
		testTempDir, err := os.MkdirTemp("", "whisp_theme_test_persistence_*")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(testTempDir)

		// Test that preferences and custom themes are saved and loaded correctly
		manager1 := NewDefaultThemeManager(testTempDir)
		app := test.NewApp()
		manager1.Initialize(app)

		// Create custom theme and set preferences
		customTheme := CustomTheme{
			Name:        "Persistent Theme",
			Description: "A theme that should persist",
			ColorScheme: LightColorScheme,
		}

		manager1.CreateCustomTheme(customTheme)
		manager1.SetTheme(ThemeDark)

		// Create new manager instance to test loading
		manager2 := NewDefaultThemeManager(testTempDir)
		manager2.Initialize(app)

		// Check that preferences were loaded
		if manager2.GetThemeType() != ThemeDark {
			t.Error("Theme preference should be persisted")
		}

		// Check that custom theme was loaded
		themes := manager2.ListCustomThemes()
		if len(themes) != 1 || themes[0].Name != "Persistent Theme" {
			t.Error("Custom theme should be persisted")
		}
	})
}

func TestCustomThemeFromColors(t *testing.T) {
	primary := color.NRGBA{R: 255, G: 0, B: 0, A: 255}
	secondary := color.NRGBA{R: 0, G: 255, B: 0, A: 255}
	background := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	surface := color.NRGBA{R: 248, G: 248, B: 248, A: 255}

	t.Run("LightCustomTheme", func(t *testing.T) {
		customTheme := CreateCustomThemeFromColors(primary, secondary, background, surface, false)

		if customTheme == nil {
			t.Fatal("Theme should not be nil")
		}
		if customTheme.IsDark() {
			t.Error("Theme should not be dark")
		}

		// Test that custom colors are used
		primaryColor := customTheme.Color(theme.ColorNamePrimary, theme.VariantLight)
		if primaryColor != primary {
			t.Error("Primary color should match custom color")
		}
	})

	t.Run("DarkCustomTheme", func(t *testing.T) {
		darkBg := color.NRGBA{R: 33, G: 33, B: 33, A: 255}
		customTheme := CreateCustomThemeFromColors(primary, secondary, darkBg, surface, true)

		if customTheme == nil {
			t.Fatal("Theme should not be nil")
		}
		if !customTheme.IsDark() {
			t.Error("Theme should be dark")
		}
	})
}

func TestColorInterpolation(t *testing.T) {
	from := color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	to := color.NRGBA{R: 255, G: 255, B: 255, A: 255}

	t.Run("InterpolateAtStart", func(t *testing.T) {
		result := InterpolateColor(from, to, 0.0)
		if result != from {
			t.Error("Interpolation at 0.0 should return from color")
		}
	})

	t.Run("InterpolateAtEnd", func(t *testing.T) {
		result := InterpolateColor(from, to, 1.0)
		if result != to {
			t.Error("Interpolation at 1.0 should return to color")
		}
	})

	t.Run("InterpolateAtMiddle", func(t *testing.T) {
		result := InterpolateColor(from, to, 0.5)
		expected := color.NRGBA{R: 127, G: 127, B: 127, A: 255}

		// Allow for small rounding differences
		r, g, b, a := result.RGBA()
		er, eg, eb, ea := expected.RGBA()

		tolerance := uint32(256) // 1 unit in 8-bit space
		if abs(r-er) > tolerance || abs(g-eg) > tolerance || abs(b-eb) > tolerance || abs(a-ea) > tolerance {
			t.Errorf("Interpolation result %v should be approximately %v", result, expected)
		}
	})
}

func TestThemeTransition(t *testing.T) {
	lightTheme := NewLightTheme()
	darkTheme := NewDarkTheme()

	t.Run("CreateThemeTransition", func(t *testing.T) {
		steps := 5
		transition := CreateThemeTransition(lightTheme, darkTheme, steps)

		if len(transition) != steps {
			t.Errorf("Expected %d transition steps, got %d", steps, len(transition))
		}

		// First step should be close to light theme
		firstScheme := transition[0].GetColorScheme()
		lightScheme := lightTheme.GetColorScheme()
		if !colorsApproximatelyEqual(firstScheme.Primary, lightScheme.Primary) {
			t.Error("First transition step should be close to light theme")
		}

		// Last step should be close to dark theme
		lastScheme := transition[steps-1].GetColorScheme()
		darkScheme := darkTheme.GetColorScheme()
		if !colorsApproximatelyEqual(lastScheme.Primary, darkScheme.Primary) {
			t.Error("Last transition step should be close to dark theme")
		}
	})
}

// Helper functions for tests

func abs(x uint32) uint32 {
	return x // uint32 is always >= 0
}

func colorsApproximatelyEqual(c1, c2 color.Color) bool {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	tolerance := uint32(1000) // Allow for some difference

	return abs(r1-r2) <= tolerance &&
		abs(g1-g2) <= tolerance &&
		abs(b1-b2) <= tolerance &&
		abs(a1-a2) <= tolerance
}
