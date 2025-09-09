package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/opd-ai/whisp/internal/core/app"
	"github.com/opd-ai/whisp/internal/storage"
	"github.com/opd-ai/whisp/platform/common"
	"github.com/opd-ai/whisp/ui/adaptive"
)

var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

func main() {
	// Parse command line flags
	var (
		debug      = flag.Bool("debug", false, "Enable debug logging")
		dataDir    = flag.String("data-dir", "", "Custom data directory")
		configPath = flag.String("config", "", "Custom config file path")
		version    = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *version {
		fmt.Printf("Whisp %s (built %s, commit %s)\n", version, buildTime, gitCommit)
		fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	// Initialize platform detection
	platform := adaptive.DetectPlatform()
	log.Printf("Detected platform: %s", platform)

	// Set up data directory
	if *dataDir == "" {
		userDataDir, err := common.GetUserDataDir()
		if err != nil {
			log.Fatal("Failed to get user data directory:", err)
		}
		*dataDir = filepath.Join(userDataDir, "whisp")
	}

	// Ensure data directory exists
	if err := os.MkdirAll(*dataDir, 0700); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	// Initialize storage
	dbPath := filepath.Join(*dataDir, "whisp.db")
	db, err := storage.NewDatabase(dbPath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize application core
	coreApp, err := core.NewApp(&core.Config{
		DataDir:    *dataDir,
		ConfigPath: *configPath,
		Debug:      *debug,
		Platform:   platform,
	})
	if err != nil {
		log.Fatal("Failed to initialize application core:", err)
	}
	defer coreApp.Cleanup()

	// Create Fyne application
	fyneApp := app.NewWithID("com.opd-ai.whisp")
	fyneApp.SetMetadata(&fyne.AppMetadata{
		ID:      "com.opd-ai.whisp",
		Name:    "Whisp",
		Version: version,
		Icon:    resourceIconPng,
	})

	// Initialize adaptive UI
	ui, err := adaptive.NewUI(fyneApp, coreApp, platform)
	if err != nil {
		log.Fatal("Failed to initialize UI:", err)
	}

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down gracefully...")
		cancel()
		fyneApp.Quit()
	}()

	// Start application
	log.Printf("Starting Whisp %s on %s", version, platform)
	
	// Platform-specific initialization
	if err := ui.Initialize(ctx); err != nil {
		log.Fatal("Failed to initialize UI:", err)
	}

	// Show main window and run
	ui.ShowMainWindow()
	fyneApp.Run()
}
