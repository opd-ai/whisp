package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/opd-ai/whisp/internal/core"
	"github.com/opd-ai/whisp/platform/common"
)

func main() {
	fmt.Println("=== Whisp Voice Message Demo ===")

	// Get application data directory
	dataDir, err := common.GetUserDataDir()
	if err != nil {
		log.Fatalf("Failed to get data directory: %v", err)
	}

	// Create demo data directory
	demoDir := filepath.Join(dataDir, "demo-voice")
	if err := os.MkdirAll(demoDir, 0o755); err != nil {
		log.Fatalf("Failed to create demo directory: %v", err)
	}

	fmt.Printf("Demo directory: %s\n", demoDir)

	// Create app configuration
	config := &core.Config{
		DataDir:    demoDir,
		ConfigPath: filepath.Join(demoDir, "config.yaml"),
		Debug:      true,
	}

	// Initialize core application
	app, err := core.NewApp(config)
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}
	defer app.Cleanup()

	// Start the application
	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		log.Fatalf("Failed to start app: %v", err)
	}

	// Wait for app to be ready
	time.Sleep(100 * time.Millisecond)

	// Demo voice message functionality
	fmt.Println("\n=== Testing Voice Message Features ===")

	// 1. Test audio manager availability
	audioMgr := app.GetAudioManager()
	if audioMgr == nil {
		log.Fatal("Audio manager not available")
	}

	if !audioMgr.IsInitialized() {
		log.Fatal("Audio manager not initialized")
	}

	fmt.Println("✓ Audio manager initialized")

	// 2. Test supported formats
	formats := audioMgr.GetSupportedFormats()
	fmt.Printf("✓ Supported audio formats: %d\n", len(formats))
	for i, format := range formats {
		fmt.Printf("  %d. %s - %dHz, %d channels, %d-bit\n",
			i+1, format.Codec, format.SampleRate, format.Channels, format.BitDepth)
	}

	// 3. Test voice recording
	fmt.Println("\n=== Testing Voice Recording ===")

	friendID := uint32(123) // Mock friend ID
	voiceDir := filepath.Join(demoDir, "voice_messages")
	if err := os.MkdirAll(voiceDir, 0o755); err != nil {
		log.Fatalf("Failed to create voice directory: %v", err)
	}

	recorder, err := app.StartVoiceRecordingFromUI(friendID, voiceDir)
	if err != nil {
		log.Fatalf("Failed to start voice recording: %v", err)
	}

	fmt.Printf("✓ Started recording, state: %v\n", recorder.GetState())

	// Record for a short time
	fmt.Println("Recording for 2 seconds...")
	for i := 0; i < 20; i++ {
		time.Sleep(100 * time.Millisecond)
		level := recorder.GetLevel()
		duration := recorder.GetDuration()
		fmt.Printf("\rRecording... Level: %.3f, Duration: %v", level, duration)
	}
	fmt.Println()

	// Stop recording
	voiceMsg, err := recorder.Stop()
	if err != nil {
		log.Fatalf("Failed to stop recording: %v", err)
	}

	fmt.Printf("✓ Recording completed!\n")
	fmt.Printf("  Duration: %v\n", voiceMsg.Duration)
	fmt.Printf("  File size: %d bytes\n", voiceMsg.FileSize)
	fmt.Printf("  Format: %s - %dHz, %d channels\n",
		voiceMsg.Format.Codec, voiceMsg.Format.SampleRate, voiceMsg.Format.Channels)

	if len(voiceMsg.Waveform) > 0 {
		fmt.Printf("  Waveform: %d points\n", len(voiceMsg.Waveform))
	}

	// 4. Test voice message sending
	fmt.Println("\n=== Testing Voice Message Sending ===")

	err = app.SendVoiceMessageFromUI(friendID, voiceMsg)
	if err != nil {
		log.Fatalf("Failed to send voice message: %v", err)
	}

	fmt.Printf("✓ Voice message sent to friend %d\n", friendID)

	// 5. Test waveform generation
	fmt.Println("\n=== Testing Waveform Generation ===")

	// Mock file path for waveform generation
	mockFile := filepath.Join(voiceDir, "test_audio.wav")
	waveform, err := app.GenerateWaveformFromUI(mockFile, 50)
	if err != nil {
		log.Printf("Waveform generation failed (expected for mock): %v", err)
	} else {
		fmt.Printf("✓ Generated waveform with %d points\n", len(waveform))

		// Display a simple ASCII waveform
		fmt.Print("  Waveform: ")
		for i, val := range waveform[:10] { // Show first 10 points
			height := int(val * 10)
			if height > 9 {
				height = 9
			}
			fmt.Printf("%d", height)
			if i < 9 {
				fmt.Print("-")
			}
		}
		fmt.Println("...")
	}

	// 6. Test audio playback
	fmt.Println("\n=== Testing Voice Playback ===")

	player, err := app.PlayVoiceMessageFromUI(mockFile)
	if err != nil {
		log.Printf("Voice playback failed (expected for mock): %v", err)
	} else {
		fmt.Printf("✓ Started playback, state: %v\n", player.GetState())
		fmt.Printf("  Duration: %v\n", player.GetDuration())
		fmt.Printf("  Volume: %.1f\n", player.GetVolume())

		// Simulate playback for a moment
		time.Sleep(500 * time.Millisecond)
		player.Stop()
		fmt.Println("✓ Playback stopped")
	}

	// 7. Test configuration integration
	fmt.Println("\n=== Testing Configuration Integration ===")

	// Show how voice settings could be configured
	settings := map[string]interface{}{
		"voice_max_duration":    "5m",
		"voice_bitrate":         32,
		"voice_noise_gate":      -30,
		"voice_format":          "wav",
		"voice_sample_rate":     48000,
		"voice_auto_send":       false,
		"voice_waveform_points": 100,
	}

	fmt.Println("Voice message configuration options:")
	for key, value := range settings {
		fmt.Printf("  %s: %v\n", key, value)
	}

	fmt.Println("\n=== Demo Completed Successfully! ===")
	fmt.Println("\nVoice message features implemented:")
	fmt.Println("  ✓ Audio recording with mock recorder")
	fmt.Println("  ✓ Voice message creation and storage")
	fmt.Println("  ✓ Audio playback with mock player")
	fmt.Println("  ✓ Waveform generation for visualization")
	fmt.Println("  ✓ Integration with existing transfer system")
	fmt.Println("  ✓ Core app integration with cleanup")
	fmt.Println("  ✓ Proper error handling and state management")

	fmt.Println("\nTo implement real audio:")
	fmt.Println("  1. Replace MockRecorder with PortAudio implementation")
	fmt.Println("  2. Replace MockPlayer with audio playback library")
	fmt.Println("  3. Add Opus encoding for compression")
	fmt.Println("  4. Implement real waveform analysis")
	fmt.Println("  5. Add UI components for recording controls")

	// Stop the application
	app.Stop()
}
