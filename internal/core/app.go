package core

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/opd-ai/toxcore"
	"github.com/opd-ai/whisp/internal/core/contact"
	"github.com/opd-ai/whisp/internal/core/message"
	"github.com/opd-ai/whisp/internal/core/security"
	"github.com/opd-ai/whisp/internal/core/tox"
	"github.com/opd-ai/whisp/internal/storage"
	"github.com/opd-ai/whisp/ui/adaptive"
)

// Config holds application configuration
type Config struct {
	DataDir    string
	ConfigPath string
	Debug      bool
	Platform   adaptive.Platform
}

// App represents the core application logic
type App struct {
	config   *Config
	tox      *tox.Manager
	storage  *storage.Database
	contacts *contact.Manager
	messages *message.Manager
	security *security.Manager

	mu       sync.RWMutex
	running  bool
	shutdown chan struct{}
}

// NewApp creates a new application instance
func NewApp(config *Config) (*App, error) {
	// Initialize database
	dbPath := filepath.Join(config.DataDir, "whisp.db")
	db, err := storage.NewDatabase(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize security manager
	securityMgr, err := security.NewManager(config.DataDir)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize security: %w", err)
	}

	// Initialize Tox manager
	toxMgr, err := tox.NewManager(&tox.Config{
		DataDir: config.DataDir,
		Debug:   config.Debug,
	})
	if err != nil {
		db.Close()
		securityMgr.Cleanup()
		return nil, fmt.Errorf("failed to initialize Tox: %w", err)
	}

	// Initialize contact manager
	contactMgr := contact.NewManager(db, toxMgr)

	// Initialize message manager
	messageMgr := message.NewManager(db, toxMgr, contactMgr)

	app := &App{
		config:   config,
		tox:      toxMgr,
		storage:  db,
		contacts: contactMgr,
		messages: messageMgr,
		security: securityMgr,
		shutdown: make(chan struct{}),
	}

	// Set up Tox callbacks
	if err := app.setupToxCallbacks(); err != nil {
		app.Cleanup()
		return nil, fmt.Errorf("failed to setup Tox callbacks: %w", err)
	}

	return app, nil
}

// Start starts the application
func (a *App) Start(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.running {
		return fmt.Errorf("application already running")
	}

	// Start Tox
	if err := a.tox.Start(); err != nil {
		return fmt.Errorf("failed to start Tox: %w", err)
	}

	a.running = true

	// Start main loop
	go a.mainLoop(ctx)

	log.Println("Application started successfully")
	return nil
}

// Stop stops the application
func (a *App) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running {
		return nil
	}

	close(a.shutdown)
	a.running = false

	if err := a.tox.Stop(); err != nil {
		log.Printf("Error stopping Tox: %v", err)
	}

	log.Println("Application stopped")
	return nil
}

// Cleanup cleans up resources
func (a *App) Cleanup() {
	if a.tox != nil {
		a.tox.Cleanup()
	}
	if a.storage != nil {
		a.storage.Close()
	}
	if a.security != nil {
		a.security.Cleanup()
	}
}

// IsRunning returns whether the application is running
func (a *App) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.running
}

// GetToxID returns the current Tox ID
func (a *App) GetToxID() string {
	return a.tox.GetToxID()
}

// GetContacts returns the contact manager
func (a *App) GetContacts() *contact.Manager {
	return a.contacts
}

// GetMessages returns the message manager
func (a *App) GetMessages() *message.Manager {
	return a.messages
}

// GetSecurity returns the security manager
func (a *App) GetSecurity() *security.Manager {
	return a.security
}

// mainLoop runs the main application loop
func (a *App) mainLoop(ctx context.Context) {
	ticker := time.NewTicker(50 * time.Millisecond) // 20 FPS
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.shutdown:
			return
		case <-ticker.C:
			// Update Tox
			a.tox.Iterate()

			// Process pending messages
			a.messages.ProcessPending()
		}
	}
}

// setupToxCallbacks sets up Tox event callbacks
func (a *App) setupToxCallbacks() error {
	// Friend request callback
	a.tox.OnFriendRequest(func(publicKey [32]byte, message string) {
		log.Printf("Friend request received: %s", message)
		// Add to pending friend requests
		a.contacts.HandleFriendRequest(publicKey, message)
	})

	// Friend message callback
	a.tox.OnFriendMessage(func(friendID uint32, msg string) {
		log.Printf("Message from friend %d: %s", friendID, msg)
		// Handle incoming message
		a.messages.HandleIncomingMessage(friendID, msg, message.MessageTypeNormal)
	})

	// Friend status callback
	a.tox.OnFriendStatus(func(friendID uint32, status toxcore.FriendStatus) {
		log.Printf("Friend %d status: %v", friendID, status)
		a.contacts.UpdateStatus(friendID, status)
	})

	// Friend name callback
	a.tox.OnFriendName(func(friendID uint32, name string) {
		log.Printf("Friend %d name: %s", friendID, name)
		a.contacts.UpdateName(friendID, name)
	})

	return nil
}
