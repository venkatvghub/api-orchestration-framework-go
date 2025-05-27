package utils

import (
	"net/url"
	"regexp"
	"strings"
)

// SanitizeURL converts a URL into a safe identifier string
func SanitizeURL(urlStr string) string {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return "invalid_url"
	}

	path := strings.Trim(parsed.Path, "/")
	path = SanitizeIdentifier(path)

	if path == "" {
		return "root"
	}

	return path
}

// SanitizeIdentifier converts a string into a safe identifier (alphanumeric + underscore)
func SanitizeIdentifier(input string) string {
	// Replace non-alphanumeric characters with underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	sanitized := reg.ReplaceAllString(input, "_")

	// Remove multiple consecutive underscores
	reg = regexp.MustCompile(`_+`)
	sanitized = reg.ReplaceAllString(sanitized, "_")

	// Remove leading/trailing underscores
	sanitized = strings.Trim(sanitized, "_")

	// Ensure it's not empty
	if sanitized == "" {
		return "unnamed"
	}

	return sanitized
}

// SanitizeFilename converts a string into a safe filename
func SanitizeFilename(filename string) string {
	// Remove or replace unsafe characters for filenames
	unsafe := regexp.MustCompile(`[<>:"/\\|?*]`)
	sanitized := unsafe.ReplaceAllString(filename, "_")

	// Remove control characters
	control := regexp.MustCompile(`[\x00-\x1f\x7f]`)
	sanitized = control.ReplaceAllString(sanitized, "")

	// Trim whitespace and dots
	sanitized = strings.Trim(sanitized, " .")

	// Ensure it's not empty and not a reserved name
	if sanitized == "" || isReservedFilename(sanitized) {
		return "file"
	}

	return sanitized
}

// SanitizeLogMessage removes sensitive information from log messages
func SanitizeLogMessage(message string) string {
	// Remove potential passwords, tokens, and keys
	patterns := []string{
		`(?i)(password|pwd|pass)\s*[:=]\s*[^\s]+`,
		`(?i)(token|key|secret)\s*[:=]\s*[^\s]+`,
		`(?i)(authorization|auth)\s*:\s*[^\s]+`,
		`(?i)bearer\s+[^\s]+`,
	}

	sanitized := message
	for _, pattern := range patterns {
		reg := regexp.MustCompile(pattern)
		sanitized = reg.ReplaceAllString(sanitized, "${1}=***")
	}

	return sanitized
}

// SanitizeHeaders removes sensitive headers for logging
func SanitizeHeaders(headers map[string]interface{}) map[string]interface{} {
	sensitiveHeaders := map[string]bool{
		"authorization": true,
		"cookie":        true,
		"set-cookie":    true,
		"x-api-key":     true,
		"x-auth-token":  true,
	}

	sanitized := make(map[string]interface{})
	for key, value := range headers {
		lowerKey := strings.ToLower(key)
		if sensitiveHeaders[lowerKey] {
			sanitized[key] = "***"
		} else {
			sanitized[key] = value
		}
	}

	return sanitized
}

// TruncateString truncates a string to a maximum length with ellipsis
func TruncateString(input string, maxLength int) string {
	if len(input) <= maxLength {
		return input
	}

	if maxLength <= 3 {
		return input[:maxLength]
	}

	return input[:maxLength-3] + "..."
}

// isReservedFilename checks if a filename is reserved on Windows
func isReservedFilename(filename string) bool {
	reserved := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}

	upper := strings.ToUpper(filename)
	for _, res := range reserved {
		if upper == res {
			return true
		}
	}

	return false
}
