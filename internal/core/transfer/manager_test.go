package transfer

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/opd-ai/toxcore"
)

// MockToxManager implements ToxManager for testing
type MockToxManager struct {
	fileSendFunc               func(friendID uint32, kind uint32, fileSize uint64, fileID [32]byte, fileName string) (uint32, error)
	fileSendChunkFunc          func(friendID uint32, fileID uint32, position uint64, data []byte) error
	fileControlFunc            func(friendID uint32, fileID uint32, control toxcore.FileControl) error
	onFileRecvCallback         func(friendID uint32, fileID uint32, kind uint32, fileSize uint64, fileName string)
	onFileRecvChunkCallback    func(friendID uint32, fileID uint32, position uint64, data []byte)
	onFileChunkRequestCallback func(friendID uint32, fileID uint32, position uint64, length int)
}

func (m *MockToxManager) FileSend(friendID uint32, kind uint32, fileSize uint64, fileID [32]byte, fileName string) (uint32, error) {
	if m.fileSendFunc != nil {
		return m.fileSendFunc(friendID, kind, fileSize, fileID, fileName)
	}
	return 1, nil // Return mock file ID
}

func (m *MockToxManager) FileSendChunk(friendID uint32, fileID uint32, position uint64, data []byte) error {
	if m.fileSendChunkFunc != nil {
		return m.fileSendChunkFunc(friendID, fileID, position, data)
	}
	return nil
}

func (m *MockToxManager) FileControl(friendID uint32, fileID uint32, control toxcore.FileControl) error {
	if m.fileControlFunc != nil {
		return m.fileControlFunc(friendID, fileID, control)
	}
	return nil
}

func (m *MockToxManager) OnFileRecv(callback func(friendID uint32, fileID uint32, kind uint32, fileSize uint64, fileName string)) {
	m.onFileRecvCallback = callback
}

func (m *MockToxManager) OnFileRecvChunk(callback func(friendID uint32, fileID uint32, position uint64, data []byte)) {
	m.onFileRecvChunkCallback = callback
}

func (m *MockToxManager) OnFileChunkRequest(callback func(friendID uint32, fileID uint32, position uint64, length int)) {
	m.onFileChunkRequestCallback = callback
}

// TriggerFileRecv simulates an incoming file transfer request
func (m *MockToxManager) TriggerFileRecv(friendID uint32, fileID uint32, kind uint32, fileSize uint64, fileName string) {
	if m.onFileRecvCallback != nil {
		m.onFileRecvCallback(friendID, fileID, kind, fileSize, fileName)
	}
}

// TriggerFileRecvChunk simulates receiving a file chunk
func (m *MockToxManager) TriggerFileRecvChunk(friendID uint32, fileID uint32, position uint64, data []byte) {
	if m.onFileRecvChunkCallback != nil {
		m.onFileRecvChunkCallback(friendID, fileID, position, data)
	}
}

// TriggerFileChunkRequest simulates a request for a file chunk
func (m *MockToxManager) TriggerFileChunkRequest(friendID uint32, fileID uint32, position uint64, length int) {
	if m.onFileChunkRequestCallback != nil {
		m.onFileChunkRequestCallback(friendID, fileID, position, length)
	}
}

func TestNewManager(t *testing.T) {
	tempDir := t.TempDir()

	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create transfer manager: %v", err)
	}

	if manager.transfersDir != filepath.Join(tempDir, "transfers") {
		t.Errorf("Expected transfersDir %s, got %s", filepath.Join(tempDir, "transfers"), manager.transfersDir)
	}

	// Check that transfers directory was created
	transfersDir := filepath.Join(tempDir, "transfers")
	if _, err := os.Stat(transfersDir); os.IsNotExist(err) {
		t.Error("Transfers directory was not created")
	}

	if manager.maxFileSize != 2*1024*1024*1024 {
		t.Errorf("Expected default max file size 2GB, got %d", manager.maxFileSize)
	}
}

func TestSetMaxFileSize(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create transfer manager: %v", err)
	}

	newSize := uint64(1024 * 1024 * 1024) // 1GB
	manager.SetMaxFileSize(newSize)

	if manager.GetMaxFileSize() != newSize {
		t.Errorf("Expected max file size %d, got %d", newSize, manager.GetMaxFileSize())
	}
}

