package main

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

// TestBuildVersionFlags tests that build flags are properly set
func TestBuildVersionFlags(t *testing.T) {
	// These variables should be set by ldflags during build
	// We test that they exist and have reasonable defaults

	tests := []struct {
		name     string
		variable string
		expected string
	}{
		{"version", version, "dev"},         // Should have default when not set by ldflags
		{"buildTime", buildTime, "unknown"}, // Should have default
		{"gitCommit", gitCommit, "unknown"}, // Should have default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.variable != tt.expected && tt.expected != "" {
				// Allow for actual values to be set by ldflags, but verify defaults exist
				if tt.variable == "" {
					t.Errorf("Expected %s to have default value %s, got empty string", tt.name, tt.expected)
				}
			}
		})
	}
}

// TestVersionCommand tests the version flag functionality
func TestVersionCommand(t *testing.T) {
	// Test that version information contains expected components

	if version == "" {
		t.Error("Version should not be empty")
	}

	if buildTime == "" {
		t.Error("Build time should not be empty")
	}

	if gitCommit == "" {
		t.Error("Git commit should not be empty")
	}
}

// TestPlatformInformation tests platform detection availability
func TestPlatformInformation(t *testing.T) {
	// Test that runtime platform information is accessible
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	if goos == "" {
		t.Error("GOOS should not be empty")
	}

	if goarch == "" {
		t.Error("GOARCH should not be empty")
	}

	// Test common platforms
	validPlatforms := []string{"windows", "darwin", "linux", "android", "ios"}
	found := false
	for _, platform := range validPlatforms {
		if goos == platform {
			found = true
			break
		}
	}

	if !found {
		t.Logf("Warning: Unknown platform %s, this may need additional support", goos)
	}
}

// TestApplicationInitialization tests basic application setup
func TestApplicationInitialization(t *testing.T) {
	t.Run("version_info", func(t *testing.T) {
		// Test version information format
		if !strings.Contains(version, "dev") && !strings.Contains(version, ".") {
			t.Logf("Version format: %s (this may be a custom format)", version)
		}
	})

	t.Run("build_flags", func(t *testing.T) {
		// Test that build flags can be read
		originalVersion := version
		originalBuildTime := buildTime
		originalGitCommit := gitCommit

		// Verify they can be accessed
		if len(originalVersion) == 0 {
			t.Error("Version should be accessible")
		}
		if len(originalBuildTime) == 0 {
			t.Error("Build time should be accessible")
		}
		if len(originalGitCommit) == 0 {
			t.Error("Git commit should be accessible")
		}
	})

	t.Run("environment", func(t *testing.T) {
		// Test that we can access environment variables that main() uses
		tempDir := os.TempDir()
		if tempDir == "" {
			t.Error("Should be able to access temp directory")
		}

		// Test that we can create directories (functionality used in main)
		testDir := os.TempDir() + "/whisp-test"
		err := os.MkdirAll(testDir, 0700)
		if err != nil {
			t.Errorf("Should be able to create directories: %v", err)
		}
		defer os.RemoveAll(testDir)
	})
}
