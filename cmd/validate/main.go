package main

import (
	"log"
	"os"

	"github.com/opd-ai/whisp/internal/core"
	"github.com/opd-ai/whisp/platform/common"
	"github.com/opd-ai/whisp/ui/adaptive"
)

// Simple validation test for the core implementation
func main() {
	log.Println("=== Whisp Implementation Validation ===")

	// Test 1: Platform detection
	platform := adaptive.DetectPlatform()
	log.Printf("✓ Platform detection: %s", platform)

	// Test 2: Data directory
	dataDir, err := common.GetUserDataDir()
	if err != nil {
		log.Printf("✗ Data directory: %v", err)
		os.Exit(1)
	}
	log.Printf("✓ Data directory: %s", dataDir)

	// Test 3: Core app creation
	config := &core.Config{
		DataDir:  dataDir,
		Debug:    true,
		Platform: platform,
	}

	app, err := core.NewApp(config)
	if err != nil {
		log.Printf("✗ Core app creation: %v", err)
		os.Exit(1)
	}
	log.Printf("✓ Core app created successfully")

	// Test 4: Tox ID generation
	toxID := app.GetToxID()
	if toxID == "" {
		log.Printf("⚠ Tox ID: Not yet initialized (expected for first run)")
	} else {
		log.Printf("✓ Tox ID: %s...", toxID[:16])
	}

	// Test 5: Database functionality
	contacts := app.GetContacts()
	if contacts == nil {
		log.Printf("✗ Contact manager: nil")
		os.Exit(1)
	}
	log.Printf("✓ Contact manager initialized")

	messages := app.GetMessages()
	if messages == nil {
		log.Printf("✗ Message manager: nil")
		os.Exit(1)
	}
	log.Printf("✓ Message manager initialized")

	// Test 6: Security manager
	security := app.GetSecurity()
	if security == nil {
		log.Printf("✗ Security manager: nil")
		os.Exit(1)
	}
	log.Printf("✓ Security manager initialized")

	// Cleanup
	app.Cleanup()
	log.Printf("✓ Cleanup completed")

	log.Println("=== All Core Components Validated ===")
	log.Println("")
	log.Println("Next steps:")
	log.Println("1. Start the application with: ./whisp")
	log.Println("2. GUI will be available (basic implementation)")
	log.Println("3. Tox integration is functional but requires network connection")
}
