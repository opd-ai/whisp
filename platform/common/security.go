package common

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// SanitizeForLogging removes or masks sensitive information from log messages
func SanitizeForLogging(message string, sensitiveFields ...string) string {
	sanitized := message

	// Remove or mask common sensitive patterns
	sensitivePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*[^,\s]+`), // password=secret
		regexp.MustCompile(`(?i)(token|key|secret)\s*[:=]\s*[^,\s]+`),    // token=abc123
		regexp.MustCompile(`(?i)(api[_-]?key)\s*[:=]\s*[^,\s]+`),         // api_key=xyz
		regexp.MustCompile(`(?i)(auth[_-]?token)\s*[:=]\s*[^,\s]+`),      // auth_token=token
		regexp.MustCompile(`[a-fA-F0-9]{32,}`),                           // Long hex strings (potential keys)
	}

	for _, pattern := range sensitivePatterns {
		sanitized = pattern.ReplaceAllStringFunc(sanitized, func(match string) string {
			// Mask the sensitive part but keep the field name
			parts := strings.SplitN(match, "=", 2)
			if len(parts) == 2 {
				return parts[0] + "=[REDACTED]"
			}
			parts = strings.SplitN(match, ":", 2)
			if len(parts) == 2 {
				return parts[0] + ":[REDACTED]"
			}
			return "[REDACTED]"
		})
	}

	// Sanitize file paths to prevent information disclosure
	if strings.Contains(sanitized, "/") || strings.Contains(sanitized, "\\") {
		sanitized = regexp.MustCompile(`(?:/[^/\s]+)+/?|\\(?:[^\\\s]+)+\\?`).ReplaceAllStringFunc(sanitized, func(path string) string {
			// Keep only the filename, replace directory path
			return filepath.Base(path)
		})
	}

	return sanitized
}

// SecurePrintf provides secure logging that sanitizes sensitive data
func SecurePrintf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	sanitized := SanitizeForLogging(message)
	fmt.Print(sanitized)
}

// SecurePrintln provides secure logging that sanitizes sensitive data
func SecurePrintln(args ...interface{}) {
	message := fmt.Sprint(args...)
	sanitized := SanitizeForLogging(message)
	fmt.Println(sanitized)
}
