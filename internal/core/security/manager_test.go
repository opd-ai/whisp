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

	// Should be in hex format
	if len(dbKey) != 64 { // 64 hex chars
		t.Errorf("Expected database key length 64, got %d", len(dbKey))
	}

	// Should be valid hex
	for i, c := range dbKey {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("Database key should be lowercase hex, got invalid char at position %d: %c", i, c)
		}
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

func TestSecureStorage(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Generate and set master key for testing
	masterKey, err := manager.GenerateMasterKey()
	if err != nil {
		t.Fatalf("Failed to generate master key: %v", err)
	}
	manager.SetMasterKey(masterKey)

	// Test basic store and retrieve
	testKey := "test_key"
	testValue := "test_value"

	err = manager.SecureStore(testKey, testValue)
	if err != nil {
		t.Fatalf("Failed to store value: %v", err)
	}

	retrievedValue, err := manager.SecureRetrieve(testKey)
	if err != nil {
		t.Fatalf("Failed to retrieve value: %v", err)
	}

	if retrievedValue != testValue {
		t.Errorf("Expected value %s, got %s", testValue, retrievedValue)
	}

	// Test secure delete
	err = manager.SecureDelete(testKey)
	if err != nil {
		t.Fatalf("Failed to delete value: %v", err)
	}

	// Verify value is deleted
	_, err = manager.SecureRetrieve(testKey)
	if err == nil {
		t.Error("Expected error when retrieving deleted value")
	}
}

func TestSecureStorageEmptyKey(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Test empty key scenarios
	err = manager.SecureStore("", "value")
	if err == nil {
		t.Error("Expected error when storing with empty key")
	}

	_, err = manager.SecureRetrieve("")
	if err == nil {
		t.Error("Expected error when retrieving with empty key")
	}

	err = manager.SecureDelete("")
	if err == nil {
		t.Error("Expected error when deleting with empty key")
	}
}

func TestMasterKeyStorage(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Generate test master key
	originalKey, err := manager.GenerateMasterKey()
	if err != nil {
		t.Fatalf("Failed to generate master key: %v", err)
	}

	// Store master key
	err = manager.StoreMasterKey(originalKey)
	if err != nil {
		t.Fatalf("Failed to store master key: %v", err)
	}

	// Load master key
	loadedKey, err := manager.LoadMasterKey()
	if err != nil {
		t.Fatalf("Failed to load master key: %v", err)
	}

	// Verify keys match
	if !bytes.Equal(originalKey, loadedKey) {
		t.Error("Loaded master key does not match original")
	}

	// Test delete master key
	err = manager.DeleteMasterKey()
	if err != nil {
		t.Fatalf("Failed to delete master key: %v", err)
	}

	// Verify key is deleted
	_, err = manager.LoadMasterKey()
	if err == nil {
		t.Error("Expected error when loading deleted master key")
	}
}

func TestMasterKeyStorageEmptyKey(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Test empty master key
	err = manager.StoreMasterKey([]byte{})
	if err == nil {
		t.Error("Expected error when storing empty master key")
	}

	err = manager.StoreMasterKey(nil)
	if err == nil {
		t.Error("Expected error when storing nil master key")
	}
}

func TestSecureStorageFallback(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Generate and set master key for encrypted file fallback
	masterKey, err := manager.GenerateMasterKey()
	if err != nil {
		t.Fatalf("Failed to generate master key: %v", err)
	}
	manager.SetMasterKey(masterKey)

	// Test file fallback methods directly
	testKey := "fallback_test"
	testValue := "fallback_value"

	err = manager.secureFileStore(testKey, testValue)
	if err != nil {
		t.Fatalf("Failed to store in file fallback: %v", err)
	}

	retrievedValue, err := manager.secureFileRetrieve(testKey)
	if err != nil {
		t.Fatalf("Failed to retrieve from file fallback: %v", err)
	}

	if retrievedValue != testValue {
		t.Errorf("Expected value %s, got %s", testValue, retrievedValue)
	}

	// Test delete
	err = manager.secureFileDelete(testKey)
	if err != nil {
		t.Fatalf("Failed to delete from file fallback: %v", err)
	}

	// Verify file is deleted
	_, err = manager.secureFileRetrieve(testKey)
	if err == nil {
		t.Error("Expected error when retrieving deleted file")
	}
}

