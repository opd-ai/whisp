package theme

import (
	"encoding/json"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
)

// SerializableColor is a JSON-serializable color type
type SerializableColor struct {
	R, G, B, A uint8
}

// RGBA implements the color.Color interface
func (c SerializableColor) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R)
	r |= r << 8
	g = uint32(c.G)
	g |= g << 8
	b = uint32(c.B)
	b |= b << 8
	a = uint32(c.A)
	a |= a << 8
	return r, g, b, a
}

// MarshalJSON implements json.Marshaler
func (c SerializableColor) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]uint8{
		"r": c.R,
		"g": c.G,
		"b": c.B,
		"a": c.A,
	})
}

// UnmarshalJSON implements json.Unmarshaler
func (c *SerializableColor) UnmarshalJSON(data []byte) error {
	var rgba map[string]uint8
	if err := json.Unmarshal(data, &rgba); err != nil {
		return err
	}

	c.R = rgba["r"]
	c.G = rgba["g"]
	c.B = rgba["b"]
	c.A = rgba["a"]
	return nil
}

// ToNRGBA converts to color.NRGBA
func (c SerializableColor) ToNRGBA() color.NRGBA {
	return color.NRGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}

// ToColor converts to color.Color interface
func (c SerializableColor) ToColor() color.Color {
	return c.ToNRGBA()
} // NewSerializableColor creates a SerializableColor from color.Color
func NewSerializableColor(c color.Color) SerializableColor {
	if nrgba, ok := c.(color.NRGBA); ok {
		return SerializableColor{R: nrgba.R, G: nrgba.G, B: nrgba.B, A: nrgba.A}
	}

	// Convert any color.Color to NRGBA
	r, g, b, a := c.RGBA()
	return SerializableColor{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}
}

