package theme

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"fyne.io/fyne/v2"
)

// DefaultThemeManager implements ThemeManager interface
type DefaultThemeManager struct {
	mu               sync.RWMutex
	currentTheme     *WhispTheme
	currentThemeType ThemeType
	preferences      ThemePreferences
	customThemes     map[string]CustomTheme
	systemDetector   *SystemThemeDetector
	app              fyne.App
	configDir        string
	changeCallbacks  []func(ThemeType)
	autoSwitchTimer  *time.Timer
}

// NewDefaultThemeManager creates a new default theme manager
func NewDefaultThemeManager(configDir string) *DefaultThemeManager {
	return &DefaultThemeManager{
		currentTheme:     NewLightTheme(),
		currentThemeType: ThemeSystem, // Start with system theme
		preferences: ThemePreferences{
			ThemeType:         ThemeSystem,
			FollowSystemTheme: true,
		},
		customThemes:   make(map[string]CustomTheme),
		systemDetector: NewSystemThemeDetector(),
		configDir:      configDir,
	}
}

// Initialize initializes the theme manager
func (tm *DefaultThemeManager) Initialize(app fyne.App) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.app = app

	// Load preferences and custom themes from disk
	if err := tm.loadPreferences(); err != nil {
		return fmt.Errorf("failed to load theme preferences: %w", err)
	}

	if err := tm.loadCustomThemes(); err != nil {
		return fmt.Errorf("failed to load custom themes: %w", err)
	}

	// Apply initial theme based on preferences
	if err := tm.applyThemeBasedOnPreferences(); err != nil {
		return fmt.Errorf("failed to apply initial theme: %w", err)
	}

	// Start auto switch timer if enabled
	if tm.preferences.AutoSwitchEnabled {
		tm.startAutoSwitchTimer()
	}

	return nil
}

// GetCurrentTheme returns the current Fyne theme
func (tm *DefaultThemeManager) GetCurrentTheme() fyne.Theme {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.currentTheme
}

// SetTheme sets the theme type
func (tm *DefaultThemeManager) SetTheme(themeType ThemeType) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	oldTheme := tm.currentThemeType
	tm.currentThemeType = themeType
	tm.preferences.ThemeType = themeType

	if err := tm.updateCurrentTheme(); err != nil {
		// Revert on error
		tm.currentThemeType = oldTheme
		tm.preferences.ThemeType = oldTheme
		return fmt.Errorf("failed to update theme: %w", err)
	}

	// Apply theme to app
	if tm.app != nil {
		tm.app.Settings().SetTheme(tm.currentTheme)
	}

	// Save preferences
	if err := tm.savePreferences(); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to save theme preferences: %v\n", err)
	}

	// Notify callbacks
	tm.notifyThemeChange(oldTheme, themeType, "user")

	return nil
}

// GetThemeType returns the current theme type
func (tm *DefaultThemeManager) GetThemeType() ThemeType {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.currentThemeType
}

// CreateCustomTheme creates a new custom theme
func (tm *DefaultThemeManager) CreateCustomTheme(theme CustomTheme) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if theme.Name == "" {
		return fmt.Errorf("custom theme name cannot be empty")
	}

	theme.CreatedAt = time.Now()
	theme.ModifiedAt = time.Now()

	tm.customThemes[theme.Name] = theme

	return tm.saveCustomThemes()
}

// GetCustomTheme retrieves a custom theme by name
func (tm *DefaultThemeManager) GetCustomTheme(name string) (*CustomTheme, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	theme, exists := tm.customThemes[name]
	if !exists {
		return nil, fmt.Errorf("custom theme '%s' not found", name)
	}

	return &theme, nil
}

// ListCustomThemes returns all custom themes
func (tm *DefaultThemeManager) ListCustomThemes() []CustomTheme {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	themes := make([]CustomTheme, 0, len(tm.customThemes))
	for _, theme := range tm.customThemes {
		themes = append(themes, theme)
	}

	return themes
}

// UpdateCustomTheme updates an existing custom theme
func (tm *DefaultThemeManager) UpdateCustomTheme(theme CustomTheme) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.customThemes[theme.Name]; !exists {
		return fmt.Errorf("custom theme '%s' not found", theme.Name)
	}

	// Preserve creation time
	if existing, exists := tm.customThemes[theme.Name]; exists {
		theme.CreatedAt = existing.CreatedAt
	}
	theme.ModifiedAt = time.Now()

	tm.customThemes[theme.Name] = theme

	// If this is the currently active custom theme, update it
	if tm.currentThemeType == ThemeCustom && tm.preferences.CustomThemeName == theme.Name {
		if err := tm.updateCurrentTheme(); err != nil {
			return fmt.Errorf("failed to apply updated theme: %w", err)
		}
		if tm.app != nil {
			tm.app.Settings().SetTheme(tm.currentTheme)
		}
	}

	return tm.saveCustomThemes()
}

