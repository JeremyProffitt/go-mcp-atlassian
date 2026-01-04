package atlassian

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nicholasgriffintn/go-mcp-atlassian/pkg/logging"
)

// TestAPICallLogging verifies that API requests and responses are logged with headers
func TestAPICallLogging(t *testing.T) {
	// Create a temporary directory for logs
	tmpDir, err := os.MkdirTemp("", "mcp-atlassian-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize logger with DEBUG level to capture HTTP logs
	logger, err := logging.NewLogger(logging.Config{
		LogDir:  tmpDir,
		AppName: "test-atlassian",
		Level:   logging.LevelDebug,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Create a buffer to capture log output
	var logBuffer bytes.Buffer
	logger.SetOutput(&logBuffer)

	// Create a mock HTTP server that returns a response with custom headers
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set some response headers
		w.Header().Set("X-Request-Id", "test-request-123")
		w.Header().Set("X-Custom-Header", "custom-value")
		w.Header().Set("Content-Type", "application/json")

		// Return a JSON response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer mockServer.Close()

	// Create a client with the mock server URL
	config := &Config{
		URL:           mockServer.URL,
		PersonalToken: "test-token-secret-1234",
		SSLVerify:     true,
		IsCloud:       false,
	}

	client, err := NewClient(config, WithLogger(logger))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Make an API request
	var result map[string]string
	err = client.Get(context.Background(), "/test/endpoint", &result)
	if err != nil {
		t.Fatalf("API request failed: %v", err)
	}

	// Give logger time to flush
	time.Sleep(100 * time.Millisecond)

	// Read the log output
	logOutput := logBuffer.String()

	// Verify request logging
	t.Run("LogsRequestMethod", func(t *testing.T) {
		if !strings.Contains(logOutput, "HTTP_REQUEST") {
			t.Error("Log should contain HTTP_REQUEST")
		}
		if !strings.Contains(logOutput, "method=GET") {
			t.Error("Log should contain request method")
		}
	})

	t.Run("LogsRequestURL", func(t *testing.T) {
		if !strings.Contains(logOutput, "/test/endpoint") {
			t.Error("Log should contain request URL/endpoint")
		}
	})

	t.Run("LogsRequestHeaders", func(t *testing.T) {
		if !strings.Contains(logOutput, "Authorization=") {
			t.Error("Log should contain Authorization header")
		}
		if !strings.Contains(logOutput, "Content-Type=") {
			t.Error("Log should contain Content-Type header")
		}
	})

	t.Run("RedactsAuthToken", func(t *testing.T) {
		// The token "test-token-secret-1234" should be redacted to show only last 4 chars
		if strings.Contains(logOutput, "test-token-secret-1234") {
			t.Error("Log should NOT contain full token - should be redacted")
		}
		// Should contain masked version (xxx + last 4 chars)
		if !strings.Contains(logOutput, "xxx") {
			t.Error("Log should contain masked token (xxx prefix)")
		}
	})

	t.Run("LogsResponseStatus", func(t *testing.T) {
		if !strings.Contains(logOutput, "HTTP_RESPONSE") {
			t.Error("Log should contain HTTP_RESPONSE")
		}
		if !strings.Contains(logOutput, "status=200") {
			t.Error("Log should contain response status code")
		}
	})

	t.Run("LogsResponseHeaders", func(t *testing.T) {
		if !strings.Contains(logOutput, "X-Request-Id=") {
			t.Error("Log should contain X-Request-Id response header")
		}
		if !strings.Contains(logOutput, "X-Custom-Header=") {
			t.Error("Log should contain X-Custom-Header response header")
		}
	})

	t.Run("LogsResponseBody", func(t *testing.T) {
		if !strings.Contains(logOutput, "body=") {
			t.Error("Log should contain response body")
		}
	})

	// Print log output for debugging if any test fails
	if t.Failed() {
		t.Logf("Full log output:\n%s", logOutput)
	}
}

// TestAPICallLogging401Error verifies that 401 errors are logged with full details
func TestAPICallLogging401Error(t *testing.T) {
	// Create a temporary directory for logs
	tmpDir, err := os.MkdirTemp("", "mcp-atlassian-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize logger with DEBUG level
	logger, err := logging.NewLogger(logging.Config{
		LogDir:  tmpDir,
		AppName: "test-atlassian",
		Level:   logging.LevelDebug,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Create a buffer to capture log output
	var logBuffer bytes.Buffer
	logger.SetOutput(&logBuffer)

	// Create a mock server that returns 401 Unauthorized
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Www-Authenticate", "Bearer realm=\"test\"")
		w.Header().Set("X-Request-Id", "failed-request-456")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"errorMessages": "You do not have permission",
			"errors":        "Unauthorized",
		})
	}))
	defer mockServer.Close()

	// Create a client
	config := &Config{
		URL:           mockServer.URL,
		PersonalToken: "invalid-token-abcd1234",
		SSLVerify:     true,
		IsCloud:       false,
	}

	client, err := NewClient(config, WithLogger(logger))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Make an API request that will fail
	var result map[string]interface{}
	err = client.Get(context.Background(), "/rest/api/2/issue/TEST-123", &result)

	// Should return an error
	if err == nil {
		t.Fatal("Expected error for 401 response, got nil")
	}

	// Give logger time to flush
	time.Sleep(100 * time.Millisecond)

	logOutput := logBuffer.String()

	t.Run("LogsUnauthorizedResponse", func(t *testing.T) {
		if !strings.Contains(logOutput, "status=401") {
			t.Error("Log should contain 401 status code")
		}
	})

	t.Run("LogsWwwAuthenticateHeader", func(t *testing.T) {
		if !strings.Contains(logOutput, "Www-Authenticate=") {
			t.Error("Log should contain Www-Authenticate header for 401 debugging")
		}
	})

	t.Run("LogsErrorResponseBody", func(t *testing.T) {
		if !strings.Contains(logOutput, "errorMessages") || !strings.Contains(logOutput, "permission") {
			t.Error("Log should contain error response body")
		}
	})

	t.Run("RedactsTokenIn401Response", func(t *testing.T) {
		if strings.Contains(logOutput, "invalid-token-abcd1234") {
			t.Error("Log should NOT contain full token even in error case")
		}
	})

	if t.Failed() {
		t.Logf("Full log output:\n%s", logOutput)
	}
}

