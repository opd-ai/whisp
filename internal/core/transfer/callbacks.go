package transfer

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/opd-ai/whisp/platform/common"
)

// handleFileRecv handles incoming file transfer requests from Tox
func (m *Manager) handleFileRecv(friendID, fileID, kind uint32, fileSize uint64, fileName string) {
	common.SecurePrintf("Received file transfer request: friend=%d, fileID=%d, size=%d, name=%s",
		friendID, fileID, fileSize, fileName)

	// Validate file size
	if err := m.validateFileSize(fileSize); err != nil {
		common.SecurePrintf("Rejecting file transfer: %v", err)
		return
	}

	// Create incoming transfer record
	transfer := &Transfer{
		ID:        uuid.New().String(),
		FriendID:  friendID,
		FileID:    fileID,
		FileName:  fileName,
		FileSize:  fileSize,
		Direction: TransferDirectionIncoming,
		State:     TransferStatePending,
		StartTime: time.Now(),
	}

	// Register transfer
	m.mu.Lock()
	m.transfers[transfer.ID] = transfer
	if m.toxTransfers[friendID] == nil {
		m.toxTransfers[friendID] = make(map[uint32]*Transfer)
	}
	m.toxTransfers[friendID][fileID] = transfer
	m.mu.Unlock()

	log.Printf("Created incoming transfer record: %s", transfer.ID)
}

// handleFileRecvChunk handles incoming file data chunks from Tox
func (m *Manager) handleFileRecvChunk(friendID, fileID uint32, position uint64, data []byte) {
	// Find the transfer
	m.mu.RLock()
	transfers, exists := m.toxTransfers[friendID]
	m.mu.RUnlock()

	if !exists {
		common.SecurePrintf("No transfers found for friend %d", friendID)
		return
	}

	transfer, exists := transfers[fileID]
	if !exists {
		common.SecurePrintf("No transfer found for friend %d, fileID %d", friendID, fileID)
		return
	}

	transfer.mu.Lock()
	defer transfer.mu.Unlock()

	if transfer.State != TransferStateActive {
		common.SecurePrintf("Transfer %s is not active, ignoring chunk", transfer.ID)
		return
	}

	if transfer.file == nil {
		common.SecurePrintf("Transfer %s has no open file, ignoring chunk", transfer.ID)
		return
	}

	// Seek to the correct position
	if _, err := transfer.file.Seek(int64(position), 0); err != nil {
		common.SecurePrintf("Failed to seek to position %d in transfer %s: %v", position, transfer.ID, err)
		return
	}

	// Write the data
	if _, err := transfer.file.Write(data); err != nil {
		common.SecurePrintf("Failed to write data for transfer %s: %v", transfer.ID, err)
		return
	}

	// Update progress
	transfer.BytesTransferred += uint64(len(data))
	if transfer.BytesTransferred >= transfer.FileSize {
		transfer.State = TransferStateCompleted
		transfer.file.Close()
		transfer.file = nil
		common.SecurePrintf("Transfer %s completed successfully", transfer.ID)
	}
}

// handleFileChunkRequest handles requests for file chunks from Tox (for outgoing transfers)
func (m *Manager) handleFileChunkRequest(friendID, fileID uint32, position uint64, length int) {
	// Find the transfer
	m.mu.RLock()
	friendTransfers, exists := m.toxTransfers[friendID]
	if !exists {
		m.mu.RUnlock()
		log.Printf("No transfers found for friend %d", friendID)
		return
	}

	transfer, exists := friendTransfers[fileID]
	if !exists {
		m.mu.RUnlock()
		log.Printf("No transfer found for friend %d, fileID %d", friendID, fileID)
		return
	}
	m.mu.RUnlock()

	transfer.mu.Lock()
	defer transfer.mu.Unlock()

	// Check if transfer is active
	if transfer.State != TransferStateActive {
		log.Printf("Transfer %s is not active, ignoring chunk request", transfer.ID)
		return
	}

	// Check if file is open for reading
	if transfer.file == nil {
		log.Printf("Transfer %s has no open file, ignoring chunk request", transfer.ID)
		return
	}

	// Handle completion signal (length 0)
	if length == 0 {
		m.completeTransfer(transfer)
		return
	}

	// Seek to position
	if _, err := transfer.file.Seek(int64(position), io.SeekStart); err != nil {
		log.Printf("Failed to seek to position %d in transfer %s: %v", position, transfer.ID, err)
		transfer.State = TransferStateFailed
		return
	}

	// Read data
	data := make([]byte, length)
	bytesRead, err := transfer.file.Read(data)
	if err != nil && err != io.EOF {
		log.Printf("Failed to read data for transfer %s: %v", transfer.ID, err)
		transfer.State = TransferStateFailed
		return
	}

	// Send chunk via Tox (this requires access to ToxManager)
	// We'll need to store the ToxManager reference in the Manager
	if m.toxMgr != nil {
		if err := m.toxMgr.FileSendChunk(friendID, fileID, position, data[:bytesRead]); err != nil {
			log.Printf("Failed to send chunk for transfer %s: %v", transfer.ID, err)
			transfer.State = TransferStateFailed
			return
		}
	}

	// Update progress
	transfer.BytesTransferred = position + uint64(bytesRead)

	// Call progress callback if set
	if transfer.onProgress != nil {
		go transfer.onProgress(transfer)
	}
}

// completeTransfer marks a transfer as completed and performs cleanup
func (m *Manager) completeTransfer(transfer *Transfer) {
	// Close file
	if transfer.file != nil {
		transfer.file.Close()
		transfer.file = nil
	}

	// Verify checksum for incoming files
	if transfer.Direction == TransferDirectionIncoming && transfer.FilePath != "" {
		if checksum, err := computeFileChecksum(transfer.FilePath); err == nil {
			transfer.FileChecksum = checksum
		}
	}

	// Update state
	transfer.State = TransferStateCompleted
	now := time.Now()
	transfer.EndTime = &now

	// Call completion callback if set
	if transfer.onComplete != nil {
		go transfer.onComplete(transfer, nil)
	}

	log.Printf("Transfer %s completed successfully", transfer.ID)
}

// SetProgressCallback sets a progress callback for a transfer
func (m *Manager) SetProgressCallback(transferID string, callback func(*Transfer)) error {
	m.mu.RLock()
	transfer, exists := m.transfers[transferID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("transfer %s not found", transferID)
	}

	transfer.mu.Lock()
	transfer.onProgress = callback
	transfer.mu.Unlock()

	return nil
}

// SetCompletionCallback sets a completion callback for a transfer
func (m *Manager) SetCompletionCallback(transferID string, callback func(*Transfer, error)) error {
	m.mu.RLock()
	transfer, exists := m.transfers[transferID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("transfer %s not found", transferID)
	}

	transfer.mu.Lock()
	transfer.onComplete = callback
	transfer.mu.Unlock()

	return nil
}
