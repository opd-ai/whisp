package message

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/opd-ai/toxcore"
	"github.com/opd-ai/whisp/internal/storage"
)

// MockToxManager implements ToxManager for testing
type MockToxManager struct {
	sendError       error
	lastMessage     string
	lastFriendID    uint32
	lastMessageType toxcore.MessageType
}

func (m *MockToxManager) SendMessage(friendID uint32, message string, messageType toxcore.MessageType) error {
	m.lastFriendID = friendID
	m.lastMessage = message
	m.lastMessageType = messageType
	return m.sendError
}

// MockContactManager implements ContactManager for testing
type MockContactManager struct {
	contacts map[uint32]interface{}
}

func (m *MockContactManager) GetContact(friendID uint32) (interface{}, bool) {
	contact, exists := m.contacts[friendID]
	return contact, exists
}

func NewMockContactManager() *MockContactManager {
	return &MockContactManager{
		contacts: make(map[uint32]interface{}),
	}
}

func (m *MockContactManager) AddContact(friendID uint32, contact interface{}) {
	m.contacts[friendID] = contact
}

// Test setup helper
func setupTestManager(t *testing.T) (*Manager, *storage.Database, *MockToxManager, *MockContactManager, func()) {
	// Create temporary database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := storage.NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	toxMgr := &MockToxManager{}
	contactMgr := NewMockContactManager()

	// Add a test contact
	contactMgr.AddContact(1, map[string]string{"name": "Test Friend"})

	mgr := NewManager(db, toxMgr, contactMgr)

	cleanup := func() {
		db.Close()
		os.RemoveAll(tempDir)
	}

	return mgr, db, toxMgr, contactMgr, cleanup
}

func TestNewManager(t *testing.T) {
	mgr, _, _, _, cleanup := setupTestManager(t)
	defer cleanup()

	if mgr == nil {
		t.Fatal("Expected non-nil manager")
	}

	if mgr.pendingMessages == nil {
		t.Error("Expected pendingMessages map to be initialized")
	}
}

func TestSendMessage(t *testing.T) {
	mgr, _, toxMgr, _, cleanup := setupTestManager(t)
	defer cleanup()

	tests := []struct {
		name        string
		friendID    uint32
		content     string
		messageType MessageType
		toxError    error
		expectError bool
	}{
		{
			name:        "successful normal message",
			friendID:    1,
			content:     "Hello, world!",
			messageType: MessageTypeNormal,
			toxError:    nil,
			expectError: false,
		},
		{
			name:        "successful action message",
			friendID:    1,
			content:     "waves hello",
			messageType: MessageTypeAction,
			toxError:    nil,
			expectError: false,
		},
		{
			name:        "tox send failure",
			friendID:    1,
			content:     "Failed message",
			messageType: MessageTypeNormal,
			toxError:    fmt.Errorf("network error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toxMgr.sendError = tt.toxError

			msg, err := mgr.SendMessage(tt.friendID, tt.content, tt.messageType)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify message properties
			if msg.FriendID != tt.friendID {
				t.Errorf("Expected friend ID %d, got %d", tt.friendID, msg.FriendID)
			}
			if msg.Content != tt.content {
				t.Errorf("Expected content %q, got %q", tt.content, msg.Content)
			}
			if msg.MessageType != tt.messageType {
				t.Errorf("Expected message type %d, got %d", tt.messageType, msg.MessageType)
			}
			if !msg.IsOutgoing {
				t.Error("Expected outgoing message")
			}
			if msg.UUID == "" {
				t.Error("Expected non-empty UUID")
			}
			if msg.ID == 0 {
				t.Error("Expected non-zero message ID")
			}

			// Verify delivery status
			if msg.DeliveredAt == nil {
				t.Error("Expected delivery timestamp")
			}

			// Verify Tox manager was called correctly
			if toxMgr.lastFriendID != tt.friendID {
				t.Errorf("Expected Tox friend ID %d, got %d", tt.friendID, toxMgr.lastFriendID)
			}
			if toxMgr.lastMessage != tt.content {
				t.Errorf("Expected Tox message %q, got %q", tt.content, toxMgr.lastMessage)
			}

			expectedToxType := toxcore.MessageTypeNormal
			if tt.messageType == MessageTypeAction {
				expectedToxType = toxcore.MessageTypeAction
			}
			if toxMgr.lastMessageType != expectedToxType {
				t.Errorf("Expected Tox message type %d, got %d", expectedToxType, toxMgr.lastMessageType)
			}
		})
	}
}

