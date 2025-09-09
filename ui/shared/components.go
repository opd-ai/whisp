package shared

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/opd-ai/whisp/internal/core/contact"
	"github.com/opd-ai/whisp/internal/core/message"
)

// CoreApp interface for core application access
type CoreApp interface {
	SendMessageFromUI(friendID uint32, content string) error
	AddContactFromUI(toxID, message string) error
	GetToxID() string
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
		
		// Add message to local display immediately
		newMsg := &message.Message{
			Content:    text,
			IsOutgoing: true,
		}
		cv.messageData = append(cv.messageData, newMsg)
		cv.messages.Refresh()
	}

	cv.input.SetText("")
}

// SetCurrentFriend sets the current friend for chat
func (cv *ChatView) SetCurrentFriend(friendID uint32) {
	cv.currentFriend = friendID
	// TODO: Load message history for this friend
	cv.messageData = []*message.Message{} // Clear for now
	cv.messages.Refresh()
}

// Container returns the chat view container
func (cv *ChatView) Container() *fyne.Container {
	return cv.container
}

// ContactList represents the contact list interface
type ContactList struct {
	container   *fyne.Container
	list        *widget.List
	coreApp     CoreApp
	contactData []*contact.Contact
	onSelect    func(uint32) // Callback when contact is selected
}

// NewContactList creates a new contact list
func NewContactList(coreApp CoreApp) *ContactList {
	cl := &ContactList{
		coreApp: coreApp,
	}
	cl.initializeComponents()
	return cl
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
	// For now, just log - proper dialog implementation would need window reference
	log.Println("Add friend dialog requested - placeholder implementation")
	
	// TODO: Implement proper dialog when window reference is available
	// This is a simplified version for the initial implementation
}

// RefreshContacts refreshes the contact list
func (cl *ContactList) RefreshContacts() {
	// TODO: Get actual contacts from core app
	cl.contactData = []*contact.Contact{} // Placeholder
	cl.list.Refresh()
}

// SetOnContactSelect sets the callback for contact selection
func (cl *ContactList) SetOnContactSelect(callback func(uint32)) {
	cl.onSelect = callback
}

// Container returns the contact list container
func (cl *ContactList) Container() *fyne.Container {
	return cl.container
}
