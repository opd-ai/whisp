package transfer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/opd-ai/toxcore"
)

// SetToxManager configures the Tox manager for file transfers
func (m *Manager) SetToxManager(toxMgr ToxManager) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Store reference
	m.toxMgr = toxMgr

	// Set up Tox callbacks for incoming file transfers
	toxMgr.OnFileRecv(m.handleFileRecv)
	toxMgr.OnFileRecvChunk(m.handleFileRecvChunk)
	toxMgr.OnFileChunkRequest(m.handleFileChunkRequest)
}

// SendFile initiates a file transfer to a friend
func (m *Manager) SendFile(friendID uint32, filePath string) (*Transfer, error) {
	// Validate file exists and get info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to access file: %w", err)
	}

	if fileInfo.IsDir() {
		return nil, fmt.Errorf("cannot send directory: %s", filePath)
	}

	fileSize := uint64(fileInfo.Size())
	if err := m.validateFileSize(fileSize); err != nil {
		return nil, err
	}

	// Compute file checksum for integrity verification
	checksum, err := computeFileChecksum(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to compute file checksum: %w", err)
	}

	// Create transfer record
	transfer := &Transfer{
		ID:           uuid.New().String(),
		FriendID:     friendID,
		FileName:     filepath.Base(filePath),
		FilePath:     filePath,
		FileSize:     fileSize,
		FileChecksum: checksum,
		Direction:    TransferDirectionOutgoing,
		State:        TransferStatePending,
		StartTime:    time.Now(),
	}

	// Open file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for reading: %w", err)
	}
	transfer.file = file

	// Register transfer
	m.mu.Lock()
	m.transfers[transfer.ID] = transfer
	m.mu.Unlock()

	return transfer, nil
}

// StartSend begins the actual file transfer via Tox
func (m *Manager) StartSend(transfer *Transfer, toxMgr ToxManager) error {
	if transfer.Direction != TransferDirectionOutgoing {
		return fmt.Errorf("transfer %s is not an outgoing transfer", transfer.ID)
	}

	transfer.mu.Lock()
	if transfer.State != TransferStatePending {
		transfer.mu.Unlock()
		return fmt.Errorf("transfer %s is not in pending state", transfer.ID)
	}

	// Generate a file ID for Tox
	var fileID [32]byte
	copy(fileID[:], transfer.ID[:32])

	// Initiate Tox file transfer
	toxFileID, err := toxMgr.FileSend(transfer.FriendID, 0, transfer.FileSize, fileID, transfer.FileName)
	if err != nil {
		transfer.State = TransferStateFailed
		transfer.mu.Unlock()
		return fmt.Errorf("failed to initiate Tox file transfer: %w", err)
	}

	transfer.FileID = toxFileID
	transfer.State = TransferStateActive
	transfer.mu.Unlock()

	// Register with Tox transfer tracking
	m.mu.Lock()
	if m.toxTransfers[transfer.FriendID] == nil {
		m.toxTransfers[transfer.FriendID] = make(map[uint32]*Transfer)
	}
	m.toxTransfers[transfer.FriendID][toxFileID] = transfer
	m.mu.Unlock()

	return nil
}

// AcceptIncomingFile accepts an incoming file transfer
func (m *Manager) AcceptIncomingFile(transferID, saveDir string) error {
	m.mu.RLock()
	transfer, exists := m.transfers[transferID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("transfer %s not found", transferID)
	}

	transfer.mu.Lock()
	defer transfer.mu.Unlock()

	if transfer.Direction != TransferDirectionIncoming {
		return fmt.Errorf("transfer %s is not an incoming transfer", transferID)
	}

	if transfer.State != TransferStatePending {
		return fmt.Errorf("transfer %s is not in pending state", transferID)
	}

	// Sanitize filename to prevent path traversal attacks
	cleanFileName := filepath.Base(transfer.FileName)
	if cleanFileName != transfer.FileName || cleanFileName == "." || cleanFileName == ".." {
		return fmt.Errorf("invalid filename: contains path traversal sequences")
	}

	// Validate filename doesn't contain dangerous characters
	if strings.ContainsAny(cleanFileName, "<>:\"|?*\x00") {
		return fmt.Errorf("invalid filename: contains dangerous characters")
	}

	// Prepare file path with sanitized filename
	savePath := filepath.Join(saveDir, cleanFileName)

	// Ensure directory exists with restrictive permissions
	if err := os.MkdirAll(filepath.Dir(savePath), 0o700); err != nil {
		return fmt.Errorf("failed to create save directory: %w", err)
	}

	// Create file for writing with restrictive permissions
	file, err := os.OpenFile(savePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("failed to create file for writing: %w", err)
	}

	transfer.FilePath = savePath
	transfer.file = file
	transfer.State = TransferStateActive

	return nil
}

