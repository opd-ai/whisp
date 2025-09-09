package security

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"os"
	"path/filepath"
	"sync"

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

// Cleanup cleans up security resources
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
