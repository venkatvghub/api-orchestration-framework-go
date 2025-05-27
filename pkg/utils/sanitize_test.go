package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeURL(t *testing.T) {
	t.Run("Simple URL", func(t *testing.T) {
		result := SanitizeURL("https://example.com/api/users")
		assert.Equal(t, "api_users", result)
	})

	t.Run("URL with query parameters", func(t *testing.T) {
		result := SanitizeURL("https://example.com/api/users?id=123&name=test")
		assert.Equal(t, "api_users", result)
	})

	t.Run("URL with special characters", func(t *testing.T) {
		result := SanitizeURL("https://example.com/api/users-list/active")
		assert.Equal(t, "api_users_list_active", result)
	})

	t.Run("Root URL", func(t *testing.T) {
		result := SanitizeURL("https://example.com/")
		assert.Equal(t, "root", result)
	})

	t.Run("Root URL without slash", func(t *testing.T) {
		result := SanitizeURL("https://example.com")
		assert.Equal(t, "root", result)
	})

	t.Run("Invalid URL", func(t *testing.T) {
		result := SanitizeURL("not-a-valid-url")
		assert.Equal(t, "invalid_url", result)
	})

	t.Run("URL with multiple slashes", func(t *testing.T) {
		result := SanitizeURL("https://example.com/api//users///list")
		assert.Equal(t, "api_users_list", result)
	})

	t.Run("URL with numbers and underscores", func(t *testing.T) {
		result := SanitizeURL("https://example.com/api/v1/user_profiles")
		assert.Equal(t, "api_v1_user_profiles", result)
	})

	t.Run("Complex path", func(t *testing.T) {
		result := SanitizeURL("https://api.example.com/v2/users/123/profile.json")
		assert.Equal(t, "v2_users_123_profile_json", result)
	})
}

func TestSanitizeIdentifier(t *testing.T) {
	t.Run("Simple string", func(t *testing.T) {
		result := SanitizeIdentifier("hello_world")
		assert.Equal(t, "hello_world", result)
	})

	t.Run("String with spaces", func(t *testing.T) {
		result := SanitizeIdentifier("hello world")
		assert.Equal(t, "hello_world", result)
	})

	t.Run("String with special characters", func(t *testing.T) {
		result := SanitizeIdentifier("hello-world@example.com")
		assert.Equal(t, "hello_world_example_com", result)
	})

	t.Run("String with multiple consecutive special chars", func(t *testing.T) {
		result := SanitizeIdentifier("hello---world___test")
		assert.Equal(t, "hello_world_test", result)
	})

	t.Run("String with leading/trailing underscores", func(t *testing.T) {
		result := SanitizeIdentifier("___hello_world___")
		assert.Equal(t, "hello_world", result)
	})

	t.Run("Empty string", func(t *testing.T) {
		result := SanitizeIdentifier("")
		assert.Equal(t, "unnamed", result)
	})

	t.Run("Only special characters", func(t *testing.T) {
		result := SanitizeIdentifier("@#$%^&*()")
		assert.Equal(t, "unnamed", result)
	})

	t.Run("Numbers and letters", func(t *testing.T) {
		result := SanitizeIdentifier("test123_abc")
		assert.Equal(t, "test123_abc", result)
	})

	t.Run("Mixed case", func(t *testing.T) {
		result := SanitizeIdentifier("HelloWorld")
		assert.Equal(t, "HelloWorld", result)
	})

	t.Run("Unicode characters", func(t *testing.T) {
		result := SanitizeIdentifier("hello世界")
		assert.Equal(t, "hello_", result)
	})
}

