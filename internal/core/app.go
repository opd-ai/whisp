package core

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/opd-ai/toxcore"
	configpkg "github.com/opd-ai/whisp/internal/core/config"
	"github.com/opd-ai/whisp/internal/core/contact"
	"github.com/opd-ai/whisp/internal/core/message"
	"github.com/opd-ai/whisp/internal/core/security"
	"github.com/opd-ai/whisp/internal/core/tox"
	"github.com/opd-ai/whisp/internal/core/transfer"
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
	config        *Config
	configMgr     *configpkg.Manager
	tox           *tox.Manager
	storage       *storage.Database
	contacts      *contact.Manager
	messages      *message.Manager
	security      *security.Manager
	transfers     *transfer.Manager
	notifications *NotificationService

	mu       sync.RWMutex
	running  bool
	shutdown chan struct{}
}

// NewApp creates a new application instance
func NewApp(config *Config) (*App, error) {
	// Initialize configuration manager
	configMgr, err := configpkg.NewManager(config.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize configuration: %w", err)
	}

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

	// Initialize file transfer manager
	transferMgr, err := transfer.NewManager(config.DataDir)
	if err != nil {
		db.Close()
		securityMgr.Cleanup()
		return nil, fmt.Errorf("failed to initialize file transfer manager: %w", err)
	}

	// Connect transfer manager to Tox
	transferMgr.SetToxManager(toxMgr)

	app := &App{
		config:    config,
		configMgr: configMgr,
		tox:       toxMgr,
		storage:   db,
		contacts:  contactMgr,
		messages:  messageMgr,
		security:  securityMgr,
		transfers: transferMgr,
		shutdown:  make(chan struct{}),
	}

	// Initialize notification service
	app.notifications = NewNotificationService(app)

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

	// Start notification service
	if err := a.notifications.Start(ctx); err != nil {
		log.Printf("Warning: Failed to start notification service: %v", err)
		// Don't fail startup for notification issues
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

	// Stop notification service
	if a.notifications != nil {
		if err := a.notifications.Stop(); err != nil {
			log.Printf("Error stopping notification service: %v", err)
		}
	}

	log.Println("Application stopped")
	return nil
}

// Cleanup cleans up resources
func (a *App) Cleanup() {
	if a.notifications != nil {
		a.notifications.Stop()
	}
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

// GetNotifications returns the notification service
func (a *App) GetNotifications() *NotificationService {
	return a.notifications
}

// GetSecurity returns the security manager
func (a *App) GetSecurity() *security.Manager {
	return a.security
}

// GetTransfers returns the file transfer manager
func (a *App) GetTransfers() *transfer.Manager {
	return a.transfers
}

// GetConfigManager returns the configuration manager
func (a *App) GetConfigManager() *configpkg.Manager {
	return a.configMgr
}

// SendMessageFromUI sends a message from the UI
func (a *App) SendMessageFromUI(friendID uint32, content string) error {
	if content == "" {
		return fmt.Errorf("message content cannot be empty")
	}

	_, err := a.messages.SendMessage(friendID, content, message.MessageTypeNormal)
	return err
}

// AddContactFromUI adds a contact from the UI
func (a *App) AddContactFromUI(toxID, message string) error {
	log.Printf("Adding contact from UI: %s", toxID)

	// Validate Tox ID format (basic validation)
	if len(toxID) != 76 {
		return fmt.Errorf("invalid Tox ID length: expected 76 characters, got %d", len(toxID))
	}

	// Add contact through contact manager
	_, err := a.contacts.AddContact(toxID, message)
	if err != nil {
		return fmt.Errorf("failed to add contact: %w", err)
	}

	return nil
}

// SendFileFromUI initiates a file transfer from the UI
func (a *App) SendFileFromUI(friendID uint32, filePath string) (string, error) {
	log.Printf("Sending file from UI: friend=%d, file=%s", friendID, filePath)

	// Create file transfer through transfer manager
	transfer, err := a.transfers.SendFile(friendID, filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file transfer: %w", err)
	}

	// Start the transfer
	if err := a.transfers.StartSend(transfer, a.tox); err != nil {
		return "", fmt.Errorf("failed to start file transfer: %w", err)
	}

	return transfer.ID, nil
}

// AcceptFileFromUI accepts an incoming file transfer from the UI
func (a *App) AcceptFileFromUI(transferID, saveDir string) error {
	log.Printf("Accepting file transfer from UI: transfer=%s, saveDir=%s", transferID, saveDir)

	return a.transfers.AcceptIncomingFile(transferID, saveDir)
}

// CancelFileFromUI cancels a file transfer from the UI
func (a *App) CancelFileFromUI(transferID string) error {
	log.Printf("Cancelling file transfer from UI: transfer=%s", transferID)

	return a.transfers.CancelTransfer(transferID, a.tox)
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
