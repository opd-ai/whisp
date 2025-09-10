package theme

import (
	"image/color"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// WhispTheme implements fyne.Theme interface for Whisp
type WhispTheme struct {
	scheme  ColorScheme
	isDark  bool
	variant fyne.ThemeVariant
}

// NewWhispTheme creates a new Whisp theme with the given color scheme
func NewWhispTheme(scheme ColorScheme, isDark bool) *WhispTheme {
	variant := theme.VariantLight
	if isDark {
		variant = theme.VariantDark
	}

	return &WhispTheme{
		scheme:  scheme,
		isDark:  isDark,
		variant: variant,
	}
}

// NewLightTheme creates a new light theme
func NewLightTheme() *WhispTheme {
	return NewWhispTheme(LightColorScheme, false)
}

// NewDarkTheme creates a new dark theme
func NewDarkTheme() *WhispTheme {
	return NewWhispTheme(DarkColorScheme, true)
}

// Color returns the theme's color for the specified ColorName
func (t *WhispTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Handle variant-specific colors
	if variant == theme.VariantLight && t.isDark {
		// Use light theme for light variant
		return NewLightTheme().Color(name, variant)
	}
	if variant == theme.VariantDark && !t.isDark {
		// Use dark theme for dark variant
		return NewDarkTheme().Color(name, variant)
	}

	switch name {
	// Primary colors
	case theme.ColorNamePrimary:
		return t.scheme.Primary
	case theme.ColorNameBackground:
		return t.scheme.Background
	case theme.ColorNameForeground:
		return t.scheme.OnBackground

	// Surface colors
	case theme.ColorNameButton:
		return t.scheme.Surface
	case theme.ColorNameDisabledButton:
		return t.scheme.Disabled
	case theme.ColorNameMenuBackground:
		return t.scheme.Surface
	case theme.ColorNameOverlayBackground:
		return t.scheme.SurfaceVariant

	// Text colors
	case theme.ColorNameDisabled:
		return t.scheme.Disabled
	case theme.ColorNameError:
		return t.scheme.Error
	case theme.ColorNameFocus:
		return t.scheme.Primary
	case theme.ColorNameHover:
		return t.scheme.Highlight.ToColor()
	case theme.ColorNamePlaceHolder:
		return t.scheme.Disabled.ToColor()
	case theme.ColorNamePressed:
		return t.scheme.Primary.ToColor()
	case theme.ColorNameSelection:
		return t.scheme.Highlight.ToColor()

	// Input colors
	case theme.ColorNameInputBackground:
		return t.scheme.Surface.ToColor()
	case theme.ColorNameInputBorder:
		return t.scheme.Border.ToColor()

	// Header colors
	case theme.ColorNameHeaderBackground:
		return t.scheme.Primary.ToColor()

	// Separator colors
	case theme.ColorNameSeparator:
		return t.scheme.Divider.ToColor()

	// Scrollbar colors
	case theme.ColorNameScrollBar:
		return t.scheme.Border.ToColor()

	// Shadow colors
	case theme.ColorNameShadow:
		return t.scheme.Shadow.ToColor()

	// Success and warning colors
	case theme.ColorNameSuccess:
		return t.scheme.Success.ToColor()
	case theme.ColorNameWarning:
		return t.scheme.Warning.ToColor()

	// Hyperlink colors
	case theme.ColorNameHyperlink:
		return t.scheme.Primary.ToColor()

	default:
		// Fallback to Fyne's default theme
		if t.isDark {
			return theme.DarkTheme().Color(name, variant)
		}
		return theme.DefaultTheme().Color(name, variant)
	}
}

// Font returns the theme's font for the specified style and variant
func (t *WhispTheme) Font(style fyne.TextStyle) fyne.Resource {
	// Use Fyne's default fonts
	if t.isDark {
		return theme.DarkTheme().Font(style)
	}
	return theme.DefaultTheme().Font(style)
}

// Icon returns the theme's icon for the specified IconName and variant
func (t *WhispTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	// Use Fyne's default icons
	if t.isDark {
		return theme.DarkTheme().Icon(name)
	}
	return theme.DefaultTheme().Icon(name)
}

// Size returns the theme's size for the specified SizeName
func (t *WhispTheme) Size(name fyne.ThemeSizeName) float32 {
	// Use Fyne's default sizes
	if t.isDark {
		return theme.DarkTheme().Size(name)
	}
	return theme.DefaultTheme().Size(name)
}

// GetColorScheme returns the theme's color scheme
func (t *WhispTheme) GetColorScheme() ColorScheme {
	return t.scheme
}

// IsDark returns whether this is a dark theme
func (t *WhispTheme) IsDark() bool {
	return t.isDark
}

// GetVariant returns the theme variant
func (t *WhispTheme) GetVariant() fyne.ThemeVariant {
	return t.variant
}

// SystemThemeDetector provides system theme detection capabilities
type SystemThemeDetector struct{}

// NewSystemThemeDetector creates a new system theme detector
func NewSystemThemeDetector() *SystemThemeDetector {
	return &SystemThemeDetector{}
}

// DetectSystemTheme detects the current system theme preference
func (d *SystemThemeDetector) DetectSystemTheme() ThemeType {
	// Platform-specific system theme detection
	switch runtime.GOOS {
	case "windows":
		return d.detectWindowsTheme()
	case "darwin":
		return d.detectMacOSTheme()
	case "linux":
		return d.detectLinuxTheme()
	default:
		// Default to light theme for unknown platforms
		return ThemeLight
	}
}

// detectWindowsTheme detects Windows system theme
func (d *SystemThemeDetector) detectWindowsTheme() ThemeType {
	// For now, return light theme
	// In a real implementation, this would check Windows registry
	// or use system APIs to detect the actual theme preference
	return ThemeLight
}

// detectMacOSTheme detects macOS system theme
func (d *SystemThemeDetector) detectMacOSTheme() ThemeType {
	// For now, return light theme
	// In a real implementation, this would use macOS APIs
	// to check the system appearance preference
	return ThemeLight
}

// detectLinuxTheme detects Linux system theme
func (d *SystemThemeDetector) detectLinuxTheme() ThemeType {
	// For now, return light theme
	// In a real implementation, this would check desktop environment
	// settings (GNOME, KDE, etc.) to detect the theme preference
	return ThemeLight
}

// Helper functions for creating theme variants

// CreateCustomThemeFromColors creates a custom theme from individual colors
func CreateCustomThemeFromColors(
	primary, secondary, background, surface color.Color,
	isDark bool,
) *WhispTheme {
	scheme := ColorScheme{
		Primary:    NewSerializableColor(primary),
		Secondary:  NewSerializableColor(secondary),
		Background: NewSerializableColor(background),
		Surface:    NewSerializableColor(surface),
		// Set reasonable defaults for other colors
		Success: NewSerializableColorFromRGBA(76, 175, 80, 255),
		Warning: NewSerializableColorFromRGBA(255, 152, 0, 255),
		Error:   NewSerializableColorFromRGBA(244, 67, 54, 255),
		Info:    NewSerializableColorFromRGBA(33, 150, 243, 255),
	}

	// Set text colors based on whether it's a dark theme
	if isDark {
		scheme.OnPrimary = NewSerializableColorFromRGBA(0, 0, 0, 255)
		scheme.OnSecondary = NewSerializableColorFromRGBA(0, 0, 0, 255)
		scheme.OnBackground = NewSerializableColorFromRGBA(255, 255, 255, 255)
		scheme.OnSurface = NewSerializableColorFromRGBA(255, 255, 255, 255)
		scheme.OnError = NewSerializableColorFromRGBA(0, 0, 0, 255)
	} else {
		scheme.OnPrimary = NewSerializableColorFromRGBA(255, 255, 255, 255)
		scheme.OnSecondary = NewSerializableColorFromRGBA(255, 255, 255, 255)
		scheme.OnBackground = NewSerializableColorFromRGBA(0, 0, 0, 255)
		scheme.OnSurface = NewSerializableColorFromRGBA(0, 0, 0, 255)
		scheme.OnError = NewSerializableColorFromRGBA(255, 255, 255, 255)
	}

	return NewWhispTheme(scheme, isDark)
}

// InterpolateColor interpolates between two colors
func InterpolateColor(from, to color.Color, factor float64) color.Color {
	if factor <= 0 {
		return from
	}
	if factor >= 1 {
		return to
	}

	r1, g1, b1, a1 := from.RGBA()
	r2, g2, b2, a2 := to.RGBA()

	r := uint8(float64(r1>>8)*(1-factor) + float64(r2>>8)*factor)
	g := uint8(float64(g1>>8)*(1-factor) + float64(g2>>8)*factor)
	b := uint8(float64(b1>>8)*(1-factor) + float64(b2>>8)*factor)
	a := uint8(float64(a1>>8)*(1-factor) + float64(a2>>8)*factor)

	return color.NRGBA{R: r, G: g, B: b, A: a}
}

// CreateThemeTransition creates an animated transition between themes
func CreateThemeTransition(from, to *WhispTheme, steps int) []*WhispTheme {
	themes := make([]*WhispTheme, steps)

	for i := 0; i < steps; i++ {
		factor := float64(i) / float64(steps-1)

		scheme := ColorScheme{
			Primary:    NewSerializableColor(InterpolateColor(from.scheme.Primary, to.scheme.Primary, factor)),
			Secondary:  NewSerializableColor(InterpolateColor(from.scheme.Secondary, to.scheme.Secondary, factor)),
			Background: NewSerializableColor(InterpolateColor(from.scheme.Background, to.scheme.Background, factor)),
			Surface:    NewSerializableColor(InterpolateColor(from.scheme.Surface, to.scheme.Surface, factor)),
			// Interpolate other colors as needed
		}

		themes[i] = NewWhispTheme(scheme, to.isDark)
	}

	return themes
}
