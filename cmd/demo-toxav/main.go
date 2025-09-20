// Demo of ToxAV calling functionality using the new call manager
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/opd-ai/whisp/internal/core/calls"
	"github.com/opd-ai/whisp/internal/core/tox"
)

// EventHandler implements the CallEventHandler interface
type EventHandler struct{}

// OnCallEvent handles call events
func (h *EventHandler) OnCallEvent(event *calls.CallEvent) {
	fmt.Printf("[%s] %s: %s\n",
		event.Timestamp.Format("15:04:05"),
		event.Type,
		event.Message)

	if event.Error != nil {
		fmt.Printf("  Error: %v\n", event.Error)
	}

	// Handle specific event types
	switch event.Type {
	case calls.CallEventIncoming:
		fmt.Printf("  📞 Incoming %s call from friend %d\n",
			event.Call.Type, event.Call.FriendID)
		fmt.Println("  Use call manager to answer or reject the call")

	case calls.CallEventOutgoing:
		fmt.Printf("  📞 Outgoing %s call to friend %d\n",
			event.Call.Type, event.Call.FriendID)

	case calls.CallEventStateChanged:
		fmt.Printf("  📞 Call state changed to %s (duration: %v)\n",
			event.Call.State, event.Call.Duration())

	case calls.CallEventEnded:
		fmt.Printf("  📞 Call ended after %v\n", event.Call.Duration())

	case calls.CallEventAudioFrame:
		if event.AudioFrame != nil {
			fmt.Printf("  🎵 Audio frame: %d samples, %d channels, %d Hz\n",
				event.AudioFrame.SampleCount,
				event.AudioFrame.Channels,
				event.AudioFrame.SamplingRate)
		}

	case calls.CallEventVideoFrame:
		if event.VideoFrame != nil {
			fmt.Printf("  📹 Video frame: %dx%d\n",
				event.VideoFrame.Width,
				event.VideoFrame.Height)
		}

	case calls.CallEventBitrateChanged:
		fmt.Printf("  📊 Bitrate changed for call %s\n", event.Call.ID)

	case calls.CallEventError:
		fmt.Printf("  ❌ Call error: %v\n", event.Error)
	}
}

func main() {
	fmt.Println("ToxAV Call Manager Demo")
	fmt.Println("======================")

	// Create a basic Tox instance configuration
	toxConfig := &tox.Config{
		DataDir: "./demo_data",
		Debug:   true,
	}

	// Initialize Tox instance
	fmt.Println("Initializing Tox instance...")
	toxInstance, err := tox.NewManager(toxConfig)
	if err != nil {
		log.Fatalf("Failed to create Tox instance: %v", err)
	}
	defer toxInstance.Cleanup()

	// Start the Tox manager
	if err := toxInstance.Start(); err != nil {
		log.Fatalf("Failed to start Tox manager: %v", err)
	}
	defer toxInstance.Stop()

	// Get the underlying Tox object for ToxAV integration
	toxCore := toxInstance.GetInstance()
	if toxCore == nil {
		log.Fatal("Failed to get Tox instance for ToxAV")
	}

	// Create call manager with event handler
	eventHandler := &EventHandler{}
	callConfig := calls.DefaultConfig()

	fmt.Println("Creating call manager...")
	callManager, err := calls.NewManager(toxCore, callConfig, eventHandler)
	if err != nil {
		log.Fatalf("Failed to create call manager: %v", err)
	}

	// Start the call manager
	fmt.Println("Starting call manager...")
	if err := callManager.Start(); err != nil {
		log.Fatalf("Failed to start call manager: %v", err)
	}
	defer callManager.Stop()

	// Print Tox ID for connection
	toxID := toxInstance.GetToxID()
	fmt.Printf("\nTox ID: %s\n", toxID)
	fmt.Println("Share this ID to receive calls")

	// Print current status
	fmt.Printf("Call manager running: %t\n", callManager.IsRunning())
	fmt.Printf("Audio config: %d kbps, %d Hz, %d channels\n",
		callConfig.AudioBitRate, callConfig.AudioSampleRate, callConfig.AudioChannels)
	fmt.Printf("Video config: %d kbps, %dx%d, %d FPS\n",
		callConfig.VideoBitRate, callConfig.VideoWidth, callConfig.VideoHeight, callConfig.VideoFPS)

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Demo loop
	fmt.Println("\nDemo running - Press Ctrl+C to exit")
	fmt.Println("Waiting for calls or use external client to place calls...")

	// Simulate some activity
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Context cancelled, shutting down...")
			return

		case sig := <-sigChan:
			fmt.Printf("\nReceived signal %v, shutting down gracefully...\n", sig)
			cancel()

		case <-ticker.C:
			// Print periodic status
			activeCalls := callManager.GetActiveCalls()
			fmt.Printf("[%s] Status: %d active calls\n",
				time.Now().Format("15:04:05"), len(activeCalls))

			// Print active call details
			for _, call := range activeCalls {
				fmt.Printf("  - %s\n", call.String())
			}

			// Example: Show call history count
			history := callManager.GetCallHistory()
			if len(history) > 0 {
				fmt.Printf("  Call history: %d completed calls\n", len(history))
			}

		case <-time.After(100 * time.Millisecond):
			// Small delay to prevent busy waiting
		}
	}
}
