package common

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// InputValidator provides input validation utilities
type InputValidator struct{}

// NewInputValidator creates a new input validator
func NewInputValidator() *InputValidator {
	return &InputValidator{}
}

// ValidateToxID validates a Tox ID format
func (v *InputValidator) ValidateToxID(toxID string) error {
	if toxID == "" {
		return fmt.Errorf("Tox ID cannot be empty")
	}

	// Tox IDs are 76 characters long (64 bytes public key + 12 bytes nospam in hex)
	if len(toxID) != 76 {
		return fmt.Errorf("Tox ID must be exactly 76 characters long, got %d", len(toxID))
	}

	// Must be valid hexadecimal
	if matched, _ := regexp.MatchString("^[a-fA-F0-9]{76}$", toxID); !matched {
		return fmt.Errorf("Tox ID must contain only hexadecimal characters")
	}

	return nil
}

// ValidateMessageContent validates message content
func (v *InputValidator) ValidateMessageContent(content string) error {
	if content == "" {
		return fmt.Errorf("message content cannot be empty")
	}

	// Check UTF-8 validity
	if !utf8.ValidString(content) {
		return fmt.Errorf("message content contains invalid UTF-8 characters")
	}

	// Check maximum length (reasonable limit to prevent abuse)
	const maxMessageLength = 65536 // 64KB
	if len(content) > maxMessageLength {
		return fmt.Errorf("message content too long: %d characters (max %d)", len(content), maxMessageLength)
	}

	// Check for null bytes (potential attack vector)
	if strings.Contains(content, "\x00") {
		return fmt.Errorf("message content contains null bytes")
	}

	return nil
}

// ValidateFileName validates a filename for security
func (v *InputValidator) ValidateFileName(filename string) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// Check for path traversal attempts
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return fmt.Errorf("filename contains path traversal sequences")
	}

	// Check for dangerous characters
	dangerousChars := "<>:\"|?*\x00"
	if strings.ContainsAny(filename, dangerousChars) {
		return fmt.Errorf("filename contains dangerous characters")
	}

	// Check for control characters
	for _, r := range filename {
		if r < 32 {
			return fmt.Errorf("filename contains control characters")
		}
	}

	// Check reasonable length
	const maxFileNameLength = 255
	if len(filename) > maxFileNameLength {
		return fmt.Errorf("filename too long: %d characters (max %d)", len(filename), maxFileNameLength)
	}

	return nil
}

// ValidateFileSize validates file size limits
func (v *InputValidator) ValidateFileSize(size int64) error {
	const maxFileSize = 2 * 1024 * 1024 * 1024 // 2GB
	if size < 0 {
		return fmt.Errorf("file size cannot be negative")
	}
	if size > maxFileSize {
		return fmt.Errorf("file size too large: %d bytes (max %d)", size, maxFileSize)
	}
	return nil
}

// SanitizeString removes potentially dangerous characters from strings
func (v *InputValidator) SanitizeString(input string) string {
	// Remove null bytes
	sanitized := strings.ReplaceAll(input, "\x00", "")

	// Remove control characters except common whitespace
	var result strings.Builder
	for _, r := range sanitized {
		if r >= 32 || r == 9 || r == 10 || r == 13 { // Allow tab, LF, CR
			result.WriteRune(r)
		}
	}

	return result.String()
}
