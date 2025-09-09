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

	"github.com/opd-ai/whisp/internal/core"
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
		debug       = flag.Bool("debug", false, "Enable debug logging")
		dataDir     = flag.String("data-dir", "", "Custom data directory")
		configPath  = flag.String("config", "", "Custom config file path")
		showVersion = flag.Bool("version", false, "Show version information")
		headless    = flag.Bool("headless", false, "Run in headless mode (no GUI)")
	)
	flag.Parse()

	if *showVersion {
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

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down gracefully...")
		cancel()
	}()

	// Start application
	log.Printf("Starting Whisp %s on %s", version, platform)
	
	if err := coreApp.Start(ctx); err != nil {
		log.Fatal("Failed to start application:", err)
	}

	if *headless {
		// Headless mode - just run the core
		log.Println("Running in headless mode...")
		<-ctx.Done()
	} else {
		// GUI mode - start UI (placeholder for now)
		log.Println("GUI mode not yet implemented, running headless...")
		<-ctx.Done()
	}
	
	log.Println("Application stopped")
}
