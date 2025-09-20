// Package calls implements the call management system for ToxAV integration.
// This file contains the main CallManager that coordinates with ToxAV for P2P calling.
package calls

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/opd-ai/toxcore"
)

// Config holds configuration for the call manager
type Config struct {
	// Audio configuration
	AudioBitRate    uint32 // Audio bitrate in kbps (32-128)
	AudioSampleRate uint32 // Audio sample rate (8000, 16000, 24000, 48000)
	AudioChannels   uint8  // Number of audio channels (1 or 2)

	// Video configuration
	VideoBitRate uint32 // Video bitrate in kbps (100-2000)
	VideoWidth   uint16 // Video frame width
	VideoHeight  uint16 // Video frame height
	VideoFPS     uint8  // Video frames per second

	// Network configuration
	IterationInterval time.Duration // ToxAV iteration interval
	CallTimeout       time.Duration // Timeout for outgoing calls
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		AudioBitRate:      64,    // 64 kbps audio
		AudioSampleRate:   48000, // 48 kHz
		AudioChannels:     1,     // Mono audio
		VideoBitRate:      500,   // 500 kbps video
		VideoWidth:        640,   // 640x480 VGA
		VideoHeight:       480,
		VideoFPS:          30,                    // 30 FPS
		IterationInterval: 50 * time.Millisecond, // 50ms iteration
		CallTimeout:       30 * time.Second,      // 30 second timeout
	}
}

// CallEventHandler defines the interface for handling call events
type CallEventHandler interface {
	OnCallEvent(event *CallEvent)
}

// Manager manages ToxAV call functionality and integrates with the Tox instance
type Manager struct {
	toxAV  *toxcore.ToxAV
	config *Config

	// State management
	mu          sync.RWMutex
	activeCalls map[uint32]*Call // friendID -> Call
	callHistory []*Call          // Historical calls
	running     bool

	// Event handling
	eventHandler CallEventHandler
	eventChan    chan *CallEvent

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// NewManager creates a new call manager instance
func NewManager(tox *toxcore.Tox, config *Config, eventHandler CallEventHandler) (*Manager, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Create ToxAV instance
	toxAV, err := toxcore.NewToxAV(tox)
	if err != nil {
		return nil, fmt.Errorf("failed to create ToxAV instance: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	manager := &Manager{
		toxAV:        toxAV,
		config:       config,
		activeCalls:  make(map[uint32]*Call),
		callHistory:  make([]*Call, 0),
		eventHandler: eventHandler,
		eventChan:    make(chan *CallEvent, 100), // Buffered channel
		ctx:          ctx,
		cancel:       cancel,
	}

	// Set up ToxAV callbacks
	manager.setupCallbacks()

	return manager, nil
}

// Start begins the call manager operations
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("call manager is already running")
	}

	m.running = true

	// Start the main event loop
	go m.eventLoop()

	// Start the ToxAV iteration loop
	go m.iterationLoop()

	log.Println("Call manager started successfully")
	return nil
}

// Stop gracefully shuts down the call manager
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return fmt.Errorf("call manager is not running")
	}

	m.running = false

	// End all active calls
	for _, call := range m.activeCalls {
		call.SetState(CallStateEnding)
		m.endCall(call.FriendID, "Call manager shutting down")
	}

	// Cancel context to stop goroutines
	m.cancel()

	// Close event channel
	close(m.eventChan)

	// Clean up ToxAV
	if m.toxAV != nil {
		m.toxAV.Kill()
	}

	log.Println("Call manager stopped")
	return nil
}

// PlaceCall initiates an outgoing call to a friend
func (m *Manager) PlaceCall(friendID uint32, callType CallType) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if there's already an active call with this friend
	if existingCall, exists := m.activeCalls[friendID]; exists {
		return fmt.Errorf("call already active with friend %d: %s", friendID, existingCall.State)
	}

	// Create new outgoing call
	call := NewCall(friendID, callType, true)
	m.activeCalls[friendID] = call

	// Determine bitrates based on call type
	audioBitRate := m.config.AudioBitRate
	videoBitRate := uint32(0)
	if callType == CallTypeVideo {
		videoBitRate = m.config.VideoBitRate
	}

	// Place the call using ToxAV
	err := m.toxAV.Call(friendID, audioBitRate, videoBitRate)
	if err != nil {
		// Clean up on failure
		delete(m.activeCalls, friendID)
		call.SetState(CallStateEnded)

		event := NewCallErrorEvent(call, fmt.Errorf("failed to place call: %w", err))
		m.sendEvent(event)

		return fmt.Errorf("failed to place call to friend %d: %w", friendID, err)
	}

	// Send call event
	event := NewCallEvent(CallEventIncoming, call, "Outgoing call initiated")
	m.sendEvent(event)

	// Set timeout for the call
	go m.handleCallTimeout(call)

	log.Printf("Placed %s call to friend %d", callType, friendID)
	return nil
}

