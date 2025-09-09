package shared

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ChatView represents the chat interface
type ChatView struct {
	container *fyne.Container
	messages  *widget.List
	input     *widget.Entry
	sendBtn   *widget.Button
}

// NewChatView creates a new chat view
func NewChatView() *ChatView {
	cv := &ChatView{}
	cv.initializeComponents()
	return cv
}

// initializeComponents initializes the chat view components
func (cv *ChatView) initializeComponents() {
	// Message list
	cv.messages = widget.NewList(
		func() int { return 0 }, // length function
		func() fyne.CanvasObject { // create function
			return widget.NewLabel("Template")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) { // update function
			// Update message item
		},
	)

	// Input field
	cv.input = widget.NewEntry()
	cv.input.SetPlaceHolder("Type a message...")
	cv.input.Wrapping = fyne.TextWrapWord

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

	// TODO: Send message through core app
	cv.input.SetText("")
}

// Container returns the chat view container
func (cv *ChatView) Container() *fyne.Container {
	return cv.container
}

// ContactList represents the contact list interface
type ContactList struct {
	container *fyne.Container
	list      *widget.List
}

// NewContactList creates a new contact list
func NewContactList() *ContactList {
	cl := &ContactList{}
	cl.initializeComponents()
	return cl
}

// initializeComponents initializes the contact list components
func (cl *ContactList) initializeComponents() {
	// Contact list
	cl.list = widget.NewList(
		func() int { return 0 }, // length function
		func() fyne.CanvasObject { // create function
			return widget.NewLabel("Contact")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) { // update function
			// Update contact item
		},
	)

	// Main container
	cl.container = container.NewVBox(
		widget.NewLabel("Contacts"),
		cl.list,
	)
}

// Container returns the contact list container
func (cl *ContactList) Container() *fyne.Container {
	return cl.container
}
