package shared

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/opd-ai/whisp/internal/core/config"
	"github.com/opd-ai/whisp/internal/core/contact"
	"github.com/opd-ai/whisp/internal/core/media"
	"github.com/opd-ai/whisp/internal/core/message"
)

// CoreApp interface for core application access
type CoreApp interface {
	SendMessageFromUI(friendID uint32, content string) error
	AddContactFromUI(toxID, message string) error
	GetToxID() string
	GetMessages() *message.Manager
	GetContacts() *contact.Manager
	GetConfigManager() *config.Manager
	
	// Media-related methods
	GetMediaInfoFromUI(filePath string) (*media.MediaInfo, error)
	GenerateThumbnailFromUI(filePath string, maxWidth, maxHeight int) (string, error)
	IsMediaFileFromUI(filePath string) bool
	GetThumbnailPathFromUI(filePath string, maxWidth, maxHeight int) (string, bool)
}

// ChatView represents the chat interface
type ChatView struct {
	container     *fyne.Container
	messages      *widget.List
	input         *widget.Entry
	sendBtn       *widget.Button
	coreApp       CoreApp
	currentFriend uint32
	messageData   []*message.Message
}

// NewChatView creates a new chat view
func NewChatView(coreApp CoreApp) *ChatView {
	cv := &ChatView{
		coreApp: coreApp,
	}
	cv.initializeComponents()
	return cv
}

// initializeComponents initializes the chat view components
func (cv *ChatView) initializeComponents() {
	// Message list
	cv.messages = widget.NewList(
		func() int { return len(cv.messageData) },
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < len(cv.messageData) {
				msg := cv.messageData[i]
				label := o.(*widget.Label)
				if msg.IsOutgoing {
					label.SetText(fmt.Sprintf("You: %s", msg.Content))
				} else {
					label.SetText(fmt.Sprintf("Friend: %s", msg.Content))
				}
			}
		},
	)

	// Input field
	cv.input = widget.NewEntry()
	cv.input.SetPlaceHolder("Type a message...")
	cv.input.Wrapping = fyne.TextWrapWord
	cv.input.OnSubmitted = func(text string) {
		cv.sendMessage()
	}

	// Send button
	cv.sendBtn = widget.NewButton("Send", func() {
		cv.sendMessage()
	})

	// Input container
	inputContainer := container.NewBorder(
		nil, nil, nil, cv.sendBtn,
		cv.input,
	)

	// Main container
	cv.container = container.NewBorder(
		nil, inputContainer, nil, nil,
		cv.messages,
	)
}

// sendMessage handles sending a message
func (cv *ChatView) sendMessage() {
	text := cv.input.Text
	if text == "" {
		return
	}

	if cv.coreApp != nil && cv.currentFriend != 0 {
		if err := cv.coreApp.SendMessageFromUI(cv.currentFriend, text); err != nil {
			log.Printf("Failed to send message: %v", err)
			return
		}

		// Reload messages from database to get the actual sent message
		if cv.coreApp.GetMessages() != nil {
			messages, err := cv.coreApp.GetMessages().GetMessages(cv.currentFriend, 50, 0)
			if err != nil {
				log.Printf("Failed to reload messages: %v", err)
			} else {
				cv.messageData = messages
				cv.messages.Refresh()
			}
		}
	}

	cv.input.SetText("")
}

// SetCurrentFriend sets the current friend for chat
func (cv *ChatView) SetCurrentFriend(friendID uint32) {
	cv.currentFriend = friendID

	// Load message history for this friend
	if cv.coreApp != nil && cv.coreApp.GetMessages() != nil {
		messages, err := cv.coreApp.GetMessages().GetMessages(friendID, 50, 0) // Load last 50 messages
		if err != nil {
			log.Printf("Failed to load message history: %v", err)
			cv.messageData = []*message.Message{} // Clear on error
		} else {
			cv.messageData = messages
		}
	} else {
		cv.messageData = []*message.Message{} // Clear if no core app
	}

	cv.messages.Refresh()
}

// Container returns the chat view container
func (cv *ChatView) Container() *fyne.Container {
	return cv.container
}

// ContactList represents the contact list interface
type ContactList struct {
	container    *fyne.Container
	list         *widget.List
	coreApp      CoreApp
	contactData  []*contact.Contact
	onSelect     func(uint32) // Callback when contact is selected
	parentWindow fyne.Window  // Reference to parent window for dialogs
}

// NewContactList creates a new contact list
func NewContactList(coreApp CoreApp) *ContactList {
	cl := &ContactList{
		coreApp: coreApp,
	}
	cl.initializeComponents()
	return cl
}

// SetParentWindow sets the parent window for dialogs
func (cl *ContactList) SetParentWindow(window fyne.Window) {
	cl.parentWindow = window
}

