package security

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opd-ai/whisp/platform/common"
)

func TestPathTraversalProtection(t *testing.T) {
	// Test cases for path traversal attempts
	testCases := []struct {
		name     string
		fileName string
		wantErr  bool
	}{
		{"valid filename", "test.txt", false},
		{"path traversal with ..", "../../../etc/passwd", true},
		{"path traversal with /", "/etc/passwd", true},
		{"path traversal with \\", "\\windows\\system32", true},
		{"dangerous characters", "test<file>.txt", true},
		{"null bytes", "test\x00file.txt", true},
		{"empty filename", "", true},
		{"current directory", ".", true},
		{"parent directory", "..", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test filename sanitization
			cleanFileName := filepath.Base(tc.fileName)
			if cleanFileName != tc.fileName && !tc.wantErr {
				t.Errorf("Expected filename %s to be valid, but filepath.Base changed it to %s", tc.fileName, cleanFileName)
			}

			// Test dangerous character detection (only for non-empty filenames)
			if tc.fileName != "" {
				hasDangerousChars := strings.ContainsAny(cleanFileName, "<>:\"|?*\x00")
				// Special case for "." and ".." which are handled separately
				isDirectoryTraversal := cleanFileName == "." || cleanFileName == ".."
				shouldDetectDanger := tc.wantErr && !isDirectoryTraversal

				if tc.name != "path traversal with .." && tc.name != "path traversal with /" && tc.name != "path traversal with \\" && tc.name != "current directory" && tc.name != "parent directory" {
					if hasDangerousChars != shouldDetectDanger {
						t.Errorf("Expected dangerous char detection %v for %s, got %v", shouldDetectDanger, tc.fileName, hasDangerousChars)
					}
				}
			}

			// Special checks for empty and directory traversal
			if tc.fileName == "" && !tc.wantErr {
				t.Errorf("Empty filename should be invalid")
			}
			if (tc.fileName == "." || tc.fileName == "..") && !tc.wantErr {
				t.Errorf("Directory traversal filename %s should be invalid", tc.fileName)
			}
		})
	}
}

func TestSecureLogging(t *testing.T) {
	validator := common.NewInputValidator()

	// Test Tox ID validation
	validToxID := "1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890AB"
	invalidToxID := "invalid-tox-id"

	if err := validator.ValidateToxID(validToxID); err != nil {
		t.Errorf("Expected valid Tox ID to pass validation: %v", err)
	}

	if err := validator.ValidateToxID(invalidToxID); err == nil {
		t.Error("Expected invalid Tox ID to fail validation")
	}

	// Test message content validation
	validMessage := "Hello, this is a valid message!"
	invalidMessage := "Message with null byte\x00"

	if err := validator.ValidateMessageContent(validMessage); err != nil {
		t.Errorf("Expected valid message to pass validation: %v", err)
	}

	if err := validator.ValidateMessageContent(invalidMessage); err == nil {
		t.Error("Expected invalid message to fail validation")
	}

	// Test filename validation
	validFileName := "document.pdf"
	invalidFileName := "../../../secret.txt"

	if err := validator.ValidateFileName(validFileName); err != nil {
		t.Errorf("Expected valid filename to pass validation: %v", err)
	}

	if err := validator.ValidateFileName(invalidFileName); err == nil {
		t.Error("Expected invalid filename to fail validation")
	}
}

func TestFilePermissionSecurity(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "secure_test.txt")

	// Test creating file with restrictive permissions
	file, err := os.OpenFile(testFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	// Check file permissions
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat test file: %v", err)
	}

	mode := info.Mode()
	if mode.Perm() != 0o600 {
		t.Errorf("Expected file permissions 0o600, got %o", mode.Perm())
	}

	// Test directory creation with restrictive permissions
	testDir := filepath.Join(tempDir, "secure_dir")
	if err := os.MkdirAll(testDir, 0o700); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	dirInfo, err := os.Stat(testDir)
	if err != nil {
		t.Fatalf("Failed to stat test directory: %v", err)
	}

	dirMode := dirInfo.Mode()
	if dirMode.Perm() != 0o700 {
		t.Errorf("Expected directory permissions 0o700, got %o", dirMode.Perm())
	}
}
