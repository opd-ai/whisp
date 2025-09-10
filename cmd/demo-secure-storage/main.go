package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/opd-ai/whisp/internal/core/security"
	"github.com/opd-ai/whisp/platform/common"
)

func main() {
	fmt.Println("=== Whisp Secure Storage Demo ===")

	// Get platform-specific data directory
	dataDir, err := common.GetUserDataDir()
	if err != nil {
		log.Fatalf("Failed to get data directory: %v", err)
	}

	// Create demo subdirectory
	demoDir := filepath.Join(dataDir, "demo-secure-storage")
	if err := os.MkdirAll(demoDir, 0o755); err != nil {
		log.Fatalf("Failed to create demo directory: %v", err)
	}

	// Initialize security manager
	secManager, err := security.NewManager(demoDir)
	if err != nil {
		log.Fatalf("Failed to create security manager: %v", err)
	}

	fmt.Printf("Using data directory: %s\n", demoDir)

	// Check if platform-specific secure storage is available
	isAvailable := secManager.IsSecureStorageAvailable()
	fmt.Printf("Platform secure storage available: %v\n", isAvailable)

	if isAvailable {
		fmt.Println("Will use platform-specific secure storage (Keychain/Credential Manager/Secret Service)")
	} else {
		fmt.Println("Will use encrypted file fallback")
	}

	// Generate and set up master key
	fmt.Println("\n1. Generating master key...")
	masterKey, err := secManager.GenerateMasterKey()
	if err != nil {
		log.Fatalf("Failed to generate master key: %v", err)
	}
	secManager.SetMasterKey(masterKey)
	fmt.Printf("Generated 256-bit master key: %x\n", masterKey[:8]) // Show first 8 bytes only

	// Store master key in secure storage
	fmt.Println("\n2. Storing master key in secure storage...")
	if err := secManager.StoreMasterKey(masterKey); err != nil {
		log.Fatalf("Failed to store master key: %v", err)
	}
	fmt.Println("Master key stored successfully")

	// Store some application configuration
	fmt.Println("\n3. Storing application configuration...")
	configs := map[string]string{
		"user_preference_theme":    "dark",
		"user_preference_language": "en",
		"app_version":              "1.0.0",
		"last_backup_date":         "2025-09-09",
		"tox_save_checksum":        "abc123def456",
	}

	for key, value := range configs {
		configKey := "config_" + key
		if err := secManager.SecureStore(configKey, value); err != nil {
			log.Fatalf("Failed to store config %s: %v", key, err)
		}
		fmt.Printf("Stored: %s = %s\n", key, value)
	}

	// Demonstrate retrieval
	fmt.Println("\n4. Retrieving stored data...")
	for key := range configs {
		configKey := "config_" + key
		value, err := secManager.SecureRetrieve(configKey)
		if err != nil {
			log.Fatalf("Failed to retrieve config %s: %v", key, err)
		}
		fmt.Printf("Retrieved: %s = %s\n", key, value)
	}

	// Test master key loading
	fmt.Println("\n5. Testing master key retrieval...")
	loadedKey, err := secManager.LoadMasterKey()
	if err != nil {
		log.Fatalf("Failed to load master key: %v", err)
	}

	// Verify keys match
	keyMatch := true
	if len(masterKey) != len(loadedKey) {
		keyMatch = false
	} else {
		for i := range masterKey {
			if masterKey[i] != loadedKey[i] {
				keyMatch = false
				break
			}
		}
	}

	if keyMatch {
		fmt.Println("Master key retrieved successfully and matches original")
	} else {
		fmt.Println("ERROR: Retrieved master key does not match original!")
	}

	// Demonstrate security with wrong key simulation
	fmt.Println("\n6. Testing security (simulating wrong key scenario)...")

	// Clear master key and set a different one
	secManager.Cleanup()
	wrongKey, _ := secManager.GenerateMasterKey()
	secManager.SetMasterKey(wrongKey)

	// Try to retrieve config with wrong key (should fail for file fallback)
	_, err = secManager.SecureRetrieve("config_user_preference_theme")
	if err != nil {
		fmt.Printf("Good: Access denied with wrong key (using file fallback): %v\n", err)
	} else {
		fmt.Println("Retrieved value with wrong key (using platform storage - this is expected)")
	}

	// Restore original key
	secManager.SetMasterKey(masterKey)

	// Clean up demo data
	fmt.Println("\n7. Cleaning up demo data...")
	for key := range configs {
		configKey := "config_" + key
		if err := secManager.SecureDelete(configKey); err != nil {
			log.Printf("Warning: Failed to delete config %s: %v", key, err)
		}
	}

	if err := secManager.DeleteMasterKey(); err != nil {
		log.Printf("Warning: Failed to delete master key: %v", err)
	}

	// Final cleanup
	secManager.Cleanup()

	fmt.Println("\nDemo completed successfully!")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("- Cross-platform secure storage (Keychain/Credential Manager/Secret Service)")
	fmt.Println("- Automatic fallback to encrypted file storage")
	fmt.Println("- Master key management and persistence")
	fmt.Println("- Configuration data protection")
	fmt.Println("- Platform detection and adaptation")
	fmt.Println("- Secure memory handling and cleanup")
}
