package core

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/opd-ai/toxcore"
	"github.com/opd-ai/whisp/internal/core/audio"
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
	audio         audio.Manager
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

	// Initialize audio manager
	audioMgr := audio.NewMockManager()
	if err := audioMgr.Initialize(); err != nil {
		db.Close()
		securityMgr.Cleanup()
		return nil, fmt.Errorf("failed to initialize audio manager: %w", err)
	}

	app := &App{
		config:    config,
		configMgr: configMgr,
		tox:       toxMgr,
		storage:   db,
		contacts:  contactMgr,
		messages:  messageMgr,
		security:  securityMgr,
		transfers: transferMgr,
		audio:     audioMgr,
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
	if a.audio != nil {
		a.audio.Shutdown()
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

// === Voice Message Methods ===

// GetAudioManager returns the audio manager
func (a *App) GetAudioManager() audio.Manager {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.audio
}

// StartVoiceRecordingFromUI starts voice recording from the UI
func (a *App) StartVoiceRecordingFromUI(friendID uint32, outputDir string) (audio.Recorder, error) {
	log.Printf("Starting voice recording from UI: friend=%d, outputDir=%s", friendID, outputDir)

	recorder, err := a.audio.GetRecorder()
	if err != nil {
		return nil, fmt.Errorf("failed to get recorder: %w", err)
	}

	if !recorder.IsSupported() {
		return nil, fmt.Errorf("voice recording not supported on this system")
	}

	// Create output path
	outputPath := filepath.Join(outputDir, fmt.Sprintf("voice_%d_%d.wav", friendID, time.Now().Unix()))

	// Configure recording options
	options := audio.DefaultRecordingOptions()
	options.OutputPath = outputPath
	options.MaxDuration = 5 * time.Minute // 5 minute max for voice messages

	// Start recording
	ctx := context.Background()
	if err := recorder.Start(ctx, options, nil); err != nil {
		return nil, fmt.Errorf("failed to start recording: %w", err)
	}

	return recorder, nil
}

// SendVoiceMessageFromUI sends a completed voice recording as a message
func (a *App) SendVoiceMessageFromUI(friendID uint32, voiceMsg *audio.VoiceMessage) error {
	log.Printf("Sending voice message from UI: friend=%d, file=%s, duration=%v",
		friendID, voiceMsg.FilePath, voiceMsg.Duration)

	// Create voice message in database
	content := fmt.Sprintf("Voice message (%.1fs)", voiceMsg.Duration.Seconds())
	msg, err := a.messages.SendMessage(friendID, content, message.MessageTypeVoice)
	if err != nil {
		return fmt.Errorf("failed to create voice message: %w", err)
	}

	// Update message with file metadata
	// Note: In a full implementation, we'd need a method to update message file info
	log.Printf("Voice message created with ID: %d", msg.ID)

	// Send file through transfer system
	transferID, err := a.SendFileFromUI(friendID, voiceMsg.FilePath)
	if err != nil {
		return fmt.Errorf("failed to send voice file: %w", err)
	}

	log.Printf("Voice message sent with transfer ID: %s", transferID)
	return nil
}

// PlayVoiceMessageFromUI plays a voice message from the UI
func (a *App) PlayVoiceMessageFromUI(filePath string) (audio.Player, error) {
	log.Printf("Playing voice message from UI: file=%s", filePath)

	player, err := a.audio.GetPlayer()
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	if !player.IsSupported() {
		return nil, fmt.Errorf("voice playback not supported on this system")
	}

	// Load the audio file
	if err := player.Load(filePath); err != nil {
		return nil, fmt.Errorf("failed to load audio file: %w", err)
	}

	// Start playback
	options := audio.DefaultPlaybackOptions()
	if err := player.Play(options, nil); err != nil {
		return nil, fmt.Errorf("failed to start playback: %w", err)
	}

	return player, nil
}

// GenerateWaveformFromUI generates waveform data for UI visualization
func (a *App) GenerateWaveformFromUI(filePath string, points int) ([]float32, error) {
	log.Printf("Generating waveform from UI: file=%s, points=%d", filePath, points)

	generator := a.audio.GetWaveformGenerator()
	return generator.GenerateWaveformFromFile(filePath, points)
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
