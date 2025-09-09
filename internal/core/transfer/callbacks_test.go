package transfer

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestPauseTransfer(t *testing.T) {
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

	mockTox := &MockToxManager{}
	manager.SetToxManager(mockTox)

	// Start the transfer first
	if err := manager.StartSend(transfer, mockTox); err != nil {
		t.Fatalf("Failed to start transfer: %v", err)
	}

	// Pause the transfer
	err = manager.PauseTransfer(transfer.ID, mockTox)
	if err != nil {
		t.Fatalf("Failed to pause transfer: %v", err)
	}

	if transfer.State != TransferStatePaused {
		t.Error("Expected transfer state to be paused")
	}
}

func TestResumeTransfer(t *testing.T) {
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

	mockTox := &MockToxManager{}
	manager.SetToxManager(mockTox)

	// Start and pause the transfer
	if err := manager.StartSend(transfer, mockTox); err != nil {
		t.Fatalf("Failed to start transfer: %v", err)
	}
	if err := manager.PauseTransfer(transfer.ID, mockTox); err != nil {
		t.Fatalf("Failed to pause transfer: %v", err)
	}

	// Resume the transfer
	err = manager.ResumeTransfer(transfer.ID, mockTox)
	if err != nil {
		t.Fatalf("Failed to resume transfer: %v", err)
	}

	if transfer.State != TransferStateActive {
		t.Error("Expected transfer state to be active after resume")
	}
}

func TestCancelTransfer(t *testing.T) {
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

	mockTox := &MockToxManager{}
	manager.SetToxManager(mockTox)

	// Start the transfer
	if err := manager.StartSend(transfer, mockTox); err != nil {
		t.Fatalf("Failed to start transfer: %v", err)
	}

	// Cancel the transfer
	err = manager.CancelTransfer(transfer.ID, mockTox)
	if err != nil {
		t.Fatalf("Failed to cancel transfer: %v", err)
	}

	if transfer.State != TransferStateCancelled {
		t.Error("Expected transfer state to be cancelled")
	}

	if transfer.EndTime == nil {
		t.Error("Expected end time to be set after cancellation")
	}
}

func TestGetActiveTransfers(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create transfer manager: %v", err)
	}

	// Create multiple test files and transfers
	for i := 0; i < 3; i++ {
		testFile := filepath.Join(tempDir, fmt.Sprintf("test%d.txt", i))
		if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		transfer, err := manager.SendFile(uint32(i), testFile)
		if err != nil {
			t.Fatalf("Failed to create transfer: %v", err)
		}

		// Make one active and one paused
		if i < 2 {
			transfer.mu.Lock()
			if i == 0 {
				transfer.State = TransferStateActive
			} else {
				transfer.State = TransferStatePaused
			}
			transfer.mu.Unlock()
		}
	}

	activeTransfers := manager.GetActiveTransfers()
	if len(activeTransfers) != 2 {
		t.Errorf("Expected 2 active transfers, got %d", len(activeTransfers))
	}
}

func TestHandleFileRecvChunk(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create transfer manager: %v", err)
	}

	mockTox := &MockToxManager{}
	manager.SetToxManager(mockTox)

	// Create incoming transfer
	friendID := uint32(456)
	fileID := uint32(789)
	fileName := "incoming.txt"
	fileSize := uint64(13) // "Hello, World!" length

	mockTox.TriggerFileRecv(friendID, fileID, 0, fileSize, fileName)

	// Get the transfer and accept it
	transfers := manager.GetTransfersByFriend(friendID)
	transfer := transfers[0]

	saveDir := filepath.Join(tempDir, "downloads")
	err = manager.AcceptIncomingFile(transfer.ID, saveDir)
	if err != nil {
		t.Fatalf("Failed to accept incoming file: %v", err)
	}

	// Simulate receiving file chunks
	testData := []byte("Hello, World!")
	mockTox.TriggerFileRecvChunk(friendID, fileID, 0, testData)

	// Verify file was written
	expectedPath := filepath.Join(saveDir, fileName)
	writtenData, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}

	if string(writtenData) != string(testData) {
		t.Errorf("Expected written data %s, got %s", testData, writtenData)
	}

	if transfer.State != TransferStateCompleted {
		t.Error("Expected transfer to be completed after receiving all data")
	}
}

