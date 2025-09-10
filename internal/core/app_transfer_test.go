package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/opd-ai/whisp/ui/adaptive"
)

// TestFileTransferIntegration tests the integration of file transfer functionality with the core app
func TestFileTransferIntegration(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Initialize core application with file transfer support
	config := &Config{
		DataDir:    tempDir,
		ConfigPath: filepath.Join(tempDir, "config.yaml"),
		Debug:      true,
		Platform:   adaptive.PlatformLinux,
	}

	app, err := NewApp(config)
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}
	defer app.Cleanup()

	// Test that transfer manager is initialized
	transferMgr := app.GetTransfers()
	if transferMgr == nil {
		t.Fatal("Transfer manager is nil")
	}

	// Create a test file
	testContent := "Hello, World! File transfer test."
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test sending file via UI method
	friendID := uint32(123)
	transferID, err := app.SendFileFromUI(friendID, testFile)
	if err != nil {
		t.Fatalf("Failed to send file via UI: %v", err)
	}

	if transferID == "" {
		t.Fatal("Transfer ID is empty")
	}

	// Verify transfer was created
	transfer, exists := transferMgr.GetTransfer(transferID)
	if !exists {
		t.Fatal("Transfer not found in manager")
	}

	// Verify transfer properties
	if transfer.FriendID != friendID {
		t.Errorf("Expected friend ID %d, got %d", friendID, transfer.FriendID)
	}

	if transfer.FileName != "test.txt" {
		t.Errorf("Expected file name test.txt, got %s", transfer.FileName)
	}

	if transfer.FileSize != uint64(len(testContent)) {
		t.Errorf("Expected file size %d, got %d", len(testContent), transfer.FileSize)
	}

	// Test file size limits
	originalLimit := transferMgr.GetMaxFileSize()
	transferMgr.SetMaxFileSize(10) // Very small limit

	_, err = app.SendFileFromUI(456, testFile)
	if err == nil {
		t.Error("Expected error for file size exceeding limit")
	}

	// Restore original limit
	transferMgr.SetMaxFileSize(originalLimit)

	// Test with non-existent file
	_, err = app.SendFileFromUI(789, "/non/existent/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

// TestFileTransferUIIntegration tests the UI integration methods
func TestFileTransferUIIntegration(t *testing.T) {
	tempDir := t.TempDir()

	config := &Config{
		DataDir:    tempDir,
		ConfigPath: filepath.Join(tempDir, "config.yaml"),
		Debug:      true,
		Platform:   adaptive.PlatformLinux,
	}

	app, err := NewApp(config)
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}
	defer app.Cleanup()

	// Test that all required methods exist and work
	transferMgr := app.GetTransfers()
	if transferMgr == nil {
		t.Fatal("Transfer manager not available")
	}

	// Create test file
	testFile := filepath.Join(tempDir, "ui_test.txt")
	testContent := "UI integration test content"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test SendFileFromUI
	transferID, err := app.SendFileFromUI(100, testFile)
	if err != nil {
		t.Errorf("SendFileFromUI failed: %v", err)
	}
	if transferID == "" {
		t.Error("Transfer ID should not be empty")
	}

	// Test CancelFileFromUI
	err = app.CancelFileFromUI(transferID)
	if err != nil {
		t.Errorf("CancelFileFromUI failed: %v", err)
	}

	// Test AcceptFileFromUI with invalid transfer ID
	err = app.AcceptFileFromUI("invalid-id", tempDir)
	if err == nil {
		t.Error("Expected error for invalid transfer ID")
	}

	// Test CancelFileFromUI with invalid transfer ID
	err = app.CancelFileFromUI("invalid-id")
	if err == nil {
		t.Error("Expected error for invalid transfer ID")
	}
}
