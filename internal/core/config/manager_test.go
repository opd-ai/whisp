package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "whisp-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")

	// Test creating a new manager with non-existent config
	mgr, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Check that config file was created with defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Verify some default values
	cfg := mgr.GetConfig()
	if cfg.UI.Theme != "system" {
		t.Errorf("Expected default theme 'system', got '%s'", cfg.UI.Theme)
	}
	if cfg.Storage.EnableEncryption != true {
		t.Error("Expected encryption to be enabled by default")
	}
	if cfg.Privacy.SaveMessageHistory != true {
		t.Error("Expected message history to be saved by default")
	}
}

func TestLoadSave(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "whisp-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")

	// Create manager and modify config
	mgr, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	cfg := mgr.GetConfig()
	cfg.UI.Theme = "dark"
	cfg.UI.FontSize = "large"
	cfg.Privacy.SaveMessageHistory = false

	// Save modified config
	err = mgr.UpdateConfig(cfg)
	if err != nil {
		t.Fatalf("UpdateConfig failed: %v", err)
	}

	// Create new manager and verify changes persisted
	mgr2, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("NewManager for reload failed: %v", err)
	}

	cfg2 := mgr2.GetConfig()
	if cfg2.UI.Theme != "dark" {
		t.Errorf("Expected theme 'dark', got '%s'", cfg2.UI.Theme)
	}
	if cfg2.UI.FontSize != "large" {
		t.Errorf("Expected font size 'large', got '%s'", cfg2.UI.FontSize)
	}
	if cfg2.Privacy.SaveMessageHistory != false {
		t.Error("Expected message history to be disabled")
	}
}

func TestValidateConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "whisp-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")
	mgr, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	tests := []struct {
		name      string
		modify    func(*Config)
		expectErr bool
	}{
		{
			name: "valid config",
			modify: func(cfg *Config) {
				cfg.UI.Theme = "light"
				cfg.UI.FontSize = "medium"
			},
			expectErr: false,
		},
		{
			name: "invalid theme",
			modify: func(cfg *Config) {
				cfg.UI.Theme = "invalid"
			},
			expectErr: true,
		},
		{
			name: "invalid font size",
			modify: func(cfg *Config) {
				cfg.UI.FontSize = "invalid"
			},
			expectErr: true,
		},
		{
			name: "invalid file size",
			modify: func(cfg *Config) {
				cfg.Storage.MaxFileSize = -1
			},
			expectErr: true,
		},
		{
			name: "invalid download limit",
			modify: func(cfg *Config) {
				cfg.Privacy.AutoDownloadLimit = 0
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := mgr.GetConfig()
			tt.modify(&cfg)

			err := mgr.UpdateConfig(cfg)
			if (err != nil) != tt.expectErr {
				t.Errorf("UpdateConfig() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "whisp-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")
	mgr, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	cfg := mgr.GetConfig()

	// Test network defaults
	if !cfg.Network.EnableIPv6 {
		t.Error("Expected IPv6 to be enabled by default")
	}
	if !cfg.Network.EnableUDP {
		t.Error("Expected UDP to be enabled by default")
	}
	if cfg.Network.Proxy.Type != "none" {
		t.Errorf("Expected proxy type 'none', got '%s'", cfg.Network.Proxy.Type)
	}

	// Test UI defaults
	if cfg.UI.Theme != "system" {
		t.Errorf("Expected theme 'system', got '%s'", cfg.UI.Theme)
	}
	if cfg.UI.Language != "en" {
		t.Errorf("Expected language 'en', got '%s'", cfg.UI.Language)
	}
	if cfg.UI.FontSize != "medium" {
		t.Errorf("Expected font size 'medium', got '%s'", cfg.UI.FontSize)
	}

	// Test storage defaults
	if cfg.Storage.MaxFileSize != 2147483648 {
		t.Errorf("Expected max file size 2GB, got %d", cfg.Storage.MaxFileSize)
	}
	if !cfg.Storage.EnableEncryption {
		t.Error("Expected encryption to be enabled by default")
	}

	// Test privacy defaults
	if !cfg.Privacy.SaveMessageHistory {
		t.Error("Expected message history to be saved by default")
	}
	if cfg.Privacy.AutoDownloadLimit != 10485760 {
		t.Errorf("Expected auto download limit 10MB, got %d", cfg.Privacy.AutoDownloadLimit)
	}

	// Test notification defaults
	if !cfg.Notifications.Enabled {
		t.Error("Expected notifications to be enabled by default")
	}
	if cfg.Notifications.Mobile.LEDColor != "#0066CC" {
		t.Errorf("Expected LED color '#0066CC', got '%s'", cfg.Notifications.Mobile.LEDColor)
	}

	// Test advanced defaults
	if cfg.Advanced.LogLevel != "info" {
		t.Errorf("Expected log level 'info', got '%s'", cfg.Advanced.LogLevel)
	}
	if cfg.Advanced.MaxConcurrentDownloads != 3 {
		t.Errorf("Expected max downloads 3, got %d", cfg.Advanced.MaxConcurrentDownloads)
	}
}

// TestConfigFileFormat tests that we can read existing config.yaml format
func TestConfigFileFormat(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "whisp-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")

	// Write a sample config that matches the project's config.yaml format
	sampleConfig := `
ui:
  theme: "dark"
  font_size: "large"
  enable_animations: false

privacy:
  save_message_history: false
  auto_download_limit: 5242880

notifications:
  enabled: false
  mobile:
    led_color: "#FF0000"

advanced:
  log_level: "debug"
  max_concurrent_downloads: 5
`

	err = os.WriteFile(configPath, []byte(sampleConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write sample config: %v", err)
	}

	// Load config and verify values
	mgr, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	cfg := mgr.GetConfig()

	if cfg.UI.Theme != "dark" {
		t.Errorf("Expected theme 'dark', got '%s'", cfg.UI.Theme)
	}
	if cfg.UI.FontSize != "large" {
		t.Errorf("Expected font size 'large', got '%s'", cfg.UI.FontSize)
	}
	if cfg.UI.EnableAnimations != false {
		t.Error("Expected animations to be disabled")
	}
	if cfg.Privacy.SaveMessageHistory != false {
		t.Error("Expected message history to be disabled")
	}
	if cfg.Privacy.AutoDownloadLimit != 5242880 {
		t.Errorf("Expected auto download limit 5MB, got %d", cfg.Privacy.AutoDownloadLimit)
	}
	if cfg.Notifications.Enabled != false {
		t.Error("Expected notifications to be disabled")
	}
	if cfg.Notifications.Mobile.LEDColor != "#FF0000" {
		t.Errorf("Expected LED color '#FF0000', got '%s'", cfg.Notifications.Mobile.LEDColor)
	}
	if cfg.Advanced.LogLevel != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", cfg.Advanced.LogLevel)
	}
	if cfg.Advanced.MaxConcurrentDownloads != 5 {
		t.Errorf("Expected max downloads 5, got %d", cfg.Advanced.MaxConcurrentDownloads)
	}
}

// Benchmark configuration operations
func BenchmarkConfigLoad(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "whisp-config-bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")

	// Create initial config
	if _, err := NewManager(configPath); err != nil {
		b.Fatalf("NewManager failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := NewManager(configPath); err != nil {
			b.Fatalf("NewManager failed: %v", err)
		}
	}
}

func BenchmarkConfigSave(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "whisp-config-bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")
	mgr, err := NewManager(configPath)
	if err != nil {
		b.Fatalf("NewManager failed: %v", err)
	}

	cfg := mgr.GetConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := mgr.UpdateConfig(cfg)
		if err != nil {
			b.Fatalf("UpdateConfig failed: %v", err)
		}
	}
}
