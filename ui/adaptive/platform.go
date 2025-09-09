package adaptive

import (
	"runtime"
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
		// TODO: Distinguish between macOS and iOS
		return PlatformMacOS
	case "linux":
		// TODO: Distinguish between Linux and Android
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
