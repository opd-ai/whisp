package theme

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
)

// ThemeType represents different theme modes
type ThemeType int

const (
	ThemeLight ThemeType = iota
	ThemeDark
	ThemeSystem // Follows system preference
	ThemeCustom // User-defined custom theme
)

// String returns string representation of theme type
func (t ThemeType) String() string {
	switch t {
	case ThemeLight:
		return "light"
	case ThemeDark:
		return "dark"
	case ThemeSystem:
		return "system"
	case ThemeCustom:
		return "custom"
	default:
		return "unknown"
	}
}

// ParseThemeType parses theme type from string
func ParseThemeType(s string) ThemeType {
	switch s {
	case "light":
		return ThemeLight
	case "dark":
		return ThemeDark
	case "system":
		return ThemeSystem
	case "custom":
		return ThemeCustom
	default:
		return ThemeSystem // Default to system
	}
}

// ColorScheme defines a complete color scheme for the application
type ColorScheme struct {
	// Basic colors
	Primary   color.Color
	Secondary color.Color
	Success   color.Color
	Warning   color.Color
	Error     color.Color
	Info      color.Color

	// Background colors
	Background     color.Color
	Surface        color.Color
	SurfaceVariant color.Color

	// Text colors
	OnPrimary    color.Color
	OnSecondary  color.Color
	OnBackground color.Color
	OnSurface    color.Color
	OnError      color.Color

	// Chat-specific colors
	MessageSent      color.Color
	MessageReceived  color.Color
	MessageText      color.Color
	MessageTime      color.Color
	OnlineIndicator  color.Color
	OfflineIndicator color.Color

	// UI element colors
	Border    color.Color
	Divider   color.Color
	Highlight color.Color
	Disabled  color.Color
	Shadow    color.Color
}

// CustomTheme represents a user-defined theme
type CustomTheme struct {
	Name        string      `json:"name" yaml:"name"`
	Description string      `json:"description" yaml:"description"`
	ColorScheme ColorScheme `json:"color_scheme" yaml:"color_scheme"`
	CreatedAt   time.Time   `json:"created_at" yaml:"created_at"`
	ModifiedAt  time.Time   `json:"modified_at" yaml:"modified_at"`
}

// ThemePreferences stores user theme preferences
type ThemePreferences struct {
	ThemeType       ThemeType `json:"theme_type" yaml:"theme_type"`
	CustomThemeName string    `json:"custom_theme_name,omitempty" yaml:"custom_theme_name,omitempty"`

	// System theme detection settings
	FollowSystemTheme bool `json:"follow_system_theme" yaml:"follow_system_theme"`

	// Auto theme switching
	AutoSwitchEnabled bool      `json:"auto_switch_enabled" yaml:"auto_switch_enabled"`
	LightThemeStart   time.Time `json:"light_theme_start,omitempty" yaml:"light_theme_start,omitempty"`
	DarkThemeStart    time.Time `json:"dark_theme_start,omitempty" yaml:"dark_theme_start,omitempty"`
}

// ThemeManager interface for managing themes
type ThemeManager interface {
	// Theme management
	GetCurrentTheme() fyne.Theme
	SetTheme(themeType ThemeType) error
	GetThemeType() ThemeType

	// Custom themes
	CreateCustomTheme(theme CustomTheme) error
	GetCustomTheme(name string) (*CustomTheme, error)
	ListCustomThemes() []CustomTheme
	UpdateCustomTheme(theme CustomTheme) error
	DeleteCustomTheme(name string) error

	// System theme detection
	DetectSystemTheme() ThemeType
	EnableSystemThemeFollowing(enabled bool)

	// Auto switching
	EnableAutoSwitch(enabled bool, lightStart, darkStart time.Time)
	CheckAutoSwitch() // Check if theme should be switched based on time

	// Preferences
	GetPreferences() ThemePreferences
	SetPreferences(prefs ThemePreferences) error

	// Theme application
	ApplyTheme(app fyne.App) error

	// Events
	OnThemeChanged(callback func(ThemeType))
}

// ThemeChangeEvent represents a theme change event
type ThemeChangeEvent struct {
	OldTheme ThemeType
	NewTheme ThemeType
	Reason   string // "user", "system", "auto_switch"
	Time     time.Time
}

