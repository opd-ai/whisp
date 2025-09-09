package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/opd-ai/whisp/platform/notifications"
)

func main() {
	fmt.Println("=== Whisp Notification System Demo ===")
	fmt.Println()

	// Create notification manager
	manager := notifications.NewManager("")
	defer manager.Close()

	fmt.Printf("Platform supported: %t\n", manager.IsSupported())
	fmt.Printf("Initial config: %+v\n", manager.GetConfig())
	fmt.Println()

	// Request permission (mainly for mobile)
	if err := manager.RequestPermission(context.Background()); err != nil {
		log.Fatalf("Failed to request permission: %v", err)
	}
	fmt.Println("✓ Permission requested successfully")

	// Test 1: Basic message notification
	fmt.Println("\n--- Test 1: Basic Message Notification ---")
	messageNotif := notifications.NewMessageNotification("Alice", "Hey there! How are you doing?")
	if err := manager.Show(context.Background(), messageNotif); err != nil {
		log.Printf("Error showing message notification: %v", err)
	} else {
		fmt.Println("✓ Message notification sent")
	}
	time.Sleep(2 * time.Second)

	// Test 2: Friend request notification
	fmt.Println("\n--- Test 2: Friend Request Notification ---")
	friendReqNotif := notifications.NewFriendRequestNotification("Bob", "I'd like to add you as a friend!")
	if err := manager.Show(context.Background(), friendReqNotif); err != nil {
		log.Printf("Error showing friend request notification: %v", err)
	} else {
		fmt.Println("✓ Friend request notification sent")
	}
	time.Sleep(2 * time.Second)

	// Test 3: Status update notification
	fmt.Println("\n--- Test 3: Status Update Notification ---")
	statusNotif := notifications.NewStatusNotification("Charlie", "online")
	if err := manager.Show(context.Background(), statusNotif); err != nil {
		log.Printf("Error showing status notification: %v", err)
	} else {
		fmt.Println("✓ Status notification sent")
	}
	time.Sleep(2 * time.Second)

	// Test 4: File transfer notification
	fmt.Println("\n--- Test 4: File Transfer Notification ---")
	fileNotif := notifications.NewFileTransferNotification("Dave", "important_document.pdf", true)
	if err := manager.Show(context.Background(), fileNotif); err != nil {
		log.Printf("Error showing file transfer notification: %v", err)
	} else {
		fmt.Println("✓ File transfer notification sent")
	}
	time.Sleep(2 * time.Second)

	// Test 5: Privacy settings
	fmt.Println("\n--- Test 5: Privacy Settings ---")
	privacyConfig := notifications.NotificationConfig{
		Enabled:     true,
		ShowPreview: false, // Hide message content
		PlaySound:   true,
		ShowSender:  false, // Hide sender name
	}
	if err := manager.SetConfig(privacyConfig); err != nil {
		log.Printf("Error setting privacy config: %v", err)
	} else {
		fmt.Println("✓ Privacy config applied")
	}

	// Show a message with privacy settings
	privateMessageNotif := notifications.NewMessageNotification("Secret Sender", "This is sensitive information!")
	if err := manager.Show(context.Background(), privateMessageNotif); err != nil {
		log.Printf("Error showing private message notification: %v", err)
	} else {
		fmt.Println("✓ Private message notification sent (should hide details)")
	}
	time.Sleep(2 * time.Second)

	// Test 6: Disabled notifications
	fmt.Println("\n--- Test 6: Disabled Notifications ---")
	disabledConfig := notifications.NotificationConfig{
		Enabled: false,
	}
	if err := manager.SetConfig(disabledConfig); err != nil {
		log.Printf("Error setting disabled config: %v", err)
	} else {
		fmt.Println("✓ Notifications disabled")
	}

	// Try to show a notification (should be silently ignored)
	disabledNotif := notifications.NewMessageNotification("Should Not Appear", "This notification should not appear")
	if err := manager.Show(context.Background(), disabledNotif); err != nil {
		log.Printf("Error (expected none): %v", err)
	} else {
		fmt.Println("✓ Disabled notification silently ignored")
	}

	// Test 7: Quiet hours
	fmt.Println("\n--- Test 7: Quiet Hours ---")
	quietConfig := notifications.NotificationConfig{
		Enabled:     true,
		ShowPreview: true,
		PlaySound:   true,
		ShowSender:  true,
		QuietHours: notifications.QuietHours{
			Enabled:   true,
			StartTime: time.Date(0, 1, 1, time.Now().Hour(), time.Now().Minute(), 0, 0, time.UTC),
			EndTime:   time.Date(0, 1, 1, time.Now().Hour(), time.Now().Minute()+1, 0, 0, time.UTC),
		},
	}
	if err := manager.SetConfig(quietConfig); err != nil {
		log.Printf("Error setting quiet hours config: %v", err)
	} else {
		fmt.Println("✓ Quiet hours enabled for current time")
	}

	// Try to show a notification during quiet hours
	quietNotif := notifications.NewMessageNotification("Quiet Sender", "This should be suppressed during quiet hours")
	if err := manager.Show(context.Background(), quietNotif); err != nil {
		log.Printf("Error (expected none): %v", err)
	} else {
		fmt.Println("✓ Quiet hours notification silently ignored")
	}

	// Test 8: Error handling
	fmt.Println("\n--- Test 8: Error Handling ---")

	// Re-enable notifications for error tests
	normalConfig := notifications.NotificationConfig{Enabled: true}
	manager.SetConfig(normalConfig)

	// Test nil notification
	if err := manager.Show(context.Background(), nil); err != nil {
		fmt.Printf("✓ Correctly handled nil notification: %v\n", err)
	} else {
		fmt.Println("✗ Should have errored on nil notification")
	}

	// Test empty title
	emptyTitleNotif := &notifications.Notification{Body: "Body without title"}
	if err := manager.Show(context.Background(), emptyTitleNotif); err != nil {
		fmt.Printf("✓ Correctly handled empty title: %v\n", err)
	} else {
		fmt.Println("✗ Should have errored on empty title")
	}

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("Check your system notifications to see the results!")
	fmt.Printf("Final config: %+v\n", manager.GetConfig())
}
