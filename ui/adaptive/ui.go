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
	"github.com/opd-ai/whisp/ui/theme"
)

// UI manages the adaptive user interface
type UI struct {
	app          fyne.App
	coreApp      CoreApp
	platform     Platform
	themeManager theme.ThemeManager

	mainWindow    fyne.Window
	chatView      *shared.ChatView
	contactList   *shared.ContactList
	mobileTabsRef *container.AppTabs // Reference for mobile navigation
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
	// Initialize theme manager
	configDir := "/home/user/.local/share/whisp" // This should come from config
	themeManager := theme.NewDefaultThemeManager(configDir)
	if err := themeManager.Initialize(app); err != nil {
		return nil, fmt.Errorf("failed to initialize theme manager: %w", err)
	}

	// Apply theme to app
	themeManager.ApplyTheme(app)

	ui := &UI{
		app:          app,
		coreApp:      coreApp,
		platform:     platform,
		themeManager: themeManager,
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

	// Set up contact selection callback with mobile navigation
	ui.contactList.SetOnContactSelect(func(friendID uint32) {
		ui.chatView.SetCurrentFriend(friendID)

		// On mobile, automatically navigate to chat tab when contact is selected
		if ui.platform.IsMobile() {
			ui.NavigateToChat()
		}
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

// createMobileLayout creates the mobile layout with touch-optimized navigation
func (ui *UI) createMobileLayout() fyne.CanvasObject {
	// Create pull-to-refresh container for contact list
	contactsWithRefresh := ui.createPullToRefreshContacts()

	// Create mobile-optimized tabs with larger touch targets
	tabs := container.NewAppTabs(
		container.NewTabItem("Contacts", contactsWithRefresh),
		container.NewTabItem("Chat", ui.chatView.Container()),
		container.NewTabItem("Settings", ui.createMobileSettingsView()),
	)

	// Configure tab bar for mobile
	tabs.SetTabLocation(container.TabLocationBottom)

	// Store reference for navigation
	ui.mobileTabsRef = tabs

	// Add gesture support for swipe navigation
	ui.setupMobileGestures(tabs)

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

// createPullToRefreshContacts creates a pull-to-refresh container for contacts
func (ui *UI) createPullToRefreshContacts() fyne.CanvasObject {
	// Create refresh button for mobile
	refreshBtn := widget.NewButton("Pull to Refresh", func() {
		ui.contactList.RefreshContacts()
	})
	refreshBtn.Importance = widget.LowImportance

	// Create container with refresh button at top
	return container.NewVBox(
		refreshBtn,
		ui.contactList.Container(),
	)
}

// createMobileSettingsView creates a mobile-optimized settings view
func (ui *UI) createMobileSettingsView() fyne.CanvasObject {
	// Create mobile settings with larger touch targets
	toxIDBtn := widget.NewButton("Show Tox ID", func() {
		toxID := ui.coreApp.GetToxID()
		dialog.ShowInformation("Your Tox ID", toxID, ui.mainWindow)
	})

	settingsBtn := widget.NewButton("Application Settings", func() {
		configMgr := ui.coreApp.GetConfigManager()
		settingsDialog := shared.NewSettingsDialog(configMgr, ui.mainWindow)
		settingsDialog.Show()
	})

	aboutBtn := widget.NewButton("About Whisp", func() {
		ui.showAboutDialog()
	})

	// Create larger buttons for mobile
	toxIDBtn.Resize(fyne.NewSize(300, 60))
	settingsBtn.Resize(fyne.NewSize(300, 60))
	aboutBtn.Resize(fyne.NewSize(300, 60))

	return container.NewVBox(
		widget.NewCard("", "Quick Actions", container.NewVBox(
			toxIDBtn,
			settingsBtn,
			aboutBtn,
		)),
	)
}

// setupMobileGestures sets up touch gestures for mobile navigation
func (ui *UI) setupMobileGestures(tabs *container.AppTabs) {
	// Note: Fyne doesn't have built-in swipe gesture support yet
	// This is a placeholder for future gesture implementation
	// When Fyne adds gesture support, implement swipe between tabs here

	// For now, we ensure tabs are touch-friendly with bottom placement
	// and adequate spacing
}

// NavigateToChat programmatically switches to chat tab (for mobile navigation)
func (ui *UI) NavigateToChat() {
	if ui.mobileTabsRef != nil && ui.platform.IsMobile() {
		ui.mobileTabsRef.SelectTab(ui.mobileTabsRef.Items[1]) // Chat tab
	}
}

// NavigateToContacts programmatically switches to contacts tab
func (ui *UI) NavigateToContacts() {
	if ui.mobileTabsRef != nil && ui.platform.IsMobile() {
		ui.mobileTabsRef.SelectTab(ui.mobileTabsRef.Items[0]) // Contacts tab
	}
}

// configureMobileWindow configures window settings optimized for mobile
func (ui *UI) configureMobileWindow() {
	if ui.mainWindow == nil {
		return
	}

	// Set mobile-optimized window properties
	ui.mainWindow.SetFixedSize(false) // Allow resizing for different screen sizes

	// Set minimum size suitable for mobile screens
	ui.mainWindow.Resize(fyne.NewSize(360, 640)) // Common mobile screen ratio

	// Center the window (useful for mobile simulators)
	ui.mainWindow.CenterOnScreen()
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
	// Create mobile layout with enhanced features
	mobileContent := ui.createMobileLayout()
	ui.mainWindow.SetContent(mobileContent)

	// Configure window for mobile
	ui.configureMobileWindow()
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

	// Create menu bar and set it on the main window if available
	mainMenu := fyne.NewMainMenu(fileMenu, friendsMenu, helpMenu)
	if ui.mainWindow != nil {
		ui.mainWindow.SetMainMenu(mainMenu)
	}

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
	if ui.mainWindow == nil {
		return // Cannot show dialog without main window
	}

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
	if ui.mainWindow == nil {
		return // Cannot show dialog without main window
	}

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

// GetThemeManager returns the theme manager
func (ui *UI) GetThemeManager() theme.ThemeManager {
	return ui.themeManager
}

// ShowThemeDialog displays the theme selection dialog
func (ui *UI) ShowThemeDialog() {
	if ui.mainWindow == nil {
		return // Cannot show dialog without main window
	}

	// Current theme info
	currentTheme := ui.themeManager.GetThemeType()
	themeLabel := widget.NewLabel(fmt.Sprintf("Current theme: %v", currentTheme))

	// Theme selection buttons
	lightBtn := widget.NewButton("Light Theme", func() {
		ui.themeManager.SetTheme(theme.ThemeLight)
		themeLabel.SetText(fmt.Sprintf("Current theme: %v", theme.ThemeLight))
	})

	darkBtn := widget.NewButton("Dark Theme", func() {
		ui.themeManager.SetTheme(theme.ThemeDark)
		themeLabel.SetText(fmt.Sprintf("Current theme: %v", theme.ThemeDark))
	})

	systemBtn := widget.NewButton("System Theme", func() {
		ui.themeManager.SetTheme(theme.ThemeSystem)
		themeLabel.SetText(fmt.Sprintf("Current theme: %v", theme.ThemeSystem))
	})

	// Custom themes section
	customThemes := ui.themeManager.ListCustomThemes()
	var customButtons []fyne.CanvasObject

	if len(customThemes) > 0 {
		customButtons = append(customButtons, widget.NewSeparator())
		customButtons = append(customButtons, widget.NewLabel("Custom Themes:"))

		for _, ct := range customThemes {
			themeName := ct.Name // Capture for closure
			btn := widget.NewButton(themeName, func() {
				// Set custom theme by updating preferences
				prefs := ui.themeManager.GetPreferences()
				prefs.CustomThemeName = themeName
				prefs.ThemeType = theme.ThemeCustom
				ui.themeManager.SetPreferences(prefs)
				themeLabel.SetText(fmt.Sprintf("Current theme: %s (custom)", themeName))
			})
			customButtons = append(customButtons, btn)
		}
	}

	// Preferences
	prefs := ui.themeManager.GetPreferences()
	followSystemCheck := widget.NewCheck("Follow system theme", func(checked bool) {
		ui.themeManager.EnableSystemThemeFollowing(checked)
	})
	followSystemCheck.SetChecked(prefs.FollowSystemTheme)

	// Build content
	content := container.NewVBox(
		themeLabel,
		widget.NewSeparator(),
		widget.NewLabel("Theme Selection:"),
		container.NewHBox(lightBtn, darkBtn, systemBtn),
	)

	// Add custom theme buttons if any
	for _, btn := range customButtons {
		content.Add(btn)
	}

	content.Add(widget.NewSeparator())
	content.Add(followSystemCheck)

	dialog.ShowCustom("Theme Settings", "Close", content, ui.mainWindow)
}
