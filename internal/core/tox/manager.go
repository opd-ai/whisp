package tox

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/opd-ai/toxcore"
)

// Config holds Tox manager configuration
type Config struct {
	DataDir string
	Debug   bool
}

// Manager manages the Tox instance and protocol operations
type Manager struct {
	tox      *toxcore.Tox
	config   *Config
	mu       sync.RWMutex
	running  bool
	saveFile string

	// Event callbacks
	onFriendRequest       func([32]byte, string)
	onFriendMessage       func(uint32, string)
	onFriendConnectionStatus func(uint32, toxcore.ConnectionStatus)
	onFriendStatus        func(uint32, toxcore.UserStatus)
	onFriendName          func(uint32, string)
	onFriendStatusMessage func(uint32, string)
}

// NewManager creates a new Tox manager
func NewManager(config *Config) (*Manager, error) {
	m := &Manager{
		config:   config,
		saveFile: filepath.Join(config.DataDir, "tox.save"),
	}

	if err := m.initializeTox(); err != nil {
		return nil, fmt.Errorf("failed to initialize Tox: %w", err)
	}

	return m, nil
}

// initializeTox initializes the Tox instance
func (m *Manager) initializeTox() error {
	options := toxcore.NewOptions()
	options.UDPEnabled = true
	options.LocalDiscoveryEnabled = true
	options.HolePunchingEnabled = true

	// Try to load existing save data
	var savedata []byte
	if data, err := ioutil.ReadFile(m.saveFile); err == nil {
		savedata = data
		log.Printf("Loaded Tox save data from %s", m.saveFile)
	} else {
		log.Printf("No existing save data found, creating new Tox instance")
	}

	// Create Tox instance
	var err error
	if len(savedata) > 0 {
		m.tox, err = toxcore.NewFromSavedata(options, savedata)
	} else {
		m.tox, err = toxcore.New(options)
	}

	if err != nil {
		return fmt.Errorf("failed to create Tox instance: %w", err)
	}

	// Set up callbacks
	m.setupCallbacks()

	// Bootstrap to the network
	if err := m.bootstrap(); err != nil {
		log.Printf("Warning: Bootstrap failed: %v", err)
	}

	// Save initial state
	if err := m.save(); err != nil {
		log.Printf("Warning: Failed to save initial state: %v", err)
	}

	log.Printf("Tox initialized. ID: %s", m.tox.SelfGetAddress())
	return nil
}

// setupCallbacks sets up Tox event callbacks
func (m *Manager) setupCallbacks() {
	m.tox.OnFriendRequest(func(publicKey [32]byte, message string) {
		if m.onFriendRequest != nil {
			m.onFriendRequest(publicKey, message)
		}
	})

	m.tox.OnFriendMessage(func(friendID uint32, message string) {
		if m.onFriendMessage != nil {
			m.onFriendMessage(friendID, message)
		}
	})

	m.tox.OnFriendConnectionStatus(func(friendID uint32, status toxcore.ConnectionStatus) {
		if m.onFriendConnectionStatus != nil {
			m.onFriendConnectionStatus(friendID, status)
		}
	})

	m.tox.OnFriendStatus(func(friendID uint32, status toxcore.UserStatus) {
		if m.onFriendStatus != nil {
			m.onFriendStatus(friendID, status)
		}
	})

	m.tox.OnFriendName(func(friendID uint32, name string) {
		if m.onFriendName != nil {
			m.onFriendName(friendID, name)
		}
	})

	m.tox.OnFriendStatusMessage(func(friendID uint32, statusMessage string) {
		if m.onFriendStatusMessage != nil {
			m.onFriendStatusMessage(friendID, statusMessage)
		}
	})
}

// bootstrap connects to the Tox network
func (m *Manager) bootstrap() error {
	// Default bootstrap nodes
	bootstrapNodes := []struct {
		address string
		port    uint16
		pubkey  string
	}{
		{"node.tox.biribiri.org", 33445, "F404ABAA1C99A9D37D61AB54898F56793E1DEF8BD46B1038B9D822E8460FAB67"},
		{"tox.initramfs.io", 33445, "3F0A45A268367C1BEA652F258C85F4A66DA76BCAA667A49E770BCC4917AB6A25"},
		{"tox.kurnevsky.net", 33445, "82EF82BA33445A1F91A7DB27189ECFC0C013E06E3DA71F588ED692BED625EC23"},
	}

	var lastErr error
	for _, node := range bootstrapNodes {
		err := m.tox.Bootstrap(node.address, node.port, node.pubkey)
		if err != nil {
			lastErr = err
			log.Printf("Failed to bootstrap from %s: %v", node.address, err)
			continue
		}
		log.Printf("Successfully bootstrapped from %s", node.address)
	}

	return lastErr
}

// Start starts the Tox manager
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("Tox manager already running")
	}

	m.running = true
	log.Println("Tox manager started")
	return nil
}

