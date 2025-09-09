package tox

import (
	"fmt"
	"log"
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
	onFriendRequest func([32]byte, string)
	onFriendMessage func(uint32, string)
	onFriendStatus  func(uint32, toxcore.FriendStatus)
	onFriendName    func(uint32, string)
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
	log.Println("Initializing Tox...")
	
	// Create options
	options := toxcore.NewOptions()
	options.UDPEnabled = true
	options.IPv6Enabled = true
	
	// Try to load existing savedata
	var tox *toxcore.Tox
	var err error
	
	// Check if save file exists
	if savedata, err := m.loadSavedata(); err == nil && len(savedata) > 0 {
		log.Println("Loading existing Tox profile...")
		tox, err = toxcore.NewFromSavedata(options, savedata)
	} else {
		log.Println("Creating new Tox profile...")
		tox, err = toxcore.New(options)
	}
	
	if err != nil {
		return fmt.Errorf("failed to create Tox instance: %w", err)
	}
	
	m.tox = tox
	
	// Set up callbacks
	if err := m.setupCallbacks(); err != nil {
		return fmt.Errorf("failed to setup callbacks: %w", err)
	}
	
	// Bootstrap to network
	if err := m.bootstrap(); err != nil {
		log.Printf("Warning: Bootstrap failed: %v", err)
		// Don't fail initialization if bootstrap fails
	}
	
	log.Printf("Tox initialized. ID: %s", m.GetToxID())
	return nil
}

// loadSavedata loads savedata from file
func (m *Manager) loadSavedata() ([]byte, error) {
	// Implementation will read from saveFile
	// For now return empty to create new profile
	return nil, fmt.Errorf("no savedata")
}

// setupCallbacks sets up Tox event callbacks
func (m *Manager) setupCallbacks() error {
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

	m.tox.OnFriendStatus(func(friendID uint32, status toxcore.FriendStatus) {
		if m.onFriendStatus != nil {
			m.onFriendStatus(friendID, status)
		}
	})

	m.tox.OnFriendName(func(friendID uint32, name string) {
		if m.onFriendName != nil {
			m.onFriendName(friendID, name)
		}
	})

	return nil
}

// bootstrap connects to the Tox network  
func (m *Manager) bootstrap() error {
	// Bootstrap to well-known nodes
	bootstrapNodes := []struct {
		address   string
		port      uint16
		publicKey string
	}{
		{"node.tox.biribiri.org", 33445, "F404ABAA1C99A9D37D61AB54898F56793E1DEF8BD46B1038B9D822E8460FAB67"},
		{"tox.initramfs.io", 33445, "3F0A45A268367C1BEA652F258C85F4A66DA76BCAA667A49E770BCC4917AB6A25"},
		{"tox2.abilinski.com", 33445, "7A6098B590BDC73F9723FC59F82B3F9085A64D1B213AAF8E610FD351930D052D"},
	}

	var lastErr error
	for _, node := range bootstrapNodes {
		err := m.tox.Bootstrap(node.address, node.port, node.publicKey)
		if err != nil {
			lastErr = err
			log.Printf("Failed to bootstrap to %s: %v", node.address, err)
		} else {
			log.Printf("Successfully bootstrapped to %s", node.address)
			return nil
		}
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
	log.Println("Tox manager stopped")
	return nil
}

// Cleanup cleans up resources
func (m *Manager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.tox != nil {
		m.tox.Kill()
		m.tox = nil
	}
	log.Println("Tox manager cleanup")
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
		return "PLACEHOLDER_TOX_ID_NOT_INITIALIZED"
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

	friends := m.tox.GetFriends()
	friendIDs := make([]uint32, 0, len(friends))
	for friendID := range friends {
		friendIDs = append(friendIDs, friendID)
	}
	return friendIDs
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

	return m.tox.SelfSetName(name)
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

	return m.tox.SelfSetStatusMessage(message)
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
	if len(savedata) == 0 {
		return fmt.Errorf("no savedata to save")
	}

	// TODO: Implement actual file saving
	// For now just log that we would save
	log.Printf("Would save %d bytes of Tox state to %s", len(savedata), m.saveFile)
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

func (m *Manager) OnFriendStatus(callback func(uint32, toxcore.FriendStatus)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onFriendStatus = callback
}

func (m *Manager) OnFriendName(callback func(uint32, string)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onFriendName = callback
}
