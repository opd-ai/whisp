// Package calls implements P2P voice and video calling functionality using ToxAV protocol.
// This package provides a high-level interface for managing call sessions, audio/video streams,
// and call state management across the Whisp application.
package calls

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// CallState represents the current state of a call
type CallState int

const (
	// CallStateNone indicates no active call
	CallStateNone CallState = iota
	// CallStateIncoming indicates an incoming call waiting to be answered
	CallStateIncoming
	// CallStateOutgoing indicates an outgoing call being placed
	CallStateOutgoing
	// CallStateActive indicates an active call in progress
	CallStateActive
	// CallStateHolding indicates a call on hold
	CallStateHolding
	// CallStateEnding indicates a call being terminated
	CallStateEnding
	// CallStateEnded indicates a call has ended
	CallStateEnded
)

// String returns the string representation of CallState
func (s CallState) String() string {
	switch s {
	case CallStateNone:
		return "none"
	case CallStateIncoming:
		return "incoming"
	case CallStateOutgoing:
		return "outgoing"
	case CallStateActive:
		return "active"
	case CallStateHolding:
		return "holding"
	case CallStateEnding:
		return "ending"
	case CallStateEnded:
		return "ended"
	default:
		return "unknown"
	}
}

// CallType represents the type of call (audio only or audio+video)
type CallType int

const (
	// CallTypeAudio indicates an audio-only call
	CallTypeAudio CallType = iota
	// CallTypeVideo indicates an audio+video call
	CallTypeVideo
)

// String returns the string representation of CallType
func (t CallType) String() string {
	switch t {
	case CallTypeAudio:
		return "audio"
	case CallTypeVideo:
		return "video"
	default:
		return "unknown"
	}
}

// CallEventType represents different types of call events
type CallEventType string

const (
	CallEventIncoming        CallEventType = "incoming"         // Incoming call received
	CallEventOutgoing        CallEventType = "outgoing"         // Outgoing call initiated
	CallEventStateChanged    CallEventType = "state_changed"    // Call state changed
	CallEventEnded           CallEventType = "ended"            // Call ended
	CallEventError           CallEventType = "error"            // Call error occurred
	CallEventAudioFrame      CallEventType = "audio_frame"      // Audio frame received
	CallEventVideoFrame      CallEventType = "video_frame"      // Video frame received
	CallEventBitrateChanged  CallEventType = "bitrate_changed"  // Bitrate changed
)

// AudioFrame represents an audio frame received during a call
type AudioFrame struct {
	FriendID     uint32    // Friend who sent the frame
	PCMData      []int16   // PCM audio data
	SampleCount  int       // Number of audio samples
	Channels     uint8     // Number of audio channels
	SamplingRate uint32    // Audio sampling rate
	Timestamp    time.Time // When the frame was received
}

// VideoFrame represents a video frame received during a call
type VideoFrame struct {
	FriendID  uint32    // Friend who sent the frame
	Width     uint16    // Frame width
	Height    uint16    // Frame height
	YPlane    []byte    // Y (luminance) plane data
	UPlane    []byte    // U (chrominance) plane data
	VPlane    []byte    // V (chrominance) plane data
	YStride   int       // Y plane stride
	UStride   int       // U plane stride
	VStride   int       // V plane stride
	Timestamp time.Time // When the frame was received
}

// Call represents an active or historical voice/video call
type Call struct {
	// Call identification
	ID       string    // Unique call ID
	FriendID uint32    // Tox friend number
	Type     CallType  // Audio or video call
	
	// Call status
	State       CallState // Current call state
	IsOutgoing  bool      // true if we initiated the call
	StartTime   time.Time // When the call started
	EndTime     *time.Time // When the call ended (nil if active)
	
	// Media settings
	audioEnabled bool   // Whether audio is enabled
	videoEnabled bool   // Whether video is enabled
	audioBitrate uint32 // Current audio bitrate
	videoBitrate uint32 // Current video bitrate
	
	// Statistics
	audioFrameCount uint64 // Number of audio frames processed
	videoFrameCount uint64 // Number of video frames processed
	
	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
	
	// Synchronization
	mu sync.RWMutex
}

// NewCall creates a new call instance
func NewCall(friendID uint32, callType CallType, isOutgoing bool) *Call {
	ctx, cancel := context.WithCancel(context.Background())
	
	call := &Call{
		ID:         uuid.New().String(),
		FriendID:   friendID,
		Type:       callType,
		IsOutgoing: isOutgoing,
		StartTime:  time.Now(),
		
		audioEnabled: true, // Audio enabled by default
		videoEnabled: callType == CallTypeVideo, // Video enabled only for video calls
		
		ctx:    ctx,
		cancel: cancel,
	}
	
	// Set initial state based on direction
	if isOutgoing {
		call.State = CallStateOutgoing
	} else {
		call.State = CallStateIncoming
	}
	
	return call
}

