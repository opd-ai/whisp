package audio

import (
	"fmt"
	"math"
)

// MockWaveformGenerator implements WaveformGenerator for demonstration
type MockWaveformGenerator struct{}

// NewMockWaveformGenerator creates a new mock waveform generator
func NewMockWaveformGenerator() *MockWaveformGenerator {
	return &MockWaveformGenerator{}
}

// GenerateWaveform generates simplified waveform data from audio
func (w *MockWaveformGenerator) GenerateWaveform(audioData *AudioData, points int) ([]float32, error) {
	if audioData == nil || len(audioData.Samples) == 0 || points <= 0 {
		return nil, fmt.Errorf("invalid audio data or points")
	}
	
	waveform := make([]float32, points)
	samplesPerPoint := len(audioData.Samples) / points
	
	if samplesPerPoint == 0 {
		samplesPerPoint = 1
	}
	
	for i := 0; i < points; i++ {
		start := i * samplesPerPoint
		end := start + samplesPerPoint
		if end > len(audioData.Samples) {
			end = len(audioData.Samples)
		}
		
		// Calculate RMS for this segment
		var sum float64
		for j := start; j < end; j++ {
			sample := float64(audioData.Samples[j])
			sum += sample * sample
		}
		rms := math.Sqrt(sum / float64(end-start))
		waveform[i] = float32(rms)
	}
	
	return waveform, nil
}

// GenerateWaveformFromFile generates waveform from audio file
func (w *MockWaveformGenerator) GenerateWaveformFromFile(filePath string, points int) ([]float32, error) {
	// For demo purposes, generate a synthetic waveform
	// In a real implementation, this would read and parse the audio file
	
	if points <= 0 {
		return nil, fmt.Errorf("invalid points count")
	}
	
	waveform := make([]float32, points)
	
	// Generate a synthetic waveform that looks realistic
	for i := 0; i < points; i++ {
		t := float64(i) / float64(points)
		
		// Create a waveform with peaks and valleys
		amplitude := 0.3 + 0.7*math.Sin(t*math.Pi*4) // Base wave
		amplitude *= (1 - t*0.5) // Fade out over time
		amplitude += 0.1*math.Sin(t*math.Pi*20) // Add some variation
		
		// Add some randomness
		amplitude += (math.Sin(t*123.456) * 0.1)
		
		// Ensure positive and reasonable bounds
		if amplitude < 0 {
			amplitude = 0
		}
		if amplitude > 1 {
			amplitude = 1
		}
		
		waveform[i] = float32(amplitude)
	}
	
	return waveform, nil
}
