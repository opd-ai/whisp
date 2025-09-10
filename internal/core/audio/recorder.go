package audio

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/hraban/opus"
)

// PortAudioRecorder implements Recorder using PortAudio
type PortAudioRecorder struct {
	mu           sync.RWMutex
	stream       *portaudio.Stream
	state        RecordingState
	options      RecordingOptions
	callback     RecordingCallback
	
	// Recording data
	audioData    []float32
	startTime    time.Time
	pausedTime   time.Duration
	currentLevel float32
	
	// Context and cancellation
	ctx    context.Context
	cancel context.CancelFunc
	
	// Opus encoder
	encoder *opus.Encoder
	
	// Output file
	outputFile *os.File
}

// NewPortAudioRecorder creates a new PortAudio-based recorder
func NewPortAudioRecorder() *PortAudioRecorder {
	return &PortAudioRecorder{
		state: RecordingStateIdle,
	}
}

// Start begins recording
func (r *PortAudioRecorder) Start(ctx context.Context, options RecordingOptions, callback RecordingCallback) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.state != RecordingStateIdle {
		return fmt.Errorf("recorder is not idle")
	}
	
	// Store options and callback
	r.options = options
	r.callback = callback
	r.audioData = make([]float32, 0)
	r.startTime = time.Now()
	r.pausedTime = 0
	
	// Create context with cancellation
	r.ctx, r.cancel = context.WithCancel(ctx)
	
	// Create output directory if needed
	if r.options.OutputPath != "" {
		if err := os.MkdirAll(filepath.Dir(r.options.OutputPath), 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}
	
	// Initialize Opus encoder for compression
	var err error
	r.encoder, err = opus.NewEncoder(
		r.options.Format.SampleRate,
		r.options.Format.Channels,
		opus.AppVoIP, // Use AppVoIP for voice
	)
	if err != nil {
		return fmt.Errorf("failed to create opus encoder: %w", err)
	}
	
	// Set bitrate
	if r.options.BitrateKbps > 0 {
		if err := r.encoder.SetBitrate(r.options.BitrateKbps * 1000); err != nil {
			return fmt.Errorf("failed to set custom bitrate: %w", err)
		}
	}
	
	// Create output file if specified
	if r.options.OutputPath != "" {
		r.outputFile, err = os.Create(r.options.OutputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
	}
	
	// Get default input device
	defaultInput, err := portaudio.DefaultInputDevice()
	if err != nil {
		return fmt.Errorf("failed to get default input device: %w", err)
	}
	
	// Configure PortAudio stream
	streamParams := portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   defaultInput,
			Channels: r.options.Format.Channels,
		},
		SampleRate:      float64(r.options.Format.SampleRate),
		FramesPerBuffer: 1024, // Small buffer for low latency
	}
	
	// Create audio stream
	r.stream, err = portaudio.OpenStream(streamParams, r.audioCallback)
	if err != nil {
		return fmt.Errorf("failed to open audio stream: %w", err)
	}
	
	// Start the stream
	if err := r.stream.Start(); err != nil {
		r.stream.Close()
		return fmt.Errorf("failed to start audio stream: %w", err)
	}
	
	r.state = RecordingStateRecording
	
	// Start duration monitoring goroutine
	go r.monitorDuration()
	
	return nil
}

// audioCallback is called by PortAudio with audio data
func (r *PortAudioRecorder) audioCallback(in []float32) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.state != RecordingStateRecording {
		return
	}
	
	// Calculate audio level (RMS)
	var sum float64
	for _, sample := range in {
		sum += float64(sample * sample)
	}
	rms := math.Sqrt(sum / float64(len(in)))
	r.currentLevel = float32(rms)
	
	// Apply noise gate if enabled
	if r.options.NoiseGateLevel > -1 && rms < math.Pow(10, float64(r.options.NoiseGateLevel)/20) {
		// Below noise gate - silence the audio
		for i := range in {
			in[i] = 0
		}
	}
	
	// Store audio data
	r.audioData = append(r.audioData, in...)
	
	// Call user callback
	if r.callback != nil {
		r.callback(in, float32(rms))
	}
}

// monitorDuration monitors recording duration and enforces limits
func (r *PortAudioRecorder) monitorDuration() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-r.ctx.Done():
			return
		case <-ticker.C:
			duration := r.GetDuration()
			
			// Check max duration
			if r.options.MaxDuration > 0 && duration >= r.options.MaxDuration {
				r.Stop() // This will handle the cleanup
				return
			}
		}
	}
}

// Stop stops recording and returns the voice message
func (r *PortAudioRecorder) Stop() (*VoiceMessage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.state != RecordingStateRecording && r.state != RecordingStatePaused {
		return nil, fmt.Errorf("not currently recording")
	}
	
	// Cancel monitoring
	if r.cancel != nil {
		r.cancel()
	}
	
	// Stop and close stream
	if r.stream != nil {
		r.stream.Stop()
		r.stream.Close()
		r.stream = nil
	}
	
	// Check minimum duration
	duration := r.GetDuration()
	if duration < r.options.MinDuration {
		r.cleanup()
		return nil, fmt.Errorf("recording too short: %v (minimum: %v)", duration, r.options.MinDuration)
	}
	
	// Create voice message
	voiceMsg := &VoiceMessage{
		Duration:  duration,
		Format:    r.options.Format,
		CreatedAt: r.startTime,
	}
	
	// Save audio data if output path specified
	if r.options.OutputPath != "" {
		if err := r.saveToFile(); err != nil {
			r.cleanup()
			return nil, fmt.Errorf("failed to save audio: %w", err)
		}
		
		voiceMsg.FilePath = r.options.OutputPath
		
		// Get file size
		if stat, err := os.Stat(r.options.OutputPath); err == nil {
			voiceMsg.FileSize = stat.Size()
		}
		
		// Generate simplified waveform
		voiceMsg.Waveform = r.generateSimpleWaveform(100) // 100 points
	}
	
	r.state = RecordingStateStopped
	r.cleanup()
	
	return voiceMsg, nil
}