// GetState returns the current call state (thread-safe)
func (c *Call) GetState() CallState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.State
}

// SetState updates the call state (thread-safe)
func (c *Call) SetState(state CallState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.State = state
	
	// Set end time when call ends
	if state == CallStateEnded {
		now := time.Now()
		c.EndTime = &now
		c.cancel() // Cancel the context
	}
}

// IsAudioEnabled returns whether audio is enabled
func (c *Call) IsAudioEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.audioEnabled
}

// SetAudioEnabled sets the audio enabled state
func (c *Call) SetAudioEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.audioEnabled = enabled
}

// IsVideoEnabled returns whether video is enabled
func (c *Call) IsVideoEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.videoEnabled
}

// SetVideoEnabled sets the video enabled state
func (c *Call) SetVideoEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.videoEnabled = enabled
}

// GetAudioBitrate returns the current audio bitrate
func (c *Call) GetAudioBitrate() uint32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.audioBitrate
}

// SetAudioBitrate sets the audio bitrate
func (c *Call) SetAudioBitrate(bitrate uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.audioBitrate = bitrate
}

// GetVideoBitrate returns the current video bitrate
func (c *Call) GetVideoBitrate() uint32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.videoBitrate
}

// SetVideoBitrate sets the video bitrate
func (c *Call) SetVideoBitrate(bitrate uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.videoBitrate = bitrate
}

// ToggleAudio toggles the audio enabled state
func (c *Call) ToggleAudio() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.audioEnabled = !c.audioEnabled
	return c.audioEnabled
}

// ToggleVideo toggles the video enabled state
func (c *Call) ToggleVideo() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.videoEnabled = !c.videoEnabled
	return c.videoEnabled
}

// Duration returns the duration of the call
func (c *Call) Duration() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.EndTime != nil {
		return c.EndTime.Sub(c.StartTime)
	}
	
	// Call is still active
	return time.Since(c.StartTime)
}

// Context returns the call's context for cancellation
func (c *Call) Context() context.Context {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ctx
}

// String returns a string representation of the call
func (c *Call) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	direction := "incoming"
	if c.IsOutgoing {
		direction = "outgoing"
	}
	
	duration := c.Duration()
	
	return fmt.Sprintf("Call{ID:%s, Friend:%d, Type:%s, State:%s, Direction:%s, Duration:%v}",
		c.ID, c.FriendID, c.Type, c.State, direction, duration)
}

// CallEvent represents an event that occurs during a call
type CallEvent struct {
	Type       CallEventType // Type of the event
	Call       *Call         // Associated call
	Message    string        // Human-readable message
	Error      error         // Error, if any
	AudioFrame *AudioFrame   // Audio frame data, if applicable
	VideoFrame *VideoFrame   // Video frame data, if applicable
	Timestamp  time.Time     // When the event occurred
}

// NewCallEvent creates a new call event
func NewCallEvent(eventType CallEventType, call *Call, message string) *CallEvent {
	return &CallEvent{
		Type:      eventType,
		Call:      call,
		Message:   message,
		Timestamp: time.Now(),
	}
}

// NewCallErrorEvent creates a new call error event
func NewCallErrorEvent(call *Call, err error) *CallEvent {
	return &CallEvent{
		Type:      CallEventError,
		Call:      call,
		Message:   err.Error(),
		Error:     err,
		Timestamp: time.Now(),
	}
}

// NewAudioFrameEvent creates a new audio frame event
func NewAudioFrameEvent(call *Call, frame *AudioFrame) *CallEvent {
	return &CallEvent{
		Type:       CallEventAudioFrame,
		Call:       call,
		Message:    fmt.Sprintf("Audio frame received (%d samples)", frame.SampleCount),
		AudioFrame: frame,
		Timestamp:  time.Now(),
	}
}

// NewVideoFrameEvent creates a new video frame event
func NewVideoFrameEvent(call *Call, frame *VideoFrame) *CallEvent {
	return &CallEvent{
		Type:       CallEventVideoFrame,
		Call:       call,
		Message:    fmt.Sprintf("Video frame received (%dx%d)", frame.Width, frame.Height),
		VideoFrame: frame,
		Timestamp:  time.Now(),
	}
}

// String returns a string representation of the call event
func (ce *CallEvent) String() string {
	return fmt.Sprintf("CallEvent{Type:%s, Call:%s, Message:%s, Time:%v}",
		ce.Type, ce.Call.ID, ce.Message, ce.Timestamp)
}