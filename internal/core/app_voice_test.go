package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/opd-ai/whisp/internal/core/audio"
	"github.com/opd-ai/whisp/ui/adaptive"
)

func TestVoiceMessageIntegration(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "whisp_voice_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test app configuration
	config := &Config{
		DataDir:    tempDir,
		ConfigPath: filepath.Join(tempDir, "config.yaml"),
		Debug:      true,
		Platform:   adaptive.PlatformLinux,
	}

	// Create and start app
	app, err := NewApp(config)
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}
	defer app.Cleanup()

	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		t.Fatalf("Failed to start app: %v", err)
	}
	defer app.Stop()

	// Wait for app to be ready
	time.Sleep(100 * time.Millisecond)

	t.Run("AudioManagerInitialization", func(t *testing.T) {
		audioMgr := app.GetAudioManager()
		if audioMgr == nil {
			t.Fatal("Audio manager should not be nil")
		}

		if !audioMgr.IsInitialized() {
			t.Fatal("Audio manager should be initialized")
		}

		formats := audioMgr.GetSupportedFormats()
		if len(formats) == 0 {
			t.Fatal("Should have supported audio formats")
		}

		t.Logf("Supported formats: %d", len(formats))
		for i, format := range formats {
			t.Logf("  %d. %s - %dHz, %d channels, %d-bit",
				i+1, format.Codec, format.SampleRate, format.Channels, format.BitDepth)
		}
	})

	t.Run("VoiceRecording", func(t *testing.T) {
		voiceDir := filepath.Join(tempDir, "voice_test")
		if err := os.MkdirAll(voiceDir, 0o755); err != nil {
			t.Fatalf("Failed to create voice dir: %v", err)
		}

		// Start recording
		recorder, err := app.StartVoiceRecordingFromUI(123, voiceDir)
		if err != nil {
			t.Fatalf("Failed to start recording: %v", err)
		}

		// Check initial state
		if recorder.GetState() != audio.RecordingStateRecording {
			t.Errorf("Expected recording state, got %v", recorder.GetState())
		}

		// Record for a short time
		time.Sleep(500 * time.Millisecond)

		// Check duration
		duration := recorder.GetDuration()
		if duration < 400*time.Millisecond {
			t.Errorf("Expected at least 400ms recording, got %v", duration)
		}

		// Check level
		level := recorder.GetLevel()
		if level < 0 {
			t.Errorf("Audio level should be non-negative, got %f", level)
		}

		// Stop recording
		voiceMsg, err := recorder.Stop()
		if err != nil {
			t.Fatalf("Failed to stop recording: %v", err)
		}

		// Validate voice message
		if voiceMsg.Duration < 400*time.Millisecond {
			t.Errorf("Voice message duration too short: %v", voiceMsg.Duration)
		}

		if voiceMsg.FileSize <= 0 {
			t.Errorf("Voice message should have positive file size, got %d", voiceMsg.FileSize)
		}

		if len(voiceMsg.Waveform) == 0 {
			t.Error("Voice message should have waveform data")
		}

		t.Logf("Voice message: duration=%v, size=%d, waveform_points=%d",
			voiceMsg.Duration, voiceMsg.FileSize, len(voiceMsg.Waveform))
	})

	t.Run("VoiceRecordingStates", func(t *testing.T) {
		audioMgr := app.GetAudioManager()
		recorder, err := audioMgr.GetRecorder()
		if err != nil {
			t.Fatalf("Failed to get recorder: %v", err)
		}

		// Test initial state
		if recorder.GetState() != audio.RecordingStateIdle {
			t.Errorf("Initial state should be idle, got %v", recorder.GetState())
		}

		// Start recording
		ctx := context.Background()
		options := audio.DefaultRecordingOptions()
		options.MaxDuration = 2 * time.Second

		err = recorder.Start(ctx, options, nil)
		if err != nil {
			t.Fatalf("Failed to start recording: %v", err)
		}

		if recorder.GetState() != audio.RecordingStateRecording {
			t.Errorf("Should be recording, got %v", recorder.GetState())
		}

		// Test pause
		err = recorder.Pause()
		if err != nil {
			t.Fatalf("Failed to pause recording: %v", err)
		}

		if recorder.GetState() != audio.RecordingStatePaused {
			t.Errorf("Should be paused, got %v", recorder.GetState())
		}

		// Test resume
		err = recorder.Resume()
		if err != nil {
			t.Fatalf("Failed to resume recording: %v", err)
		}

		if recorder.GetState() != audio.RecordingStateRecording {
			t.Errorf("Should be recording after resume, got %v", recorder.GetState())
		}

		// Cancel recording
		err = recorder.Cancel()
		if err != nil {
			t.Fatalf("Failed to cancel recording: %v", err)
		}

		if recorder.GetState() != audio.RecordingStateIdle {
			t.Errorf("Should be idle after cancel, got %v", recorder.GetState())
		}
	})

	t.Run("VoicePlayback", func(t *testing.T) {
		audioMgr := app.GetAudioManager()
		player, err := audioMgr.GetPlayer()
		if err != nil {
			t.Fatalf("Failed to get player: %v", err)
		}

		// Test supported check
		if !player.IsSupported() {
			t.Skip("Voice playback not supported in test environment")
		}

		// Load mock file (this will work with our mock implementation)
		mockFile := filepath.Join(tempDir, "test.wav")
		err = player.Load(mockFile)
		if err != nil {
			t.Fatalf("Failed to load mock file: %v", err)
		}

		// Test playback
		options := audio.DefaultPlaybackOptions()
		err = player.Play(options, nil)
		if err != nil {
			t.Fatalf("Failed to start playback: %v", err)
		}

		if player.GetState() != audio.PlaybackStatePlaying {
			t.Errorf("Should be playing, got %v", player.GetState())
		}

		// Test volume
		initialVolume := player.GetVolume()
		if initialVolume != 1.0 {
			t.Errorf("Expected initial volume 1.0, got %f", initialVolume)
		}

		err = player.SetVolume(0.5)
		if err != nil {
			t.Fatalf("Failed to set volume: %v", err)
		}

		if player.GetVolume() != 0.5 {
			t.Errorf("Expected volume 0.5, got %f", player.GetVolume())
		}

		// Test pause
		err = player.Pause()
		if err != nil {
			t.Fatalf("Failed to pause playback: %v", err)
		}

		if player.GetState() != audio.PlaybackStatePaused {
			t.Errorf("Should be paused, got %v", player.GetState())
		}

		// Test resume
		err = player.Resume()
		if err != nil {
			t.Fatalf("Failed to resume playback: %v", err)
		}

		if player.GetState() != audio.PlaybackStatePlaying {
			t.Errorf("Should be playing after resume, got %v", player.GetState())
		}

		// Stop playback
		err = player.Stop()
		if err != nil {
			t.Fatalf("Failed to stop playback: %v", err)
		}

		if player.GetState() != audio.PlaybackStateStopped {
			t.Errorf("Should be stopped, got %v", player.GetState())
		}
	})

	t.Run("WaveformGeneration", func(t *testing.T) {
		mockFile := filepath.Join(tempDir, "test_waveform.wav")

		waveform, err := app.GenerateWaveformFromUI(mockFile, 50)
		if err != nil {
			t.Fatalf("Failed to generate waveform: %v", err)
		}

		if len(waveform) != 50 {
			t.Errorf("Expected 50 waveform points, got %d", len(waveform))
		}

		// Check that waveform values are reasonable
		for i, val := range waveform {
			if val < 0 || val > 1 {
				t.Errorf("Waveform value %d out of range [0,1]: %f", i, val)
			}
		}

		t.Logf("Generated waveform with %d points", len(waveform))
	})
}

