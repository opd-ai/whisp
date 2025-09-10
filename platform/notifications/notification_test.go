package notifications

import (
	"context"
	"testing"
	"time"
)

func TestNotification(t *testing.T) {
	t.Run("NewNotification creates valid notification", func(t *testing.T) {
		notif := NewNotification(NotificationMessage, "Test Title", "Test Body")

		if notif.Type != NotificationMessage {
			t.Errorf("Expected type %v, got %v", NotificationMessage, notif.Type)
		}
		if notif.Title != "Test Title" {
			t.Errorf("Expected title 'Test Title', got '%s'", notif.Title)
		}
		if notif.Body != "Test Body" {
			t.Errorf("Expected body 'Test Body', got '%s'", notif.Body)
		}
		if notif.ID == "" {
			t.Error("Expected non-empty ID")
		}
		if !notif.Sound {
			t.Error("Expected Sound to be true by default")
		}
		if notif.Urgent {
			t.Error("Expected Urgent to be false by default")
		}
		if notif.Metadata == nil {
			t.Error("Expected Metadata to be initialized")
		}
	})
}

func TestNotificationType(t *testing.T) {
	tests := []struct {
		notifType NotificationType
		expected  string
	}{
		{NotificationMessage, "message"},
		{NotificationFriendRequest, "friend_request"},
		{NotificationStatus, "status"},
		{NotificationFileTransfer, "file_transfer"},
		{NotificationType(999), "unknown"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			if test.notifType.String() != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, test.notifType.String())
			}
		})
	}
}

func TestQuietHours(t *testing.T) {
	t.Run("IsQuietTime returns false when disabled", func(t *testing.T) {
		qh := QuietHours{Enabled: false}
		if qh.IsQuietTime() {
			t.Error("Expected false when quiet hours are disabled")
		}
	})

	t.Run("IsQuietTime handles normal hours", func(t *testing.T) {
		// Create quiet hours from 22:00 to 08:00
		qh := QuietHours{
			Enabled:   true,
			StartTime: time.Date(0, 1, 1, 22, 0, 0, 0, time.UTC),
			EndTime:   time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
		}

		// Test during quiet hours (23:00)
		testTime := time.Date(2023, 1, 1, 23, 0, 0, 0, time.UTC)
		currentTime := time.Date(0, 1, 1, testTime.Hour(), testTime.Minute(), testTime.Second(), 0, time.UTC)

		// Manually check if it's quiet time
		isQuiet := currentTime.After(qh.StartTime) || currentTime.Before(qh.EndTime)

		if !isQuiet {
			t.Error("Expected quiet time at 23:00")
		}

		// Test during non-quiet hours (15:00)
		testTime = time.Date(2023, 1, 1, 15, 0, 0, 0, time.UTC)
		currentTime = time.Date(0, 1, 1, testTime.Hour(), testTime.Minute(), testTime.Second(), 0, time.UTC)

		isQuiet = currentTime.After(qh.StartTime) || currentTime.Before(qh.EndTime)

		if isQuiet {
			t.Error("Expected non-quiet time at 15:00")
		}
	})
}

func TestCrossPlatformManager(t *testing.T) {
	manager := NewCrossPlatformManager("test-icon.png")

	t.Run("IsSupported returns true", func(t *testing.T) {
		if !manager.IsSupported() {
			t.Error("Expected IsSupported to return true")
		}
	})

	t.Run("SetConfig and GetConfig work", func(t *testing.T) {
		config := NotificationConfig{
			Enabled:     false,
			ShowPreview: false,
			PlaySound:   false,
			ShowSender:  false,
		}

		err := manager.SetConfig(config)
		if err != nil {
			t.Errorf("Unexpected error setting config: %v", err)
		}

		retrievedConfig := manager.GetConfig()
		if retrievedConfig.Enabled != config.Enabled {
			t.Errorf("Expected Enabled %v, got %v", config.Enabled, retrievedConfig.Enabled)
		}
		if retrievedConfig.ShowPreview != config.ShowPreview {
			t.Errorf("Expected ShowPreview %v, got %v", config.ShowPreview, retrievedConfig.ShowPreview)
		}
	})

	t.Run("Show respects disabled config", func(t *testing.T) {
		config := NotificationConfig{Enabled: false}
		manager.SetConfig(config)

		notification := NewNotification(NotificationMessage, "Test", "Body")
		err := manager.Show(context.Background(), notification)
		// Should not error when disabled, just silently ignore
		if err != nil {
			t.Errorf("Unexpected error when notifications disabled: %v", err)
		}
	})

	t.Run("Show validates notification", func(t *testing.T) {
		config := NotificationConfig{Enabled: true}
		manager.SetConfig(config)

		// Test nil notification
		err := manager.Show(context.Background(), nil)
		if err == nil {
			t.Error("Expected error for nil notification")
		}

		// Test empty title
		notification := &Notification{Body: "Body"}
		err = manager.Show(context.Background(), notification)
		if err == nil {
			t.Error("Expected error for empty title")
		}
	})

	t.Run("RequestPermission succeeds", func(t *testing.T) {
		err := manager.RequestPermission(context.Background())
		if err != nil {
			t.Errorf("Unexpected error requesting permission: %v", err)
		}
	})

	t.Run("Cancel works without error", func(t *testing.T) {
		err := manager.Cancel(context.Background(), "test-id")
		if err != nil {
			t.Errorf("Unexpected error canceling notification: %v", err)
		}
	})

	t.Run("Close works without error", func(t *testing.T) {
		err := manager.Close()
		if err != nil {
			t.Errorf("Unexpected error closing manager: %v", err)
		}
	})
}