// DefaultColorSchemes provides default color schemes
var (
	// Light theme color scheme
	LightColorScheme = ColorScheme{
		Primary:   color.NRGBA{R: 25, G: 118, B: 210, A: 255}, // Blue
		Secondary: color.NRGBA{R: 156, G: 39, B: 176, A: 255}, // Purple
		Success:   color.NRGBA{R: 76, G: 175, B: 80, A: 255},  // Green
		Warning:   color.NRGBA{R: 255, G: 152, B: 0, A: 255},  // Orange
		Error:     color.NRGBA{R: 244, G: 67, B: 54, A: 255},  // Red
		Info:      color.NRGBA{R: 33, G: 150, B: 243, A: 255}, // Light Blue

		Background:     color.NRGBA{R: 250, G: 250, B: 250, A: 255}, // Very Light Gray
		Surface:        color.NRGBA{R: 255, G: 255, B: 255, A: 255}, // White
		SurfaceVariant: color.NRGBA{R: 245, G: 245, B: 245, A: 255}, // Light Gray

		OnPrimary:    color.NRGBA{R: 255, G: 255, B: 255, A: 255}, // White
		OnSecondary:  color.NRGBA{R: 255, G: 255, B: 255, A: 255}, // White
		OnBackground: color.NRGBA{R: 33, G: 33, B: 33, A: 255},    // Dark Gray
		OnSurface:    color.NRGBA{R: 33, G: 33, B: 33, A: 255},    // Dark Gray
		OnError:      color.NRGBA{R: 255, G: 255, B: 255, A: 255}, // White

		MessageSent:      color.NRGBA{R: 227, G: 242, B: 253, A: 255}, // Light Blue
		MessageReceived:  color.NRGBA{R: 245, G: 245, B: 245, A: 255}, // Light Gray
		MessageText:      color.NRGBA{R: 33, G: 33, B: 33, A: 255},    // Dark Gray
		MessageTime:      color.NRGBA{R: 117, G: 117, B: 117, A: 255}, // Medium Gray
		OnlineIndicator:  color.NRGBA{R: 76, G: 175, B: 80, A: 255},   // Green
		OfflineIndicator: color.NRGBA{R: 158, G: 158, B: 158, A: 255}, // Gray

		Border:    color.NRGBA{R: 224, G: 224, B: 224, A: 255}, // Light Border
		Divider:   color.NRGBA{R: 238, G: 238, B: 238, A: 255}, // Very Light Gray
		Highlight: color.NRGBA{R: 25, G: 118, B: 210, A: 50},   // Transparent Blue
		Disabled:  color.NRGBA{R: 189, G: 189, B: 189, A: 255}, // Medium Gray
		Shadow:    color.NRGBA{R: 0, G: 0, B: 0, A: 20},        // Transparent Black
	}

	// Dark theme color scheme
	DarkColorScheme = ColorScheme{
		Primary:   color.NRGBA{R: 144, G: 202, B: 249, A: 255}, // Light Blue
		Secondary: color.NRGBA{R: 206, G: 147, B: 216, A: 255}, // Light Purple
		Success:   color.NRGBA{R: 129, G: 199, B: 132, A: 255}, // Light Green
		Warning:   color.NRGBA{R: 255, G: 183, B: 77, A: 255},  // Light Orange
		Error:     color.NRGBA{R: 239, G: 83, B: 80, A: 255},   // Light Red
		Info:      color.NRGBA{R: 79, G: 195, B: 247, A: 255},  // Cyan

		Background:     color.NRGBA{R: 18, G: 18, B: 18, A: 255}, // Very Dark Gray
		Surface:        color.NRGBA{R: 30, G: 30, B: 30, A: 255}, // Dark Gray
		SurfaceVariant: color.NRGBA{R: 42, G: 42, B: 42, A: 255}, // Medium Dark Gray

		OnPrimary:    color.NRGBA{R: 0, G: 0, B: 0, A: 255},       // Black
		OnSecondary:  color.NRGBA{R: 0, G: 0, B: 0, A: 255},       // Black
		OnBackground: color.NRGBA{R: 224, G: 224, B: 224, A: 255}, // Light Gray
		OnSurface:    color.NRGBA{R: 224, G: 224, B: 224, A: 255}, // Light Gray
		OnError:      color.NRGBA{R: 0, G: 0, B: 0, A: 255},       // Black

		MessageSent:      color.NRGBA{R: 37, G: 47, B: 62, A: 255},    // Dark Blue
		MessageReceived:  color.NRGBA{R: 42, G: 42, B: 42, A: 255},    // Medium Dark Gray
		MessageText:      color.NRGBA{R: 224, G: 224, B: 224, A: 255}, // Light Gray
		MessageTime:      color.NRGBA{R: 158, G: 158, B: 158, A: 255}, // Medium Gray
		OnlineIndicator:  color.NRGBA{R: 129, G: 199, B: 132, A: 255}, // Light Green
		OfflineIndicator: color.NRGBA{R: 117, G: 117, B: 117, A: 255}, // Dark Gray

		Border:    color.NRGBA{R: 66, G: 66, B: 66, A: 255},   // Dark Border
		Divider:   color.NRGBA{R: 54, G: 54, B: 54, A: 255},   // Dark Divider
		Highlight: color.NRGBA{R: 144, G: 202, B: 249, A: 50}, // Transparent Light Blue
		Disabled:  color.NRGBA{R: 97, G: 97, B: 97, A: 255},   // Dark Medium Gray
		Shadow:    color.NRGBA{R: 0, G: 0, B: 0, A: 80},       // Darker Shadow
	}
)
