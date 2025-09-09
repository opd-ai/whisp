package security

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	tempDir := t.TempDir()
	
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}
	
	if manager.dataDir != tempDir {
		t.Errorf("Expected dataDir %s, got %s", tempDir, manager.dataDir)
	}
	
	// Check that security directory was created
	securityDir := filepath.Join(tempDir, "security")
	if _, err := os.Stat(securityDir); os.IsNotExist(err) {
		t.Error("Security directory was not created")
	}
}

func TestGenerateMasterKey(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}
	
	key1, err := manager.GenerateMasterKey()
	if err != nil {
		t.Fatalf("Failed to generate master key: %v", err)
	}
	
	if len(key1) != 32 {
		t.Errorf("Expected key length 32, got %d", len(key1))
	}
	
	// Generate another key and ensure they're different
	key2, err := manager.GenerateMasterKey()
	if err != nil {
		t.Fatalf("Failed to generate second master key: %v", err)
	}
	
	if bytes.Equal(key1, key2) {
		t.Error("Generated keys should be different")
	}
}

func TestSetAndGetMasterKey(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}
	
	originalKey := []byte("test-key-32-bytes-long-123456789")
	if len(originalKey) != 32 {
		t.Fatalf("Test key should be 32 bytes, got %d", len(originalKey))
	}
	
	// Initially, manager should not be unlocked
	key := manager.GetMasterKey()
	if key != nil {
		t.Error("Expected nil key when manager is not unlocked")
	}
	
	// Set the master key
	manager.SetMasterKey(originalKey)
	
	// Get the key back
	retrievedKey := manager.GetMasterKey()
	if retrievedKey == nil {
		t.Fatal("Expected non-nil key after setting master key")
	}
	
	if !bytes.Equal(originalKey, retrievedKey) {
		t.Error("Retrieved key doesn't match original key")
	}
	
	// Ensure we get a copy, not the original
	retrievedKey[0] = 0x00
	if originalKey[0] == 0x00 {
		t.Error("Modifying retrieved key should not affect original")
	}
}

func TestDeriveContextKey(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}
	
	masterKey := []byte("test-master-key-32-bytes-long-12")
	if len(masterKey) != 32 {
		t.Fatalf("Master key should be 32 bytes, got %d", len(masterKey))
	}
	
	// Should fail before setting master key
	_, err = manager.DeriveContextKey("test-context")
	if err == nil {
		t.Error("Expected error when deriving key without master key")
	}
	
	// Set master key
	manager.SetMasterKey(masterKey)
	
	// Derive context keys
	key1, err := manager.DeriveContextKey("database")
	if err != nil {
		t.Fatalf("Failed to derive database key: %v", err)
	}
	
	key2, err := manager.DeriveContextKey("files")
	if err != nil {
		t.Fatalf("Failed to derive files key: %v", err)
	}
	
	// Keys should be 32 bytes
	if len(key1) != 32 {
		t.Errorf("Expected key length 32, got %d", len(key1))
	}
	
	// Different contexts should produce different keys
	if bytes.Equal(key1, key2) {
		t.Error("Different contexts should produce different keys")
	}
	
	// Same context should produce same key
	key1Again, err := manager.DeriveContextKey("database")
	if err != nil {
		t.Fatalf("Failed to derive database key again: %v", err)
	}
	
	if !bytes.Equal(key1, key1Again) {
		t.Error("Same context should produce same key")
	}
}

func TestEncryptDecryptData(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}
	
	masterKey := []byte("test-master-key-32-bytes-long-12")
	manager.SetMasterKey(masterKey)
	
	testData := []byte("This is sensitive data that needs encryption!")
	context := "test-context"
	
	// Encrypt data
	encryptedData, err := manager.EncryptData(testData, context)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}
	
	// Encrypted data should be different and longer (due to nonce and auth tag)
	if bytes.Equal(testData, encryptedData) {
		t.Error("Encrypted data should be different from original")
	}
	
	if len(encryptedData) <= len(testData) {
		t.Error("Encrypted data should be longer than original (includes nonce and auth tag)")
	}
	
	// Decrypt data
	decryptedData, err := manager.DecryptData(encryptedData, context)
	if err != nil {
		t.Fatalf("Failed to decrypt data: %v", err)
	}
	
	// Decrypted data should match original
	if !bytes.Equal(testData, decryptedData) {
		t.Error("Decrypted data doesn't match original")
	}
	
	// Decryption with wrong context should fail
	_, err = manager.DecryptData(encryptedData, "wrong-context")
	if err == nil {
		t.Error("Expected error when decrypting with wrong context")
	}
	
	// Decryption of tampered data should fail
	tamperedData := make([]byte, len(encryptedData))
	copy(tamperedData, encryptedData)
	tamperedData[len(tamperedData)-1] ^= 0x01 // Flip last bit
	
	_, err = manager.DecryptData(tamperedData, context)
	if err == nil {
		t.Error("Expected error when decrypting tampered data")
	}
}

func TestGetDatabaseKey(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}
	
	// Should fail before setting master key
	_, err = manager.GetDatabaseKey()
	if err == nil {
		t.Error("Expected error when getting database key without master key")
	}
	
	masterKey := []byte("test-master-key-32-bytes-long-12")
	manager.SetMasterKey(masterKey)
	
	// Get database key
	dbKey, err := manager.GetDatabaseKey()
	if err != nil {
		t.Fatalf("Failed to get database key: %v", err)
	}
	
	// Should be in hex format with x' prefix and ' suffix
	if len(dbKey) != 67 { // x' + 64 hex chars + '
		t.Errorf("Expected database key length 67, got %d", len(dbKey))
	}
	
	if dbKey[:2] != "x'" || dbKey[len(dbKey)-1:] != "'" {
		t.Errorf("Database key should be in format x'...', got %s", dbKey)
	}
	
	// Same call should produce same key
	dbKey2, err := manager.GetDatabaseKey()
	if err != nil {
		t.Fatalf("Failed to get database key second time: %v", err)
	}
	
	if dbKey != dbKey2 {
		t.Error("Database key should be consistent")
	}
}

func TestCleanup(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}
	
	masterKey := []byte("test-master-key-32-bytes-long-12")
	manager.SetMasterKey(masterKey)
	
	// Verify key is available
	key := manager.GetMasterKey()
	if key == nil {
		t.Fatal("Expected master key to be available")
	}
	
	// Cleanup
	manager.Cleanup()
	
	// Key should no longer be available
	key = manager.GetMasterKey()
	if key != nil {
		t.Error("Expected master key to be cleared after cleanup")
	}
	
	// Operations requiring master key should fail
	_, err = manager.GetDatabaseKey()
	if err == nil {
		t.Error("Expected error when getting database key after cleanup")
	}
}

func TestPasswordHashing(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}
	
	password := "test-password-123"
	
	// Hash password
	hash, salt, err := manager.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	
	if len(hash) != 32 {
		t.Errorf("Expected hash length 32, got %d", len(hash))
	}
	
	if len(salt) != 32 {
		t.Errorf("Expected salt length 32, got %d", len(salt))
	}
	
	// Verify correct password
	if !manager.VerifyPassword(password, hash, salt) {
		t.Error("Password verification failed for correct password")
	}
	
	// Verify incorrect password
	if manager.VerifyPassword("wrong-password", hash, salt) {
		t.Error("Password verification should fail for incorrect password")
	}
	
	// Different calls should produce different salts/hashes
	hash2, salt2, err := manager.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password second time: %v", err)
	}
	
	if bytes.Equal(hash, hash2) || bytes.Equal(salt, salt2) {
		t.Error("Multiple hashing of same password should produce different results")
	}
}