func TestSanitizeFilename(t *testing.T) {
	t.Run("Simple filename", func(t *testing.T) {
		result := SanitizeFilename("document.txt")
		assert.Equal(t, "document.txt", result)
	})

	t.Run("Filename with unsafe characters", func(t *testing.T) {
		result := SanitizeFilename("my<file>name.txt")
		assert.Equal(t, "my_file_name.txt", result)
	})

	t.Run("Filename with all unsafe characters", func(t *testing.T) {
		result := SanitizeFilename(`file<>:"/\|?*.txt`)
		assert.Equal(t, "file_________.txt", result)
	})

	t.Run("Filename with control characters", func(t *testing.T) {
		result := SanitizeFilename("file\x00\x1f\x7fname.txt")
		assert.Equal(t, "filename.txt", result)
	})

	t.Run("Filename with leading/trailing spaces and dots", func(t *testing.T) {
		result := SanitizeFilename("  ..filename.txt..  ")
		assert.Equal(t, "filename.txt", result)
	})

	t.Run("Empty filename", func(t *testing.T) {
		result := SanitizeFilename("")
		assert.Equal(t, "file", result)
	})

	t.Run("Reserved filename - CON", func(t *testing.T) {
		result := SanitizeFilename("CON")
		assert.Equal(t, "file", result)
	})

	t.Run("Reserved filename - PRN", func(t *testing.T) {
		result := SanitizeFilename("PRN")
		assert.Equal(t, "file", result)
	})

	t.Run("Reserved filename - COM1", func(t *testing.T) {
		result := SanitizeFilename("COM1")
		assert.Equal(t, "file", result)
	})

	t.Run("Reserved filename - LPT1", func(t *testing.T) {
		result := SanitizeFilename("LPT1")
		assert.Equal(t, "file", result)
	})

	t.Run("Reserved filename case insensitive", func(t *testing.T) {
		result := SanitizeFilename("con")
		assert.Equal(t, "file", result)
	})

	t.Run("Non-reserved similar name", func(t *testing.T) {
		result := SanitizeFilename("CONFIG")
		assert.Equal(t, "CONFIG", result)
	})

	t.Run("Only whitespace and dots", func(t *testing.T) {
		result := SanitizeFilename("   ...   ")
		assert.Equal(t, "file", result)
	})
}

func TestSanitizeLogMessage(t *testing.T) {
	t.Run("Message with password", func(t *testing.T) {
		result := SanitizeLogMessage("User login with password: secret123")
		assert.Equal(t, "User login with password=***", result)
	})

	t.Run("Message with token", func(t *testing.T) {
		result := SanitizeLogMessage("API call with token: abc123xyz")
		assert.Equal(t, "API call with token=***", result)
	})

	t.Run("Message with authorization header", func(t *testing.T) {
		result := SanitizeLogMessage("Request with Authorization: Bearer token123")
		assert.Equal(t, "Request with Authorization=***", result)
	})

	t.Run("Message with bearer token", func(t *testing.T) {
		result := SanitizeLogMessage("Using bearer abc123def456")
		assert.Equal(t, "Using bearer=***", result)
	})

	t.Run("Message with key", func(t *testing.T) {
		result := SanitizeLogMessage("API key=myapikey123")
		assert.Equal(t, "API key=***", result)
	})

	t.Run("Message with secret", func(t *testing.T) {
		result := SanitizeLogMessage("Client secret: mysecret456")
		assert.Equal(t, "Client secret=***", result)
	})

	t.Run("Message with pwd", func(t *testing.T) {
		result := SanitizeLogMessage("Database pwd=dbpassword")
		assert.Equal(t, "Database pwd=***", result)
	})

	t.Run("Message with pass", func(t *testing.T) {
		result := SanitizeLogMessage("Login pass: userpass")
		assert.Equal(t, "Login pass=***", result)
	})

	t.Run("Message without sensitive data", func(t *testing.T) {
		result := SanitizeLogMessage("User successfully logged in")
		assert.Equal(t, "User successfully logged in", result)
	})

	t.Run("Message with multiple sensitive fields", func(t *testing.T) {
		result := SanitizeLogMessage("Login with password: secret and token: abc123")
		assert.Equal(t, "Login with password=*** and token=***", result)
	})

	t.Run("Case insensitive matching", func(t *testing.T) {
		result := SanitizeLogMessage("Using PASSWORD: secret and TOKEN: abc123")
		assert.Equal(t, "Using PASSWORD=*** and TOKEN=***", result)
	})

	t.Run("Different separators", func(t *testing.T) {
		result := SanitizeLogMessage("password=secret token:abc123 key secret123")
		assert.Equal(t, "password=*** token=*** key=***", result)
	})
}

