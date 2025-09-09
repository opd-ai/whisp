package notifications

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/gen2brain/beeep"
	"github.com/opd-ai/whisp/ui/adaptive"
)

// CrossPlatformManager implements the Manager interface using beeep for cross-platform notifications
type CrossPlatformManager struct {
	config       NotificationConfig
	mu           sync.RWMutex
	platform     adaptive.Platform
	activeNotifs map[string]*Notification
	iconPath     string
}

// NewCrossPlatformManager creates a new cross-platform notification manager
func NewCrossPlatformManager(iconPath string) *CrossPlatformManager {
	return &CrossPlatformManager{
		config: NotificationConfig{
			Enabled:     true,
			ShowPreview: true,
			PlaySound:   true,
			ShowSender:  true,
		},
		platform:     adaptive.DetectPlatform(),
		activeNotifs: make(map[string]*Notification),
		iconPath:     iconPath,
	}
}

// Show displays a notification using beeep
func (m *CrossPlatformManager) Show(ctx context.Context, notification *Notification) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if notifications are enabled
	if !m.config.Enabled {
		return nil // Silently ignore if disabled
	}

	// Check quiet hours
	if m.config.QuietHours.IsQuietTime() {
		return nil // Silently ignore during quiet hours
	}

	// Validate notification
	if notification == nil {
		return errors.New("notification cannot be nil")
	}
	if notification.Title == "" {
		return errors.New("notification title cannot be empty")
	}

	// Prepare notification content based on config
	title := notification.Title
	body := notification.Body

	// Respect privacy settings
	if !m.config.ShowPreview {
		body = "New message received"
	}
	if !m.config.ShowSender && notification.Type == NotificationMessage {
		title = "New Message"
	}

	// Choose icon
	iconPath := notification.Icon
	if iconPath == "" {
		iconPath = m.iconPath
	}

	// Store active notification
	m.activeNotifs[notification.ID] = notification

	// Show notification based on platform
	var err error
	switch m.platform {
	case adaptive.PlatformWindows:
		err = m.showWindows(title, body, iconPath)
	case adaptive.PlatformMacOS:
		err = m.showMacOS(title, body, iconPath)
	case adaptive.PlatformLinux:
		err = m.showLinux(title, body, iconPath)
	case adaptive.PlatformAndroid, adaptive.PlatformIOS:
		err = m.showMobile(title, body, iconPath)
	default:
		err = m.showGeneric(title, body, iconPath)
	}

	if err != nil {
		delete(m.activeNotifs, notification.ID)
		return fmt.Errorf("failed to show notification: %w", err)
	}

	return nil
}

// Cancel cancels a notification by ID
func (m *CrossPlatformManager) Cancel(ctx context.Context, notificationID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.activeNotifs, notificationID)

	// Note: beeep doesn't provide a way to cancel notifications,
	// so we just remove it from our tracking
	return nil
}

// SetConfig updates the notification configuration
func (m *CrossPlatformManager) SetConfig(config NotificationConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = config
	return nil
}

// GetConfig returns the current notification configuration
func (m *CrossPlatformManager) GetConfig() NotificationConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.config
}

// IsSupported returns true if notifications are supported on this platform
func (m *CrossPlatformManager) IsSupported() bool {
	// beeep supports all major platforms
	return m.platform != adaptive.PlatformUnknown
}

// RequestPermission requests notification permission (primarily for mobile)
func (m *CrossPlatformManager) RequestPermission(ctx context.Context) error {
	// Desktop platforms don't require explicit permission for basic notifications
	if m.platform.IsDesktop() {
		return nil
	}

	// For mobile platforms, this would need platform-specific implementation
	// For now, we assume permission is granted
	return nil
}

// Close cleans up resources
func (m *CrossPlatformManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear active notifications
	m.activeNotifs = make(map[string]*Notification)

	return nil
}

// Platform-specific notification methods

func (m *CrossPlatformManager) showWindows(title, body, iconPath string) error {
	return beeep.Notify(title, body, iconPath)
}

func (m *CrossPlatformManager) showMacOS(title, body, iconPath string) error {
	return beeep.Notify(title, body, iconPath)
}

func (m *CrossPlatformManager) showLinux(title, body, iconPath string) error {
	return beeep.Notify(title, body, iconPath)
}

func (m *CrossPlatformManager) showMobile(title, body, iconPath string) error {
	// For mobile platforms, beeep may not work optimally
	// This is a placeholder for platform-specific mobile notification implementation
	return beeep.Notify(title, body, iconPath)
}

func (m *CrossPlatformManager) showGeneric(title, body, iconPath string) error {
	return beeep.Notify(title, body, iconPath)
}

// Helper function to get default icon path
func GetDefaultIconPath() string {
	// Try to find app icon in common locations
	iconPaths := []string{
		"assets/icon.png",
		"resources/icon.png",
		"icon.png",
		"whisp.png",
	}

	for _, path := range iconPaths {
		if exists, _ := fileExists(path); exists {
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	// Return empty string if no icon found - beeep will use system default
	return ""
}

// fileExists checks if a file exists
func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