func TestSecureStorageFallbackUnlocked(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Test fallback without unlocked manager
	err = manager.secureFileStore("key", "value")
	if err == nil {
		t.Error("Expected error when using file fallback without unlocked manager")
	}

	_, err = manager.secureFileRetrieve("key")
	if err == nil {
		t.Error("Expected error when retrieving from file fallback without unlocked manager")
	}
}

func TestIsSecureStorageAvailable(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Note: This test may fail in CI environments without GUI/keyring services
	// The method should return either true or false without error
	available := manager.IsSecureStorageAvailable()
	t.Logf("Secure storage available: %v", available)

	// Just verify the method doesn't panic or cause errors
	// The actual availability depends on the platform and environment
}

func TestSecureStorageCleanup(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Setup master key
	masterKey, err := manager.GenerateMasterKey()
	if err != nil {
		t.Fatalf("Failed to generate master key: %v", err)
	}
	manager.SetMasterKey(masterKey)

	// Store some test data
	err = manager.SecureStore("cleanup_test", "test_value")
	if err != nil {
		t.Fatalf("Failed to store test data: %v", err)
	}

	// Cleanup should clear the master key
	manager.Cleanup()

	// Verify manager is locked
	if manager.isUnlocked {
		t.Error("Manager should be locked after cleanup")
	}

	if manager.masterKey != nil {
		t.Error("Master key should be nil after cleanup")
	}

	// Should not be able to use file fallback after cleanup
	err = manager.secureFileStore("post_cleanup", "value")
	if err == nil {
		t.Error("Expected error when using file fallback after cleanup")
	}
}

func TestSecureStorageIntegration(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Generate and set master key
	masterKey, err := manager.GenerateMasterKey()
	if err != nil {
		t.Fatalf("Failed to generate master key: %v", err)
	}
	manager.SetMasterKey(masterKey)

	// Test multiple key-value pairs
	testData := map[string]string{
		"config_theme":    "dark",
		"config_language": "en",
		"user_token":      "abc123def456",
		"backup_key":      "backup_secret_key",
		"session_id":      "session_12345",
	}

	// Store all data
	for key, value := range testData {
		if err := manager.SecureStore(key, value); err != nil {
			t.Fatalf("Failed to store %s: %v", key, err)
		}
	}

	// Retrieve and verify all data
	for key, expectedValue := range testData {
		actualValue, err := manager.SecureRetrieve(key)
		if err != nil {
			t.Fatalf("Failed to retrieve %s: %v", key, err)
		}
		if actualValue != expectedValue {
			t.Errorf("For key %s: expected %s, got %s", key, expectedValue, actualValue)
		}
	}

	// Test overwriting existing values
	newValue := "new_dark_theme"
	if err := manager.SecureStore("config_theme", newValue); err != nil {
		t.Fatalf("Failed to overwrite value: %v", err)
	}

	retrievedValue, err := manager.SecureRetrieve("config_theme")
	if err != nil {
		t.Fatalf("Failed to retrieve overwritten value: %v", err)
	}
	if retrievedValue != newValue {
		t.Errorf("Expected overwritten value %s, got %s", newValue, retrievedValue)
	}

	// Clean up
	for key := range testData {
		if err := manager.SecureDelete(key); err != nil {
			t.Errorf("Failed to delete %s: %v", key, err)
		}
	}
}

func TestSecureFileStorageDirectoryCreation(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Generate and set master key
	masterKey, err := manager.GenerateMasterKey()
	if err != nil {
		t.Fatalf("Failed to generate master key: %v", err)
	}
	manager.SetMasterKey(masterKey)

	// Store data which should create keystore directory
	err = manager.secureFileStore("test_key", "test_value")
	if err != nil {
		t.Fatalf("Failed to store in file fallback: %v", err)
	}

	// Verify keystore directory was created
	keystoreDir := filepath.Join(tempDir, "security", "keystore")
	if _, err := os.Stat(keystoreDir); os.IsNotExist(err) {
		t.Error("Keystore directory was not created")
	}

	// Verify encrypted file exists
	keyFile := filepath.Join(keystoreDir, "test_key.enc")
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		t.Error("Encrypted key file was not created")
	}
}

func TestSecureFileDeleteNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Try to delete non-existent file - should not error
	err = manager.secureFileDelete("non_existent_key")
	if err != nil {
		t.Errorf("Unexpected error when deleting non-existent file: %v", err)
	}
}

func TestMasterKeyHexEncoding(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Test with known key to verify hex encoding/decoding
	testKey := []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF}

	err = manager.StoreMasterKey(testKey)
	if err != nil {
		t.Fatalf("Failed to store master key: %v", err)
	}

	loadedKey, err := manager.LoadMasterKey()
	if err != nil {
		t.Fatalf("Failed to load master key: %v", err)
	}

	if !bytes.Equal(testKey, loadedKey) {
		t.Errorf("Hex encoding/decoding failed. Original: %x, Loaded: %x", testKey, loadedKey)
	}
}

func TestSecureStorageAvailabilityCheck(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Test the availability check multiple times to ensure it's stable
	available1 := manager.IsSecureStorageAvailable()
	available2 := manager.IsSecureStorageAvailable()

	if available1 != available2 {
		t.Error("Secure storage availability should be consistent across calls")
	}

	// The actual value depends on the test environment
	t.Logf("Secure storage available: %v", available1)
}

func TestDeriveKey(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	password := []byte("test_password")
	salt := []byte("test_salt_1234567890abcdef1234567890abcdef")

	key, err := manager.DeriveKey(password, salt)
	if err != nil {
		t.Fatalf("Failed to derive key: %v", err)
	}

	if len(key) != 32 {
		t.Errorf("Expected derived key length 32, got %d", len(key))
	}

	// Test with same inputs should produce same key
	key2, err := manager.DeriveKey(password, salt)
	if err != nil {
		t.Fatalf("Failed to derive key second time: %v", err)
	}

	if !bytes.Equal(key, key2) {
		t.Error("Same password and salt should produce same derived key")
	}

	// Different password should produce different key
	key3, err := manager.DeriveKey([]byte("different_password"), salt)
	if err != nil {
		t.Fatalf("Failed to derive key with different password: %v", err)
	}

	if bytes.Equal(key, key3) {
		t.Error("Different passwords should produce different derived keys")
	}
}

func TestGetDatabaseKeyBytes(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Generate and set master key
	masterKey, err := manager.GenerateMasterKey()
	if err != nil {
		t.Fatalf("Failed to generate master key: %v", err)
	}
	manager.SetMasterKey(masterKey)

	// Get database key bytes
	keyBytes, err := manager.GetDatabaseKeyBytes()
	if err != nil {
		t.Fatalf("Failed to get database key bytes: %v", err)
	}

	if len(keyBytes) != 32 {
		t.Errorf("Expected database key length 32, got %d", len(keyBytes))
	}

	// Test multiple calls return same key
	keyBytes2, err := manager.GetDatabaseKeyBytes()
	if err != nil {
		t.Fatalf("Failed to get database key bytes second time: %v", err)
	}

	if !bytes.Equal(keyBytes, keyBytes2) {
		t.Error("Multiple calls should return same database key bytes")
	}

	// Test with unlocked manager
	manager.Cleanup()
	_, err = manager.GetDatabaseKeyBytes()
	if err == nil {
		t.Error("Expected error when getting database key bytes from unlocked manager")
	}
}

func TestSecureStorageErrorHandling(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Test storing with invalid directory permissions (simulate error)
	// This is a bit tricky to test portably, so we'll test other error paths

	// Test with unlocked manager for file fallback
	err = manager.secureFileStore("test", "value")
	if err == nil {
		t.Error("Expected error when using file store without unlocked manager")
	}

	// Test invalid hex in LoadMasterKey by directly manipulating storage
	// First, set up a master key for testing
	masterKey, _ := manager.GenerateMasterKey()
	manager.SetMasterKey(masterKey)

	// Store invalid hex data
	err = manager.SecureStore(MasterKeyName, "invalid_hex_data")
	if err != nil {
		t.Fatalf("Failed to store invalid hex: %v", err)
	}

	// Try to load - should fail
	_, err = manager.LoadMasterKey()
	if err == nil {
		t.Error("Expected error when loading invalid hex master key")
	}

	// Clean up
	manager.SecureDelete(MasterKeyName)
}