// DeleteCustomTheme deletes a custom theme
func (tm *DefaultThemeManager) DeleteCustomTheme(name string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.customThemes[name]; !exists {
		return fmt.Errorf("custom theme '%s' not found", name)
	}

	// If this is the currently active theme, switch to system theme
	if tm.currentThemeType == ThemeCustom && tm.preferences.CustomThemeName == name {
		tm.currentThemeType = ThemeSystem
		tm.preferences.ThemeType = ThemeSystem
		tm.preferences.CustomThemeName = ""

		if err := tm.updateCurrentTheme(); err != nil {
			return fmt.Errorf("failed to switch theme after deletion: %w", err)
		}
		if tm.app != nil {
			tm.app.Settings().SetTheme(tm.currentTheme)
		}
	}

	delete(tm.customThemes, name)

	return tm.saveCustomThemes()
}

// DetectSystemTheme detects the current system theme
func (tm *DefaultThemeManager) DetectSystemTheme() ThemeType {
	return tm.systemDetector.DetectSystemTheme()
}

// EnableSystemThemeFollowing enables or disables system theme following
func (tm *DefaultThemeManager) EnableSystemThemeFollowing(enabled bool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.preferences.FollowSystemTheme = enabled

	// If enabled and current theme is system, update to actual system theme
	if enabled && tm.currentThemeType == ThemeSystem {
		systemTheme := tm.systemDetector.DetectSystemTheme()
		if systemTheme != tm.currentThemeType {
			oldTheme := tm.currentThemeType
			tm.updateCurrentTheme()
			if tm.app != nil {
				tm.app.Settings().SetTheme(tm.currentTheme)
			}
			tm.notifyThemeChange(oldTheme, tm.currentThemeType, "system")
		}
	}

	tm.savePreferences()
}

// EnableAutoSwitch enables or disables automatic theme switching
func (tm *DefaultThemeManager) EnableAutoSwitch(enabled bool, lightStart, darkStart time.Time) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.preferences.AutoSwitchEnabled = enabled
	tm.preferences.LightThemeStart = lightStart
	tm.preferences.DarkThemeStart = darkStart

	if enabled {
		tm.startAutoSwitchTimer()
	} else {
		tm.stopAutoSwitchTimer()
	}

	tm.savePreferences()
}

// CheckAutoSwitch checks if theme should be switched based on time
func (tm *DefaultThemeManager) CheckAutoSwitch() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if !tm.preferences.AutoSwitchEnabled {
		return
	}

	now := time.Now()
	currentTime := time.Date(0, 1, 1, now.Hour(), now.Minute(), now.Second(), 0, time.UTC)

	lightStart := time.Date(0, 1, 1,
		tm.preferences.LightThemeStart.Hour(),
		tm.preferences.LightThemeStart.Minute(),
		tm.preferences.LightThemeStart.Second(), 0, time.UTC)

	darkStart := time.Date(0, 1, 1,
		tm.preferences.DarkThemeStart.Hour(),
		tm.preferences.DarkThemeStart.Minute(),
		tm.preferences.DarkThemeStart.Second(), 0, time.UTC)

	var targetTheme ThemeType

	// Determine which theme should be active
	if lightStart.Before(darkStart) {
		// Normal case: light theme during day, dark at night
		if currentTime.After(lightStart) && currentTime.Before(darkStart) {
			targetTheme = ThemeLight
		} else {
			targetTheme = ThemeDark
		}
	} else {
		// Inverted case: light theme starts after dark theme (e.g., light at 6 PM, dark at 6 AM)
		if currentTime.After(lightStart) || currentTime.Before(darkStart) {
			targetTheme = ThemeLight
		} else {
			targetTheme = ThemeDark
		}
	}

	// Switch theme if needed
	if targetTheme != tm.currentThemeType {
		oldTheme := tm.currentThemeType
		tm.currentThemeType = targetTheme
		tm.updateCurrentTheme()
		if tm.app != nil {
			tm.app.Settings().SetTheme(tm.currentTheme)
		}
		tm.notifyThemeChange(oldTheme, targetTheme, "auto_switch")
	}
}

// GetPreferences returns current theme preferences
func (tm *DefaultThemeManager) GetPreferences() ThemePreferences {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.preferences
}

// SetPreferences sets theme preferences
func (tm *DefaultThemeManager) SetPreferences(prefs ThemePreferences) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	oldTheme := tm.currentThemeType
	tm.preferences = prefs
	tm.currentThemeType = prefs.ThemeType

	if err := tm.updateCurrentTheme(); err != nil {
		tm.currentThemeType = oldTheme
		return fmt.Errorf("failed to apply new preferences: %w", err)
	}

	if tm.app != nil {
		tm.app.Settings().SetTheme(tm.currentTheme)
	}

	// Update auto switch timer
	if prefs.AutoSwitchEnabled {
		tm.startAutoSwitchTimer()
	} else {
		tm.stopAutoSwitchTimer()
	}

	tm.notifyThemeChange(oldTheme, tm.currentThemeType, "preferences")

	return tm.savePreferences()
}

// ApplyTheme applies the current theme to the app
func (tm *DefaultThemeManager) ApplyTheme(app fyne.App) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.app = app
	app.Settings().SetTheme(tm.currentTheme)

	return nil
}