// saveToFile saves the recorded audio to file
func (r *PortAudioRecorder) saveToFile() error {
	if r.outputFile == nil {
		return fmt.Errorf("no output file")
	}
	defer r.outputFile.Close()
	
	// Convert float32 samples to int16 for Opus encoding
	pcmData := make([]int16, len(r.audioData))
	for i, sample := range r.audioData {
		// Clamp and convert to int16
		if sample > 1.0 {
			sample = 1.0
		} else if sample < -1.0 {
			sample = -1.0
		}
		pcmData[i] = int16(sample * 32767)
	}
	
	// Encode with Opus
	frameSize := 960 // 20ms at 48kHz
	encodedBuffer := make([]byte, 4000) // Buffer for encoded data
	var encodedData []byte
	
	for i := 0; i < len(pcmData); i += frameSize {
		end := i + frameSize
		if end > len(pcmData) {
			// Pad the last frame with zeros
			frame := make([]int16, frameSize)
			copy(frame, pcmData[i:])
			pcmData = append(pcmData[:i], frame...)
			end = i + frameSize
		}
		
		frame := pcmData[i:end]
		bytesWritten, err := r.encoder.Encode(frame, encodedBuffer)
		if err != nil {
			return fmt.Errorf("failed to encode frame: %w", err)
		}
		encodedData = append(encodedData, encodedBuffer[:bytesWritten]...)
	}
	
	// Write encoded data to file
	if _, err := r.outputFile.Write(encodedData); err != nil {
		return fmt.Errorf("failed to write audio data: %w", err)
	}
	
	return nil
}

// generateSimpleWaveform creates a simplified waveform for visualization
func (r *PortAudioRecorder) generateSimpleWaveform(points int) []float32 {
	if len(r.audioData) == 0 || points <= 0 {
		return nil
	}
	
	waveform := make([]float32, points)
	samplesPerPoint := len(r.audioData) / points
	
	if samplesPerPoint == 0 {
		samplesPerPoint = 1
	}
	
	for i := 0; i < points; i++ {
		start := i * samplesPerPoint
		end := start + samplesPerPoint
		if end > len(r.audioData) {
			end = len(r.audioData)
		}
		
		// Calculate RMS for this segment
		var sum float64
		for j := start; j < end; j++ {
			sum += float64(r.audioData[j] * r.audioData[j])
		}
		rms := math.Sqrt(sum / float64(end-start))
		waveform[i] = float32(rms)
	}
	
	return waveform
}

// Pause pauses recording
func (r *PortAudioRecorder) Pause() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.state != RecordingStateRecording {
		return fmt.Errorf("not currently recording")
	}
	
	if r.stream != nil {
		if err := r.stream.Stop(); err != nil {
			return fmt.Errorf("failed to pause stream: %w", err)
		}
	}
	
	r.state = RecordingStatePaused
	return nil
}

// Resume resumes recording
func (r *PortAudioRecorder) Resume() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.state != RecordingStatePaused {
		return fmt.Errorf("not currently paused")
	}
	
	if r.stream != nil {
		if err := r.stream.Start(); err != nil {
			return fmt.Errorf("failed to resume stream: %w", err)
		}
	}
	
	r.state = RecordingStateRecording
	return nil
}

// GetState returns current recording state
func (r *PortAudioRecorder) GetState() RecordingState {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state
}

// GetLevel returns current audio input level
func (r *PortAudioRecorder) GetLevel() float32 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.currentLevel
}

// GetDuration returns current recording duration
func (r *PortAudioRecorder) GetDuration() time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if r.state == RecordingStateIdle || r.state == RecordingStateStopped {
		return 0
	}
	
	elapsed := time.Since(r.startTime) - r.pausedTime
	if r.state == RecordingStatePaused {
		// Don't count time since pause started
		return elapsed
	}
	
	return elapsed
}

// Cancel cancels recording without saving
func (r *PortAudioRecorder) Cancel() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.state == RecordingStateIdle || r.state == RecordingStateStopped {
		return nil
	}
	
	// Cancel monitoring
	if r.cancel != nil {
		r.cancel()
	}
	
	// Stop and close stream
	if r.stream != nil {
		r.stream.Stop()
		r.stream.Close()
		r.stream = nil
	}
	
	r.state = RecordingStateIdle
	r.cleanup()
	
	return nil
}

// cleanup cleans up resources
func (r *PortAudioRecorder) cleanup() {
	if r.outputFile != nil {
		r.outputFile.Close()
		r.outputFile = nil
	}
	
	if r.encoder != nil {
		r.encoder = nil
	}
	
	r.audioData = nil
}

// IsSupported returns true if recording is supported
func (r *PortAudioRecorder) IsSupported() bool {
	// Check if PortAudio is available and has input devices
	device, err := portaudio.DefaultInputDevice()
	if err != nil || device == nil {
		return false
	}
	return true
}
