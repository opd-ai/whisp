package common

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// GetUserDataDir returns the appropriate user data directory for the platform
func GetUserDataDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		// Windows: %APPDATA%\Whisp
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		return filepath.Join(appData, "Whisp"), nil

	case "darwin":
		// macOS: ~/Library/Application Support/Whisp
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, "Library", "Application Support", "Whisp"), nil

	case "linux":
		// Linux: ~/.local/share/whisp or $XDG_DATA_HOME/whisp
		xdgDataHome := os.Getenv("XDG_DATA_HOME")
		if xdgDataHome != "" {
			return filepath.Join(xdgDataHome, "whisp"), nil
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, ".local", "share", "whisp"), nil

	case "android":
		// Android: Use internal storage
		return "/data/data/com.opd-ai.whisp/files", nil

	case "ios":
		// iOS: Use documents directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, "Documents"), nil

	default:
		// Fallback: ~/.whisp
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, ".whisp"), nil
	}
}

// GetCacheDir returns the appropriate cache directory for the platform
func GetCacheDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		// Windows: %LOCALAPPDATA%\Whisp\Cache
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			return "", fmt.Errorf("LOCALAPPDATA environment variable not set")
		}
		return filepath.Join(localAppData, "Whisp", "Cache"), nil

	case "darwin":
		// macOS: ~/Library/Caches/Whisp
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, "Library", "Caches", "Whisp"), nil

	case "linux":
		// Linux: ~/.cache/whisp or $XDG_CACHE_HOME/whisp
		xdgCacheHome := os.Getenv("XDG_CACHE_HOME")
		if xdgCacheHome != "" {
			return filepath.Join(xdgCacheHome, "whisp"), nil
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, ".cache", "whisp"), nil

	default:
		// Fallback: Use data dir + cache
		dataDir, err := GetUserDataDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(dataDir, "cache"), nil
	}
}

// GetConfigDir returns the appropriate config directory for the platform
func GetConfigDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		// Windows: Same as data dir
		return GetUserDataDir()

	case "darwin":
		// macOS: Same as data dir
		return GetUserDataDir()

	case "linux":
		// Linux: ~/.config/whisp or $XDG_CONFIG_HOME/whisp
		xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfigHome != "" {
			return filepath.Join(xdgConfigHome, "whisp"), nil
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, ".config", "whisp"), nil

	default:
		// Fallback: Use data dir
		return GetUserDataDir()
	}
}
