package audio

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MockPlayer implements Player for demonstration/testing
type MockPlayer struct {
	mu       sync.RWMutex
	state    PlaybackState
	filePath string
	options  PlaybackOptions
	callback PlaybackCallback

	// Playback state
	duration  time.Duration
	position  time.Duration
	startTime time.Time
	volume    float32

	// Context and cancellation
	ctx    context.Context
	cancel context.CancelFunc

	// Simulation
	ticker *time.Ticker
}

// NewMockPlayer creates a new mock player for demo/testing
func NewMockPlayer() *MockPlayer {
	return &MockPlayer{
		state:  PlaybackStateIdle,
		volume: 1.0,
	}
}

// Load loads an audio file for playback
func (p *MockPlayer) Load(filePath string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.filePath = filePath

	// Simulate loading file and getting duration
	// For demo purposes, we'll use a fixed duration based on file name
	p.duration = 5 * time.Second // Default 5 seconds

	// Could analyze WAV file header here for real implementation

	p.position = 0
	p.state = PlaybackStateIdle

	return nil
}

// Play starts playback
func (p *MockPlayer) Play(options PlaybackOptions, callback PlaybackCallback) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.filePath == "" {
		return fmt.Errorf("no file loaded")
	}

	if p.state == PlaybackStatePlaying {
		return fmt.Errorf("already playing")
	}

	p.options = options
	p.callback = callback
	p.volume = options.Volume

	// Create context for cancellation
	p.ctx, p.cancel = context.WithCancel(context.Background())

	// Start from specified offset
	p.position = options.StartOffset
	p.startTime = time.Now().Add(-p.position)

	p.state = PlaybackStatePlaying

	// Start playback simulation
	p.ticker = time.NewTicker(100 * time.Millisecond) // 10Hz updates
	go p.simulatePlayback()

	return nil
}

// simulatePlayback simulates audio playback
func (p *MockPlayer) simulatePlayback() {
	defer func() {
		if p.ticker != nil {
			p.ticker.Stop()
		}
	}()

	if p.ctx == nil || p.ticker == nil {
		return
	}

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-p.ticker.C:
			p.mu.Lock()

			if p.state != PlaybackStatePlaying {
				p.mu.Unlock()
				continue
			}

			// Update position based on playback speed
			elapsed := time.Since(p.startTime)
			p.position = time.Duration(float64(elapsed) * float64(p.options.Speed))

			// Check if playback finished
			if p.position >= p.duration {
				p.position = p.duration
				if p.options.AutoStop {
					p.state = PlaybackStateStopped
					if p.ticker != nil {
						p.ticker.Stop()
						p.ticker = nil
					}
				}
			}

			// Call user callback
			if p.callback != nil {
				p.callback(p.position, p.duration)
			}

			p.mu.Unlock()

			// Stop if reached end
			if p.position >= p.duration && p.options.AutoStop {
				return
			}
		}
	}
}

// Pause pauses playback
func (p *MockPlayer) Pause() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state != PlaybackStatePlaying {
		return fmt.Errorf("not currently playing")
	}

	p.state = PlaybackStatePaused
	return nil
}

// Resume resumes playback
func (p *MockPlayer) Resume() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state != PlaybackStatePaused {
		return fmt.Errorf("not currently paused")
	}

	// Adjust start time to account for pause
	p.startTime = time.Now().Add(-p.position)
	p.state = PlaybackStatePlaying
	return nil
}

// Stop stops playback
func (p *MockPlayer) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancel != nil {
		p.cancel()
	}

	if p.ticker != nil {
		p.ticker.Stop()
		p.ticker = nil
	}

	p.state = PlaybackStateStopped
	p.position = 0

	return nil
}

// Seek seeks to a specific position
func (p *MockPlayer) Seek(position time.Duration) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if position < 0 {
		position = 0
	}
	if position > p.duration {
		position = p.duration
	}

	p.position = position

	// Adjust start time if playing
	if p.state == PlaybackStatePlaying {
		p.startTime = time.Now().Add(-position)
	}

	return nil
}

// GetState returns current playback state
func (p *MockPlayer) GetState() PlaybackState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}

// GetPosition returns current playback position
func (p *MockPlayer) GetPosition() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.state == PlaybackStatePlaying {
		// Calculate real-time position
		elapsed := time.Since(p.startTime)
		return time.Duration(float64(elapsed) * float64(p.options.Speed))
	}

	return p.position
}

// GetDuration returns total audio duration
func (p *MockPlayer) GetDuration() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.duration
}

// SetVolume sets playback volume
func (p *MockPlayer) SetVolume(volume float32) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if volume < 0 {
		volume = 0
	}
	if volume > 1 {
		volume = 1
	}

	p.volume = volume
	return nil
}

// GetVolume returns current volume
func (p *MockPlayer) GetVolume() float32 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.volume
}

// IsSupported returns true (mock always supports playback)
func (p *MockPlayer) IsSupported() bool {
	return true
}
