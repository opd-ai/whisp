package audio

import (
	"context"
	"time"
)

// RecordingState represents the current state of audio recording
type RecordingState int

const (
	RecordingStateIdle RecordingState = iota
	RecordingStateRecording
	RecordingStatePaused
	RecordingStateStopped
)

// PlaybackState represents the current state of audio playback
type PlaybackState int

const (
	PlaybackStateIdle PlaybackState = iota
	PlaybackStatePlaying
	PlaybackStatePaused
	PlaybackStateStopped
)

// AudioFormat represents audio format parameters
type AudioFormat struct {
	SampleRate int    // Hz (e.g., 48000)
	Channels   int    // 1 for mono, 2 for stereo
	BitDepth   int    // bits per sample (e.g., 16)
	Codec      string // "opus", "wav", "mp3"
}

// DefaultVoiceFormat returns the default format for voice messages
func DefaultVoiceFormat() AudioFormat {
	return AudioFormat{
		SampleRate: 48000,  // Opus native sample rate
		Channels:   1,      // Mono for voice
		BitDepth:   16,     // 16-bit samples
		Codec:      "opus", // Efficient compression
	}
}

// RecordingOptions contains options for audio recording
type RecordingOptions struct {
	Format         AudioFormat
	MaxDuration    time.Duration // Maximum recording duration
	MinDuration    time.Duration // Minimum recording duration
	OutputPath     string        // Where to save the recording
	BitrateKbps    int           // Bitrate for compression (kbps)
	NoiseGateLevel float32       // Noise gate threshold (-1 to disable)
}

// DefaultRecordingOptions returns sensible defaults for voice recording
func DefaultRecordingOptions() RecordingOptions {
	return RecordingOptions{
		Format:         DefaultVoiceFormat(),
		MaxDuration:    5 * time.Minute, // 5 minute max
		MinDuration:    time.Second,     // 1 second min
		BitrateKbps:    32,              // 32kbps good for voice
		NoiseGateLevel: -30.0,           // -30dB noise gate
	}
}

// PlaybackOptions contains options for audio playback
type PlaybackOptions struct {
	Volume      float32 // 0.0 to 1.0
	Speed       float32 // 0.5 to 2.0 (playback speed)
	StartOffset time.Duration
	AutoStop    bool // Stop at end of file
}

// DefaultPlaybackOptions returns sensible defaults for voice playback
func DefaultPlaybackOptions() PlaybackOptions {
	return PlaybackOptions{
		Volume:   1.0,
		Speed:    1.0,
		AutoStop: true,
	}
}

// AudioData represents raw audio data with metadata
type AudioData struct {
	Samples    []float32
	Format     AudioFormat
	Duration   time.Duration
	SampleRate int
	Channels   int
}

// VoiceMessage represents a voice message with metadata
type VoiceMessage struct {
	ID         string
	FilePath   string
	Duration   time.Duration
	Format     AudioFormat
	FileSize   int64
	CreatedAt  time.Time
	Waveform   []float32 // Simplified waveform for visualization
	Transcript string    // Optional transcript
}

// RecordingCallback is called during recording with audio data and level
type RecordingCallback func(data []float32, level float32)

// PlaybackCallback is called during playback with progress
type PlaybackCallback func(position time.Duration, duration time.Duration)

// Recorder interface for audio recording
type Recorder interface {
	// Start begins recording with the given options
	Start(ctx context.Context, options RecordingOptions, callback RecordingCallback) error

	// Stop stops recording and returns the final audio data
	Stop() (*VoiceMessage, error)

	// Pause temporarily stops recording (can be resumed)
	Pause() error

	// Resume continues recording after pause
	Resume() error

	// GetState returns current recording state
	GetState() RecordingState

	// GetLevel returns current audio input level (0.0 to 1.0)
	GetLevel() float32

	// GetDuration returns current recording duration
	GetDuration() time.Duration

	// Cancel cancels recording without saving
	Cancel() error

	// IsSupported returns true if recording is supported on this platform
	IsSupported() bool
}

// Player interface for audio playback
type Player interface {
	// Load loads an audio file for playback
	Load(filePath string) error

	// Play starts playback with options and callback
	Play(options PlaybackOptions, callback PlaybackCallback) error

	// Pause pauses playback
	Pause() error

	// Resume resumes playback
	Resume() error

	// Stop stops playback
	Stop() error

	// Seek seeks to a specific position
	Seek(position time.Duration) error

	// GetState returns current playback state
	GetState() PlaybackState

	// GetPosition returns current playback position
	GetPosition() time.Duration

	// GetDuration returns total audio duration
	GetDuration() time.Duration

	// SetVolume sets playback volume (0.0 to 1.0)
	SetVolume(volume float32) error

	// GetVolume returns current volume
	GetVolume() float32

	// IsSupported returns true if playback is supported on this platform
	IsSupported() bool
}

// WaveformGenerator generates waveform data for visualization
type WaveformGenerator interface {
	// GenerateWaveform generates simplified waveform data from audio
	GenerateWaveform(audioData *AudioData, points int) ([]float32, error)

	// GenerateWaveformFromFile generates waveform from audio file
	GenerateWaveformFromFile(filePath string, points int) ([]float32, error)
}

// Manager manages audio recording and playback
type Manager interface {
	// GetRecorder returns a new recorder instance
	GetRecorder() (Recorder, error)

	// GetPlayer returns a new player instance
	GetPlayer() (Player, error)

	// GetWaveformGenerator returns waveform generator
	GetWaveformGenerator() WaveformGenerator

	// GetSupportedFormats returns list of supported audio formats
	GetSupportedFormats() []AudioFormat

	// Initialize initializes the audio system
	Initialize() error

	// Shutdown shuts down the audio system
	Shutdown() error

	// IsInitialized returns true if audio system is initialized
	IsInitialized() bool
}
