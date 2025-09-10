package audio

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MockRecorder implements Recorder interface for demo/testing
type MockRecorder struct {
	mu           sync.RWMutex
	state        RecordingState
	duration     time.Duration
	startTime    time.Time
	currentLevel float32
}

// NewMockRecorder creates a new mock recorder
func NewMockRecorder() *MockRecorder {
	return &MockRecorder{
		state: RecordingStateIdle,
	}
}

// Start begins recording simulation
func (r *MockRecorder) Start(ctx context.Context, options RecordingOptions, callback RecordingCallback) error {
	r.mu.Lock()
	if r.state != RecordingStateIdle {
		r.mu.Unlock()
		return fmt.Errorf("recorder is not idle")
	}

	r.state = RecordingStateRecording
	r.startTime = time.Now()
	r.currentLevel = 0.0
	r.mu.Unlock()

	// Simulate recording in background
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.mu.RLock()
				currentState := r.state
				startTime := r.startTime
				r.mu.RUnlock()

				if currentState != RecordingStateRecording {
					return
				}

				// Simulate audio level changes
				newLevel := 0.1 + 0.05*float32(time.Now().UnixNano()%100)/100.0

				r.mu.Lock()
				r.currentLevel = newLevel
				r.mu.Unlock()

				// Call callback if provided
				if callback != nil {
					samples := make([]float32, 100) // Mock samples
					callback(samples, newLevel)
				}

				// Check duration limits
				if options.MaxDuration > 0 && time.Since(startTime) >= options.MaxDuration {
					r.mu.Lock()
					r.state = RecordingStateStopped
					r.mu.Unlock()
					return
				}
			}
		}
	}()

	return nil
}

// Stop stops recording and returns the voice message
func (r *MockRecorder) Stop() (*VoiceMessage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.state != RecordingStateRecording && r.state != RecordingStatePaused {
		return nil, fmt.Errorf("not currently recording")
	}

	r.duration = time.Since(r.startTime)
	r.state = RecordingStateStopped

	// Create mock voice message
	voiceMsg := &VoiceMessage{
		Duration:  r.duration,
		Format:    DefaultVoiceFormat(),
		CreatedAt: r.startTime,
		FileSize:  int64(r.duration.Seconds() * 8000), // Approximate size
		Waveform:  []float32{0.1, 0.3, 0.5, 0.2, 0.1}, // Mock waveform
	}

	return voiceMsg, nil
}

// Pause pauses recording
func (r *MockRecorder) Pause() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.state != RecordingStateRecording {
		return fmt.Errorf("not currently recording")
	}
	r.state = RecordingStatePaused
	return nil
}

// Resume resumes recording
func (r *MockRecorder) Resume() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.state != RecordingStatePaused {
		return fmt.Errorf("not currently paused")
	}
	r.state = RecordingStateRecording
	return nil
}

// GetState returns current recording state
func (r *MockRecorder) GetState() RecordingState {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state
}

// GetLevel returns current audio input level
func (r *MockRecorder) GetLevel() float32 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.currentLevel
}

// GetDuration returns current recording duration
func (r *MockRecorder) GetDuration() time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.state == RecordingStateIdle || r.state == RecordingStateStopped {
		return r.duration
	}
	return time.Since(r.startTime)
}

// Cancel cancels recording without saving
func (r *MockRecorder) Cancel() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.state = RecordingStateIdle
	r.duration = 0
	return nil
}

// IsSupported returns true (mock always supports recording)
func (r *MockRecorder) IsSupported() bool {
	return true
}