// initializeComponents initializes the contact list components
func (cl *ContactList) initializeComponents() {
	// Contact list
	cl.list = widget.NewList(
		func() int { return len(cl.contactData) },
		func() fyne.CanvasObject {
			return widget.NewButton("Contact", nil)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < len(cl.contactData) {
				contact := cl.contactData[i]
				button := o.(*widget.Button)
				displayName := contact.Name
				if displayName == "" || displayName == "Unknown" {
					displayName = fmt.Sprintf("Friend %d", contact.FriendID)
				}
				button.SetText(displayName)
				button.OnTapped = func() {
					if cl.onSelect != nil {
						cl.onSelect(contact.FriendID)
					}
				}
			}
		},
	)

	// Add friend button
	addFriendBtn := widget.NewButton("Add Friend", func() {
		cl.showAddFriendDialog()
	})

	// Main container
	cl.container = container.NewVBox(
		widget.NewLabel("Contacts"),
		addFriendBtn,
		cl.list,
	)
}

// showAddFriendDialog shows the add friend dialog
func (cl *ContactList) showAddFriendDialog() {
	if cl.parentWindow == nil {
		log.Println("No parent window available for add friend dialog")
		return
	}

	// Create input fields
	toxIDEntry := widget.NewEntry()
	toxIDEntry.SetPlaceHolder("Enter Tox ID...")
	toxIDEntry.Wrapping = fyne.TextWrapWord

	messageEntry := widget.NewEntry()
	messageEntry.SetText("Hello! I'd like to add you as a friend.")
	messageEntry.SetPlaceHolder("Friend request message...")
	messageEntry.Wrapping = fyne.TextWrapWord

	// Create buttons
	var dialog *widget.PopUp

	addButton := widget.NewButton("Add Friend", func() {
		toxID := toxIDEntry.Text
		message := messageEntry.Text

		if toxID == "" {
			// Show error - invalid Tox ID
			cl.showErrorDialog("Please enter a valid Tox ID")
			return
		}

		// Try to add the contact
		if cl.coreApp != nil {
			if err := cl.coreApp.AddContactFromUI(toxID, message); err != nil {
				log.Printf("Failed to add contact: %v", err)
				// Show error dialog
				cl.showErrorDialog(fmt.Sprintf("Failed to add contact: %v", err))
			} else {
				log.Println("Friend request sent successfully")
				// Refresh contact list
				cl.RefreshContacts()
				dialog.Hide()
			}
		}
	})

	cancelButton := widget.NewButton("Cancel", func() {
		dialog.Hide()
	})

	// Create dialog content
	content := container.NewVBox(
		widget.NewLabel("Add Friend"),
		widget.NewSeparator(),
		widget.NewLabel("Tox ID:"),
		toxIDEntry,
		widget.NewLabel("Message:"),
		messageEntry,
		widget.NewSeparator(),
		container.NewHBox(
			cancelButton,
			addButton,
		),
	)

	// Create and show dialog
	dialog = widget.NewModalPopUp(content, cl.parentWindow.Canvas())
	dialog.Resize(fyne.NewSize(400, 300))
	dialog.Show()
}

// RefreshContacts refreshes the contact list
func (cl *ContactList) RefreshContacts() {
	if cl.coreApp != nil && cl.coreApp.GetContacts() != nil {
		contacts := cl.coreApp.GetContacts().GetAllContacts()
		cl.contactData = contacts
	} else {
		cl.contactData = []*contact.Contact{} // Clear if no core app
	}
	cl.list.Refresh()
}

// SetOnContactSelect sets the callback for contact selection
func (cl *ContactList) SetOnContactSelect(callback func(uint32)) {
	cl.onSelect = callback
}

// ShowAddFriendDialog shows the add friend dialog (public method)
func (cl *ContactList) ShowAddFriendDialog() {
	cl.showAddFriendDialog()
}

// Container returns the contact list container
func (cl *ContactList) Container() *fyne.Container {
	return cl.container
}

// showErrorDialog shows an error dialog
func (cl *ContactList) showErrorDialog(message string) {
	if cl.parentWindow == nil {
		log.Printf("Error: %s", message)
		return
	}

	errorLabel := widget.NewLabel(message)
	var errorPopup *widget.PopUp
	errorPopup = widget.NewModalPopUp(
		container.NewVBox(
			errorLabel,
			widget.NewButton("OK", func() {
				errorPopup.Hide()
			}),
		),
		cl.parentWindow.Canvas(),
	)
	errorPopup.Show()
}

// MediaPreview represents a media preview widget for displaying image/video thumbnails
type MediaPreview struct {
	container    *fyne.Container
	image        *widget.Card
	videoIcon    *widget.Card
	mediaInfo    *media.MediaInfo
	thumbnailPath string
	coreApp      CoreApp
}

// NewMediaPreview creates a new media preview for a file
func NewMediaPreview(coreApp CoreApp, filePath string, maxWidth, maxHeight int) *MediaPreview {
	mp := &MediaPreview{
		coreApp: coreApp,
	}

	mp.initializePreview(filePath, maxWidth, maxHeight)
	return mp
}

