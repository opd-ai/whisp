package tox

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/opd-ai/toxcore"
)

// TestManager_NewManager tests creating a new Tox manager
func TestManager_NewManager(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				DataDir: t.TempDir(),
				Debug:   false,
			},
			wantErr: false,
		},
		{
			name: "valid config with debug",
			config: &Config{
				DataDir: t.TempDir(),
				Debug:   true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewManager(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if manager == nil {
					t.Error("NewManager() returned nil manager")
				}
				// Cleanup
				if manager != nil {
					manager.Cleanup()
				}
			}
		})
	}
}

// TestManager_SaveAndLoad tests saving and loading Tox state
func TestManager_SaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	config := &Config{
		DataDir: tempDir,
		Debug:   false,
	}

	// Create first manager and save state
	manager1, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create first manager: %v", err)
	}
	defer manager1.Cleanup()

	// Get initial Tox ID
	toxID1 := manager1.GetToxID()
	if toxID1 == "" {
		t.Error("Expected non-empty Tox ID")
	}

	// Save state
	err = manager1.Save()
	if err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Verify save file exists
	saveFile := filepath.Join(tempDir, "tox.save")
	if _, err := os.Stat(saveFile); os.IsNotExist(err) {
		t.Error("Save file was not created")
	}

	// Create second manager with same config (should load existing state)
	manager2, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create second manager: %v", err)
	}
	defer manager2.Cleanup()

	// Verify Tox ID is preserved
	toxID2 := manager2.GetToxID()
	if toxID2 != toxID1 {
		t.Errorf("Tox ID not preserved. Expected %s, got %s", toxID1, toxID2)
	}
}

// TestManager_LifecycleMethods tests Start, Stop, and Cleanup
func TestManager_LifecycleMethods(t *testing.T) {
	config := &Config{
		DataDir: t.TempDir(),
		Debug:   false,
	}

	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Cleanup()

	// Test Start
	err = manager.Start()
	if err != nil {
		t.Errorf("Start() failed: %v", err)
	}

	// Test double start (should fail)
	err = manager.Start()
	if err == nil {
		t.Error("Expected error on double start")
	}

	// Test Stop
	err = manager.Stop()
	if err != nil {
		t.Errorf("Stop() failed: %v", err)
	}

	// Test double stop (should not fail)
	err = manager.Stop()
	if err != nil {
		t.Errorf("Double stop failed: %v", err)
	}
}

// TestManager_FriendOperations tests friend management
func TestManager_FriendOperations(t *testing.T) {
	config := &Config{
		DataDir: t.TempDir(),
		Debug:   false,
	}

	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Cleanup()

	err = manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	// Test GetFriends (should be empty initially)
	friends := manager.GetFriends()
	if len(friends) != 0 {
		t.Errorf("Expected 0 friends, got %d", len(friends))
	}

	// Note: We can't easily test actual friend operations without a second Tox instance
	// These would require integration tests with actual Tox network
}

// TestManager_SelfMethods tests self information methods
func TestManager_SelfMethods(t *testing.T) {
	config := &Config{
		DataDir: t.TempDir(),
		Debug:   false,
	}

	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Cleanup()

	// Test SetName and GetName
	testName := "Test User"
	err = manager.SetName(testName)
	if err != nil {
		t.Errorf("SetName() failed: %v", err)
	}

	name := manager.GetName()
	if name != testName {
		t.Errorf("Expected name %s, got %s", testName, name)
	}

	// Test SetStatusMessage and GetStatusMessage
	testStatus := "Testing Whisp"
	err = manager.SetStatusMessage(testStatus)
	if err != nil {
		t.Errorf("SetStatusMessage() failed: %v", err)
	}

	status := manager.GetStatusMessage()
	if status != testStatus {
		t.Errorf("Expected status %s, got %s", testStatus, status)
	}

	// Test GetToxID
	toxID := manager.GetToxID()
	if toxID == "" {
		t.Error("Expected non-empty Tox ID")
	}
	if len(toxID) != 76 { // Tox ID is 76 hex characters
		t.Errorf("Expected Tox ID length 76, got %d", len(toxID))
	}
}

