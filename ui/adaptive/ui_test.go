package adaptive

import (
	"context"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/test"

	"github.com/opd-ai/whisp/internal/core/config"
	"github.com/opd-ai/whisp/internal/core/contact"
	"github.com/opd-ai/whisp/internal/core/message"
)

// MockCoreApp implements CoreApp interface for testing
type MockCoreApp struct {
	toxID      string
	configMgr  *config.Manager
	contacts   *contact.Manager
	messages   *message.Manager
	startError error
}

func (m *MockCoreApp) Start(ctx context.Context) error {
	return m.startError
}

func (m *MockCoreApp) GetToxID() string {
	if m.toxID == "" {
		return "7E1A1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF"
	}
	return m.toxID
}

func (m *MockCoreApp) GetContacts() *contact.Manager {
	return m.contacts
}

func (m *MockCoreApp) GetMessages() *message.Manager {
	return m.messages
}

func (m *MockCoreApp) GetConfigManager() *config.Manager {
	return m.configMgr
}

func (m *MockCoreApp) SendMessageFromUI(friendID uint32, content string) error {
	return nil
}

func (m *MockCoreApp) AddContactFromUI(toxID, message string) error {
	return nil
}

// MockPlatform implements Platform interface for testing
type MockPlatform struct {
	isMobile bool
}

func (m *MockPlatform) IsMobile() bool {
	return m.isMobile
}

func (m *MockPlatform) GetNativeControls() any {
	return nil
}

func TestNewUI(t *testing.T) {
	// Create test app
	testApp := app.New()
	
	// Create temporary config for testing
	configMgr, err := config.NewManager("/tmp/test_whisp_config.yaml")
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Create mock dependencies
	mockCore := &MockCoreApp{
		configMgr: configMgr,
	}
	mockPlatform := &MockPlatform{isMobile: false}

	// Test UI creation
	ui, err := NewUI(testApp, mockCore, mockPlatform)
	if err != nil {
		t.Fatalf("NewUI failed: %v", err)
	}

	if ui.app != testApp {
		t.Error("UI app not set correctly")
	}

	if ui.coreApp != mockCore {
		t.Error("UI coreApp not set correctly")
	}

	if ui.platform != mockPlatform {
		t.Error("UI platform not set correctly")
	}
}

func TestUI_Initialize(t *testing.T) {
	// Create test app
	testApp := app.New()
	
	// Create temporary config for testing
	configMgr, err := config.NewManager("/tmp/test_whisp_config_init.yaml")
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Create mock dependencies
	mockCore := &MockCoreApp{
		configMgr: configMgr,
	}
	mockPlatform := &MockPlatform{isMobile: false}

	ui, err := NewUI(testApp, mockCore, mockPlatform)
	if err != nil {
		t.Fatalf("NewUI failed: %v", err)
	}

	// Test initialization
	ctx := context.Background()
	err = ui.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Check that components were created
	if ui.chatView == nil {
		t.Error("ChatView not created")
	}

	if ui.contactList == nil {
		t.Error("ContactList not created")
	}
}

func TestUI_Initialize_StartError(t *testing.T) {
	// Create test app
	testApp := app.New()
	
	// Create temporary config for testing
	configMgr, err := config.NewManager("/tmp/test_whisp_config_error.yaml")
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Create mock with start error
	mockCore := &MockCoreApp{
		configMgr:  configMgr,
		startError: context.Canceled,
	}
	mockPlatform := &MockPlatform{isMobile: false}

	ui, err := NewUI(testApp, mockCore, mockPlatform)
	if err != nil {
		t.Fatalf("NewUI failed: %v", err)
	}

	// Test initialization with error
	ctx := context.Background()
	err = ui.Initialize(ctx)
	if err == nil {
		t.Error("Expected error from Initialize when core app start fails")
	}

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got: %v", err)
	}
}

func TestUI_CreateMainContent_Desktop(t *testing.T) {
	// Create test app
	testApp := app.New()
	
	// Create temporary config for testing
	configMgr, err := config.NewManager("/tmp/test_whisp_config_desktop.yaml")
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Create mock dependencies
	mockCore := &MockCoreApp{
		configMgr: configMgr,
	}
	mockPlatform := &MockPlatform{isMobile: false}

	ui, err := NewUI(testApp, mockCore, mockPlatform)
	if err != nil {
		t.Fatalf("NewUI failed: %v", err)
	}

	// Initialize UI
	ctx := context.Background()
	err = ui.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test desktop layout creation
	content := ui.CreateMainContent()
	if content == nil {
		t.Error("CreateMainContent returned nil")
	}
}

