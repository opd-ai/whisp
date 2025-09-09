package adaptive

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"github.com/opd-ai/whisp/internal/core/config"
	"github.com/opd-ai/whisp/internal/core/contact"
	"github.com/opd-ai/whisp/internal/core/message"
	"github.com/opd-ai/whisp/ui/shared"
)

// UI manages the adaptive user interface
type UI struct {
	app      fyne.App
	coreApp  CoreApp
	platform Platform

	mainWindow  fyne.Window
	chatView    *shared.ChatView
	contactList *shared.ContactList
}

// CoreApp interface for the core application
type CoreApp interface {
	Start(ctx context.Context) error
	GetToxID() string
	GetContacts() *contact.Manager
	GetMessages() *message.Manager
	GetConfigManager() *config.Manager
	SendMessageFromUI(friendID uint32, content string) error
	AddContactFromUI(toxID, message string) error
}

// NewUI creates a new adaptive UI
func NewUI(app fyne.App, coreApp CoreApp, platform Platform) (*UI, error) {
	ui := &UI{
		app:      app,
		coreApp:  coreApp,
		platform: platform,
	}

	return ui, nil
}

// Initialize initializes the UI
func (ui *UI) Initialize(ctx context.Context) error {
	// Start core application
	if err := ui.coreApp.Start(ctx); err != nil {
		return fmt.Errorf("failed to start core app: %w", err)
	}

	// Create UI components
	ui.chatView = shared.NewChatView(ui.coreApp)
	ui.contactList = shared.NewContactList(ui.coreApp)

	// Set up contact selection callback
	ui.contactList.SetOnContactSelect(func(friendID uint32) {
		ui.chatView.SetCurrentFriend(friendID)
	})

	return nil
}

// CreateMainContent creates the main content for the window
func (ui *UI) CreateMainContent() fyne.CanvasObject {
	// Create menu bar
	ui.createMenuBar()

	if ui.platform.IsMobile() {
		return ui.createMobileLayout()
	} else {
		return ui.createDesktopLayout()
	}
}

// createMobileLayout creates the mobile layout
func (ui *UI) createMobileLayout() fyne.CanvasObject {
	// Mobile layout with tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Chats", ui.contactList.Container()),
		container.NewTabItem("Messages", ui.chatView.Container()),
	)

	return tabs
}

// ShowMainWindow shows the main application window
func (ui *UI) ShowMainWindow() {
	ui.mainWindow = ui.app.NewWindow("Whisp")
	
	// Load window state from configuration
	ui.loadWindowState()

	// Set parent window for contact list dialogs
	if ui.contactList != nil {
		ui.contactList.SetParentWindow(ui.mainWindow)
		// Initial refresh of contacts
		ui.contactList.RefreshContacts()
	}

	// Setup keyboard shortcuts for desktop platforms
	if !ui.platform.IsMobile() {
		ui.setupKeyboardShortcuts()
	}

	// Create layout based on platform
	if ui.platform.IsMobile() {
		ui.setupMobileLayout()
	} else {
		ui.setupDesktopLayout()
	}

	// Setup window close handler to save state
	ui.mainWindow.SetCloseIntercept(func() {
		ui.saveWindowState()
		ui.app.Quit()
	})

	ui.mainWindow.ShowAndRun()
}

// createDesktopLayout creates the desktop layout
func (ui *UI) createDesktopLayout() fyne.CanvasObject {
	// Create main content
	content := container.NewHSplit(
		ui.contactList.Container(),
		ui.chatView.Container(),
	)
	content.SetOffset(0.3) // 30% for contacts, 70% for chat

	// Create menu bar
	menuBar := ui.createMenuBar()

	// Return the content layout
	return container.NewBorder(
		menuBar, // top
		nil,     // bottom
		nil,     // left
		nil,     // right
		content, // center
	)
}

// setupDesktopLayout sets up the desktop layout and assigns it to the window
func (ui *UI) setupDesktopLayout() {
	content := ui.createDesktopLayout()
	ui.mainWindow.SetContent(content)
}

// setupMobileLayout sets up the mobile layout and assigns it to the window
func (ui *UI) setupMobileLayout() {
	// Mobile layout with tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Chats", ui.contactList.Container()),
		container.NewTabItem("Messages", ui.chatView.Container()),
	)

	ui.mainWindow.SetContent(tabs)
}