// TestManager_CallbackMethods tests callback registration
func TestManager_CallbackMethods(t *testing.T) {
	config := &Config{
		DataDir: t.TempDir(),
		Debug:   false,
	}

	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Cleanup()

	// Test callback registration (these should not panic)
	manager.OnFriendRequest(func(publicKey [32]byte, message string) {
		// Test callback
	})

	manager.OnFriendMessage(func(friendID uint32, message string) {
		// Test callback
	})

	manager.OnFriendStatus(func(friendID uint32, status toxcore.FriendStatus) {
		// Test callback
	})

	manager.OnFriendName(func(friendID uint32, name string) {
		// Test callback
	})
}

// TestManager_SaveFileHandling tests file I/O error handling
func TestManager_SaveFileHandling(t *testing.T) {
	tempDir := t.TempDir()
	config := &Config{
		DataDir: tempDir,
		Debug:   false,
	}

	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Cleanup()

	// Test saving to read-only directory
	readOnlyDir := filepath.Join(tempDir, "readonly")
	err = os.Mkdir(readOnlyDir, 0555) // Read-only directory
	if err != nil {
		t.Fatalf("Failed to create read-only directory: %v", err)
	}
	defer os.Chmod(readOnlyDir, 0755) // Restore permissions for cleanup

	// Change save file to read-only directory
	manager.saveFile = filepath.Join(readOnlyDir, "tox.save")

	// This should fail due to permissions
	err = manager.Save()
	if err == nil {
		t.Error("Expected error when saving to read-only directory")
	}
}

// TestManager_LoadSavedataErrors tests loadSavedata error conditions
func TestManager_LoadSavedataErrors(t *testing.T) {
	tempDir := t.TempDir()

	manager := &Manager{
		saveFile: filepath.Join(tempDir, "nonexistent.save"),
	}

	// Test loading non-existent file
	_, err := manager.loadSavedata()
	if err == nil {
		t.Error("Expected error when loading non-existent file")
	}

	// Test loading empty file
	emptyFile := filepath.Join(tempDir, "empty.save")
	err = os.WriteFile(emptyFile, []byte{}, 0600)
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	manager.saveFile = emptyFile
	_, err = manager.loadSavedata()
	if err == nil {
		t.Error("Expected error when loading empty file")
	}
}

// TestManager_Iterate tests the iteration method
func TestManager_Iterate(t *testing.T) {
	config := &Config{
		DataDir: t.TempDir(),
		Debug:   false,
	}

	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Cleanup()

	err = manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	// Test iteration (should not panic)
	for i := 0; i < 5; i++ {
		manager.Iterate()
		time.Sleep(10 * time.Millisecond)
	}

	// Test iteration when stopped
	err = manager.Stop()
	if err != nil {
		t.Errorf("Failed to stop manager: %v", err)
	}

	// Should not panic even when stopped
	manager.Iterate()
}

// TestManager_SaveStateOnCleanup tests that state is saved during cleanup
func TestManager_SaveStateOnCleanup(t *testing.T) {
	tempDir := t.TempDir()
	config := &Config{
		DataDir: tempDir,
		Debug:   false,
	}

	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Set some state
	testName := "Cleanup Test User"
	err = manager.SetName(testName)
	if err != nil {
		t.Fatalf("Failed to set name: %v", err)
	}

	// Cleanup (should save state)
	manager.Cleanup()

	// Verify save file was created
	saveFile := filepath.Join(tempDir, "tox.save")
	if _, err := os.Stat(saveFile); os.IsNotExist(err) {
		t.Error("Save file was not created during cleanup")
	}

	// Create new manager and verify state was preserved
	manager2, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create second manager: %v", err)
	}
	defer manager2.Cleanup()

	name := manager2.GetName()
	if name != testName {
		t.Errorf("Name not preserved after cleanup. Expected %s, got %s", testName, name)
	}
}

// Benchmark tests
func BenchmarkManager_Save(b *testing.B) {
	tempDir := b.TempDir()
	config := &Config{
		DataDir: tempDir,
		Debug:   false,
	}

	manager, err := NewManager(config)
	if err != nil {
		b.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = manager.Save()
		if err != nil {
			b.Fatalf("Save failed: %v", err)
		}
	}
}

func BenchmarkManager_GetToxID(b *testing.B) {
	tempDir := b.TempDir()
	config := &Config{
		DataDir: tempDir,
		Debug:   false,
	}

	manager, err := NewManager(config)
	if err != nil {
		b.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.GetToxID()
	}
}
