package notifications

import (
	"time"

	"github.com/opd-ai/whisp/ui/adaptive"
)

// NewManager creates a new notification manager for the current platform
func NewManager(iconPath string) Manager {
	platform := adaptive.DetectPlatform()

	// For now, we use the cross-platform manager for all platforms
	// In the future, we could add platform-specific managers here
	switch platform {
	case adaptive.PlatformWindows:
		return NewCrossPlatformManager(iconPath)
	case adaptive.PlatformMacOS:
		return NewCrossPlatformManager(iconPath)
	case adaptive.PlatformLinux:
		return NewCrossPlatformManager(iconPath)
	case adaptive.PlatformAndroid:
		return NewCrossPlatformManager(iconPath)
	case adaptive.PlatformIOS:
		return NewCrossPlatformManager(iconPath)
	default:
		return NewCrossPlatformManager(iconPath)
	}
}

// ConfigFromYAML converts YAML configuration to NotificationConfig
func ConfigFromYAML(yamlConfig any) (NotificationConfig, error) {
	// This function would convert from the YAML structure defined in config.yaml
	// to our NotificationConfig struct. For now, we'll provide defaults.

	config := NotificationConfig{
		Enabled:     true,
		ShowPreview: true,
		PlaySound:   true,
		ShowSender:  true,
		QuietHours: QuietHours{
			Enabled:   false,
			StartTime: time.Date(0, 1, 1, 22, 0, 0, 0, time.UTC), // 10 PM
			EndTime:   time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),  // 8 AM
		},
		PlatformOpts: make(map[string]any),
	}

	return config, nil
}

// Helper functions for creating common notification types

// NewMessageNotification creates a notification for a new message
func NewMessageNotification(senderName, messageContent string) *Notification {
	notification := NewNotification(NotificationMessage, senderName, messageContent)
	notification.Sound = true
	notification.Urgent = false
	return notification
}

// NewFriendRequestNotification creates a notification for a friend request
func NewFriendRequestNotification(senderName, message string) *Notification {
	title := "New Friend Request"
	body := senderName
	if message != "" {
		body += ": " + message
	}

	notification := NewNotification(NotificationFriendRequest, title, body)
	notification.Sound = true
	notification.Urgent = false
	return notification
}

// NewStatusNotification creates a notification for a status update
func NewStatusNotification(friendName, status string) *Notification {
	title := "Status Update"
	body := friendName + " is now " + status

	notification := NewNotification(NotificationStatus, title, body)
	notification.Sound = false
	notification.Urgent = false
	return notification
}

// NewFileTransferNotification creates a notification for a file transfer
func NewFileTransferNotification(friendName, fileName string, isIncoming bool) *Notification {
	var title, body string

	if isIncoming {
		title = "File Received"
		body = friendName + " sent you " + fileName
	} else {
		title = "File Sent"
		body = "Successfully sent " + fileName + " to " + friendName
	}

	notification := NewNotification(NotificationFileTransfer, title, body)
	notification.Sound = isIncoming
	notification.Urgent = false
	return notification
}