// createMenuBar creates the application menu bar
func (ui *UI) createMenuBar() *fyne.Container {
	// File menu
	settingsItem := fyne.NewMenuItem("Settings", func() {
		configMgr := ui.coreApp.GetConfigManager()
		settingsDialog := shared.NewSettingsDialog(configMgr, ui.mainWindow)
		settingsDialog.Show()
	})

	quitItem := fyne.NewMenuItem("Quit", func() {
		ui.saveWindowState()
		ui.app.Quit()
	})

	fileMenu := fyne.NewMenu("File",
		settingsItem,
		fyne.NewMenuItemSeparator(),
		quitItem,
	)

	// Friends menu
	addFriendItem := fyne.NewMenuItem("Add Friend", func() {
		if ui.contactList != nil {
			ui.contactList.ShowAddFriendDialog()
		}
	})

	showToxIDItem := fyne.NewMenuItem("Show My Tox ID", func() {
		ui.showToxIDDialog()
	})

	friendsMenu := fyne.NewMenu("Friends",
		addFriendItem,
		showToxIDItem,
	)

	// Help menu
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About", func() {
			ui.showAboutDialog()
		}),
	)

	// Set accelerator keys for desktop
	if !ui.platform.IsMobile() {
		settingsItem.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyComma, Modifier: fyne.KeyModifierControl}
		quitItem.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyQ, Modifier: fyne.KeyModifierControl}
		addFriendItem.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyN, Modifier: fyne.KeyModifierControl}
	}

	// Create menu bar
	mainMenu := fyne.NewMainMenu(fileMenu, friendsMenu, helpMenu)
	ui.mainWindow.SetMainMenu(mainMenu)

	// Return empty container as Fyne handles menu internally
	return container.NewHBox()
}

// setupKeyboardShortcuts configures keyboard shortcuts for desktop platforms
func (ui *UI) setupKeyboardShortcuts() {
	if ui.mainWindow == nil {
		return
	}

	// Set up canvas shortcuts (these work on all windows)
	canvas := ui.mainWindow.Canvas()

	// Ctrl+Q: Quit application
	canvas.AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyQ,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		ui.saveWindowState()
		ui.app.Quit()
	})

	// Ctrl+N: Add new friend
	canvas.AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyN,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		if ui.contactList != nil {
			ui.contactList.ShowAddFriendDialog()
		}
	})

	// Ctrl+Comma: Open settings
	canvas.AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyComma,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		configMgr := ui.coreApp.GetConfigManager()
		settingsDialog := shared.NewSettingsDialog(configMgr, ui.mainWindow)
		settingsDialog.Show()
	})

	// Escape: Close current dialog (handled by Fyne automatically)
}

// loadWindowState loads window size and position from configuration
func (ui *UI) loadWindowState() {
	configMgr := ui.coreApp.GetConfigManager()
	config := configMgr.GetConfig()

	// Set default size
	defaultSize := fyne.NewSize(1000, 700)
	ui.mainWindow.Resize(defaultSize)

	// TODO: Load saved size and position when configuration structure is extended
	// For now, use defaults based on config flags
	if config.UI.Window.RememberSize {
		// Use saved size when available (future enhancement)
		ui.mainWindow.Resize(defaultSize)
	}

	if config.UI.Window.RememberPosition {
		// Center window for now (future enhancement will use saved position)
		ui.mainWindow.CenterOnScreen()
	} else {
		ui.mainWindow.CenterOnScreen()
	}

	// Handle minimize to tray setting
	if config.UI.Window.MinimizeToTray {
		// TODO: Implement system tray functionality
		// For now, just set the behavior flag
	}
}

// saveWindowState saves current window size and position to configuration
func (ui *UI) saveWindowState() {
	if ui.mainWindow == nil {
		return
	}

	configMgr := ui.coreApp.GetConfigManager()
	config := configMgr.GetConfig()

	// Only save if remember settings are enabled
	if config.UI.Window.RememberSize || config.UI.Window.RememberPosition {
		// TODO: Save actual window size and position when configuration structure is extended
		// For now, just ensure the flags are preserved
		if err := configMgr.Save(); err != nil {
			fmt.Printf("Warning: Failed to save window state: %v\n", err)
		}
	}
}

// showToxIDDialog displays a dialog with the user's Tox ID
func (ui *UI) showToxIDDialog() {
	toxID := ui.coreApp.GetToxID()
	
	entry := widget.NewEntry()
	entry.SetText(toxID)
	entry.Disable()

	copyButton := widget.NewButton("Copy to Clipboard", func() {
		ui.mainWindow.Clipboard().SetContent(toxID)
		// Show brief confirmation
		dialog.ShowInformation("Copied", "Tox ID copied to clipboard", ui.mainWindow)
	})

	content := container.NewVBox(
		widget.NewLabel("Your Tox ID:"),
		entry,
		copyButton,
	)

	dialog.ShowCustom("My Tox ID", "Close", content, ui.mainWindow)
}

// showAboutDialog displays the about dialog
func (ui *UI) showAboutDialog() {
	content := container.NewVBox(
		widget.NewLabel("Whisp"),
		widget.NewLabel("Secure Cross-Platform Messaging"),
		widget.NewLabel(""),
		widget.NewLabel("Built with Go and Fyne"),
		widget.NewLabel("Uses Tox protocol for P2P messaging"),
		widget.NewLabel(""),
		widget.NewLabel("Version: 1.0.0-dev"),
	)

	dialog.ShowCustom("About Whisp", "Close", content, ui.mainWindow)
}
