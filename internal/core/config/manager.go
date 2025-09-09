package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Manager handles application configuration
// Uses established libraries: yaml.v3 for parsing, standard library for file I/O
type Manager struct {
	configPath string
	config     *Config
}

// Config represents the complete application configuration
// Maps directly to config.yaml structure for simplicity
type Config struct {
	Network struct {
		BootstrapNodes []struct {
			Address   string `yaml:"address"`
			Port      int    `yaml:"port"`
			PublicKey string `yaml:"public_key"`
		} `yaml:"bootstrap_nodes"`
		EnableIPv6           bool `yaml:"enable_ipv6"`
		EnableUDP            bool `yaml:"enable_udp"`
		EnableTCP            bool `yaml:"enable_tcp"`
		EnableLocalDiscovery bool `yaml:"enable_local_discovery"`
		EnableHolePunching   bool `yaml:"enable_hole_punching"`
		Proxy                struct {
			Type     string `yaml:"type"`
			Address  string `yaml:"address"`
			Port     int    `yaml:"port"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"proxy"`
	} `yaml:"network"`

	Storage struct {
		DataDir               string `yaml:"data_dir"`
		EnableEncryption      bool   `yaml:"enable_encryption"`
		MaxFileSize           int64  `yaml:"max_file_size"`
		DownloadDir           string `yaml:"download_dir"`
		MaxMessageHistoryDays int    `yaml:"max_message_history_days"`
		AutoDeleteMediaDays   int    `yaml:"auto_delete_media_days"`
	} `yaml:"storage"`

	UI struct {
		Theme              string `yaml:"theme"`
		Language           string `yaml:"language"`
		FontFamily         string `yaml:"font_family"`
		FontSize           string `yaml:"font_size"`
		EnableAnimations   bool   `yaml:"enable_animations"`
		EnableSoundEffects bool   `yaml:"enable_sound_effects"`
		Window             struct {
			RememberSize     bool `yaml:"remember_size"`
			RememberPosition bool `yaml:"remember_position"`
			MinimizeToTray   bool `yaml:"minimize_to_tray"`
			StartMinimized   bool `yaml:"start_minimized"`
		} `yaml:"window"`
		Mobile struct {
			VibrateOnMessage   bool `yaml:"vibrate_on_message"`
			ShowMessagePreview bool `yaml:"show_message_preview"`
		} `yaml:"mobile"`
	} `yaml:"ui"`

	Privacy struct {
		SaveMessageHistory           bool   `yaml:"save_message_history"`
		EnableDisappearingMessages   bool   `yaml:"enable_disappearing_messages"`
		DefaultDisappearingTimer     string `yaml:"default_disappearing_timer"`
		ShowTypingIndicators         bool   `yaml:"show_typing_indicators"`
		SendTypingIndicators         bool   `yaml:"send_typing_indicators"`
		ShowReadReceipts             bool   `yaml:"show_read_receipts"`
		SendReadReceipts             bool   `yaml:"send_read_receipts"`
		ShowLastSeen                 bool   `yaml:"show_last_seen"`
		AutoAcceptFiles              bool   `yaml:"auto_accept_files"`
		AutoDownloadLimit            int64  `yaml:"auto_download_limit"`
		PreventScreenshots           bool   `yaml:"prevent_screenshots"`
		AutoAcceptFriendRequests     bool   `yaml:"auto_accept_friend_requests"`
		RequireFriendRequestsMessage bool   `yaml:"require_friend_requests_message"`
	} `yaml:"privacy"`

	Notifications struct {
		Enabled bool `yaml:"enabled"`
		Desktop struct {
			ShowPreview bool `yaml:"show_preview"`
			PlaySound   bool `yaml:"play_sound"`
			ShowSender  bool `yaml:"show_sender"`
		} `yaml:"desktop"`
		Mobile struct {
			ShowOnLockScreen bool   `yaml:"show_on_lock_screen"`
			ShowPreview      bool   `yaml:"show_preview"`
			Vibrate          bool   `yaml:"vibrate"`
			LEDColor         string `yaml:"led_color"`
		} `yaml:"mobile"`
		QuietHours struct {
			Enabled   bool   `yaml:"enabled"`
			StartTime string `yaml:"start_time"`
			EndTime   string `yaml:"end_time"`
		} `yaml:"quiet_hours"`
	} `yaml:"notifications"`

	Advanced struct {
		LogLevel               string `yaml:"log_level"`
		LogToFile              bool   `yaml:"log_to_file"`
		MaxLogSize             int64  `yaml:"max_log_size"`
		MaxConcurrentDownloads int    `yaml:"max_concurrent_downloads"`
		MaxConcurrentUploads   int    `yaml:"max_concurrent_uploads"`
		MessageCacheSize       int    `yaml:"message_cache_size"`
		EnableDebugMode        bool   `yaml:"enable_debug_mode"`
		ShowInternalIDs        bool   `yaml:"show_internal_ids"`
		Experimental           struct {
			EnableVoiceCalls bool `yaml:"enable_voice_calls"`
			EnableVideoCalls bool `yaml:"enable_video_calls"`
			EnableGroupChats bool `yaml:"enable_group_chats"`
		} `yaml:"experimental"`
	} `yaml:"advanced"`
}

// NewManager creates a new configuration manager
// Takes config path to support different config files (user vs system defaults)
func NewManager(configPath string) (*Manager, error) {
	mgr := &Manager{
		configPath: configPath,
		config:     &Config{},
	}

	// Load configuration from file
	if err := mgr.Load(); err != nil {
		// If config doesn't exist, create default and save it
		if os.IsNotExist(err) {
			mgr.setDefaults()
			if saveErr := mgr.Save(); saveErr != nil {
				return nil, fmt.Errorf("failed to create default config: %w", saveErr)
			}
		} else {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	return mgr, nil
}

// Load reads configuration from the file
// Uses yaml.v3 which is the standard choice for Go YAML parsing
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, m.config)
}

// Save writes configuration to the file
// Creates parent directories if they don't exist for user convenience
func (m *Manager) Save() error {
	// Ensure config directory exists
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(m.configPath, data, 0644)
}

// GetConfig returns the current configuration
// Returns copy to prevent external modification of internal state
func (m *Manager) GetConfig() Config {
	return *m.config
}

// UpdateConfig updates the configuration and saves it
// Takes full config for simplicity, validates before saving
func (m *Manager) UpdateConfig(config Config) error {
	// Basic validation
	if err := m.validateConfig(&config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	m.config = &config
	return m.Save()
}

// validateConfig performs basic validation on configuration values
// Ensures values are within reasonable ranges to prevent runtime errors
func (m *Manager) validateConfig(config *Config) error {
	// Validate theme values
	validThemes := map[string]bool{
		"system": true, "light": true, "dark": true, "amoled": true, "custom": true,
	}
	if !validThemes[config.UI.Theme] {
		return fmt.Errorf("invalid theme: %s", config.UI.Theme)
	}

	// Validate font size
	validFontSizes := map[string]bool{
		"small": true, "medium": true, "large": true, "extra_large": true,
	}
	if !validFontSizes[config.UI.FontSize] {
		return fmt.Errorf("invalid font size: %s", config.UI.FontSize)
	}

	// Validate file size limits (must be positive)
	if config.Storage.MaxFileSize <= 0 {
		return fmt.Errorf("max file size must be positive")
	}

	if config.Privacy.AutoDownloadLimit <= 0 {
		return fmt.Errorf("auto download limit must be positive")
	}

	return nil
}

// setDefaults sets reasonable default values
// Matches the defaults in config.yaml for consistency
func (m *Manager) setDefaults() {
	// Network defaults
	m.config.Network.EnableIPv6 = true
	m.config.Network.EnableUDP = true
	m.config.Network.EnableTCP = true
	m.config.Network.EnableLocalDiscovery = true
	m.config.Network.EnableHolePunching = true
	m.config.Network.Proxy.Type = "none"

	// Storage defaults
	m.config.Storage.EnableEncryption = true
	m.config.Storage.MaxFileSize = 2147483648 // 2GB
	m.config.Storage.DownloadDir = "Downloads"
	m.config.Storage.MaxMessageHistoryDays = 365
	m.config.Storage.AutoDeleteMediaDays = 30

	// UI defaults
	m.config.UI.Theme = "system"
	m.config.UI.Language = "en"
	m.config.UI.FontSize = "medium"
	m.config.UI.EnableAnimations = true
	m.config.UI.EnableSoundEffects = true
	m.config.UI.Window.RememberSize = true
	m.config.UI.Window.RememberPosition = true
	m.config.UI.Window.MinimizeToTray = true
	m.config.UI.Mobile.VibrateOnMessage = true
	m.config.UI.Mobile.ShowMessagePreview = true

	// Privacy defaults
	m.config.Privacy.SaveMessageHistory = true
	m.config.Privacy.ShowTypingIndicators = true
	m.config.Privacy.SendTypingIndicators = true
	m.config.Privacy.ShowReadReceipts = true
	m.config.Privacy.SendReadReceipts = true
	m.config.Privacy.ShowLastSeen = true
	m.config.Privacy.AutoDownloadLimit = 10485760 // 10MB

	// Notification defaults
	m.config.Notifications.Enabled = true
	m.config.Notifications.Desktop.ShowPreview = true
	m.config.Notifications.Desktop.PlaySound = true
	m.config.Notifications.Desktop.ShowSender = true
	m.config.Notifications.Mobile.ShowOnLockScreen = true
	m.config.Notifications.Mobile.ShowPreview = true
	m.config.Notifications.Mobile.Vibrate = true
	m.config.Notifications.Mobile.LEDColor = "#0066CC"

	// Advanced defaults
	m.config.Advanced.LogLevel = "info"
	m.config.Advanced.MaxLogSize = 10485760 // 10MB
	m.config.Advanced.MaxConcurrentDownloads = 3
	m.config.Advanced.MaxConcurrentUploads = 3
	m.config.Advanced.MessageCacheSize = 1000
}