// OnThemeChanged registers a callback for theme changes
func (tm *DefaultThemeManager) OnThemeChanged(callback func(ThemeType)) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.changeCallbacks = append(tm.changeCallbacks, callback)
}

// Private helper methods

func (tm *DefaultThemeManager) updateCurrentTheme() error {
	switch tm.currentThemeType {
	case ThemeLight:
		tm.currentTheme = NewLightTheme()
	case ThemeDark:
		tm.currentTheme = NewDarkTheme()
	case ThemeSystem:
		systemTheme := tm.systemDetector.DetectSystemTheme()
		if systemTheme == ThemeDark {
			tm.currentTheme = NewDarkTheme()
		} else {
			tm.currentTheme = NewLightTheme()
		}
	case ThemeCustom:
		if tm.preferences.CustomThemeName == "" {
			return fmt.Errorf("custom theme name not specified")
		}
		customTheme, exists := tm.customThemes[tm.preferences.CustomThemeName]
		if !exists {
			return fmt.Errorf("custom theme '%s' not found", tm.preferences.CustomThemeName)
		}
		tm.currentTheme = NewWhispTheme(customTheme.ColorScheme, tm.isDarkScheme(customTheme.ColorScheme))
	default:
		return fmt.Errorf("unknown theme type: %v", tm.currentThemeType)
	}

	return nil
}

func (tm *DefaultThemeManager) isDarkScheme(scheme ColorScheme) bool {
	// Simple heuristic: if background is darker than foreground, it's a dark theme
	r1, g1, b1, _ := scheme.Background.RGBA()
	r2, g2, b2, _ := scheme.OnBackground.RGBA()

	bgBrightness := (r1 + g1 + b1) / 3
	fgBrightness := (r2 + g2 + b2) / 3

	return bgBrightness < fgBrightness
}

func (tm *DefaultThemeManager) applyThemeBasedOnPreferences() error {
	return tm.updateCurrentTheme()
}

func (tm *DefaultThemeManager) notifyThemeChange(oldTheme, newTheme ThemeType, reason string) {
	for _, callback := range tm.changeCallbacks {
		go callback(newTheme) // Run callbacks in goroutines to avoid blocking
	}
}

func (tm *DefaultThemeManager) startAutoSwitchTimer() {
	tm.stopAutoSwitchTimer() // Stop existing timer if any

	// Calculate next switch time
	nextSwitch := tm.calculateNextSwitchTime()
	duration := time.Until(nextSwitch)

	tm.autoSwitchTimer = time.AfterFunc(duration, func() {
		tm.CheckAutoSwitch()
		tm.startAutoSwitchTimer() // Reschedule for next switch
	})
}

func (tm *DefaultThemeManager) stopAutoSwitchTimer() {
	if tm.autoSwitchTimer != nil {
		tm.autoSwitchTimer.Stop()
		tm.autoSwitchTimer = nil
	}
}

func (tm *DefaultThemeManager) calculateNextSwitchTime() time.Time {
	now := time.Now()

	// Create today's switch times
	lightToday := time.Date(now.Year(), now.Month(), now.Day(),
		tm.preferences.LightThemeStart.Hour(),
		tm.preferences.LightThemeStart.Minute(),
		tm.preferences.LightThemeStart.Second(), 0, now.Location())

	darkToday := time.Date(now.Year(), now.Month(), now.Day(),
		tm.preferences.DarkThemeStart.Hour(),
		tm.preferences.DarkThemeStart.Minute(),
		tm.preferences.DarkThemeStart.Second(), 0, now.Location())

	// Find next switch time
	if now.Before(lightToday) {
		return lightToday
	} else if now.Before(darkToday) {
		return darkToday
	} else {
		// Both times have passed today, return tomorrow's light theme start
		return lightToday.AddDate(0, 0, 1)
	}
}

func (tm *DefaultThemeManager) getConfigFilePath(filename string) string {
	return filepath.Join(tm.configDir, filename)
}

func (tm *DefaultThemeManager) loadPreferences() error {
	path := tm.getConfigFilePath("theme_preferences.json")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, use defaults
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &tm.preferences)
}

func (tm *DefaultThemeManager) savePreferences() error {
	// Ensure config directory exists
	if err := os.MkdirAll(tm.configDir, 0755); err != nil {
		return err
	}

	path := tm.getConfigFilePath("theme_preferences.json")
	data, err := json.MarshalIndent(tm.preferences, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (tm *DefaultThemeManager) loadCustomThemes() error {
	path := tm.getConfigFilePath("custom_themes.json")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, no custom themes
			return nil
		}
		return err
	}

	var themes map[string]CustomTheme
	if err := json.Unmarshal(data, &themes); err != nil {
		return err
	}

	tm.customThemes = themes
	return nil
}

func (tm *DefaultThemeManager) saveCustomThemes() error {
	// Ensure config directory exists
	if err := os.MkdirAll(tm.configDir, 0755); err != nil {
		return err
	}

	path := tm.getConfigFilePath("custom_themes.json")
	data, err := json.MarshalIndent(tm.customThemes, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