// TestLoggerNotInitialized verifies client works even without logger
func TestLoggerNotInitialized(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer mockServer.Close()

	config := &Config{
		URL:           mockServer.URL,
		PersonalToken: "test-token",
		SSLVerify:     true,
		IsCloud:       false,
	}

	// Create client without initializing global logger
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Explicitly set logger to nil to test nil-safety
	client.logger = nil

	// This should not panic even with nil logger
	var result map[string]string
	err = client.Get(context.Background(), "/test", &result)
	if err != nil {
		t.Fatalf("Request should succeed even without logger: %v", err)
	}
}

// TestLogFileCreation verifies that log files are created on disk
func TestLogFileCreation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcp-atlassian-logfile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logger, err := logging.NewLogger(logging.Config{
		LogDir:  tmpDir,
		AppName: "test-app",
		Level:   logging.LevelDebug,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Log something
	logger.Debug("Test log message")
	logger.Close()

	// Check that log file was created
	files, err := filepath.Glob(filepath.Join(tmpDir, "test-app-*.log"))
	if err != nil {
		t.Fatalf("Failed to glob log files: %v", err)
	}

	if len(files) == 0 {
		t.Error("Expected log file to be created")
	}

	// Read and verify content
	if len(files) > 0 {
		content, err := os.ReadFile(files[0])
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}
		if !strings.Contains(string(content), "Test log message") {
			t.Error("Log file should contain test message")
		}
	}
}

// TestMaskSecret verifies token masking shows only last 4 characters
func TestMaskSecret(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"mysecrettoken1234", "xxx1234"},
		{"short", "xxxhort"},
		{"1234", "xxx1234"},
		{"abc", "xxxabc"},
		{"", ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := logging.MaskSecret(tc.input)
			if result != tc.expected {
				t.Errorf("MaskSecret(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}