func TestSendFile(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create transfer manager: %v", err)
	}

	// Create a test file
	testContent := "Hello, World! This is a test file for transfer."
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	friendID := uint32(123)
	transfer, err := manager.SendFile(friendID, testFile)
	if err != nil {
		t.Fatalf("Failed to send file: %v", err)
	}

	if transfer.FriendID != friendID {
		t.Errorf("Expected friend ID %d, got %d", friendID, transfer.FriendID)
	}

	if transfer.FileName != "test.txt" {
		t.Errorf("Expected file name test.txt, got %s", transfer.FileName)
	}

	if transfer.FileSize != uint64(len(testContent)) {
		t.Errorf("Expected file size %d, got %d", len(testContent), transfer.FileSize)
	}

	if transfer.Direction != TransferDirectionOutgoing {
		t.Error("Expected outgoing transfer direction")
	}

	if transfer.State != TransferStatePending {
		t.Error("Expected transfer state to be pending")
	}

	// Verify checksum
	expectedChecksum := fmt.Sprintf("%x", sha256.Sum256([]byte(testContent)))
	if transfer.FileChecksum != expectedChecksum {
		t.Errorf("Expected checksum %s, got %s", expectedChecksum, transfer.FileChecksum)
	}

	// Verify transfer is registered
	retrieved, exists := manager.GetTransfer(transfer.ID)
	if !exists {
		t.Error("Transfer not found in manager")
	}
	if retrieved.ID != transfer.ID {
		t.Error("Retrieved transfer ID mismatch")
	}
}

func TestSendFileNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create transfer manager: %v", err)
	}

	_, err = manager.SendFile(123, "/non/existent/file.txt")
	if err == nil {
		t.Error("Expected error when sending non-existent file")
	}
}

func TestSendFileDirectory(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create transfer manager: %v", err)
	}

	_, err = manager.SendFile(123, tempDir)
	if err == nil {
		t.Error("Expected error when trying to send directory")
	}
}

func TestSendFileTooLarge(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create transfer manager: %v", err)
	}

	// Set very small file size limit
	manager.SetMaxFileSize(10)

	// Create a larger test file
	testContent := "This content is longer than 10 bytes"
	testFile := filepath.Join(tempDir, "large.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = manager.SendFile(123, testFile)
	if err == nil {
		t.Error("Expected error when sending file that exceeds size limit")
	}
}

func TestStartSend(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create transfer manager: %v", err)
	}

	// Create test file and transfer
	testContent := "Test file content"
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	transfer, err := manager.SendFile(123, testFile)
	if err != nil {
		t.Fatalf("Failed to create transfer: %v", err)
	}

	// Set up mock Tox manager
	mockTox := &MockToxManager{}
	manager.SetToxManager(mockTox)

	// Start the transfer
	err = manager.StartSend(transfer, mockTox)
	if err != nil {
		t.Fatalf("Failed to start send: %v", err)
	}

	if transfer.State != TransferStateActive {
		t.Error("Expected transfer state to be active after starting")
	}

	if transfer.FileID != 1 {
		t.Errorf("Expected file ID 1, got %d", transfer.FileID)
	}
}

func TestIncomingFileTransfer(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create transfer manager: %v", err)
	}

	mockTox := &MockToxManager{}
	manager.SetToxManager(mockTox)

	// Simulate incoming file transfer
	friendID := uint32(456)
	fileID := uint32(789)
	fileName := "incoming.txt"
	fileSize := uint64(100)

	mockTox.TriggerFileRecv(friendID, fileID, 0, fileSize, fileName)

	// Check that transfer was created
	transfers := manager.GetTransfersByFriend(friendID)
	if len(transfers) != 1 {
		t.Fatalf("Expected 1 transfer, got %d", len(transfers))
	}

	transfer := transfers[0]
	if transfer.Direction != TransferDirectionIncoming {
		t.Error("Expected incoming transfer direction")
	}

	if transfer.State != TransferStatePending {
		t.Error("Expected transfer state to be pending")
	}

	if transfer.FileName != fileName {
		t.Errorf("Expected file name %s, got %s", fileName, transfer.FileName)
	}

	if transfer.FileSize != fileSize {
		t.Errorf("Expected file size %d, got %d", fileSize, transfer.FileSize)
	}
}