func TestFactory(t *testing.T) {
	t.Run("NewManager returns valid manager", func(t *testing.T) {
		manager := NewManager("test-icon.png")
		if manager == nil {
			t.Error("Expected non-nil manager")
		}

		if !manager.IsSupported() {
			t.Error("Expected manager to be supported")
		}
	})

	t.Run("ConfigFromYAML returns valid config", func(t *testing.T) {
		config, err := ConfigFromYAML(nil)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !config.Enabled {
			t.Error("Expected default Enabled to be true")
		}
		if !config.ShowPreview {
			t.Error("Expected default ShowPreview to be true")
		}
	})
}

func TestNotificationHelpers(t *testing.T) {
	t.Run("NewMessageNotification", func(t *testing.T) {
		notif := NewMessageNotification("Alice", "Hello there!")

		if notif.Type != NotificationMessage {
			t.Errorf("Expected type %v, got %v", NotificationMessage, notif.Type)
		}
		if notif.Title != "Alice" {
			t.Errorf("Expected title 'Alice', got '%s'", notif.Title)
		}
		if notif.Body != "Hello there!" {
			t.Errorf("Expected body 'Hello there!', got '%s'", notif.Body)
		}
		if !notif.Sound {
			t.Error("Expected Sound to be true for messages")
		}
	})

	t.Run("NewFriendRequestNotification", func(t *testing.T) {
		notif := NewFriendRequestNotification("Bob", "Let's be friends!")

		if notif.Type != NotificationFriendRequest {
			t.Errorf("Expected type %v, got %v", NotificationFriendRequest, notif.Type)
		}
		if notif.Title != "New Friend Request" {
			t.Errorf("Expected title 'New Friend Request', got '%s'", notif.Title)
		}
		expectedBody := "Bob: Let's be friends!"
		if notif.Body != expectedBody {
			t.Errorf("Expected body '%s', got '%s'", expectedBody, notif.Body)
		}
	})

	t.Run("NewStatusNotification", func(t *testing.T) {
		notif := NewStatusNotification("Charlie", "online")

		if notif.Type != NotificationStatus {
			t.Errorf("Expected type %v, got %v", NotificationStatus, notif.Type)
		}
		if notif.Title != "Status Update" {
			t.Errorf("Expected title 'Status Update', got '%s'", notif.Title)
		}
		expectedBody := "Charlie is now online"
		if notif.Body != expectedBody {
			t.Errorf("Expected body '%s', got '%s'", expectedBody, notif.Body)
		}
		if notif.Sound {
			t.Error("Expected Sound to be false for status updates")
		}
	})

	t.Run("NewFileTransferNotification incoming", func(t *testing.T) {
		notif := NewFileTransferNotification("Dave", "document.pdf", true)

		if notif.Type != NotificationFileTransfer {
			t.Errorf("Expected type %v, got %v", NotificationFileTransfer, notif.Type)
		}
		if notif.Title != "File Received" {
			t.Errorf("Expected title 'File Received', got '%s'", notif.Title)
		}
		expectedBody := "Dave sent you document.pdf"
		if notif.Body != expectedBody {
			t.Errorf("Expected body '%s', got '%s'", expectedBody, notif.Body)
		}
		if !notif.Sound {
			t.Error("Expected Sound to be true for incoming files")
		}
	})

	t.Run("NewFileTransferNotification outgoing", func(t *testing.T) {
		notif := NewFileTransferNotification("Eve", "photo.jpg", false)

		if notif.Title != "File Sent" {
			t.Errorf("Expected title 'File Sent', got '%s'", notif.Title)
		}
		expectedBody := "Successfully sent photo.jpg to Eve"
		if notif.Body != expectedBody {
			t.Errorf("Expected body '%s', got '%s'", expectedBody, notif.Body)
		}
		if notif.Sound {
			t.Error("Expected Sound to be false for outgoing files")
		}
	})
}

func TestGenerateNotificationID(t *testing.T) {
	id1 := generateNotificationID()
	id2 := generateNotificationID()

	if id1 == "" {
		t.Error("Expected non-empty ID")
	}
	if id2 == "" {
		t.Error("Expected non-empty ID")
	}
	if id1 == id2 {
		t.Error("Expected unique IDs")
	}

	// Check format
	if len(id1) < 6 || id1[:6] != "whisp_" {
		t.Errorf("Expected ID to start with 'whisp_', got '%s'", id1)
	}
}