// PauseTransfer pauses an active transfer
func (m *Manager) PauseTransfer(transferID string, toxMgr ToxManager) error {
	m.mu.RLock()
	transfer, exists := m.transfers[transferID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("transfer %s not found", transferID)
	}

	transfer.mu.Lock()
	defer transfer.mu.Unlock()

	if transfer.State != TransferStateActive {
		return fmt.Errorf("transfer %s is not active", transferID)
	}

	// Send pause control to Tox
	if err := toxMgr.FileControl(transfer.FriendID, transfer.FileID, toxcore.FileControlPause); err != nil {
		return fmt.Errorf("failed to pause transfer via Tox: %w", err)
	}

	transfer.State = TransferStatePaused
	return nil
}

// ResumeTransfer resumes a paused transfer
func (m *Manager) ResumeTransfer(transferID string, toxMgr ToxManager) error {
	m.mu.RLock()
	transfer, exists := m.transfers[transferID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("transfer %s not found", transferID)
	}

	transfer.mu.Lock()
	defer transfer.mu.Unlock()

	if transfer.State != TransferStatePaused {
		return fmt.Errorf("transfer %s is not paused", transferID)
	}

	// Send resume control to Tox
	if err := toxMgr.FileControl(transfer.FriendID, transfer.FileID, toxcore.FileControlResume); err != nil {
		return fmt.Errorf("failed to resume transfer via Tox: %w", err)
	}

	transfer.State = TransferStateActive
	return nil
}

// CancelTransfer cancels an active or paused transfer
func (m *Manager) CancelTransfer(transferID string, toxMgr ToxManager) error {
	m.mu.RLock()
	transfer, exists := m.transfers[transferID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("transfer %s not found", transferID)
	}

	transfer.mu.Lock()
	defer transfer.mu.Unlock()

	// Check if transfer is already in a terminal state (avoid recursive lock)
	if transfer.State == TransferStateCompleted ||
		transfer.State == TransferStateFailed ||
		transfer.State == TransferStateCancelled {
		return fmt.Errorf("transfer %s is already complete", transferID)
	}

	// Send cancel control to Tox
	if err := toxMgr.FileControl(transfer.FriendID, transfer.FileID, toxcore.FileControlCancel); err != nil {
		return fmt.Errorf("failed to cancel transfer via Tox: %w", err)
	}

	// Close file if open
	if transfer.file != nil {
		transfer.file.Close()
		transfer.file = nil
	}

	// Remove incomplete incoming file
	if transfer.Direction == TransferDirectionIncoming && transfer.FilePath != "" {
		os.Remove(transfer.FilePath)
	}

	transfer.State = TransferStateCancelled
	now := time.Now()
	transfer.EndTime = &now

	return nil
}

// GetTransfer retrieves a transfer by ID
func (m *Manager) GetTransfer(transferID string) (*Transfer, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	transfer, exists := m.transfers[transferID]
	return transfer, exists
}

// GetActiveTransfers returns all active transfers
func (m *Manager) GetActiveTransfers() []*Transfer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var active []*Transfer
	for _, transfer := range m.transfers {
		transfer.mu.RLock()
		if transfer.State == TransferStateActive || transfer.State == TransferStatePaused {
			active = append(active, transfer)
		}
		transfer.mu.RUnlock()
	}

	return active
}

// GetTransfersByFriend returns all transfers for a specific friend
func (m *Manager) GetTransfersByFriend(friendID uint32) []*Transfer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var transfers []*Transfer
	for _, transfer := range m.transfers {
		if transfer.FriendID == friendID {
			transfers = append(transfers, transfer)
		}
	}

	return transfers
}
