package contact

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/opd-ai/whisp/internal/storage"
)

// Status represents contact status
type Status int

const (
	StatusOffline Status = iota
	StatusOnline
	StatusAway
	StatusBusy
)

// Contact represents a contact/friend
type Contact struct {
	ID            int64     `json:"id"`
	ToxID         string    `json:"tox_id"`
	PublicKey     []byte    `json:"public_key"`
	FriendID      uint32    `json:"friend_id"`
	Name          string    `json:"name"`
	StatusMessage string    `json:"status_message"`
	Avatar        []byte    `json:"avatar,omitempty"`
	Status        Status    `json:"status"`
	IsBlocked     bool      `json:"is_blocked"`
	IsFavorite    bool      `json:"is_favorite"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	LastSeenAt    time.Time `json:"last_seen_at"`
}

// Manager manages contacts and friend relationships
type Manager struct {
	db     *storage.Database
	toxMgr ToxManager // Interface to avoid circular dependency
	
	mu       sync.RWMutex
	contacts map[uint32]*Contact // friendID -> Contact
	pending  []PendingRequest
}

// ToxManager interface for Tox operations
type ToxManager interface {
	GetFriends() []uint32
	GetFriendPublicKey(friendID uint32) ([32]byte, error)
	AddFriend(toxID, message string) (uint32, error)
	AcceptFriendRequest(publicKey [32]byte) (uint32, error)
	DeleteFriend(friendID uint32) error
}

// PendingRequest represents a pending friend request
type PendingRequest struct {
	PublicKey [32]byte  `json:"public_key"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// NewManager creates a new contact manager
func NewManager(db *storage.Database, toxMgr ToxManager) *Manager {
	m := &Manager{
		db:       db,
		toxMgr:   toxMgr,
		contacts: make(map[uint32]*Contact),
		pending:  make([]PendingRequest, 0),
	}

	// Load existing contacts
	if err := m.loadContacts(); err != nil {
		log.Printf("Warning: Failed to load contacts: %v", err)
	}

	return m
}

// loadContacts loads contacts from database
func (m *Manager) loadContacts() error {
	query := `
		SELECT id, tox_id, public_key, friend_id, name, status_message, 
		       avatar, status, is_blocked, is_favorite, created_at, updated_at, last_seen_at
		FROM contacts WHERE is_blocked = 0
	`
	
	rows, err := m.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query contacts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		contact := &Contact{}
		var avatar sql.NullString
		
		err := rows.Scan(
			&contact.ID, &contact.ToxID, &contact.PublicKey, &contact.FriendID,
			&contact.Name, &contact.StatusMessage, &avatar, &contact.Status,
			&contact.IsBlocked, &contact.IsFavorite, &contact.CreatedAt,
			&contact.UpdatedAt, &contact.LastSeenAt,
		)
		if err != nil {
			return fmt.Errorf("failed to scan contact: %w", err)
		}

		if avatar.Valid {
			contact.Avatar = []byte(avatar.String)
		}

		m.contacts[contact.FriendID] = contact
	}

	return rows.Err()
}

// GetAllContacts returns all contacts
func (m *Manager) GetAllContacts() []*Contact {
	m.mu.RLock()
	defer m.mu.RUnlock()

	contacts := make([]*Contact, 0, len(m.contacts))
	for _, contact := range m.contacts {
		contacts = append(contacts, contact)
	}
	return contacts
}

// GetContact returns a contact by friend ID
func (m *Manager) GetContact(friendID uint32) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	contact, exists := m.contacts[friendID]
	return contact, exists
}

