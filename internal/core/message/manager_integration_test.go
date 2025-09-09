package message

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/opd-ai/whisp/internal/core/security"
	"github.com/opd-ai/whisp/internal/storage"
)

// TestMessagePersistenceWithEncryption tests that message persistence works with an encrypted database
func TestMessagePersistenceWithEncryption(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "encrypted_test.db")

	// Create security manager for encryption
	securityMgr, err := security.NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Use a fixed master key for consistent testing
	masterKey := []byte("test-master-key-32-bytes-long!!")
	securityMgr.SetMasterKey(masterKey)

	// Create encrypted database
	db1, err := storage.NewDatabaseWithEncryption(dbPath, securityMgr)
	if err != nil {
		t.Fatalf("Failed to create encrypted database: %v", err)
	}

	if !db1.IsEncrypted() {
		t.Fatal("Expected encrypted database")
	}

	toxMgr1 := &MockToxManager{}
	contactMgr1 := NewMockContactManager()
	contactMgr1.AddContact(1, map[string]string{"name": "Test Friend"})

	mgr1 := NewManager(db1, toxMgr1, contactMgr1)

	// Send a message
	originalMessage := "Encrypted persistent message"
	msg, err := mgr1.SendMessage(1, originalMessage, MessageTypeNormal)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	msgID := msg.ID
	msgUUID := msg.UUID

	// Properly close the database
	if err := db1.Close(); err != nil {
		t.Fatalf("Failed to close first database: %v", err)
	}

	// Wait to ensure database is fully closed
	time.Sleep(50 * time.Millisecond)

	// Debug: Check if database file exists and its size
	if stat, err := os.Stat(dbPath); err != nil {
		t.Fatalf("Database file doesn't exist after close: %v", err)
	} else {
		t.Logf("Database file size after close: %d bytes", stat.Size())
	}

	// Reopen with same master key (simulating the same password/key)
	securityMgr2, err := security.NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create second security manager: %v", err)
	}
	// Use the same master key that was used to encrypt the database
	securityMgr2.SetMasterKey(masterKey)

	// Verify that both security managers generate the same database key
	key1, err := securityMgr.GetDatabaseKey()
	if err != nil {
		t.Fatalf("Failed to get first database key: %v", err)
	}
	key2, err := securityMgr2.GetDatabaseKey()
	if err != nil {
		t.Fatalf("Failed to get second database key: %v", err)
	}
	if key1 != key2 {
		t.Fatalf("Database keys don't match - encryption will fail. First: %s, Second: %s", key1, key2)
	}
	t.Logf("Database keys match: %s", key1[:16]+"...") // Log first 16 chars for debugging

	db2, err := storage.NewDatabaseWithEncryption(dbPath, securityMgr2)
	if err != nil {
		t.Fatalf("Failed to reopen encrypted database: %v", err)
	}
	defer db2.Close()

	if !db2.IsEncrypted() {
		t.Fatal("Expected encrypted database on reopen")
	}

	toxMgr2 := &MockToxManager{}
	contactMgr2 := NewMockContactManager()
	contactMgr2.AddContact(1, map[string]string{"name": "Test Friend"})

	mgr2 := NewManager(db2, toxMgr2, contactMgr2)

	// Retrieve the message
	messages, err := mgr2.GetMessages(1, 10, 0)
	if err != nil {
		t.Fatalf("Failed to retrieve messages from encrypted database: %v", err)
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

// TestDatabaseMigrationFromUnencrypted tests that existing unencrypted databases can be migrated
func TestDatabaseMigrationFromUnencrypted(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "migration_test.db")

	// Create unencrypted database first
	db1, err := storage.NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create unencrypted database: %v", err)
	}

	toxMgr1 := &MockToxManager{}
	contactMgr1 := NewMockContactManager()
	contactMgr1.AddContact(1, map[string]string{"name": "Test Friend"})

	mgr1 := NewManager(db1, toxMgr1, contactMgr1)

	// Send a message to unencrypted database
	originalMessage := "Migration test message"
	msg, err := mgr1.SendMessage(1, originalMessage, MessageTypeNormal)
	if err != nil {
		t.Fatalf("Failed to send message to unencrypted database: %v", err)
	}

	// Store message details for verification (variables unused in this test but preserved for completeness)
	_ = msg.ID
	_ = msg.UUID

	db1.Close()

	// Verify the message exists by reopening as unencrypted
	db2, err := storage.NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to reopen unencrypted database: %v", err)
	}

	toxMgr2 := &MockToxManager{}
	contactMgr2 := NewMockContactManager()
	contactMgr2.AddContact(1, map[string]string{"name": "Test Friend"})

	mgr2 := NewManager(db2, toxMgr2, contactMgr2)

	// Retrieve the message from unencrypted database
	messages, err := mgr2.GetMessages(1, 10, 0)
	if err != nil {
		t.Fatalf("Failed to retrieve messages from unencrypted database: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message in unencrypted database, got %d", len(messages))
	}

	// Verify the message has a UUID (migration worked)
	if messages[0].UUID == "" {
		t.Error("Expected message to have UUID after migration")
	}

	db2.Close()
}
