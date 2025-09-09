package message

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/opd-ai/toxcore"
	"github.com/opd-ai/whisp/internal/storage"
)

// MessageType represents the type of message
type MessageType int

const (
	MessageTypeNormal MessageType = iota
	MessageTypeAction
	MessageTypeFile
	MessageTypeVoice
	MessageTypeImage
	MessageTypeVideo
)

// Message represents a chat message
type Message struct {
	ID              int64       `json:"id"`
	UUID            string      `json:"uuid"`
	FriendID        uint32      `json:"friend_id"`
	Content         string      `json:"content"`
	MessageType     MessageType `json:"message_type"`
	IsOutgoing      bool        `json:"is_outgoing"`
	Timestamp       time.Time   `json:"timestamp"`
	DeliveredAt     *time.Time  `json:"delivered_at,omitempty"`
	ReadAt          *time.Time  `json:"read_at,omitempty"`
	EditedAt        *time.Time  `json:"edited_at,omitempty"`
	OriginalContent string      `json:"original_content,omitempty"`
	FilePath        string      `json:"file_path,omitempty"`
	FileSize        int64       `json:"file_size,omitempty"`
	FileType        string      `json:"file_type,omitempty"`
	IsDeleted       bool        `json:"is_deleted"`
	ReplyToID       *int64      `json:"reply_to_id,omitempty"`
}

// Manager manages messages and conversations
type Manager struct {
	db       *storage.Database
	toxMgr   ToxManager
	contacts ContactManager

	mu              sync.RWMutex
	pendingMessages map[string]*Message // UUID -> Message
}

// ToxManager interface for Tox operations
type ToxManager interface {
	SendMessage(friendID uint32, message string, messageType toxcore.MessageType) error
}

// ContactManager interface for contact operations
type ContactManager interface {
	GetContact(friendID uint32) (interface{}, bool)
}

// NewManager creates a new message manager
func NewManager(db *storage.Database, toxMgr ToxManager, contacts ContactManager) *Manager {
	return &Manager{
		db:              db,
		toxMgr:          toxMgr,
		contacts:        contacts,
		pendingMessages: make(map[string]*Message),
	}
}

