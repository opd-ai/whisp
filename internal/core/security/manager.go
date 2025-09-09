package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/zalando/go-keyring"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/scrypt"
)

// Manager handles security operations
type Manager struct {
	dataDir    string
	mu         sync.RWMutex
	masterKey  []byte
	isUnlocked bool
}

// NewManager creates a new security manager
func NewManager(dataDir string) (*Manager, error) {
	m := &Manager{
		dataDir: dataDir,
	}

	// Ensure security directory exists
	securityDir := filepath.Join(dataDir, "security")
	if err := os.MkdirAll(securityDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create security directory: %w", err)
	}

	return m, nil
}

// GenerateMasterKey generates a new master key
func (m *Manager) GenerateMasterKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate master key: %w", err)
	}
	return key, nil
}

// DeriveKey derives a key from password using scrypt
func (m *Manager) DeriveKey(password, salt []byte) ([]byte, error) {
	return scrypt.Key(password, salt, 32768, 8, 1, 32)
}

// GenerateSalt generates a random salt
func (m *Manager) GenerateSalt() ([]byte, error) {
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

// HashPassword hashes a password with salt
func (m *Manager) HashPassword(password string) ([]byte, []byte, error) {
	salt, err := m.GenerateSalt()
	if err != nil {
		return nil, nil, err
	}

	hash := sha256.Sum256(append([]byte(password), salt...))
	return hash[:], salt, nil
}

// VerifyPassword verifies a password against hash and salt
func (m *Manager) VerifyPassword(password string, hash, salt []byte) bool {
	testHash := sha256.Sum256(append([]byte(password), salt...))
	return subtle.ConstantTimeCompare(hash, testHash[:]) == 1
}

// SetMasterKey sets the master key for encryption operations
func (m *Manager) SetMasterKey(key []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear existing key
	if m.masterKey != nil {
		for i := range m.masterKey {
			m.masterKey[i] = 0
		}
	}

	// Set new key
	m.masterKey = make([]byte, len(key))
	copy(m.masterKey, key)
	m.isUnlocked = true
}

// GetMasterKey returns a copy of the master key (for internal use)
func (m *Manager) GetMasterKey() []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.isUnlocked || m.masterKey == nil {
		return nil
	}

	key := make([]byte, len(m.masterKey))
	copy(key, m.masterKey)
	return key
}

// DeriveContextKey derives a key for a specific context using HKDF
func (m *Manager) DeriveContextKey(context string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.isUnlocked || m.masterKey == nil {
		return nil, fmt.Errorf("security manager not unlocked")
	}

	// Use HKDF to derive context-specific key
	hkdf := hkdf.New(sha256.New, m.masterKey, nil, []byte(context))

	key := make([]byte, 32) // 256-bit key
	if _, err := io.ReadFull(hkdf, key); err != nil {
		return nil, fmt.Errorf("failed to derive context key: %w", err)
	}

	return key, nil
}

// EncryptData encrypts data using AES-256-GCM with a context-derived key
func (m *Manager) EncryptData(data []byte, context string) ([]byte, error) {
	key, err := m.DeriveContextKey(context)
	if err != nil {
		return nil, err
	}
	defer func() {
		// Clear key from memory
		for i := range key {
			key[i] = 0
		}
	}()

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt data (nonce is prepended to ciphertext)
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// DecryptData decrypts data using AES-256-GCM with a context-derived key
func (m *Manager) DecryptData(encryptedData []byte, context string) ([]byte, error) {
	key, err := m.DeriveContextKey(context)
	if err != nil {
		return nil, err
	}
	defer func() {
		// Clear key from memory
		for i := range key {
			key[i] = 0
		}
	}()

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("encrypted data too short")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}

// GetDatabaseKey derives a database encryption key in hex format
func (m *Manager) GetDatabaseKey() (string, error) {
	key, err := m.DeriveContextKey("database")
	if err != nil {
		return "", err
	}
	defer func() {
		// Clear key from memory
		for i := range key {
			key[i] = 0
		}
	}()

	// Return key as hex string for SQLCipher PRAGMA key
	return fmt.Sprintf("%x", key), nil
}

// GetDatabaseKeyBytes derives a database encryption key as raw bytes
func (m *Manager) GetDatabaseKeyBytes() ([]byte, error) {
	key, err := m.DeriveContextKey("database")
	if err != nil {
		return nil, err
	}

	// Return a copy of the key bytes
	result := make([]byte, len(key))
	copy(result, key)

	// Clear original key from memory
	for i := range key {
		key[i] = 0
	}

	return result, nil
}

// Secure storage constants
const (
	// KeyringService is the service name used for keyring operations
	KeyringService = "com.opd-ai.whisp"
	// MasterKeyName is the key name for the master key in secure storage
	MasterKeyName = "master_key"
	// ConfigKeyPrefix is the prefix for configuration keys in secure storage
	ConfigKeyPrefix = "config_"
)

// SecureStore stores a key-value pair in platform-specific secure storage.
// This method attempts to use platform-specific secure storage first (Keychain on macOS,
// Credential Manager on Windows, Secret Service on Linux), and falls back to
// encrypted file storage if platform storage is unavailable.
//
// The value is stored as-is in platform storage, or encrypted with AES-256-GCM
// and stored in a file when using the fallback method.
//
// Parameters:
//   - key: The key to store the value under (cannot be empty)
//   - value: The value to store securely
//
// Returns an error if the key is empty or if both platform storage and file fallback fail.
func (m *Manager) SecureStore(key, value string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	// Try to store in platform-specific secure storage first
	err := keyring.Set(KeyringService, key, value)
	if err != nil {
		// If keyring fails, fall back to encrypted file storage
		return m.secureFileStore(key, value)
	}

	return nil
}

// SecureRetrieve retrieves a value from platform-specific secure storage.
// This method attempts to retrieve from platform-specific secure storage first,
// and falls back to encrypted file storage if platform storage is unavailable.
//
// Parameters:
//   - key: The key to retrieve the value for (cannot be empty)
//
// Returns the stored value and an error if the key is empty, not found, or
// if both platform storage and file fallback fail.
func (m *Manager) SecureRetrieve(key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("key cannot be empty")
	}

	// Try to retrieve from platform-specific secure storage first
	value, err := keyring.Get(KeyringService, key)
	if err != nil {
		// If keyring fails, try encrypted file storage fallback
		return m.secureFileRetrieve(key)
	}

	return value, nil
}