func TestHandleFileChunkRequest(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create transfer manager: %v", err)
	}

	// Create test file and transfer
	testContent := "Hello, World!"
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	transfer, err := manager.SendFile(123, testFile)
	if err != nil {
		t.Fatalf("Failed to create transfer: %v", err)
	}

	var sentData []byte
	mockTox := &MockToxManager{
		fileSendChunkFunc: func(friendID uint32, fileID uint32, position uint64, data []byte) error {
			sentData = make([]byte, len(data))
			copy(sentData, data)
			return nil
		},
	}
	manager.SetToxManager(mockTox)

	// Start the transfer
	if err := manager.StartSend(transfer, mockTox); err != nil {
		t.Fatalf("Failed to start transfer: %v", err)
	}

	// Simulate chunk request
	mockTox.TriggerFileChunkRequest(transfer.FriendID, transfer.FileID, 0, 5)

	// Verify chunk was sent
	expectedChunk := testContent[:5]
	if string(sentData) != expectedChunk {
		t.Errorf("Expected sent chunk %s, got %s", expectedChunk, sentData)
	}
}

func TestErrorCases(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create transfer manager: %v", err)
	}

	mockTox := &MockToxManager{}

	// Test operations on non-existent transfer
	err = manager.PauseTransfer("non-existent", mockTox)
	if err == nil {
		t.Error("Expected error when pausing non-existent transfer")
	}

	err = manager.ResumeTransfer("non-existent", mockTox)
	if err == nil {
		t.Error("Expected error when resuming non-existent transfer")
	}

	err = manager.CancelTransfer("non-existent", mockTox)
	if err == nil {
		t.Error("Expected error when cancelling non-existent transfer")
	}

	err = manager.AcceptIncomingFile("non-existent", tempDir)
	if err == nil {
		t.Error("Expected error when accepting non-existent transfer")
	}

	// Test GetTransfer with non-existent ID
	_, exists := manager.GetTransfer("non-existent")
	if exists {
		t.Error("Expected false when getting non-existent transfer")
	}
}

func TestStartSendErrors(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create transfer manager: %v", err)
	}

	// Create incoming transfer to test wrong direction
	transfer := &Transfer{
		ID:        "test-id",
		Direction: TransferDirectionIncoming,
		State:     TransferStatePending,
	}
	manager.transfers[transfer.ID] = transfer

	mockTox := &MockToxManager{}
	err = manager.StartSend(transfer, mockTox)
	if err == nil {
		t.Error("Expected error when starting send on incoming transfer")
	}

	// Test wrong state
	transfer.Direction = TransferDirectionOutgoing
	transfer.State = TransferStateCompleted
	err = manager.StartSend(transfer, mockTox)
	if err == nil {
		t.Error("Expected error when starting send on completed transfer")
	}
}

func TestAcceptIncomingFileErrors(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create transfer manager: %v", err)
	}

	// Create outgoing transfer to test wrong direction
	transfer := &Transfer{
		ID:        "test-id",
		Direction: TransferDirectionOutgoing,
		State:     TransferStatePending,
	}
	manager.transfers[transfer.ID] = transfer

	err = manager.AcceptIncomingFile(transfer.ID, tempDir)
	if err == nil {
		t.Error("Expected error when accepting outgoing transfer")
	}

	// Test wrong state
	transfer.Direction = TransferDirectionIncoming
	transfer.State = TransferStateCompleted
	err = manager.AcceptIncomingFile(transfer.ID, tempDir)
	if err == nil {
		t.Error("Expected error when accepting completed transfer")
	}
}

func TestTransferControlErrors(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create transfer manager: %v", err)
	}

	// Create transfer in wrong state for operations
	transfer := &Transfer{
		ID:    "test-id",
		State: TransferStatePending,
	}
	manager.transfers[transfer.ID] = transfer

	mockTox := &MockToxManager{}

	// Test pause on pending transfer
	err = manager.PauseTransfer(transfer.ID, mockTox)
	if err == nil {
		t.Error("Expected error when pausing pending transfer")
	}

	// Test resume on active transfer
	transfer.State = TransferStateActive
	err = manager.ResumeTransfer(transfer.ID, mockTox)
	if err == nil {
		t.Error("Expected error when resuming active transfer")
	}

	// Test cancel on completed transfer
	transfer.State = TransferStateCompleted
	err = manager.CancelTransfer(transfer.ID, mockTox)
	if err == nil {
		t.Error("Expected error when cancelling completed transfer")
	}
}
