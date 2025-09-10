package audio

import (
	"fmt"
	"sync"
)

// MockManager implements Manager for demonstration/testing
type MockManager struct {
	mu          sync.RWMutex
	initialized bool
	
	// Supported formats
	supportedFormats []AudioFormat
}

// NewMockManager creates a new mock audio manager
func NewMockManager() *MockManager {
	return &MockManager{
		supportedFormats: []AudioFormat{
			DefaultVoiceFormat(),
			{
				SampleRate: 44100,
				Channels:   2,
				BitDepth:   16,
				Codec:      "wav",
			},
			{
				SampleRate: 22050,
				Channels:   1,
				BitDepth:   16,
				Codec:      "wav",
			},
		},
	}
}

// Initialize initializes the audio system
func (m *MockManager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.initialized {
		return fmt.Errorf("already initialized")
	}
	
	// For mock implementation, no actual initialization needed
	m.initialized = true
	return nil
}

// Shutdown shuts down the audio system
func (m *MockManager) Shutdown() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.initialized {
		return fmt.Errorf("not initialized")
	}
	
	m.initialized = false
	return nil
}

// IsInitialized returns true if audio system is initialized
func (m *MockManager) IsInitialized() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.initialized
}

// GetRecorder returns a new recorder instance
func (m *MockManager) GetRecorder() (Recorder, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if !m.initialized {
		return nil, fmt.Errorf("audio system not initialized")
	}
	
	return NewMockRecorder(), nil
}

// GetPlayer returns a new player instance
func (m *MockManager) GetPlayer() (Player, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if !m.initialized {
		return nil, fmt.Errorf("audio system not initialized")
	}
	
	return NewMockPlayer(), nil
}

// GetWaveformGenerator returns waveform generator
func (m *MockManager) GetWaveformGenerator() WaveformGenerator {
	return NewMockWaveformGenerator()
}

// GetSupportedFormats returns list of supported audio formats
func (m *MockManager) GetSupportedFormats() []AudioFormat {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Return a copy to prevent modification
	formats := make([]AudioFormat, len(m.supportedFormats))
	copy(formats, m.supportedFormats)
	return formats
}
