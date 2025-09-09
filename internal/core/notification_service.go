package core

import (
	"context"
	"log"
	"path/filepath"

	"github.com/opd-ai/toxcore"
	"github.com/opd-ai/whisp/platform/notifications"
)

// NotificationService provides notification functionality to the core app
type NotificationService struct {
	manager notifications.Manager
	app     *App
	config  notifications.NotificationConfig
	enabled bool
}

// NewNotificationService creates a new notification service
func NewNotificationService(app *App) *NotificationService {
	// Get icon path from app data directory
	iconPath := notifications.GetDefaultIconPath()
	if iconPath == "" {
		// Try app-specific icon
		if app.config != nil && app.config.DataDir != "" {
			iconPath = filepath.Join(app.config.DataDir, "whisp.png")
		}
	}

	manager := notifications.NewManager(iconPath)

	// Create default config from YAML config
	config, err := notifications.ConfigFromYAML(nil)
	if err != nil {
		log.Printf("Warning: Failed to load notification config: %v", err)
	}

	service := &NotificationService{
		manager: manager,
		app:     app,
		config:  config,
		enabled: true,
	}

	// Apply config to manager
	if err := manager.SetConfig(config); err != nil {
		log.Printf("Warning: Failed to set notification config: %v", err)
	}

	return service
}

// Start initializes the notification service and sets up callbacks
func (ns *NotificationService) Start(ctx context.Context) error {
	if !ns.enabled {
		return nil
	}

	// Request permission (mainly for mobile platforms)
	if err := ns.manager.RequestPermission(ctx); err != nil {
		log.Printf("Warning: Failed to request notification permission: %v", err)
		// Don't fail startup for permission issues
	}

	// Set up callbacks if we have a tox manager
	if ns.app.tox != nil {
		ns.setupToxCallbacks()
	}

	return nil
}

// Stop shuts down the notification service
func (ns *NotificationService) Stop() error {
	return ns.manager.Close()
}

// setupToxCallbacks sets up callbacks to show notifications for Tox events
func (ns *NotificationService) setupToxCallbacks() {
	// Set up message callback
	ns.app.tox.OnFriendMessage(func(friendID uint32, message string) {
		if !ns.enabled {
			return
		}

		// Get friend name from contact manager
		friendName := ns.getFriendName(friendID)

		// Create and show notification
		notification := notifications.NewMessageNotification(friendName, message)
		if err := ns.manager.Show(context.Background(), notification); err != nil {
			log.Printf("Failed to show message notification: %v", err)
		}
	})

	// Set up friend request callback
	ns.app.tox.OnFriendRequest(func(publicKey [32]byte, message string) {
		if !ns.enabled {
			return
		}

		// Convert public key to display format
		senderName := "Unknown User" // In a real app, we might decode the key

		// Create and show notification
		notification := notifications.NewFriendRequestNotification(senderName, message)
		if err := ns.manager.Show(context.Background(), notification); err != nil {
			log.Printf("Failed to show friend request notification: %v", err)
		}
	})

	// Set up status change callback
	ns.app.tox.OnFriendStatus(func(friendID uint32, status toxcore.FriendStatus) {
		if !ns.enabled {
			return
		}

		// Get friend name and convert status
		friendName := ns.getFriendName(friendID)
		statusStr := "unknown"

		// Convert status to string
		switch status {
		case toxcore.FriendStatusNone:
			statusStr = "offline"
		case toxcore.FriendStatusAway:
			statusStr = "away"
		case toxcore.FriendStatusBusy:
			statusStr = "busy"
		default:
			statusStr = "online"
		}

		// Only notify for online status to avoid spam
		if statusStr == "online" {
			notification := notifications.NewStatusNotification(friendName, statusStr)
			if err := ns.manager.Show(context.Background(), notification); err != nil {
				log.Printf("Failed to show status notification: %v", err)
			}
		}
	})
}

// ShowFileTransferNotification shows a notification for file transfers
func (ns *NotificationService) ShowFileTransferNotification(friendID uint32, fileName string, isIncoming bool) error {
	if !ns.enabled {
		return nil
	}

	friendName := ns.getFriendName(friendID)
	notification := notifications.NewFileTransferNotification(friendName, fileName, isIncoming)

	return ns.manager.Show(context.Background(), notification)
}

// ShowCustomNotification shows a custom notification
func (ns *NotificationService) ShowCustomNotification(notificationType notifications.NotificationType, title, body string) error {
	if !ns.enabled {
		return nil
	}

	notification := notifications.NewNotification(notificationType, title, body)
	return ns.manager.Show(context.Background(), notification)
}

// UpdateConfig updates the notification configuration
func (ns *NotificationService) UpdateConfig(config notifications.NotificationConfig) error {
	ns.config = config
	return ns.manager.SetConfig(config)
}

// GetConfig returns the current notification configuration
func (ns *NotificationService) GetConfig() notifications.NotificationConfig {
	return ns.manager.GetConfig()
}

// SetEnabled enables or disables notifications
func (ns *NotificationService) SetEnabled(enabled bool) error {
	ns.enabled = enabled

	// Update manager config
	config := ns.config
	config.Enabled = enabled
	return ns.manager.SetConfig(config)
}

// IsEnabled returns whether notifications are enabled
func (ns *NotificationService) IsEnabled() bool {
	return ns.enabled && ns.config.Enabled
}

// IsSupported returns whether notifications are supported on this platform
func (ns *NotificationService) IsSupported() bool {
	return ns.manager.IsSupported()
}

// getFriendName gets the display name for a friend ID
func (ns *NotificationService) getFriendName(friendID uint32) string {
	if ns.app.contacts == nil {
		return "Friend"
	}

	// Get contact by friend ID
	contacts := ns.app.contacts.GetAllContacts()
	for _, contact := range contacts {
		if contact.FriendID == friendID {
			if contact.Name != "" {
				return contact.Name
			}
			return "Friend " + contact.ToxID[:8] // Show first 8 chars of ToxID
		}
	}

	return "Unknown Friend"
}
