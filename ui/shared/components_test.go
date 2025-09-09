package shared

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"github.com/opd-ai/whisp/internal/core/config"
	"github.com/opd-ai/whisp/internal/core/contact"
	"github.com/opd-ai/whisp/internal/core/message"
)

// MockCoreApp implements the CoreApp interface for testing
type MockCoreApp struct {
	messages []*message.Message
	contacts []*contact.Contact
}

func (m *MockCoreApp) SendMessageFromUI(friendID uint32, content string) error {
	return nil
}

func (m *MockCoreApp) AddContactFromUI(toxID, message string) error {
	return nil
}

func (m *MockCoreApp) GetToxID() string {
	return "test-tox-id"
}

func (m *MockCoreApp) GetMessages() *message.Manager {
	return nil // Simple mock
}

func (m *MockCoreApp) GetContacts() *contact.Manager {
	return nil // Simple mock
}

func (m *MockCoreApp) GetConfigManager() *config.Manager {
	return nil // Simple mock
}

// TestChatViewCreation tests that ChatView can be created
func TestChatViewCreation(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockCore := &MockCoreApp{}
	chatView := NewChatView(mockCore)

	if chatView == nil {
		t.Error("ChatView should not be nil")
	}

	if chatView.Container() == nil {
		t.Error("ChatView container should not be nil")
	}
}

// TestContactListCreation tests that ContactList can be created
func TestContactListCreation(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockCore := &MockCoreApp{}
	contactList := NewContactList(mockCore)

	if contactList == nil {
		t.Error("ContactList should not be nil")
	}

	if contactList.Container() == nil {
		t.Error("ContactList container should not be nil")
	}
}

// TestChatViewSetCurrentFriend tests setting current friend
func TestChatViewSetCurrentFriend(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockCore := &MockCoreApp{}
	chatView := NewChatView(mockCore)

	// This should not panic
	chatView.SetCurrentFriend(123)

	if chatView.currentFriend != 123 {
		t.Errorf("Expected current friend to be 123, got %d", chatView.currentFriend)
	}
}
