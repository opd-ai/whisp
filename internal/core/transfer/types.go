package transfer

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/opd-ai/toxcore"
)

// ToxManager interface for Tox file transfer operations
type ToxManager interface {
	FileSend(friendID, kind uint32, fileSize uint64, fileID [32]byte, fileName string) (uint32, error)
	FileSendChunk(friendID, fileID uint32, position uint64, data []byte) error
	FileControl(friendID, fileID uint32, control toxcore.FileControl) error
	OnFileRecv(callback func(friendID, fileID, kind uint32, fileSize uint64, fileName string))
	OnFileRecvChunk(callback func(friendID, fileID uint32, position uint64, data []byte))
	OnFileChunkRequest(callback func(friendID, fileID uint32, position uint64, length int))
}

// TransferState represents the current state of a file transfer
type TransferState int

const (
	// TransferStatePending indicates the transfer is waiting to start
	TransferStatePending TransferState = iota
	// TransferStateActive indicates the transfer is in progress
	TransferStateActive
	// TransferStatePaused indicates the transfer is paused
	TransferStatePaused
	// TransferStateCompleted indicates the transfer completed successfully
	TransferStateCompleted
	// TransferStateFailed indicates the transfer failed
	TransferStateFailed
	// TransferStateCancelled indicates the transfer was cancelled
	TransferStateCancelled
)

// TransferDirection indicates if this is an incoming or outgoing transfer
type TransferDirection int

const (
	// TransferDirectionOutgoing for files being sent
	TransferDirectionOutgoing TransferDirection = iota
	// TransferDirectionIncoming for files being received
	TransferDirectionIncoming
)

// Transfer represents a file transfer operation
type Transfer struct {
	// Unique identifier for this transfer
	ID string

	// Tox-specific identifiers
	FriendID uint32
	FileID   uint32

	// File metadata
	FileName     string
	FilePath     string
	FileSize     uint64
	FileChecksum string // SHA256 hash

	// Transfer metadata
	Direction        TransferDirection
	State            TransferState
	BytesTransferred uint64
	StartTime        time.Time
	EndTime          *time.Time

	// File handle for active transfers
	file *os.File

	// Progress callback
	onProgress func(transfer *Transfer)
	onComplete func(transfer *Transfer, err error)

	// Synchronization
	mu sync.RWMutex
}

// Manager handles file transfer operations
type Manager struct {
	transfersDir string

	// Active transfers by ID
	transfers map[string]*Transfer

	// Active transfers by Tox file ID
	toxTransfers map[uint32]map[uint32]*Transfer // FriendID -> FileID -> Transfer

	// File size limits
	maxFileSize uint64

	// Tox manager for file operations
	toxMgr ToxManager

	mu sync.RWMutex
}

// NewManager creates a new file transfer manager
func NewManager(dataDir string) (*Manager, error) {
	transfersDir := filepath.Join(dataDir, "transfers")
	if err := os.MkdirAll(transfersDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create transfers directory: %w", err)
	}

	return &Manager{
		transfersDir: transfersDir,
		transfers:    make(map[string]*Transfer),
		toxTransfers: make(map[uint32]map[uint32]*Transfer),
		maxFileSize:  2 * 1024 * 1024 * 1024, // 2GB default limit
	}, nil
}

// SetMaxFileSize sets the maximum allowed file size for transfers
func (m *Manager) SetMaxFileSize(size uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.maxFileSize = size
}

// GetMaxFileSize returns the current maximum file size limit
func (m *Manager) GetMaxFileSize() uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.maxFileSize
}

// Progress returns the transfer progress as a percentage (0.0 to 1.0)
func (t *Transfer) Progress() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.FileSize == 0 {
		return 0.0
	}

	return float64(t.BytesTransferred) / float64(t.FileSize)
}

// IsComplete returns true if the transfer is in a terminal state
func (t *Transfer) IsComplete() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.State == TransferStateCompleted ||
		t.State == TransferStateFailed ||
		t.State == TransferStateCancelled
}

// computeFileChecksum calculates the SHA256 checksum of a file
func computeFileChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// validateFileSize checks if file size is within limits
func (m *Manager) validateFileSize(size uint64) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if size > m.maxFileSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size %d", size, m.maxFileSize)
	}

	return nil
}