func TestVoiceMessageUIIntegration(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "whisp_voice_ui_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test app
	config := &Config{
		DataDir:    tempDir,
		ConfigPath: filepath.Join(tempDir, "config.yaml"),
		Debug:      true,
		Platform:   adaptive.PlatformLinux,
	}

	app, err := NewApp(config)
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}
	defer app.Cleanup()

	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		t.Fatalf("Failed to start app: %v", err)
	}
	defer app.Stop()

	time.Sleep(100 * time.Millisecond)

	t.Run("CompleteVoiceMessageWorkflow", func(t *testing.T) {
		friendID := uint32(999) // Non-existent friend for testing
		voiceDir := filepath.Join(tempDir, "voice_workflow")

		// 1. Start recording
		recorder, err := app.StartVoiceRecordingFromUI(friendID, voiceDir)
		if err != nil {
			t.Fatalf("Failed to start recording: %v", err)
		}

		// 2. Record for a moment
		time.Sleep(300 * time.Millisecond)

		// 3. Stop and get voice message
		voiceMsg, err := recorder.Stop()
		if err != nil {
			t.Fatalf("Failed to stop recording: %v", err)
		}

		// 4. Verify voice message structure
		if voiceMsg.Duration < 200*time.Millisecond {
			t.Errorf("Voice message too short: %v", voiceMsg.Duration)
		}

		if voiceMsg.FileSize <= 0 {
			t.Errorf("Voice message should have file size > 0, got %d", voiceMsg.FileSize)
		}

		if len(voiceMsg.Waveform) == 0 {
			t.Error("Voice message should have waveform data")
		}

		// 5. Test sending (will fail due to no friend, but validates the pipeline)
		err = app.SendVoiceMessageFromUI(friendID, voiceMsg)
		if err == nil {
			t.Error("Expected error when sending to non-existent friend")
		}

		t.Logf("Complete workflow tested: duration=%v, size=%d bytes",
			voiceMsg.Duration, voiceMsg.FileSize)
	})
}