// initializePreview sets up the media preview based on file type
func (mp *MediaPreview) initializePreview(filePath string, maxWidth, maxHeight int) {
	// Check if file is a media file
	if !mp.coreApp.IsMediaFileFromUI(filePath) {
		mp.createNonMediaPreview(filePath)
		return
	}

	// Get media info
	mediaInfo, err := mp.coreApp.GetMediaInfoFromUI(filePath)
	if err != nil {
		log.Printf("Failed to get media info for %s: %v", filePath, err)
		mp.createErrorPreview(filePath, err)
		return
	}

	mp.mediaInfo = mediaInfo

	// Generate or get cached thumbnail
	thumbnailPath, hasCached := mp.coreApp.GetThumbnailPathFromUI(filePath, maxWidth, maxHeight)
	if !hasCached {
		// Generate thumbnail
		var err error
		thumbnailPath, err = mp.coreApp.GenerateThumbnailFromUI(filePath, maxWidth, maxHeight)
		if err != nil {
			log.Printf("Failed to generate thumbnail for %s: %v", filePath, err)
			mp.createErrorPreview(filePath, err)
			return
		}
	}

	mp.thumbnailPath = thumbnailPath

	// Create preview based on media type
	switch mediaInfo.Type {
	case media.MediaTypeImage:
		mp.createImagePreview(filePath)
	case media.MediaTypeVideo:
		mp.createVideoPreview(filePath)
	default:
		mp.createGenericMediaPreview(filePath)
	}
}

// createImagePreview creates a preview for image files
func (mp *MediaPreview) createImagePreview(filePath string) {
	// Create image card with thumbnail
	title := fmt.Sprintf("Image (%dx%d)", mp.mediaInfo.Width, mp.mediaInfo.Height)
	subtitle := mp.formatFileSize(mp.mediaInfo.Size)

	mp.image = widget.NewCard(title, subtitle, nil)
	
	// TODO: In a full implementation, load the actual thumbnail image
	// For now, show a placeholder with image info
	content := widget.NewLabel(fmt.Sprintf("📷 %s", title))
	content.Alignment = fyne.TextAlignCenter
	
	mp.image.SetContent(content)
	mp.container = container.NewVBox(mp.image)
}

// createVideoPreview creates a preview for video files
func (mp *MediaPreview) createVideoPreview(filePath string) {
	// Create video card with thumbnail
	title := "Video"
	if mp.mediaInfo.Duration > 0 {
		title = fmt.Sprintf("Video (%s)", mp.formatDuration(mp.mediaInfo.Duration))
	}
	subtitle := mp.formatFileSize(mp.mediaInfo.Size)

	mp.videoIcon = widget.NewCard(title, subtitle, nil)
	
	// Show video icon placeholder
	content := widget.NewLabel("🎬 " + title)
	content.Alignment = fyne.TextAlignCenter
	
	mp.videoIcon.SetContent(content)
	mp.container = container.NewVBox(mp.videoIcon)
}

// createGenericMediaPreview creates a preview for other media types
func (mp *MediaPreview) createGenericMediaPreview(filePath string) {
	title := fmt.Sprintf("%s File", mp.mediaInfo.Type.String())
	subtitle := mp.formatFileSize(mp.mediaInfo.Size)

	card := widget.NewCard(title, subtitle, nil)
	
	// Show generic media icon
	var icon string
	switch mp.mediaInfo.Type {
	case media.MediaTypeAudio:
		icon = "🎵"
	default:
		icon = "📄"
	}
	
	content := widget.NewLabel(icon + " " + title)
	content.Alignment = fyne.TextAlignCenter
	
	card.SetContent(content)
	mp.container = container.NewVBox(card)
}

// createNonMediaPreview creates a preview for non-media files
func (mp *MediaPreview) createNonMediaPreview(filePath string) {
	title := "File"
	subtitle := "Non-media file"

	card := widget.NewCard(title, subtitle, nil)
	
	content := widget.NewLabel("📎 " + title)
	content.Alignment = fyne.TextAlignCenter
	
	card.SetContent(content)
	mp.container = container.NewVBox(card)
}

// createErrorPreview creates a preview when there's an error
func (mp *MediaPreview) createErrorPreview(filePath string, err error) {
	title := "Error"
	subtitle := "Failed to load preview"

	card := widget.NewCard(title, subtitle, nil)
	
	content := widget.NewLabel("❌ " + title)
	content.Alignment = fyne.TextAlignCenter
	
	card.SetContent(content)
	mp.container = container.NewVBox(card)
}

// formatFileSize formats file size in human-readable format
func (mp *MediaPreview) formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// formatDuration formats duration in human-readable format
func (mp *MediaPreview) formatDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	remainingSeconds := seconds % 60
	if minutes < 60 {
		return fmt.Sprintf("%dm %ds", minutes, remainingSeconds)
	}
	hours := minutes / 60
	remainingMinutes := minutes % 60
	return fmt.Sprintf("%dh %dm", hours, remainingMinutes)
}

// Container returns the media preview container
func (mp *MediaPreview) Container() *fyne.Container {
	return mp.container
}

// GetMediaInfo returns the media information
func (mp *MediaPreview) GetMediaInfo() *media.MediaInfo {
	return mp.mediaInfo
}

// GetThumbnailPath returns the thumbnail file path
func (mp *MediaPreview) GetThumbnailPath() string {
	return mp.thumbnailPath
}
