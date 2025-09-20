// Package calls implements ToxAV callback handlers for call events.
// This file contains the ToxAV callback setup and event processing.
package calls

import (
	"fmt"
	"log"

	"github.com/opd-ai/toxcore/av"
)

// setupCallbacks configures the ToxAV callbacks for handling call events
func (m *Manager) setupCallbacks() {
	// Set call callback - handles incoming calls
	m.toxAV.CallbackCall(m.onCallReceived)

	// Set call state callback - handles call state changes
	m.toxAV.CallbackCallState(m.onCallStateChanged)

	// Set audio receive frame callback - handles incoming audio
	m.toxAV.CallbackAudioReceiveFrame(m.onAudioFrameReceived)

	// Set video receive frame callback - handles incoming video
	m.toxAV.CallbackVideoReceiveFrame(m.onVideoFrameReceived)

	// Set audio bitrate callback - handles audio bitrate changes
	m.toxAV.CallbackAudioBitRate(m.onAudioBitrateChanged)

	// Set video bitrate callback - handles video bitrate changes
	m.toxAV.CallbackVideoBitRate(m.onVideoBitrateChanged)

	log.Println("ToxAV callbacks configured successfully")
}

// onCallReceived handles incoming call events from ToxAV
func (m *Manager) onCallReceived(friendNumber uint32, audioEnabled, videoEnabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if there's already an active call with this friend
	if existingCall, exists := m.activeCalls[friendNumber]; exists {
		log.Printf("Warning: Received call from friend %d but call already exists: %s",
			friendNumber, existingCall.State)
		return
	}

	// Determine call type based on audio/video flags
	callType := CallTypeAudio
	if videoEnabled {
		callType = CallTypeVideo
	}

	// Create new incoming call
	call := NewCall(friendNumber, callType, false)
	call.SetState(CallStateIncoming)
	m.activeCalls[friendNumber] = call

	// Send call event
	event := NewCallEvent(CallEventIncoming, call, "Incoming call received")
	m.sendEvent(event)

	log.Printf("Received incoming %s call from friend %d", callType, friendNumber)
}

// onCallStateChanged handles call state change events from ToxAV
func (m *Manager) onCallStateChanged(friendNumber uint32, state av.CallState) {
	m.mu.Lock()
	defer m.mu.Unlock()

	call, exists := m.activeCalls[friendNumber]
	if !exists {
		log.Printf("Warning: Received call state change for unknown friend %d", friendNumber)
		return
	}

	// Map ToxAV states to our internal states
	var newState CallState
	var eventType CallEventType
	var message string

	// Convert toxcore.CallState to our internal state (assuming it's a bitmask)
	stateValue := uint32(state)

	switch {
	case stateValue == 0: // TOXAV_FRIEND_CALL_STATE_FINISHED
		newState = CallStateEnded
		eventType = CallEventEnded
		message = "Call ended by peer"

		// Move to history and remove from active calls
		m.callHistory = append(m.callHistory, call)
		delete(m.activeCalls, friendNumber)

	case stateValue&1 != 0: // TOXAV_FRIEND_CALL_STATE_SENDING_A
		newState = CallStateActive
		eventType = CallEventStateChanged
		message = "Audio transmission started"
		call.SetAudioEnabled(true)

	case stateValue&2 != 0: // TOXAV_FRIEND_CALL_STATE_SENDING_V
		newState = CallStateActive
		eventType = CallEventStateChanged
		message = "Video transmission started"
		call.SetVideoEnabled(true)

	case stateValue&4 != 0: // TOXAV_FRIEND_CALL_STATE_ACCEPTING_A
		newState = CallStateActive
		eventType = CallEventStateChanged
		message = "Audio reception started"

	case stateValue&8 != 0: // TOXAV_FRIEND_CALL_STATE_ACCEPTING_V
		newState = CallStateActive
		eventType = CallEventStateChanged
		message = "Video reception started"

	default:
		log.Printf("Unknown ToxAV call state %d for friend %d", stateValue, friendNumber)
		return
	}

	// Update call state
	call.SetState(newState)

	// Send call event
	event := NewCallEvent(eventType, call, message)
	m.sendEvent(event)

	log.Printf("Call state changed for friend %d: %s (%s)", friendNumber, newState, message)
}

