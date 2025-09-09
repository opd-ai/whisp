package adaptive

import (
	"os"
	"runtime"
	"strings"
)

// Platform represents the detected platform
type Platform string

const (
	PlatformWindows Platform = "windows"
	PlatformMacOS   Platform = "macos"
	PlatformLinux   Platform = "linux"
	PlatformAndroid Platform = "android"
	PlatformIOS     Platform = "ios"
	PlatformUnknown Platform = "unknown"
)

// DetectPlatform detects the current platform
func DetectPlatform() Platform {
	switch runtime.GOOS {
	case "windows":
		return PlatformWindows
	case "darwin":
		// Detect iOS vs macOS based on architecture and build tags
		if isIOSEnvironment() {
			return PlatformIOS
		}
		return PlatformMacOS
	case "linux":
		// Detect Android vs Linux based on environment
		if isAndroidEnvironment() {
			return PlatformAndroid
		}
		return PlatformLinux
	default:
		return PlatformUnknown
	}
}

// IsMobile returns true if the platform is mobile
func (p Platform) IsMobile() bool {
	return p == PlatformAndroid || p == PlatformIOS
}

// IsDesktop returns true if the platform is desktop
func (p Platform) IsDesktop() bool {
	return p == PlatformWindows || p == PlatformMacOS || p == PlatformLinux
}

// String returns the string representation of the platform
func (p Platform) String() string {
	return string(p)
}

// isIOSEnvironment detects if running on iOS
func isIOSEnvironment() bool {
	// Check for iOS-specific environment indicators
	// In a real iOS build, these would be set by the build system
	if runtime.GOARCH == "arm64" {
		// Check for iOS-specific environment variables or build tags
		if strings.Contains(strings.ToLower(os.Getenv("HOME")), "mobile") {
			return true
		}
	}
	return false
}

// isAndroidEnvironment detects if running on Android
func isAndroidEnvironment() bool {
	// Check for Android-specific environment indicators
	// In a real Android build, these would be set by the build system
	if strings.Contains(strings.ToLower(os.Getenv("ANDROID_DATA")), "android") ||
		strings.Contains(strings.ToLower(os.Getenv("ANDROID_ROOT")), "android") {
		return true
	}
	return false
}
