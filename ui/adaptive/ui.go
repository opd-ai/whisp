package adaptive

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/opd-ai/whisp/internal/core/contact"
	"github.com/opd-ai/whisp/internal/core/message"
	"github.com/opd-ai/whisp/ui/shared"
)

// UI manages the adaptive user interface
type UI struct {
	app      fyne.App
	coreApp  CoreApp
	platform Platform
	
	mainWindow fyne.Window
	chatView   *shared.ChatView
	contactList *shared.ContactList
}

// CoreApp interface for the core application
type CoreApp interface {
	Start(ctx context.Context) error
	GetToxID() string
	GetContacts() *contact.Manager  
	GetMessages() *message.Manager  
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
	if err := ui.coreApp.Start(ctx); err != nil {
		return fmt.Errorf("failed to start core app: %w", err)
	}

	// Initialize UI components with core app integration
	ui.chatView = shared.NewChatView(ui.coreApp)
	ui.contactList = shared.NewContactList(ui.coreApp)

	// Set up contact selection callback
	ui.contactList.SetOnContactSelect(func(friendID uint32) {
		ui.chatView.SetCurrentFriend(friendID)
	})

	return nil
}

// ShowMainWindow shows the main application window
func (ui *UI) ShowMainWindow() {
	ui.mainWindow = ui.app.NewWindow("Whisp")
	ui.mainWindow.Resize(fyne.NewSize(1000, 700))

	// Create layout based on platform
	if ui.platform.IsMobile() {
		ui.createMobileLayout()
	} else {
		ui.createDesktopLayout()
	}

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

// createMobileLayout creates the mobile layout
func (ui *UI) createMobileLayout() {
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
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Settings", func() {
			// TODO: Show settings dialog
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() {
			ui.app.Quit()
		}),
	)

	// Friends menu
	friendsMenu := fyne.NewMenu("Friends",
		fyne.NewMenuItem("Add Friend", func() {
			// TODO: Show add friend dialog
		}),
		fyne.NewMenuItem("Show My Tox ID", func() {
			toxID := ui.coreApp.GetToxID()
			dialog := widget.NewEntry()
			dialog.SetText(toxID)
			dialog.Disable()
			
			content := container.NewVBox(
				widget.NewLabel("Your Tox ID:"),
				dialog,
			)
			
			popup := widget.NewModalPopUp(content, ui.mainWindow.Canvas())
			popup.Show()
		}),
	)

	// Help menu
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About", func() {
			// TODO: Show about dialog
		}),
	)

	// Create menu bar
	mainMenu := fyne.NewMainMenu(fileMenu, friendsMenu, helpMenu)
	ui.mainWindow.SetMainMenu(mainMenu)

	// Return empty container as Fyne handles menu internally
	return container.NewHBox()
}