func TestUI_CreateMainContent_Mobile(t *testing.T) {
	// Create test app
	testApp := app.New()
	
	// Create temporary config for testing
	configMgr, err := config.NewManager("/tmp/test_whisp_config_mobile.yaml")
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Create mock dependencies
	mockCore := &MockCoreApp{
		configMgr: configMgr,
	}
	mockPlatform := &MockPlatform{isMobile: true}

	ui, err := NewUI(testApp, mockCore, mockPlatform)
	if err != nil {
		t.Fatalf("NewUI failed: %v", err)
	}

	// Initialize UI
	ctx := context.Background()
	err = ui.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test mobile layout creation
	content := ui.CreateMainContent()
	if content == nil {
		t.Error("CreateMainContent returned nil")
	}
}

func TestUI_LoadWindowState(t *testing.T) {
	// Create test app
	testApp := app.New()
	
	// Create temporary config for testing
	configMgr, err := config.NewManager("/tmp/test_whisp_config_window.yaml")
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Create mock dependencies
	mockCore := &MockCoreApp{
		configMgr: configMgr,
	}
	mockPlatform := &MockPlatform{isMobile: false}

	ui, err := NewUI(testApp, mockCore, mockPlatform)
	if err != nil {
		t.Fatalf("NewUI failed: %v", err)
	}

	// Create a test window
	ui.mainWindow = testApp.NewWindow("Test")

	// Test load window state (should not panic)
	ui.loadWindowState()

	// Check that window has default size
	size := ui.mainWindow.Content().Size()
	if size.Width <= 0 || size.Height <= 0 {
		t.Error("Window size not set properly")
	}
}

func TestUI_SaveWindowState(t *testing.T) {
	// Create test app
	testApp := app.New()
	
	// Create temporary config for testing
	configMgr, err := config.NewManager("/tmp/test_whisp_config_save.yaml")
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Create mock dependencies
	mockCore := &MockCoreApp{
		configMgr: configMgr,
	}
	mockPlatform := &MockPlatform{isMobile: false}

	ui, err := NewUI(testApp, mockCore, mockPlatform)
	if err != nil {
		t.Fatalf("NewUI failed: %v", err)
	}

	// Test save with no window (should not panic)
	ui.saveWindowState()

	// Create a test window
	ui.mainWindow = testApp.NewWindow("Test")

	// Test save window state (should not panic)
	ui.saveWindowState()
}

func TestUI_SetupKeyboardShortcuts(t *testing.T) {
	// Create test app with test driver
	testApp := test.NewApp()
	
	// Create temporary config for testing
	configMgr, err := config.NewManager("/tmp/test_whisp_config_shortcuts.yaml")
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Create mock dependencies
	mockCore := &MockCoreApp{
		configMgr: configMgr,
	}
	mockPlatform := &MockPlatform{isMobile: false}

	ui, err := NewUI(testApp, mockCore, mockPlatform)
	if err != nil {
		t.Fatalf("NewUI failed: %v", err)
	}

	// Test setup without window (should not panic)
	ui.setupKeyboardShortcuts()

	// Create a test window
	ui.mainWindow = testApp.NewWindow("Test")

	// Test setup keyboard shortcuts (should not panic)
	ui.setupKeyboardShortcuts()
}

func TestUI_ShowToxIDDialog(t *testing.T) {
	// Create test app with test driver
	testApp := test.NewApp()
	
	// Create temporary config for testing
	configMgr, err := config.NewManager("/tmp/test_whisp_config_toxid.yaml")
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Create mock dependencies with custom Tox ID
	mockCore := &MockCoreApp{
		configMgr: configMgr,
		toxID:     "TEST1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF",
	}
	mockPlatform := &MockPlatform{isMobile: false}

	ui, err := NewUI(testApp, mockCore, mockPlatform)
	if err != nil {
		t.Fatalf("NewUI failed: %v", err)
	}

	// Create a test window
	ui.mainWindow = testApp.NewWindow("Test")

	// Test show Tox ID dialog (should not panic)
	ui.showToxIDDialog()
}

func TestUI_ShowAboutDialog(t *testing.T) {
	// Create test app with test driver
	testApp := test.NewApp()
	
	// Create temporary config for testing
	configMgr, err := config.NewManager("/tmp/test_whisp_config_about.yaml")
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Create mock dependencies
	mockCore := &MockCoreApp{
		configMgr: configMgr,
	}
	mockPlatform := &MockPlatform{isMobile: false}

	ui, err := NewUI(testApp, mockCore, mockPlatform)
	if err != nil {
		t.Fatalf("NewUI failed: %v", err)
	}

	// Create a test window
	ui.mainWindow = testApp.NewWindow("Test")

	// Test show about dialog (should not panic)
	ui.showAboutDialog()
}