// onAudioFrameReceived handles incoming audio frames from ToxAV
func (m *Manager) onAudioFrameReceived(friendNumber uint32, pcm []int16, sampleCount int,
	channels uint8, samplingRate uint32,
) {
	call, exists := m.GetActiveCall(friendNumber)
	if !exists {
		log.Printf("Warning: Received audio frame for unknown call from friend %d", friendNumber)
		return
	}

	if call.State != CallStateActive {
		log.Printf("Warning: Received audio frame for inactive call from friend %d: %s",
			friendNumber, call.State)
		return
	}

	// Create audio frame event
	audioFrame := &AudioFrame{
		FriendID:     friendNumber,
		PCMData:      pcm,
		SampleCount:  sampleCount,
		Channels:     channels,
		SamplingRate: samplingRate,
		Timestamp:    call.StartTime,
	}

	// Send audio frame event
	event := NewAudioFrameEvent(call, audioFrame)
	m.sendEvent(event)

	// Increment frame counter (thread-safe)
	call.mu.Lock()
	call.audioFrameCount++
	frameCount := call.audioFrameCount
	call.mu.Unlock()

	// Log periodic audio frame reception (every 1000 frames to avoid spam)
	if frameCount%1000 == 0 {
		log.Printf("Received audio frame %d from friend %d (samples=%d, rate=%d)",
			frameCount, friendNumber, sampleCount, samplingRate)
	}
}

// onVideoFrameReceived handles incoming video frames from ToxAV
func (m *Manager) onVideoFrameReceived(friendNumber uint32, width, height uint16,
	y, u, v []byte, yStride, uStride, vStride int,
) {
	call, exists := m.GetActiveCall(friendNumber)
	if !exists {
		log.Printf("Warning: Received video frame for unknown call from friend %d", friendNumber)
		return
	}

	if call.State != CallStateActive {
		log.Printf("Warning: Received video frame for inactive call from friend %d: %s",
			friendNumber, call.State)
		return
	}

	if !call.IsVideoEnabled() {
		log.Printf("Warning: Received video frame but video is disabled for friend %d", friendNumber)
		return
	}

	// Create video frame event
	videoFrame := &VideoFrame{
		FriendID:  friendNumber,
		Width:     width,
		Height:    height,
		YPlane:    y,
		UPlane:    u,
		VPlane:    v,
		YStride:   yStride,
		UStride:   uStride,
		VStride:   vStride,
		Timestamp: call.StartTime,
	}

	// Send video frame event
	event := NewVideoFrameEvent(call, videoFrame)
	m.sendEvent(event)

	// Increment frame counter (thread-safe)
	call.mu.Lock()
	call.videoFrameCount++
	frameCount := call.videoFrameCount
	call.mu.Unlock()

	// Log periodic video frame reception (every 30 frames to avoid spam)
	if frameCount%30 == 0 {
		log.Printf("Received video frame %d from friend %d (%dx%d)",
			frameCount, friendNumber, width, height)
	}
}

// onAudioBitrateChanged handles audio bitrate change events from ToxAV
func (m *Manager) onAudioBitrateChanged(friendNumber, bitrate uint32) {
	call, exists := m.GetActiveCall(friendNumber)
	if !exists {
		log.Printf("Warning: Audio bitrate changed for unknown call from friend %d", friendNumber)
		return
	}

	// Update call audio bitrate
	call.SetAudioBitrate(bitrate)

	// Send bitrate change event
	event := NewCallEvent(CallEventBitrateChanged, call,
		fmt.Sprintf("Audio bitrate changed to %d kbps", bitrate))
	m.sendEvent(event)

	log.Printf("Audio bitrate changed for friend %d: %d kbps", friendNumber, bitrate)
}

// onVideoBitrateChanged handles video bitrate change events from ToxAV
func (m *Manager) onVideoBitrateChanged(friendNumber, bitrate uint32) {
	call, exists := m.GetActiveCall(friendNumber)
	if !exists {
		log.Printf("Warning: Video bitrate changed for unknown call from friend %d", friendNumber)
		return
	}

	// Update call video bitrate
	call.SetVideoBitrate(bitrate)

	// Send bitrate change event
	event := NewCallEvent(CallEventBitrateChanged, call,
		fmt.Sprintf("Video bitrate changed to %d kbps", bitrate))
	m.sendEvent(event)

	log.Printf("Video bitrate changed for friend %d: %d kbps", friendNumber, bitrate)
}