func TestHandleIncomingMessage(t *testing.T) {
	mgr, _, _, _, cleanup := setupTestManager(t)
	defer cleanup()

	friendID := uint32(1)
	content := "Hello from friend!"
	messageType := MessageTypeNormal

	msg := mgr.HandleIncomingMessage(friendID, content, messageType)

	if msg == nil {
		t.Fatal("Expected non-nil message")
	}

	// Verify message properties
	if msg.FriendID != friendID {
		t.Errorf("Expected friend ID %d, got %d", friendID, msg.FriendID)
	}
	if msg.Content != content {
		t.Errorf("Expected content %q, got %q", content, msg.Content)
	}
	if msg.MessageType != messageType {
		t.Errorf("Expected message type %d, got %d", messageType, msg.MessageType)
	}
	if msg.IsOutgoing {
		t.Error("Expected incoming message")
	}
	if msg.UUID == "" {
		t.Error("Expected non-empty UUID")
	}
	if msg.ID == 0 {
		t.Error("Expected non-zero message ID")
	}

	// Verify read status (should be automatically marked as read)
	if msg.ReadAt == nil {
		t.Error("Expected read timestamp")
	}
}

func TestGetMessages(t *testing.T) {
	mgr, _, _, _, cleanup := setupTestManager(t)
	defer cleanup()

	friendID := uint32(1)

	// Send some test messages
	messages := []struct {
		content string
		msgType MessageType
	}{
		{"First message", MessageTypeNormal},
		{"Second message", MessageTypeAction},
		{"Third message", MessageTypeNormal},
	}

	var sentMessages []*Message
	for _, msg := range messages {
		sent, err := mgr.SendMessage(friendID, msg.content, msg.msgType)
		if err != nil {
			t.Fatalf("Failed to send test message: %v", err)
		}
		sentMessages = append(sentMessages, sent)
	}

	// Test getting messages with limit
	retrieved, err := mgr.GetMessages(friendID, 2, 0)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	if len(retrieved) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(retrieved))
	}

	// Verify messages are in reverse chronological order (newest first)
	if len(retrieved) >= 2 {
		if retrieved[0].Timestamp.Before(retrieved[1].Timestamp) {
			t.Error("Expected messages in reverse chronological order")
		}
	}

	// Test getting all messages
	allMessages, err := mgr.GetMessages(friendID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get all messages: %v", err)
	}

	if len(allMessages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(allMessages))
	}

	// Test pagination
	secondPage, err := mgr.GetMessages(friendID, 2, 2)
	if err != nil {
		t.Fatalf("Failed to get second page: %v", err)
	}

	if len(secondPage) != 1 {
		t.Errorf("Expected 1 message on second page, got %d", len(secondPage))
	}
}

func TestEditMessage(t *testing.T) {
	mgr, _, _, _, cleanup := setupTestManager(t)
	defer cleanup()

	// Send a message first
	originalContent := "Original message"
	msg, err := mgr.SendMessage(1, originalContent, MessageTypeNormal)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Edit the message
	newContent := "Edited message"
	err = mgr.EditMessage(msg.ID, newContent)
	if err != nil {
		t.Fatalf("Failed to edit message: %v", err)
	}

	// Verify the edit by retrieving the message
	messages, err := mgr.GetMessages(1, 1, 0)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	editedMsg := messages[0]
	if editedMsg.Content != newContent {
		t.Errorf("Expected content %q, got %q", newContent, editedMsg.Content)
	}
	if editedMsg.OriginalContent != originalContent {
		t.Errorf("Expected original content %q, got %q", originalContent, editedMsg.OriginalContent)
	}
	if editedMsg.EditedAt == nil {
		t.Error("Expected edited timestamp")
	}
}

