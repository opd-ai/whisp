package core

import (
	"context"
	"testing"
	"time"

	"github.com/opd-ai/whisp/platform/notifications"
)

func TestNotificationService(t *testing.T) {
	// Create test app
	config := &Config{
		DataDir:    "/tmp/whisp-test",
		ConfigPath: "/tmp/whisp-test/config.yaml",
		Debug:      true,
		Platform:   "linux",
	}

	app, err := NewApp(config)
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}
	defer app.Cleanup()

	// Test that notification service was initialized
	if app.notifications == nil {
		t.Fatal("Notification service should be initialized")
	}

	t.Run("IsSupported returns true", func(t *testing.T) {
		if !app.notifications.IsSupported() {
			t.Error("Expected notification service to be supported")
		}
	})

	t.Run("GetConfig returns valid config", func(t *testing.T) {
		config := app.notifications.GetConfig()
		if !config.Enabled {
			t.Error("Expected notifications to be enabled by default")
		}
	})

	t.Run("SetEnabled works", func(t *testing.T) {
		err := app.notifications.SetEnabled(false)
		if err != nil {
			t.Errorf("Unexpected error setting enabled: %v", err)
		}

		if app.notifications.IsEnabled() {
			t.Error("Expected notifications to be disabled")
		}

		// Re-enable for other tests
		app.notifications.SetEnabled(true)
	})

	t.Run("UpdateConfig works", func(t *testing.T) {
		newConfig := notifications.NotificationConfig{
			Enabled:     true,
			ShowPreview: false,
			PlaySound:   false,
			ShowSender:  false,
		}

		err := app.notifications.UpdateConfig(newConfig)
		if err != nil {
			t.Errorf("Unexpected error updating config: %v", err)
		}

		retrievedConfig := app.notifications.GetConfig()
		if retrievedConfig.ShowPreview != newConfig.ShowPreview {
			t.Errorf("Expected ShowPreview %v, got %v", newConfig.ShowPreview, retrievedConfig.ShowPreview)
		}
	})

	t.Run("ShowCustomNotification works", func(t *testing.T) {
		err := app.notifications.ShowCustomNotification(
			notifications.NotificationMessage,
			"Test Title",
			"Test Body",
		)
		if err != nil {
			t.Errorf("Unexpected error showing custom notification: %v", err)
		}
	})

	t.Run("ShowFileTransferNotification works", func(t *testing.T) {
		err := app.notifications.ShowFileTransferNotification(
			1,
			"test.txt",
			true,
		)
		if err != nil {
			t.Errorf("Unexpected error showing file transfer notification: %v", err)
		}
	})
}

func TestNotificationServiceIntegration(t *testing.T) {
	// Create test app
	config := &Config{
		DataDir:    "/tmp/whisp-test-integration",
		ConfigPath: "/tmp/whisp-test-integration/config.yaml",
		Debug:      true,
		Platform:   "linux",
	}

	app, err := NewApp(config)
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}
	defer app.Cleanup()

	// Start the app to initialize callbacks
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = app.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start app: %v", err)
	}
	defer app.Stop()

	t.Run("Service starts without error", func(t *testing.T) {
		// The service should have started as part of app.Start()
		if !app.notifications.IsEnabled() {
			t.Error("Expected notification service to be enabled after start")
		}
	})

	t.Run("Service handles disabled state", func(t *testing.T) {
		// Disable notifications
		app.notifications.SetEnabled(false)

		// Try to show a notification - should not error
		err := app.notifications.ShowCustomNotification(
			notifications.NotificationMessage,
			"Should Not Show",
			"This should be ignored",
		)
		if err != nil {
			t.Errorf("Expected no error when notifications disabled, got: %v", err)
		}

		// Re-enable for cleanup
		app.notifications.SetEnabled(true)
	})
}

func TestNotificationServiceCallbacks(t *testing.T) {
	// Create a mock notification service to test callback setup
	config := &Config{
		DataDir:    "/tmp/whisp-test-callbacks",
		ConfigPath: "/tmp/whisp-test-callbacks/config.yaml",
		Debug:      true,
		Platform:   "linux",
	}

	app, err := NewApp(config)
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}
	defer app.Cleanup()

	// Test that notification service was created
	if app.notifications == nil {
		t.Fatal("Notification service should be initialized")
	}

	// Test callback setup (we can't easily test the actual callbacks without
	// mocking the Tox manager, but we can test that the service initializes correctly)
	t.Run("setupToxCallbacks doesn't panic", func(t *testing.T) {
		// This should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("setupToxCallbacks panicked: %v", r)
			}
		}()

		// Create a new service and try to set up callbacks
		service := NewNotificationService(app)
		service.setupToxCallbacks()
	})

	t.Run("getFriendName handles missing contact", func(t *testing.T) {
		friendName := app.notifications.getFriendName(999999)
		if friendName == "" {
			t.Error("Expected non-empty friend name for unknown friend")
		}
		if friendName != "Unknown Friend" {
			t.Errorf("Expected 'Unknown Friend', got '%s'", friendName)
		}
	})
}

// Benchmark the notification system
func BenchmarkNotificationService(b *testing.B) {
	config := &Config{
		DataDir:    "/tmp/whisp-bench",
		ConfigPath: "/tmp/whisp-bench/config.yaml",
		Debug:      false,
		Platform:   "linux",
	}

	app, err := NewApp(config)
	if err != nil {
		b.Fatalf("Failed to create app: %v", err)
	}
	defer app.Cleanup()

	b.ResetTimer()

	b.Run("ShowCustomNotification", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			app.notifications.ShowCustomNotification(
				notifications.NotificationMessage,
				"Benchmark Test",
				"This is a benchmark test message",
			)
		}
	})

	b.Run("ConfigOperations", func(b *testing.B) {
		config := notifications.NotificationConfig{
			Enabled:     true,
			ShowPreview: true,
			PlaySound:   true,
			ShowSender:  true,
		}

		for i := 0; i < b.N; i++ {
			app.notifications.UpdateConfig(config)
			_ = app.notifications.GetConfig()
		}
	})
}