// SecureDelete removes a key from platform-specific secure storage
func (m *Manager) SecureDelete(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	// Try to delete from platform-specific secure storage first
	err := keyring.Delete(KeyringService, key)
	if err != nil {
		// If keyring fails, try to delete from file storage fallback
		return m.secureFileDelete(key)
	}

	return nil
}

// StoreMasterKey stores the master key in secure storage.
// The master key is encoded as a hexadecimal string before storage to ensure
// safe handling across different storage backends.
//
// This method is typically used during application setup or when the master key
// needs to be persisted for future application sessions.
//
// Parameters:
//   - masterKey: The 32-byte master key to store (cannot be empty)
//
// Returns an error if the master key is empty or if storage fails.
func (m *Manager) StoreMasterKey(masterKey []byte) error {
	if len(masterKey) == 0 {
		return fmt.Errorf("master key cannot be empty")
	}

	// Convert to hex string for storage
	keyHex := hex.EncodeToString(masterKey)

	return m.SecureStore(MasterKeyName, keyHex)
}

// LoadMasterKey loads the master key from secure storage.
// The key is retrieved as a hexadecimal string and decoded back to bytes.
//
// This method is typically used during application startup to restore the
// master key for cryptographic operations.
//
// Returns the 32-byte master key and an error if retrieval or decoding fails.
func (m *Manager) LoadMasterKey() ([]byte, error) {
	keyHex, err := m.SecureRetrieve(MasterKeyName)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve master key: %w", err)
	}

	// Convert from hex string
	masterKey, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode master key: %w", err)
	}

	return masterKey, nil
}

// DeleteMasterKey removes the master key from secure storage
func (m *Manager) DeleteMasterKey() error {
	return m.SecureDelete(MasterKeyName)
}

// secureFileStore stores data in encrypted file as fallback
func (m *Manager) secureFileStore(key, value string) error {
	if !m.isUnlocked || m.masterKey == nil {
		return fmt.Errorf("security manager not unlocked")
	}

	// Encrypt the value
	encryptedValue, err := m.EncryptData([]byte(value), "secure_storage")
	if err != nil {
		return fmt.Errorf("failed to encrypt value: %w", err)
	}

	// Store in secure directory
	secureDir := filepath.Join(m.dataDir, "security", "keystore")
	if err := os.MkdirAll(secureDir, 0700); err != nil {
		return fmt.Errorf("failed to create secure directory: %w", err)
	}

	keyFile := filepath.Join(secureDir, key+".enc")
	if err := os.WriteFile(keyFile, encryptedValue, 0600); err != nil {
		return fmt.Errorf("failed to write encrypted file: %w", err)
	}

	return nil
}

// secureFileRetrieve retrieves data from encrypted file as fallback
func (m *Manager) secureFileRetrieve(key string) (string, error) {
	if !m.isUnlocked || m.masterKey == nil {
		return "", fmt.Errorf("security manager not unlocked")
	}

	keyFile := filepath.Join(m.dataDir, "security", "keystore", key+".enc")
	encryptedValue, err := os.ReadFile(keyFile)
	if err != nil {
		return "", fmt.Errorf("failed to read encrypted file: %w", err)
	}

	// Decrypt the value
	decryptedValue, err := m.DecryptData(encryptedValue, "secure_storage")
	if err != nil {
		return "", fmt.Errorf("failed to decrypt value: %w", err)
	}

	return string(decryptedValue), nil
}

// secureFileDelete removes encrypted file as fallback
func (m *Manager) secureFileDelete(key string) error {
	keyFile := filepath.Join(m.dataDir, "security", "keystore", key+".enc")
	err := os.Remove(keyFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete encrypted file: %w", err)
	}

	return nil
}

// IsSecureStorageAvailable checks if platform-specific secure storage is available.
// This method performs a test operation by storing and retrieving a test value
// to verify that the platform's secure storage system is functional.
//
// Platform support:
//   - Windows: Windows Credential Manager
//   - macOS: Keychain Services
//   - Linux: Secret Service API (GNOME Keyring, KDE Wallet)
//
// Returns true if platform-specific secure storage is available and functional,
// false if only encrypted file fallback can be used.
func (m *Manager) IsSecureStorageAvailable() bool {
	// Test by trying to store and retrieve a test value
	testKey := "test_availability"
	testValue := "test"

	// Clean up any existing test key
	keyring.Delete(KeyringService, testKey)

	// Try to store a test value
	if err := keyring.Set(KeyringService, testKey, testValue); err != nil {
		return false
	}

	// Try to retrieve the test value
	retrievedValue, err := keyring.Get(KeyringService, testKey)
	if err != nil || retrievedValue != testValue {
		keyring.Delete(KeyringService, testKey) // Clean up on failure
		return false
	}

	// Clean up test key
	keyring.Delete(KeyringService, testKey)

	return true
} // Cleanup cleans up security resources
func (m *Manager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear master key from memory
	if m.masterKey != nil {
		for i := range m.masterKey {
			m.masterKey[i] = 0
		}
		m.masterKey = nil
	}
	m.isUnlocked = false
}
