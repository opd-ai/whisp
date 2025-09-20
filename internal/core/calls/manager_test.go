package calls

import (
	"fmt"
	"testing"
	"time"

	"github.com/opd-ai/toxcore"
)

// MockEventHandler for testing
type MockEventHandler struct {
	events []*CallEvent
}

func (h *MockEventHandler) OnCallEvent(event *CallEvent) {
	h.events = append(h.events, event)
}

func TestCallManager_Creation(t *testing.T) {
	// Create a minimal Tox instance for testing
	opts := toxcore.NewOptionsForTesting()

	tox, err := toxcore.New(opts)
	if err != nil {
		t.Fatalf("Failed to create Tox instance: %v", err)
	}
	defer tox.Kill()

	// Create event handler
	handler := &MockEventHandler{}

	// Create call manager
	manager, err := NewManager(tox, nil, handler)
	if err != nil {
		t.Fatalf("Failed to create call manager: %v", err)
	}

	// Test manager properties
	if manager == nil {
		t.Fatal("Manager should not be nil")
	}

	if manager.IsRunning() {
		t.Error("Manager should not be running initially")
	}

	// Test configuration
	config := DefaultConfig()
	if config.AudioBitRate == 0 {
		t.Error("Default audio bitrate should not be zero")
	}

	if config.AudioSampleRate == 0 {
		t.Error("Default audio sample rate should not be zero")
	}
}

func TestCallManager_StartStop(t *testing.T) {
	// Create Tox instance
	opts := toxcore.NewOptionsForTesting()

	tox, err := toxcore.New(opts)
	if err != nil {
		t.Fatalf("Failed to create Tox instance: %v", err)
	}
	defer tox.Kill()

	// Create call manager
	handler := &MockEventHandler{}
	manager, err := NewManager(tox, nil, handler)
	if err != nil {
		t.Fatalf("Failed to create call manager: %v", err)
	}

	// Test start
	err = manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	if !manager.IsRunning() {
		t.Error("Manager should be running after start")
	}

	// Test stop
	err = manager.Stop()
	if err != nil {
		t.Fatalf("Failed to stop manager: %v", err)
	}

	if manager.IsRunning() {
		t.Error("Manager should not be running after stop")
	}
}

func TestCall_StateManagement(t *testing.T) {
	// Create a new call
	call := NewCall(123, CallTypeAudio, true)

	// Test initial state
	if call.GetState() != CallStateOutgoing {
		t.Errorf("Expected initial state CallStateOutgoing, got %v", call.GetState())
	}

	if call.FriendID != 123 {
		t.Errorf("Expected friend ID 123, got %d", call.FriendID)
	}

	if call.Type != CallTypeAudio {
		t.Errorf("Expected CallTypeAudio, got %v", call.Type)
	}

	if !call.IsOutgoing {
		t.Error("Call should be outgoing")
	}

	// Test audio/video state
	if !call.IsAudioEnabled() {
		t.Error("Audio should be enabled by default")
	}

	if call.IsVideoEnabled() {
		t.Error("Video should not be enabled for audio call")
	}

	// Test state changes
	call.SetState(CallStateActive)
	if call.GetState() != CallStateActive {
		t.Errorf("Expected state CallStateActive, got %v", call.GetState())
	}

	// Test end state
	call.SetState(CallStateEnded)
	if call.GetState() != CallStateEnded {
		t.Errorf("Expected state CallStateEnded, got %v", call.GetState())
	}

	if call.EndTime == nil {
		t.Error("EndTime should be set when call ends")
	}

	// Test duration
	duration := call.Duration()
	if duration <= 0 {
		t.Error("Duration should be positive for ended call")
	}
}