func TestAcceptIncomingFile(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create transfer manager: %v", err)
	}

	mockTox := &MockToxManager{}
	manager.SetToxManager(mockTox)

	// Simulate incoming file transfer
	friendID := uint32(456)
	fileID := uint32(789)
	fileName := "incoming.txt"
	fileSize := uint64(100)

	mockTox.TriggerFileRecv(friendID, fileID, 0, fileSize, fileName)

	// Get the transfer
	transfers := manager.GetTransfersByFriend(friendID)
	transfer := transfers[0]

	// Accept the transfer
	saveDir := filepath.Join(tempDir, "downloads")
	err = manager.AcceptIncomingFile(transfer.ID, saveDir)
	if err != nil {
		t.Fatalf("Failed to accept incoming file: %v", err)
	}

	if transfer.State != TransferStateActive {
		t.Error("Expected transfer state to be active after accepting")
	}

	expectedPath := filepath.Join(saveDir, fileName)
	if transfer.FilePath != expectedPath {
		t.Errorf("Expected file path %s, got %s", expectedPath, transfer.FilePath)
	}

	// Check that file was created
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Error("File was not created")
	}
}

func TestTransferProgress(t *testing.T) {
	transfer := &Transfer{
		FileSize:         1000,
		BytesTransferred: 250,
	}

	progress := transfer.Progress()
	expected := 0.25
	if progress != expected {
		t.Errorf("Expected progress %f, got %f", expected, progress)
	}

	// Test zero file size
	transfer.FileSize = 0
	progress = transfer.Progress()
	if progress != 0.0 {
		t.Errorf("Expected progress 0.0 for zero file size, got %f", progress)
	}
}

func TestTransferIsComplete(t *testing.T) {
	transfer := &Transfer{State: TransferStateActive}
	if transfer.IsComplete() {
		t.Error("Active transfer should not be complete")
	}

	transfer.State = TransferStateCompleted
	if !transfer.IsComplete() {
		t.Error("Completed transfer should be complete")
	}

	transfer.State = TransferStateFailed
	if !transfer.IsComplete() {
		t.Error("Failed transfer should be complete")
	}

	transfer.State = TransferStateCancelled
	if !transfer.IsComplete() {
		t.Error("Cancelled transfer should be complete")
	}
}

func TestSetCallbacks(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create transfer manager: %v", err)
	}

	// Create test file and transfer
	testContent := "Test content"
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	transfer, err := manager.SendFile(123, testFile)
	if err != nil {
		t.Fatalf("Failed to create transfer: %v", err)
	}

	// Test progress callback
	err = manager.SetProgressCallback(transfer.ID, func(t *Transfer) {
		// Progress callback set successfully
	})
	if err != nil {
		t.Fatalf("Failed to set progress callback: %v", err)
	}

	// Test completion callback
	err = manager.SetCompletionCallback(transfer.ID, func(t *Transfer, err error) {
		// Completion callback set successfully
	})
	if err != nil {
		t.Fatalf("Failed to set completion callback: %v", err)
	}

	// Verify callbacks were set (we can't easily test execution without complex setup)
	if transfer.onProgress == nil {
		t.Error("Progress callback was not set")
	}

	if transfer.onComplete == nil {
		t.Error("Completion callback was not set")
	}
}

func TestSetCallbacksNonExistentTransfer(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create transfer manager: %v", err)
	}

	err = manager.SetProgressCallback("non-existent", func(t *Transfer) {})
	if err == nil {
		t.Error("Expected error when setting callback for non-existent transfer")
	}

	err = manager.SetCompletionCallback("non-existent", func(t *Transfer, err error) {})
	if err == nil {
		t.Error("Expected error when setting callback for non-existent transfer")
	}
}

func TestComputeFileChecksum(t *testing.T) {
	tempDir := t.TempDir()
	testContent := "Hello, World!"
	testFile := filepath.Join(tempDir, "test.txt")

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	checksum, err := computeFileChecksum(testFile)
	if err != nil {
		t.Fatalf("Failed to compute checksum: %v", err)
	}

	expectedChecksum := fmt.Sprintf("%x", sha256.Sum256([]byte(testContent)))
	if checksum != expectedChecksum {
		t.Errorf("Expected checksum %s, got %s", expectedChecksum, checksum)
	}
}

func TestComputeFileChecksumNonExistent(t *testing.T) {
	_, err := computeFileChecksum("/non/existent/file.txt")
	if err == nil {
		t.Error("Expected error when computing checksum of non-existent file")
	}
}