// AddContact adds a new contact
func (m *Manager) AddContact(toxID, message string) (*Contact, error) {
	// Add friend via Tox
	friendID, err := m.toxMgr.AddFriend(toxID, message)
	if err != nil {
		return nil, fmt.Errorf("failed to add friend: %w", err)
	}

	// Get public key
	publicKey, err := m.toxMgr.GetFriendPublicKey(friendID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	// Create contact
	contact := &Contact{
		ToxID:     toxID,
		PublicKey: publicKey[:],
		FriendID:  friendID,
		Name:      "Unknown", // Will be updated when friend comes online
		Status:    StatusOffline,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to database
	if err := m.saveContact(contact); err != nil {
		return nil, fmt.Errorf("failed to save contact: %w", err)
	}

	// Add to memory
	m.mu.Lock()
	m.contacts[friendID] = contact
	m.mu.Unlock()

	return contact, nil
}

// AcceptFriendRequest accepts a pending friend request
func (m *Manager) AcceptFriendRequest(publicKey [32]byte) (*Contact, error) {
	// Accept via Tox
	friendID, err := m.toxMgr.AcceptFriendRequest(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to accept friend request: %w", err)
	}

	// Create contact
	contact := &Contact{
		PublicKey: publicKey[:],
		FriendID:  friendID,
		Name:      "Unknown",
		Status:    StatusOffline,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to database
	if err := m.saveContact(contact); err != nil {
		return nil, fmt.Errorf("failed to save contact: %w", err)
	}

	// Add to memory
	m.mu.Lock()
	m.contacts[friendID] = contact
	// Remove from pending
	m.removePendingRequest(publicKey)
	m.mu.Unlock()

	return contact, nil
}

// DeleteContact deletes a contact
func (m *Manager) DeleteContact(friendID uint32) error {
	// Delete from Tox
	if err := m.toxMgr.DeleteFriend(friendID); err != nil {
		return fmt.Errorf("failed to delete friend: %w", err)
	}

	// Mark as deleted in database (soft delete)
	query := `UPDATE contacts SET is_blocked = 1, updated_at = ? WHERE friend_id = ?`
	if _, err := m.db.Exec(query, time.Now(), friendID); err != nil {
		return fmt.Errorf("failed to delete contact from database: %w", err)
	}

	// Remove from memory
	m.mu.Lock()
	delete(m.contacts, friendID)
	m.mu.Unlock()

	return nil
}

// UpdateName updates a contact's name
func (m *Manager) UpdateName(friendID uint32, name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	contact, exists := m.contacts[friendID]
	if !exists {
		return
	}

	contact.Name = name
	contact.UpdatedAt = time.Now()

	// Update database
	go func() {
		query := `UPDATE contacts SET name = ?, updated_at = ? WHERE friend_id = ?`
		if _, err := m.db.Exec(query, name, contact.UpdatedAt, friendID); err != nil {
			log.Printf("Failed to update contact name: %v", err)
		}
	}()
}

// UpdateStatusMessage updates a contact's status message
func (m *Manager) UpdateStatusMessage(friendID uint32, statusMessage string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	contact, exists := m.contacts[friendID]
	if !exists {
		return
	}

	contact.StatusMessage = statusMessage
	contact.UpdatedAt = time.Now()

	// Update database
	go func() {
		query := `UPDATE contacts SET status_message = ?, updated_at = ? WHERE friend_id = ?`
		if _, err := m.db.Exec(query, statusMessage, contact.UpdatedAt, friendID); err != nil {
			log.Printf("Failed to update contact status message: %v", err)
		}
	}()
}

// UpdateStatus updates a contact's status
func (m *Manager) UpdateStatus(friendID uint32, status interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	contact, exists := m.contacts[friendID]
	if !exists {
		return
	}

	// Convert Tox status to our status
	var newStatus Status
	switch status {
	case 0: // UserStatusNone
		newStatus = StatusOnline
	case 1: // UserStatusAway
		newStatus = StatusAway
	case 2: // UserStatusBusy
		newStatus = StatusBusy
	default:
		newStatus = StatusOffline
	}

	contact.Status = newStatus
	contact.UpdatedAt = time.Now()
	if newStatus != StatusOffline {
		contact.LastSeenAt = time.Now()
	}

	// Update database
	go func() {
		query := `UPDATE contacts SET status = ?, updated_at = ?, last_seen_at = ? WHERE friend_id = ?`
		if _, err := m.db.Exec(query, newStatus, contact.UpdatedAt, contact.LastSeenAt, friendID); err != nil {
			log.Printf("Failed to update contact status: %v", err)
		}
	}()
}

// UpdateConnectionStatus updates a contact's connection status
func (m *Manager) UpdateConnectionStatus(friendID uint32, status interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	contact, exists := m.contacts[friendID]
	if !exists {
		return
	}

	// Convert connection status to our status
	var newStatus Status
	switch status {
	case 0: // ConnectionStatusNone
		newStatus = StatusOffline
	case 1, 2: // ConnectionStatusTCP, ConnectionStatusUDP
		newStatus = StatusOnline
	default:
		newStatus = StatusOffline
	}

	contact.Status = newStatus
	contact.UpdatedAt = time.Now()
	if newStatus != StatusOffline {
		contact.LastSeenAt = time.Now()
	}

	// Update database
	go func() {
		query := `UPDATE contacts SET status = ?, updated_at = ?, last_seen_at = ? WHERE friend_id = ?`
		if _, err := m.db.Exec(query, newStatus, contact.UpdatedAt, contact.LastSeenAt, friendID); err != nil {
			log.Printf("Failed to update contact connection status: %v", err)
		}
	}()
}

// HandleFriendRequest handles an incoming friend request
func (m *Manager) HandleFriendRequest(publicKey [32]byte, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Add to pending requests
	request := PendingRequest{
		PublicKey: publicKey,
		Message:   message,
		Timestamp: time.Now(),
	}

	m.pending = append(m.pending, request)
}

// GetPendingRequests returns pending friend requests
func (m *Manager) GetPendingRequests() []PendingRequest {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return append([]PendingRequest(nil), m.pending...)
}

// removePendingRequest removes a pending request
func (m *Manager) removePendingRequest(publicKey [32]byte) {
	for i, req := range m.pending {
		if req.PublicKey == publicKey {
			m.pending = append(m.pending[:i], m.pending[i+1:]...)
			break
		}
	}
}

// saveContact saves a contact to the database
func (m *Manager) saveContact(contact *Contact) error {
	query := `
		INSERT INTO contacts (tox_id, public_key, friend_id, name, status_message, 
		                     avatar, status, is_blocked, is_favorite, created_at, updated_at, last_seen_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := m.db.Exec(query,
		contact.ToxID, contact.PublicKey, contact.FriendID, contact.Name,
		contact.StatusMessage, contact.Avatar, contact.Status, contact.IsBlocked,
		contact.IsFavorite, contact.CreatedAt, contact.UpdatedAt, contact.LastSeenAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	contact.ID = id
	return nil
}