func TestSanitizeHeaders(t *testing.T) {
	t.Run("Headers with sensitive data", func(t *testing.T) {
		headers := map[string]interface{}{
			"Authorization": "Bearer token123",
			"Cookie":        "session=abc123",
			"Content-Type":  "application/json",
			"X-API-Key":     "apikey123",
		}
		result := SanitizeHeaders(headers)
		expected := map[string]interface{}{
			"Authorization": "***",
			"Cookie":        "***",
			"Content-Type":  "application/json",
			"X-API-Key":     "***",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("Headers case insensitive", func(t *testing.T) {
		headers := map[string]interface{}{
			"AUTHORIZATION": "Bearer token123",
			"cookie":        "session=abc123",
			"Set-Cookie":    "session=new123",
		}
		result := SanitizeHeaders(headers)
		expected := map[string]interface{}{
			"AUTHORIZATION": "***",
			"cookie":        "***",
			"Set-Cookie":    "***",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("Headers with X-Auth-Token", func(t *testing.T) {
		headers := map[string]interface{}{
			"X-Auth-Token": "authtoken123",
			"User-Agent":   "MyApp/1.0",
		}
		result := SanitizeHeaders(headers)
		expected := map[string]interface{}{
			"X-Auth-Token": "***",
			"User-Agent":   "MyApp/1.0",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("Empty headers", func(t *testing.T) {
		headers := map[string]interface{}{}
		result := SanitizeHeaders(headers)
		assert.Empty(t, result)
	})

	t.Run("No sensitive headers", func(t *testing.T) {
		headers := map[string]interface{}{
			"Content-Type":   "application/json",
			"Accept":         "application/json",
			"Content-Length": "123",
		}
		result := SanitizeHeaders(headers)
		assert.Equal(t, headers, result)
	})

	t.Run("Mixed case sensitive headers", func(t *testing.T) {
		headers := map[string]interface{}{
			"Authorization": "Bearer token",
			"authorization": "Basic auth",
			"COOKIE":        "session=123",
		}
		result := SanitizeHeaders(headers)
		expected := map[string]interface{}{
			"Authorization": "***",
			"authorization": "***",
			"COOKIE":        "***",
		}
		assert.Equal(t, expected, result)
	})
}

func TestTruncateString(t *testing.T) {
	t.Run("String shorter than max length", func(t *testing.T) {
		result := TruncateString("hello", 10)
		assert.Equal(t, "hello", result)
	})

	t.Run("String equal to max length", func(t *testing.T) {
		result := TruncateString("hello", 5)
		assert.Equal(t, "hello", result)
	})

	t.Run("String longer than max length", func(t *testing.T) {
		result := TruncateString("hello world", 8)
		assert.Equal(t, "hello...", result)
	})

	t.Run("Very short max length", func(t *testing.T) {
		result := TruncateString("hello", 3)
		assert.Equal(t, "hel", result)
	})

	t.Run("Max length of 1", func(t *testing.T) {
		result := TruncateString("hello", 1)
		assert.Equal(t, "h", result)
	})

	t.Run("Max length of 0", func(t *testing.T) {
		result := TruncateString("hello", 0)
		assert.Equal(t, "", result)
	})

	t.Run("Empty string", func(t *testing.T) {
		result := TruncateString("", 10)
		assert.Equal(t, "", result)
	})

	t.Run("Long string with ellipsis", func(t *testing.T) {
		result := TruncateString("This is a very long string that needs truncation", 20)
		assert.Equal(t, "This is a very lo...", result)
	})

	t.Run("Unicode string", func(t *testing.T) {
		result := TruncateString("Hello 世界", 8)
		assert.Equal(t, "Hello...", result)
	})
}

func TestIsReservedFilename(t *testing.T) {
	t.Run("Reserved names", func(t *testing.T) {
		reserved := []string{
			"CON", "PRN", "AUX", "NUL",
			"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
			"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
		}
		for _, name := range reserved {
			assert.True(t, isReservedFilename(name), "Should be reserved: %s", name)
		}
	})

	t.Run("Reserved names case insensitive", func(t *testing.T) {
		reserved := []string{"con", "prn", "aux", "nul", "com1", "lpt1"}
		for _, name := range reserved {
			assert.True(t, isReservedFilename(name), "Should be reserved (lowercase): %s", name)
		}
	})

	t.Run("Non-reserved names", func(t *testing.T) {
		nonReserved := []string{
			"CONFIG", "PRINT", "AUXILIARY", "NULL",
			"COM", "LPT", "COM10", "LPT10",
			"document", "file.txt", "test",
		}
		for _, name := range nonReserved {
			assert.False(t, isReservedFilename(name), "Should not be reserved: %s", name)
		}
	})

	t.Run("Empty string", func(t *testing.T) {
		assert.False(t, isReservedFilename(""))
	})
}

func TestSanitizationIntegration(t *testing.T) {
	t.Run("URL to identifier", func(t *testing.T) {
		url := "https://api.example.com/v1/users/profile"
		sanitizedURL := SanitizeURL(url)
		identifier := SanitizeIdentifier(sanitizedURL)
		assert.Equal(t, "v1_users_profile", identifier)
	})

	t.Run("Complex log sanitization", func(t *testing.T) {
		logMsg := "User login attempt with password: secret123 and token: abc456def"
		sanitized := SanitizeLogMessage(logMsg)
		assert.Equal(t, "User login attempt with password=*** and token=***", sanitized)
	})

	t.Run("Filename from URL", func(t *testing.T) {
		url := "https://example.com/api/users/report.pdf"
		sanitizedURL := SanitizeURL(url)
		filename := SanitizeFilename(sanitizedURL + ".log")
		assert.Equal(t, "api_users_report_pdf.log", filename)
	})
}