func TestDeleteMessage(t *testing.T) {
	mgr, _, _, _, cleanup := setupTestManager(t)
	defer cleanup()

	// Send a message first
	msg, err := mgr.SendMessage(1, "Message to delete", MessageTypeNormal)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Delete the message
	err = mgr.DeleteMessage(msg.ID)
	if err != nil {
		t.Fatalf("Failed to delete message: %v", err)
	}

	// Verify the message is not returned in queries
	messages, err := mgr.GetMessages(1, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	if len(messages) != 0 {
		t.Errorf("Expected 0 messages after deletion, got %d", len(messages))
	}
}

func TestMarkAsRead(t *testing.T) {
	mgr, _, _, _, cleanup := setupTestManager(t)
	defer cleanup()

	friendID := uint32(1)

	// Create some incoming messages (simulate them being unread)
	msg1 := mgr.HandleIncomingMessage(friendID, "Unread message 1", MessageTypeNormal)
	msg2 := mgr.HandleIncomingMessage(friendID, "Unread message 2", MessageTypeNormal)

	if msg1 == nil || msg2 == nil {
		t.Fatal("Failed to create test messages")
	}

	// Clear read status for testing (simulate unread messages)
	_, err := mgr.db.Exec("UPDATE messages SET read_at = NULL WHERE id IN (?, ?)", msg1.ID, msg2.ID)
	if err != nil {
		t.Fatalf("Failed to clear read status: %v", err)
	}

	// Mark messages as read
	err = mgr.MarkAsRead(friendID)
	if err != nil {
		t.Fatalf("Failed to mark as read: %v", err)
	}

	// Verify messages are marked as read
	messages, err := mgr.GetMessages(friendID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	for _, msg := range messages {
		if !msg.IsOutgoing && msg.ReadAt == nil {
			t.Error("Expected incoming message to be marked as read")
		}
	}
}

func TestSearchMessages(t *testing.T) {
	mgr, _, _, _, cleanup := setupTestManager(t)
	defer cleanup()

	friendID := uint32(1)

	// Send messages with different content
	testMessages := []string{
		"Hello world",
		"How are you today?",
		"World peace is important",
		"Random message",
	}

	for _, content := range testMessages {
		_, err := mgr.SendMessage(friendID, content, MessageTypeNormal)
		if err != nil {
			t.Fatalf("Failed to send test message: %v", err)
		}
	}

	// Test search functionality
	results, err := mgr.SearchMessages("world", 10)
	if err != nil {
		t.Fatalf("Failed to search messages: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 search results, got %d", len(results))
	}

	// Verify results contain the search term
	for _, msg := range results {
		if !contains(msg.Content, "world") && !contains(msg.Content, "World") {
			t.Errorf("Search result doesn't contain search term: %q", msg.Content)
		}
	}

	// Test case-insensitive search
	results, err = mgr.SearchMessages("HELLO", 10)
	if err != nil {
		t.Fatalf("Failed to search messages: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 search result for case-insensitive search, got %d", len(results))
	}
}

func TestMessagePersistence(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "persistence_test.db")

	// Create first manager instance
	db1, err := storage.NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	toxMgr1 := &MockToxManager{}
	contactMgr1 := NewMockContactManager()
	contactMgr1.AddContact(1, map[string]string{"name": "Test Friend"})

	mgr1 := NewManager(db1, toxMgr1, contactMgr1)

	// Send a message
	originalMessage := "Persistent message"
	msg, err := mgr1.SendMessage(1, originalMessage, MessageTypeNormal)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	msgID := msg.ID
	msgUUID := msg.UUID

	db1.Close()

	// Create second manager instance with same database
	db2, err := storage.NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to reopen database: %v", err)
	}
	defer db2.Close()

	toxMgr2 := &MockToxManager{}
	contactMgr2 := NewMockContactManager()
	contactMgr2.AddContact(1, map[string]string{"name": "Test Friend"})

	mgr2 := NewManager(db2, toxMgr2, contactMgr2)

	// Retrieve the message
	messages, err := mgr2.GetMessages(1, 10, 0)
	if err != nil {
		t.Fatalf("Failed to retrieve messages: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 persisted message, got %d", len(messages))
	}

	persistedMsg := messages[0]
	if persistedMsg.ID != msgID {
		t.Errorf("Expected message ID %d, got %d", msgID, persistedMsg.ID)
	}
	if persistedMsg.UUID != msgUUID {
		t.Errorf("Expected message UUID %s, got %s", msgUUID, persistedMsg.UUID)
	}
	if persistedMsg.Content != originalMessage {
		t.Errorf("Expected message content %q, got %q", originalMessage, persistedMsg.Content)
	}
}

// Helper function to check if string contains substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsInner(s, substr))))
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
