package storage

import (
	"os"
	"path/filepath"
	"testing"
)

// MockSecurityManager for testing database encryption
type MockSecurityManager struct {
	dbKey string
	err   error
}

func (m *MockSecurityManager) GetDatabaseKey() (string, error) {
	return m.dbKey, m.err
}

func TestNewDatabase(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if db.GetPath() != dbPath {
		t.Errorf("Expected path %s, got %s", dbPath, db.GetPath())
	}

	if db.IsEncrypted() {
		t.Error("Database should not be encrypted when created with NewDatabase")
	}

	// Verify database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}
}

func TestNewDatabaseWithEncryption(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "encrypted.db")

	// Test with valid security manager
	securityManager := &MockSecurityManager{
		dbKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	}

	db, err := NewDatabaseWithEncryption(dbPath, securityManager)
	if err != nil {
		t.Fatalf("Failed to create encrypted database: %v", err)
	}
	defer db.Close()

	if !db.IsEncrypted() {
		t.Error("Database should be encrypted when security manager is provided")
	}

	// Verify database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}
}

func TestNewDatabaseWithEncryptionErrors(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "error.db")

	// Test with security manager that returns error
	securityManager := &MockSecurityManager{
		err: ErrorMockSecurityError{},
	}

	_, err := NewDatabaseWithEncryption(dbPath, securityManager)
	if err == nil {
		t.Error("Expected error when security manager fails to provide key")
	}
}

// Custom error type for testing
type ErrorMockSecurityError struct{}

func (e ErrorMockSecurityError) Error() string {
	return "mock security error"
}

func TestDatabaseOperations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "operations.db")

	// Create encrypted database
	securityManager := &MockSecurityManager{
		dbKey: "fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210",
	}

	db, err := NewDatabaseWithEncryption(dbPath, securityManager)
	if err != nil {
		t.Fatalf("Failed to create encrypted database: %v", err)
	}
	defer db.Close()

	// Test basic database operations
	testQueries := []string{
		"INSERT INTO settings (key, value, updated_at) VALUES ('test_key', 'test_value', datetime('now'))",
		"SELECT value FROM settings WHERE key = 'test_key'",
	}

	// Insert test data
	_, err = db.Exec(testQueries[0])
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Query test data
	var value string
	err = db.QueryRow(testQueries[1]).Scan(&value)
	if err != nil {
		t.Fatalf("Failed to query test data: %v", err)
	}

	if value != "test_value" {
		t.Errorf("Expected 'test_value', got %s", value)
	}
}

func TestDatabaseSchema(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "schema.db")

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Verify all expected tables exist
	expectedTables := []string{"contacts", "messages", "settings", "file_transfers"}

	for _, table := range expectedTables {
		var count int
		query := "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
		err := db.QueryRow(query, table).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check for table %s: %v", table, err)
		}

		if count != 1 {
			t.Errorf("Expected table %s to exist", table)
		}
	}

	// Verify indexes exist
	expectedIndexes := []string{
		"idx_messages_friend_id",
		"idx_messages_timestamp",
		"idx_contacts_friend_id",
		"idx_file_transfers_friend_id",
	}

	for _, index := range expectedIndexes {
		var count int
		query := "SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?"
		err := db.QueryRow(query, index).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check for index %s: %v", index, err)
		}

		if count != 1 {
			t.Errorf("Expected index %s to exist", index)
		}
	}
}

func TestEncryptedDatabaseCompatibility(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "compat.db")

	securityManager := &MockSecurityManager{
		dbKey: "1111111111111111111111111111111111111111111111111111111111111111",
	}

	// Create encrypted database and add some data
	db1, err := NewDatabaseWithEncryption(dbPath, securityManager)
	if err != nil {
		t.Fatalf("Failed to create first database instance: %v", err)
	}

	// Insert test data
	_, err = db1.Exec("INSERT INTO settings (key, value, updated_at) VALUES ('compat_test', 'data123', datetime('now'))")
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	db1.Close()

	// Open the same database again with same key
	db2, err := NewDatabaseWithEncryption(dbPath, securityManager)
	if err != nil {
		t.Fatalf("Failed to reopen encrypted database: %v", err)
	}
	defer db2.Close()

	// Verify we can read the data
	var value string
	err = db2.QueryRow("SELECT value FROM settings WHERE key = 'compat_test'").Scan(&value)
	if err != nil {
		t.Fatalf("Failed to read data from reopened database: %v", err)
	}

	if value != "data123" {
		t.Errorf("Expected 'data123', got %s", value)
	}
}

func TestWrongKeyDetection(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "wrongkey.db")

	// Create database with one key
	securityManager1 := &MockSecurityManager{
		dbKey: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}

	db1, err := NewDatabaseWithEncryption(dbPath, securityManager1)
	if err != nil {
		t.Fatalf("Failed to create encrypted database: %v", err)
	}

	// Add some data
	_, err = db1.Exec("INSERT INTO settings (key, value, updated_at) VALUES ('test', 'value', datetime('now'))")
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	db1.Close()

	// Try to open with wrong key
	securityManager2 := &MockSecurityManager{
		dbKey: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	}

	_, err = NewDatabaseWithEncryption(dbPath, securityManager2)
	if err == nil {
		t.Error("Expected error when opening encrypted database with wrong key")
	}
}
