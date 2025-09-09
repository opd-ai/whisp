package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

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