// AnswerCall accepts an incoming call
func (m *Manager) AnswerCall(friendID uint32) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	call, exists := m.activeCalls[friendID]
	if !exists {
		return fmt.Errorf("no incoming call from friend %d", friendID)
	}

	if call.State != CallStateIncoming {
		return fmt.Errorf("call from friend %d is not in incoming state: %s", friendID, call.State)
	}

	// Determine bitrates based on call type
	audioBitRate := m.config.AudioBitRate
	videoBitRate := uint32(0)
	if call.Type == CallTypeVideo {
		videoBitRate = m.config.VideoBitRate
	}

	// Answer the call using ToxAV
	err := m.toxAV.Answer(friendID, audioBitRate, videoBitRate)
	if err != nil {
		call.SetState(CallStateEnded)
		delete(m.activeCalls, friendID)

		event := NewCallErrorEvent(call, fmt.Errorf("failed to answer call: %w", err))
		m.sendEvent(event)

		return fmt.Errorf("failed to answer call from friend %d: %w", friendID, err)
	}

	call.SetState(CallStateActive)

	// Send call event
	event := NewCallEvent(CallEventStateChanged, call, "Call answered and active")
	m.sendEvent(event)

	log.Printf("Answered %s call from friend %d", call.Type, friendID)
	return nil
}

// EndCall terminates an active call
func (m *Manager) EndCall(friendID uint32) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.endCall(friendID, "Call ended by user")
}

// endCall is the internal implementation for ending calls (requires lock)
func (m *Manager) endCall(friendID uint32, reason string) error {
	call, exists := m.activeCalls[friendID]
	if !exists {
		return fmt.Errorf("no active call with friend %d", friendID)
	}

	// Use ToxAV call control to hang up
	err := m.toxAV.CallControl(friendID, 0) // 0 = TOXAV_CALL_CONTROL_CANCEL/FINISH
	if err != nil {
		log.Printf("Warning: ToxAV call control failed for friend %d: %v", friendID, err)
	}

	// Update call state
	call.SetState(CallStateEnded)

	// Move to history and remove from active calls
	m.callHistory = append(m.callHistory, call)
	delete(m.activeCalls, friendID)

	// Send call event
	event := NewCallEvent(CallEventEnded, call, reason)
	m.sendEvent(event)

	log.Printf("Ended call with friend %d: %s", friendID, reason)
	return nil
}

// GetActiveCall returns the active call for a friend, if any
func (m *Manager) GetActiveCall(friendID uint32) (*Call, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	call, exists := m.activeCalls[friendID]
	return call, exists
}

// GetActiveCalls returns all currently active calls
func (m *Manager) GetActiveCalls() []*Call {
	m.mu.RLock()
	defer m.mu.RUnlock()

	calls := make([]*Call, 0, len(m.activeCalls))
	for _, call := range m.activeCalls {
		calls = append(calls, call)
	}

	return calls
}

// GetCallHistory returns the call history
func (m *Manager) GetCallHistory() []*Call {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	history := make([]*Call, len(m.callHistory))
	copy(history, m.callHistory)

	return history
}

// IsRunning returns whether the call manager is currently running
func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// sendEvent sends a call event to the event channel
func (m *Manager) sendEvent(event *CallEvent) {
	select {
	case m.eventChan <- event:
		// Event sent successfully
	default:
		// Channel is full, log warning
		log.Printf("Warning: Call event channel is full, dropping event: %+v", event)
	}
}

// handleCallTimeout handles timeouts for outgoing calls
func (m *Manager) handleCallTimeout(call *Call) {
	select {
	case <-time.After(m.config.CallTimeout):
		if call.GetState() == CallStateOutgoing {
			m.mu.Lock()
			m.endCall(call.FriendID, "Call timeout")
			m.mu.Unlock()
		}
	case <-call.Context().Done():
		// Call was ended before timeout
		return
	}
}

// eventLoop processes call events and forwards them to the handler
func (m *Manager) eventLoop() {
	for {
		select {
		case event, ok := <-m.eventChan:
			if !ok {
				// Channel closed, exit loop
				return
			}

			if m.eventHandler != nil {
				m.eventHandler.OnCallEvent(event)
			}

		case <-m.ctx.Done():
			// Context cancelled, exit loop
			return
		}
	}
}

// iterationLoop runs the ToxAV iteration loop
func (m *Manager) iterationLoop() {
	ticker := time.NewTicker(m.config.IterationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if m.toxAV != nil {
				m.toxAV.Iterate()
			}

		case <-m.ctx.Done():
			// Context cancelled, exit loop
			return
		}
	}
}