// NewSerializableColorFromRGBA creates a SerializableColor from RGBA values
func NewSerializableColorFromRGBA(r, g, b, a uint8) SerializableColor {
	return SerializableColor{R: r, G: g, B: b, A: a}
}

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
	Primary   SerializableColor `json:"primary"`
	Secondary SerializableColor `json:"secondary"`
	Success   SerializableColor `json:"success"`
	Warning   SerializableColor `json:"warning"`
	Error     SerializableColor `json:"error"`
	Info      SerializableColor `json:"info"`

	// Background colors
	Background     SerializableColor `json:"background"`
	Surface        SerializableColor `json:"surface"`
	SurfaceVariant SerializableColor `json:"surface_variant"`

	// Text colors
	OnPrimary    SerializableColor `json:"on_primary"`
	OnSecondary  SerializableColor `json:"on_secondary"`
	OnBackground SerializableColor `json:"on_background"`
	OnSurface    SerializableColor `json:"on_surface"`
	OnError      SerializableColor `json:"on_error"`

	// Chat-specific colors
	MessageSent      SerializableColor `json:"message_sent"`
	MessageReceived  SerializableColor `json:"message_received"`
	MessageText      SerializableColor `json:"message_text"`
	MessageTime      SerializableColor `json:"message_time"`
	OnlineIndicator  SerializableColor `json:"online_indicator"`
	OfflineIndicator SerializableColor `json:"offline_indicator"`

	// UI element colors
	Border    SerializableColor `json:"border"`
	Divider   SerializableColor `json:"divider"`
	Highlight SerializableColor `json:"highlight"`
	Disabled  SerializableColor `json:"disabled"`
	Shadow    SerializableColor `json:"shadow"`
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
		Primary:   NewSerializableColorFromRGBA(25, 118, 210, 255), // Blue
		Secondary: NewSerializableColorFromRGBA(156, 39, 176, 255), // Purple
		Success:   NewSerializableColorFromRGBA(76, 175, 80, 255),  // Green
		Warning:   NewSerializableColorFromRGBA(255, 152, 0, 255),  // Orange
		Error:     NewSerializableColorFromRGBA(244, 67, 54, 255),  // Red
		Info:      NewSerializableColorFromRGBA(33, 150, 243, 255), // Light Blue

		Background:     NewSerializableColorFromRGBA(250, 250, 250, 255), // Very Light Gray
		Surface:        NewSerializableColorFromRGBA(255, 255, 255, 255), // White
		SurfaceVariant: NewSerializableColorFromRGBA(245, 245, 245, 255), // Light Gray

		OnPrimary:    NewSerializableColorFromRGBA(255, 255, 255, 255), // White
		OnSecondary:  NewSerializableColorFromRGBA(255, 255, 255, 255), // White
		OnBackground: NewSerializableColorFromRGBA(33, 33, 33, 255),    // Dark Gray
		OnSurface:    NewSerializableColorFromRGBA(33, 33, 33, 255),    // Dark Gray
		OnError:      NewSerializableColorFromRGBA(255, 255, 255, 255), // White

		MessageSent:      NewSerializableColorFromRGBA(227, 242, 253, 255), // Light Blue
		MessageReceived:  NewSerializableColorFromRGBA(245, 245, 245, 255), // Light Gray
		MessageText:      NewSerializableColorFromRGBA(33, 33, 33, 255),    // Dark Gray
		MessageTime:      NewSerializableColorFromRGBA(117, 117, 117, 255), // Medium Gray
		OnlineIndicator:  NewSerializableColorFromRGBA(76, 175, 80, 255),   // Green
		OfflineIndicator: NewSerializableColorFromRGBA(158, 158, 158, 255), // Gray

		Border:    NewSerializableColorFromRGBA(224, 224, 224, 255), // Light Border
		Divider:   NewSerializableColorFromRGBA(238, 238, 238, 255), // Very Light Gray
		Highlight: NewSerializableColorFromRGBA(25, 118, 210, 50),   // Transparent Blue
		Disabled:  NewSerializableColorFromRGBA(189, 189, 189, 255), // Medium Gray
		Shadow:    NewSerializableColorFromRGBA(0, 0, 0, 20),        // Transparent Black
	}

	// Dark theme color scheme
	DarkColorScheme = ColorScheme{
		Primary:   NewSerializableColorFromRGBA(144, 202, 249, 255), // Light Blue
		Secondary: NewSerializableColorFromRGBA(206, 147, 216, 255), // Light Purple
		Success:   NewSerializableColorFromRGBA(129, 199, 132, 255), // Light Green
		Warning:   NewSerializableColorFromRGBA(255, 183, 77, 255),  // Light Orange
		Error:     NewSerializableColorFromRGBA(239, 83, 80, 255),   // Light Red
		Info:      NewSerializableColorFromRGBA(79, 195, 247, 255),  // Cyan

		Background:     NewSerializableColorFromRGBA(18, 18, 18, 255), // Very Dark Gray
		Surface:        NewSerializableColorFromRGBA(30, 30, 30, 255), // Dark Gray
		SurfaceVariant: NewSerializableColorFromRGBA(42, 42, 42, 255), // Medium Dark Gray

		OnPrimary:    NewSerializableColorFromRGBA(0, 0, 0, 255),       // Black
		OnSecondary:  NewSerializableColorFromRGBA(0, 0, 0, 255),       // Black
		OnBackground: NewSerializableColorFromRGBA(224, 224, 224, 255), // Light Gray
		OnSurface:    NewSerializableColorFromRGBA(224, 224, 224, 255), // Light Gray
		OnError:      NewSerializableColorFromRGBA(0, 0, 0, 255),       // Black

		MessageSent:      NewSerializableColorFromRGBA(37, 47, 62, 255),    // Dark Blue
		MessageReceived:  NewSerializableColorFromRGBA(42, 42, 42, 255),    // Medium Dark Gray
		MessageText:      NewSerializableColorFromRGBA(224, 224, 224, 255), // Light Gray
		MessageTime:      NewSerializableColorFromRGBA(158, 158, 158, 255), // Medium Gray
		OnlineIndicator:  NewSerializableColorFromRGBA(129, 199, 132, 255), // Light Green
		OfflineIndicator: NewSerializableColorFromRGBA(117, 117, 117, 255), // Dark Gray

		Border:    NewSerializableColorFromRGBA(66, 66, 66, 255),   // Dark Border
		Divider:   NewSerializableColorFromRGBA(54, 54, 54, 255),   // Dark Divider
		Highlight: NewSerializableColorFromRGBA(144, 202, 249, 50), // Transparent Light Blue
		Disabled:  NewSerializableColorFromRGBA(97, 97, 97, 255),   // Dark Medium Gray
		Shadow:    NewSerializableColorFromRGBA(0, 0, 0, 80),       // Darker Shadow
	}
)