func TestCall_VideoCall(t *testing.T) {
	// Create a video call
	call := NewCall(456, CallTypeVideo, false)

	// Test initial state for incoming video call
	if call.GetState() != CallStateIncoming {
		t.Errorf("Expected initial state CallStateIncoming, got %v", call.GetState())
	}

	if call.Type != CallTypeVideo {
		t.Errorf("Expected CallTypeVideo, got %v", call.Type)
	}

	if call.IsOutgoing {
		t.Error("Call should be incoming")
	}

	// Test video state for video call
	if !call.IsVideoEnabled() {
		t.Error("Video should be enabled for video call")
	}

	// Test toggle functionality
	result := call.ToggleVideo()
	if result || call.IsVideoEnabled() {
		t.Error("Video should be disabled after toggle")
	}

	result = call.ToggleAudio()
	if result || call.IsAudioEnabled() {
		t.Error("Audio should be disabled after toggle")
	}
}

func TestCallEvent_Creation(t *testing.T) {
	call := NewCall(789, CallTypeAudio, true)

	// Test basic event
	event := NewCallEvent(CallEventOutgoing, call, "Test message")
	if event.Type != CallEventOutgoing {
		t.Errorf("Expected CallEventOutgoing, got %v", event.Type)
	}

	if event.Call != call {
		t.Error("Event should reference the correct call")
	}

	if event.Message != "Test message" {
		t.Errorf("Expected 'Test message', got %s", event.Message)
	}

	// Test error event
	testErr := fmt.Errorf("test error")
	errorEvent := NewCallErrorEvent(call, testErr)
	if errorEvent.Type != CallEventError {
		t.Errorf("Expected CallEventError, got %v", errorEvent.Type)
	}

	if errorEvent.Error != testErr {
		t.Error("Error event should contain the original error")
	}

	// Test frame events
	audioFrame := &AudioFrame{
		FriendID:     789,
		SampleCount:  480,
		Channels:     2,
		SamplingRate: 48000,
		Timestamp:    time.Now(),
	}

	audioEvent := NewAudioFrameEvent(call, audioFrame)
	if audioEvent.Type != CallEventAudioFrame {
		t.Errorf("Expected CallEventAudioFrame, got %v", audioEvent.Type)
	}

	if audioEvent.AudioFrame != audioFrame {
		t.Error("Audio event should contain the frame")
	}

	videoFrame := &VideoFrame{
		FriendID:  789,
		Width:     640,
		Height:    480,
		Timestamp: time.Now(),
	}

	videoEvent := NewVideoFrameEvent(call, videoFrame)
	if videoEvent.Type != CallEventVideoFrame {
		t.Errorf("Expected CallEventVideoFrame, got %v", videoEvent.Type)
	}

	if videoEvent.VideoFrame != videoFrame {
		t.Error("Video event should contain the frame")
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Test audio configuration
	if config.AudioBitRate < 32 || config.AudioBitRate > 128 {
		t.Errorf("Audio bitrate %d should be between 32-128 kbps", config.AudioBitRate)
	}

	if config.AudioSampleRate != 48000 {
		t.Errorf("Expected audio sample rate 48000, got %d", config.AudioSampleRate)
	}

	if config.AudioChannels != 1 {
		t.Errorf("Expected 1 audio channel, got %d", config.AudioChannels)
	}

	// Test video configuration
	if config.VideoBitRate < 100 || config.VideoBitRate > 2000 {
		t.Errorf("Video bitrate %d should be between 100-2000 kbps", config.VideoBitRate)
	}

	if config.VideoWidth != 640 || config.VideoHeight != 480 {
		t.Errorf("Expected 640x480 video, got %dx%d", config.VideoWidth, config.VideoHeight)
	}

	if config.VideoFPS != 30 {
		t.Errorf("Expected 30 FPS, got %d", config.VideoFPS)
	}

	// Test timing configuration
	if config.IterationInterval != 50*time.Millisecond {
		t.Errorf("Expected 50ms iteration interval, got %v", config.IterationInterval)
	}

	if config.CallTimeout != 30*time.Second {
		t.Errorf("Expected 30s call timeout, got %v", config.CallTimeout)
	}
}
