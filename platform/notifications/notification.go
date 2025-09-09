package notifications

import (
	"context"
	"fmt"
	"time"
)

// NotificationType represents the type of notification
type NotificationType int

const (
	// NotificationMessage represents a new message notification
	NotificationMessage NotificationType = iota
	// NotificationFriendRequest represents a friend request notification
	NotificationFriendRequest
	// NotificationStatus represents a status update notification
	NotificationStatus
	// NotificationFileTransfer represents a file transfer notification
	NotificationFileTransfer
)

// String returns the string representation of the notification type
func (nt NotificationType) String() string {
	switch nt {
	case NotificationMessage:
		return "message"
	case NotificationFriendRequest:
		return "friend_request"
	case NotificationStatus:
		return "status"
	case NotificationFileTransfer:
		return "file_transfer"
	default:
		return "unknown"
	}
}

// Notification represents a single notification to be displayed
type Notification struct {
	ID       string           `json:"id"`
	Type     NotificationType `json:"type"`
	Title    string           `json:"title"`
	Body     string           `json:"body"`
	Icon     string           `json:"icon,omitempty"`
	Sound    bool             `json:"sound"`
	Urgent   bool             `json:"urgent"`
	Actions  []Action         `json:"actions,omitempty"`
	Metadata map[string]any   `json:"metadata,omitempty"`
}

// Action represents a notification action button
type Action struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// NotificationConfig holds configuration for notifications
type NotificationConfig struct {
	Enabled      bool
	ShowPreview  bool
	PlaySound    bool
	ShowSender   bool
	QuietHours   QuietHours
	PlatformOpts map[string]any
}

// QuietHours represents quiet hours configuration
type QuietHours struct {
	Enabled   bool
	StartTime time.Time
	EndTime   time.Time
}

// IsQuietTime checks if current time is within quiet hours
func (qh QuietHours) IsQuietTime() bool {
	if !qh.Enabled {
		return false
	}

	now := time.Now()
	currentTime := time.Date(0, 1, 1, now.Hour(), now.Minute(), now.Second(), 0, time.UTC)

	// Handle quiet hours that span midnight
	if qh.StartTime.After(qh.EndTime) {
		return currentTime.After(qh.StartTime) || currentTime.Before(qh.EndTime)
	}

	return currentTime.After(qh.StartTime) && currentTime.Before(qh.EndTime)
}

// Manager defines the interface for notification management
type Manager interface {
	// Show displays a notification
	Show(ctx context.Context, notification *Notification) error

	// Cancel cancels a notification by ID
	Cancel(ctx context.Context, notificationID string) error

	// SetConfig updates the notification configuration
	SetConfig(config NotificationConfig) error

	// GetConfig returns the current notification configuration
	GetConfig() NotificationConfig

	// IsSupported returns true if notifications are supported on this platform
	IsSupported() bool

	// RequestPermission requests notification permission (primarily for mobile)
	RequestPermission(ctx context.Context) error

	// Close cleans up resources
	Close() error
}

// NewNotification creates a new notification with sensible defaults
func NewNotification(notificationType NotificationType, title, body string) *Notification {
	return &Notification{
		ID:       generateNotificationID(),
		Type:     notificationType,
		Title:    title,
		Body:     body,
		Sound:    true,
		Urgent:   false,
		Metadata: make(map[string]any),
	}
}

// generateNotificationID generates a unique notification ID
func generateNotificationID() string {
	return fmt.Sprintf("whisp_%d", time.Now().UnixNano())
}