// SendMessage sends a message to a friend
func (m *Manager) SendMessage(friendID uint32, content string, messageType MessageType) (*Message, error) {
	// Create message
	msg := &Message{
		UUID:        uuid.New().String(),
		FriendID:    friendID,
		Content:     content,
		MessageType: messageType,
		IsOutgoing:  true,
		Timestamp:   time.Now(),
	}

	// Save to database first
	if err := m.saveMessage(msg); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	// Add to pending
	m.mu.Lock()
	m.pendingMessages[msg.UUID] = msg
	m.mu.Unlock()

	// Convert message type for Tox
	var toxMsgType toxcore.MessageType
	switch messageType {
	case MessageTypeAction:
		toxMsgType = toxcore.MessageTypeAction
	default:
		toxMsgType = toxcore.MessageTypeNormal
	}

	// Send via Tox
	if err := m.toxMgr.SendMessage(friendID, content, toxMsgType); err != nil {
		// Mark as failed
		m.mu.Lock()
		delete(m.pendingMessages, msg.UUID)
		m.mu.Unlock()
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Mark as delivered (for now, in real implementation this would be done by callback)
	now := time.Now()
	msg.DeliveredAt = &now

	// Update database
	go func() {
		query := `UPDATE messages SET delivered_at = ? WHERE id = ?`
		if _, err := m.db.Exec(query, now, msg.ID); err != nil {
			log.Printf("Failed to update message delivery status: %v", err)
		}
	}()

	return msg, nil
}

// HandleIncomingMessage handles an incoming message
func (m *Manager) HandleIncomingMessage(friendID uint32, content string, messageType MessageType) *Message {
	msg := &Message{
		UUID:        uuid.New().String(),
		FriendID:    friendID,
		Content:     content,
		MessageType: messageType,
		IsOutgoing:  false,
		Timestamp:   time.Now(),
	}

	// Save to database
	if err := m.saveMessage(msg); err != nil {
		log.Printf("Failed to save incoming message: %v", err)
		return nil
	}

	// Mark as read immediately (for demo, in real app user would mark as read)
	now := time.Now()
	msg.ReadAt = &now

	go func() {
		query := `UPDATE messages SET read_at = ? WHERE id = ?`
		if _, err := m.db.Exec(query, now, msg.ID); err != nil {
			log.Printf("Failed to update message read status: %v", err)
		}
	}()

	return msg
}

// GetMessages returns messages for a conversation
func (m *Manager) GetMessages(friendID uint32, limit, offset int) ([]*Message, error) {
	query := `
		SELECT id, uuid, friend_id, content, message_type, is_outgoing,
		       timestamp, delivered_at, read_at, edited_at, original_content,
		       file_path, file_size, file_type, is_deleted, reply_to_id
		FROM messages 
		WHERE friend_id = ? AND is_deleted = 0
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`

	rows, err := m.db.Query(query, friendID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		msg := &Message{}
		var deliveredAt, readAt, editedAt sql.NullTime
		var originalContent, filePath, fileType sql.NullString
		var fileSize sql.NullInt64
		var replyToID sql.NullInt64

		err := rows.Scan(
			&msg.ID, &msg.UUID, &msg.FriendID, &msg.Content, &msg.MessageType,
			&msg.IsOutgoing, &msg.Timestamp, &deliveredAt, &readAt, &editedAt,
			&originalContent, &filePath, &fileSize, &fileType, &msg.IsDeleted,
			&replyToID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		if deliveredAt.Valid {
			msg.DeliveredAt = &deliveredAt.Time
		}
		if readAt.Valid {
			msg.ReadAt = &readAt.Time
		}
		if editedAt.Valid {
			msg.EditedAt = &editedAt.Time
		}
		if originalContent.Valid {
			msg.OriginalContent = originalContent.String
		}
		if filePath.Valid {
			msg.FilePath = filePath.String
		}
		if fileType.Valid {
			msg.FileType = fileType.String
		}
		if fileSize.Valid {
			msg.FileSize = fileSize.Int64
		}
		if replyToID.Valid {
			msg.ReplyToID = &replyToID.Int64
		}

		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

// EditMessage edits an existing message
func (m *Manager) EditMessage(messageID int64, newContent string) error {
	// Get original message
	query := `SELECT content FROM messages WHERE id = ?`
	var originalContent string
	if err := m.db.QueryRow(query, messageID).Scan(&originalContent); err != nil {
		return fmt.Errorf("failed to get original message: %w", err)
	}

	// Update message
	now := time.Now()
	updateQuery := `
		UPDATE messages 
		SET content = ?, original_content = ?, edited_at = ? 
		WHERE id = ?
	`

	_, err := m.db.Exec(updateQuery, newContent, originalContent, now, messageID)
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}

	return nil
}

// DeleteMessage deletes a message (soft delete)
func (m *Manager) DeleteMessage(messageID int64) error {
	query := `UPDATE messages SET is_deleted = 1 WHERE id = ?`
	_, err := m.db.Exec(query, messageID)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	return nil
}

// MarkAsRead marks messages as read
func (m *Manager) MarkAsRead(friendID uint32) error {
	now := time.Now()
	query := `
		UPDATE messages 
		SET read_at = ? 
		WHERE friend_id = ? AND is_outgoing = 0 AND read_at IS NULL
	`

	_, err := m.db.Exec(query, now, friendID)
	if err != nil {
		return fmt.Errorf("failed to mark messages as read: %w", err)
	}

	return nil
}

// SearchMessages searches for messages containing text using FTS for optimal performance
func (m *Manager) SearchMessages(query string, limit int) ([]*Message, error) {
	if query == "" {
		return []*Message{}, nil
	}

	// First try FTS search if available
	if m.isFTSAvailable() {
		messages, err := m.searchWithFTS(query, limit)
		if err == nil {
			return messages, nil
		}
		// If FTS fails, fall back to LIKE search
		log.Printf("FTS search failed, falling back to LIKE: %v", err)
	}

	// Fallback to LIKE query
	return m.searchWithLike(query, limit)
}

// isFTSAvailable checks if the FTS virtual table exists and is usable
func (m *Manager) isFTSAvailable() bool {
	var count int
	err := m.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='messages_fts'").Scan(&count)
	return err == nil && count > 0
}

// searchWithFTS performs search using FTS5 virtual table
func (m *Manager) searchWithFTS(query string, limit int) ([]*Message, error) {
	searchQuery := `
		SELECT m.id, m.uuid, m.friend_id, m.content, m.message_type, m.is_outgoing,
		       m.timestamp, m.delivered_at, m.read_at, m.edited_at, m.original_content,
		       m.file_path, m.file_size, m.file_type, m.is_deleted, m.reply_to_id
		FROM messages m
		INNER JOIN messages_fts fts ON m.id = fts.rowid
		WHERE messages_fts MATCH ? AND m.is_deleted = 0
		ORDER BY m.timestamp DESC
		LIMIT ?
	`

	// Escape query for FTS5 MATCH syntax - wrap in double quotes for phrase search
	ftsQuery := fmt.Sprintf(`"%s"`, query)

	rows, err := m.db.Query(searchQuery, ftsQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("FTS search failed: %w", err)
	}
	defer rows.Close()

	return m.scanMessageRows(rows)
}

// searchWithLike performs search using LIKE operator (fallback)
func (m *Manager) searchWithLike(query string, limit int) ([]*Message, error) {
	searchQuery := `
		SELECT id, uuid, friend_id, content, message_type, is_outgoing,
		       timestamp, delivered_at, read_at, edited_at, original_content,
		       file_path, file_size, file_type, is_deleted, reply_to_id
		FROM messages 
		WHERE content LIKE ? AND is_deleted = 0
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := m.db.Query(searchQuery, "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("LIKE search failed: %w", err)
	}
	defer rows.Close()

	return m.scanMessageRows(rows)
}

// scanMessageRows scans database rows into Message structs
func (m *Manager) scanMessageRows(rows *sql.Rows) ([]*Message, error) {
	var messages []*Message
	for rows.Next() {
		msg := &Message{}
		var deliveredAt, readAt, editedAt sql.NullTime
		var originalContent, filePath, fileType sql.NullString
		var fileSize sql.NullInt64
		var replyToID sql.NullInt64

		err := rows.Scan(
			&msg.ID, &msg.UUID, &msg.FriendID, &msg.Content, &msg.MessageType,
			&msg.IsOutgoing, &msg.Timestamp, &deliveredAt, &readAt, &editedAt,
			&originalContent, &filePath, &fileSize, &fileType, &msg.IsDeleted,
			&replyToID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		// Set nullable fields
		if deliveredAt.Valid {
			msg.DeliveredAt = &deliveredAt.Time
		}
		if readAt.Valid {
			msg.ReadAt = &readAt.Time
		}
		if editedAt.Valid {
			msg.EditedAt = &editedAt.Time
		}
		if originalContent.Valid {
			msg.OriginalContent = originalContent.String
		}
		if filePath.Valid {
			msg.FilePath = filePath.String
		}
		if fileType.Valid {
			msg.FileType = fileType.String
		}
		if fileSize.Valid {
			msg.FileSize = fileSize.Int64
		}
		if replyToID.Valid {
			msg.ReplyToID = &replyToID.Int64
		}

		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

// ProcessPending processes pending messages
func (m *Manager) ProcessPending() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// In a real implementation, this would handle message delivery confirmations
	// For now, we'll just clear old pending messages
	for uuid, msg := range m.pendingMessages {
		if time.Since(msg.Timestamp) > 30*time.Second {
			delete(m.pendingMessages, uuid)
		}
	}
}

// saveMessage saves a message to the database
func (m *Manager) saveMessage(msg *Message) error {
	query := `
		INSERT INTO messages (uuid, friend_id, content, message_type, is_outgoing,
		                     timestamp, delivered_at, read_at, edited_at, original_content,
		                     file_path, file_size, file_type, is_deleted, reply_to_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := m.db.Exec(query,
		msg.UUID, msg.FriendID, msg.Content, msg.MessageType, msg.IsOutgoing,
		msg.Timestamp, msg.DeliveredAt, msg.ReadAt, msg.EditedAt, msg.OriginalContent,
		msg.FilePath, msg.FileSize, msg.FileType, msg.IsDeleted, msg.ReplyToID,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	msg.ID = id
	return nil
}