// Stop stops the Tox manager
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	m.running = false

	// Save state before stopping
	if err := m.save(); err != nil {
		log.Printf("Warning: Failed to save state on stop: %v", err)
	}

	log.Println("Tox manager stopped")
	return nil
}

// Cleanup cleans up resources
func (m *Manager) Cleanup() {
	if m.tox != nil {
		m.save() // Final save
		m.tox.Kill()
	}
}

// Iterate performs one Tox iteration
func (m *Manager) Iterate() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.tox != nil && m.running {
		m.tox.Iterate()
	}
}

// GetToxID returns the current Tox ID
func (m *Manager) GetToxID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.tox == nil {
		return ""
	}
	return m.tox.SelfGetAddress()
}

// SendMessage sends a message to a friend
func (m *Manager) SendMessage(friendID uint32, message string, messageType toxcore.MessageType) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.tox == nil {
		return fmt.Errorf("Tox not initialized")
	}

	return m.tox.SendFriendMessage(friendID, message, messageType)
}

// AddFriend adds a friend by Tox ID
func (m *Manager) AddFriend(toxID, message string) (uint32, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.tox == nil {
		return 0, fmt.Errorf("Tox not initialized")
	}

	friendID, err := m.tox.AddFriend(toxID, message)
	if err != nil {
		return 0, err
	}

	// Save state after adding friend
	if err := m.save(); err != nil {
		log.Printf("Warning: Failed to save after adding friend: %v", err)
	}

	return friendID, nil
}

// AcceptFriendRequest accepts a friend request
func (m *Manager) AcceptFriendRequest(publicKey [32]byte) (uint32, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.tox == nil {
		return 0, fmt.Errorf("Tox not initialized")
	}

	friendID, err := m.tox.AddFriendByPublicKey(publicKey)
	if err != nil {
		return 0, err
	}

	// Save state after accepting friend request
	if err := m.save(); err != nil {
		log.Printf("Warning: Failed to save after accepting friend: %v", err)
	}

	return friendID, nil
}

// DeleteFriend removes a friend
func (m *Manager) DeleteFriend(friendID uint32) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.tox == nil {
		return fmt.Errorf("Tox not initialized")
	}

	err := m.tox.DeleteFriend(friendID)
	if err != nil {
		return err
	}

	// Save state after deleting friend
	if err := m.save(); err != nil {
		log.Printf("Warning: Failed to save after deleting friend: %v", err)
	}

	return nil
}

// GetFriends returns the list of friends
func (m *Manager) GetFriends() []uint32 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.tox == nil {
		return nil
	}

	return m.tox.GetFriends()
}

// GetFriendPublicKey returns a friend's public key
func (m *Manager) GetFriendPublicKey(friendID uint32) ([32]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.tox == nil {
		return [32]byte{}, fmt.Errorf("Tox not initialized")
	}

	return m.tox.GetFriendPublicKey(friendID)
}

// SetName sets our display name
func (m *Manager) SetName(name string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.tox == nil {
		return fmt.Errorf("Tox not initialized")
	}

	err := m.tox.SelfSetName(name)
	if err != nil {
		return err
	}

	return m.save()
}

// GetName returns our display name
func (m *Manager) GetName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.tox == nil {
		return ""
	}

	return m.tox.SelfGetName()
}

// SetStatusMessage sets our status message
func (m *Manager) SetStatusMessage(message string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.tox == nil {
		return fmt.Errorf("Tox not initialized")
	}

	err := m.tox.SelfSetStatusMessage(message)
	if err != nil {
		return err
	}

	return m.save()
}

// GetStatusMessage returns our status message
func (m *Manager) GetStatusMessage() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.tox == nil {
		return ""
	}

	return m.tox.SelfGetStatusMessage()
}

// save saves the Tox state to disk
func (m *Manager) save() error {
	if m.tox == nil {
		return fmt.Errorf("Tox not initialized")
	}

	savedata := m.tox.GetSavedata()
	
	// Write to temporary file first
	tmpFile := m.saveFile + ".tmp"
	if err := ioutil.WriteFile(tmpFile, savedata, 0600); err != nil {
		return fmt.Errorf("failed to write save data: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpFile, m.saveFile); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to rename save file: %w", err)
	}

	return nil
}

// Event callback setters
func (m *Manager) OnFriendRequest(callback func([32]byte, string)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onFriendRequest = callback
}

func (m *Manager) OnFriendMessage(callback func(uint32, string)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onFriendMessage = callback
}

func (m *Manager) OnFriendConnectionStatus(callback func(uint32, toxcore.ConnectionStatus)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onFriendConnectionStatus = callback
}

func (m *Manager) OnFriendStatus(callback func(uint32, toxcore.UserStatus)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onFriendStatus = callback
}

func (m *Manager) OnFriendName(callback func(uint32, string)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onFriendName = callback
}

func (m *Manager) OnFriendStatusMessage(callback func(uint32, string)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onFriendStatusMessage = callback
}
