package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

// ConfigSource indicates where a configuration value came from
type ConfigSource string

const (
	SourceDefault     ConfigSource = "default"
	SourceEnvironment ConfigSource = "environment"
	SourceFlag        ConfigSource = "flag"
)

const (
	LevelOff LogLevel = iota
	LevelError
	LevelWarn
	LevelInfo
	LevelAccess
	LevelDebug
)

func (l LogLevel) String() string {
	switch l {
	case LevelOff:
		return "OFF"
	case LevelError:
		return "ERROR"
	case LevelWarn:
		return "WARN"
	case LevelInfo:
		return "INFO"
	case LevelAccess:
		return "ACCESS"
	case LevelDebug:
		return "DEBUG"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel converts a string to a LogLevel
func ParseLevel(s string) LogLevel {
	switch strings.ToLower(s) {
	case "off":
		return LevelOff
	case "error":
		return LevelError
	case "warn", "warning":
		return LevelWarn
	case "info":
		return LevelInfo
	case "access":
		return LevelAccess
	case "debug":
		return LevelDebug
	default:
		return LevelInfo
	}
}

// ConfigValue holds a configuration value and its source
type ConfigValue struct {
	Value  string
	Source ConfigSource
}

// Logger handles structured logging to file
type Logger struct {
	mu           sync.Mutex
	level        LogLevel
	logger       *log.Logger
	file         *os.File
	logDir       string
	appName      string
	startTime    time.Time
	logQueries   bool
	queryLogDir  string
}

// Config holds logger configuration
type Config struct {
	LogDir     string
	AppName    string
	Level      LogLevel
	LogQueries bool // Enable query logging to queries subfolder
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// DefaultLogDir returns the default log directory for the application
func DefaultLogDir(appName string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", appName, "logs")
	}
	return filepath.Join(homeDir, appName, "logs")
}

// Init initializes the global default logger
func Init(cfg Config) error {
	var initErr error
	once.Do(func() {
		defaultLogger, initErr = NewLogger(cfg)
	})
	return initErr
}

// NewLogger creates a new Logger instance
func NewLogger(cfg Config) (*Logger, error) {
	if cfg.AppName == "" {
		cfg.AppName = "go-mcp-atlassian"
	}

	logDir := cfg.LogDir
	if logDir == "" {
		logDir = DefaultLogDir(cfg.AppName)
	}

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory %s: %w", logDir, err)
	}

	timestamp := time.Now().Format("2006-01-02")
	logFileName := fmt.Sprintf("%s-%s.log", cfg.AppName, timestamp)
	logPath := filepath.Join(logDir, logFileName)

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", logPath, err)
	}

	// Set up query log directory if query logging is enabled
	queryLogDir := filepath.Join(logDir, "queries")

	l := &Logger{
		level:       cfg.Level,
		logger:      log.New(file, "", 0),
		file:        file,
		logDir:      logDir,
		appName:     cfg.AppName,
		startTime:   time.Now(),
		logQueries:  cfg.LogQueries,
		queryLogDir: queryLogDir,
	}

	return l, nil
}

// Close closes the log file
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current logging level
func (l *Logger) GetLevel() LogLevel {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if l == nil || level > l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z07:00")
	message := fmt.Sprintf(format, args...)
	l.logger.Printf("[%s] [%s] %s", timestamp, level.String(), message)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Access logs an access message
func (l *Logger) Access(format string, args ...interface{}) {
	l.log(LevelAccess, format, args...)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// APIRequest logs an API request (no sensitive data)
func (l *Logger) APIRequest(method, endpoint string, statusCode int, duration time.Duration, err error) {
	if err != nil {
		l.Access("API_REQUEST method=%s endpoint=%q status=%d duration=%s error=%q", method, endpoint, statusCode, duration, err.Error())
	} else {
		l.Access("API_REQUEST method=%s endpoint=%q status=%d duration=%s", method, endpoint, statusCode, duration)
	}
}

// ToolCall logs a tool invocation
func (l *Logger) ToolCall(toolName string, args map[string]interface{}, duration time.Duration, success bool) {
	argKeys := make([]string, 0, len(args))
	for k := range args {
		argKeys = append(argKeys, k)
	}
	l.Info("TOOL_CALL tool=%q args=%v duration=%s success=%v", toolName, argKeys, duration, success)
}

// LogQuery logs a query to the queries subfolder when query logging is enabled.
// Queries are stored in: {log_dir}/queries/YYYYMMDD/{descriptive_name}.YYYYMMDD.HHmmss.query
// The queryType should be a short descriptive name like "jira_search", "confluence_search", etc.
// The query parameter contains the actual query content (JQL, CQL, etc.).
// The args parameter contains the full arguments map for additional context.
func (l *Logger) LogQuery(queryType string, query string, args map[string]interface{}) {
	if l == nil || !l.logQueries {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	dateDir := now.Format("20060102")
	timestamp := now.Format("20060102.150405")

	// Create date-specific directory: {log_dir}/queries/YYYYMMDD/
	queryDirPath := filepath.Join(l.queryLogDir, dateDir)
	if err := os.MkdirAll(queryDirPath, 0755); err != nil {
		// Log error but don't fail - query logging is optional
		l.logger.Printf("[%s] [ERROR] Failed to create query log directory %s: %v",
			now.Format("2006-01-02T15:04:05.000Z07:00"), queryDirPath, err)
		return
	}

	// Sanitize queryType to create a safe filename
	safeName := sanitizeFileName(queryType)

	// Build filename: {descriptive_name}.YYYYMMDD.HHmmss.query
	fileName := fmt.Sprintf("%s.%s.query", safeName, timestamp)
	filePath := filepath.Join(queryDirPath, fileName)

	// Build query log content
	var content strings.Builder
	content.WriteString(fmt.Sprintf("# Query Log: %s\n", queryType))
	content.WriteString(fmt.Sprintf("# Timestamp: %s\n", now.Format(time.RFC3339)))
	content.WriteString(fmt.Sprintf("# Type: %s\n", queryType))
	content.WriteString("# ----------------------------------------\n\n")
	content.WriteString("## Query:\n")
	content.WriteString(query)
	content.WriteString("\n\n")

	// Add arguments if present
	if len(args) > 0 {
		content.WriteString("## Arguments:\n")
		for k, v := range args {
			content.WriteString(fmt.Sprintf("  %s: %v\n", k, v))
		}
	}

	// Write to file
	if err := os.WriteFile(filePath, []byte(content.String()), 0644); err != nil {
		l.logger.Printf("[%s] [ERROR] Failed to write query log %s: %v",
			now.Format("2006-01-02T15:04:05.000Z07:00"), filePath, err)
		return
	}

	l.Debug("QUERY_LOGGED type=%s file=%s", queryType, filePath)
}

// IsQueryLoggingEnabled returns whether query logging is enabled
func (l *Logger) IsQueryLoggingEnabled() bool {
	if l == nil {
		return false
	}
	return l.logQueries
}

// sanitizeFileName removes or replaces characters that are unsafe for filenames
func sanitizeFileName(name string) string {
	// Replace unsafe characters with underscores
	unsafe := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " "}
	result := name
	for _, char := range unsafe {
		result = strings.ReplaceAll(result, char, "_")
	}
	// Limit length to reasonable filename size
	if len(result) > 50 {
		result = result[:50]
	}
	return result
}

// AuthToken logs authentication token operations (no secrets)
func (l *Logger) AuthToken(tokenType string, expiresIn time.Duration, err error) {
	if err != nil {
		l.Debug("AUTH_TOKEN type=%s error=%q", tokenType, err.Error())
	} else {
		l.Debug("AUTH_TOKEN type=%s expires_in=%s", tokenType, expiresIn)
	}
}

// StartupInfo contains information logged at server startup
type StartupInfo struct {
	Version      string
	GoVersion    string
	OS           string
	Arch         string
	NumCPU       int
	LogDir       ConfigValue
	LogLevel     ConfigValue
	AtlassianURL ConfigValue
	PID          int
	StartTime    time.Time
}

// LogStartup logs server startup information
func (l *Logger) LogStartup(info StartupInfo) {
	l.Info("========================================")
	l.Info("SERVER STARTUP")
	l.Info("========================================")
	l.Info("Application: %s", l.appName)
	l.Info("Version: %s", info.Version)
	l.Info("Go Version: %s", info.GoVersion)
	l.Info("OS: %s", info.OS)
	l.Info("Architecture: %s", info.Arch)
	l.Info("Number of CPUs: %d", info.NumCPU)
	l.Info("Process ID: %d", info.PID)
	l.Info("Start Time: %s", info.StartTime.Format(time.RFC3339))
	l.Info("----------------------------------------")
	l.Info("CONFIGURATION (value [source])")
	l.Info("----------------------------------------")
	l.Info("Log Directory: %s [%s]", info.LogDir.Value, info.LogDir.Source)
	l.Info("Log Level: %s [%s]", info.LogLevel.Value, info.LogLevel.Source)
	l.Info("Atlassian URL: %s [%s]", info.AtlassianURL.Value, info.AtlassianURL.Source)
	l.Info("========================================")
}

// LogShutdown logs server shutdown information
func (l *Logger) LogShutdown(reason string) {
	uptime := time.Since(l.startTime)
	l.Info("========================================")
	l.Info("SERVER SHUTDOWN")
	l.Info("========================================")
	l.Info("Reason: %s", reason)
	l.Info("Uptime: %s", uptime)
	l.Info("========================================")
}

// GetLogger returns the global default logger
func GetLogger() *Logger {
	return defaultLogger
}

// SetOutput sets the output writer for the logger
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.SetOutput(w)
}

// GetStartupInfo creates a StartupInfo struct with current system information
func GetStartupInfo(version string, logDir ConfigValue, logLevel ConfigValue, atlassianURL ConfigValue) StartupInfo {
	return StartupInfo{
		Version:      version,
		GoVersion:    runtime.Version(),
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		NumCPU:       runtime.NumCPU(),
		LogDir:       logDir,
		LogLevel:     logLevel,
		AtlassianURL: atlassianURL,
		PID:          os.Getpid(),
		StartTime:    time.Now(),
	}
}

// Global convenience functions

// Error logs an error message using the default logger
func Error(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Error(format, args...)
	}
}

// Warn logs a warning message using the default logger
func Warn(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Warn(format, args...)
	}
}

// Info logs an info message using the default logger
func Info(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Info(format, args...)
	}
}

// Access logs an access message using the default logger
func Access(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Access(format, args...)
	}
}

// Debug logs a debug message using the default logger
func Debug(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Debug(format, args...)
	}
}

// APIRequest logs an API request using the default logger
func APIRequest(method, endpoint string, statusCode int, duration time.Duration, err error) {
	if defaultLogger != nil {
		defaultLogger.APIRequest(method, endpoint, statusCode, duration, err)
	}
}

// ToolCall logs a tool invocation using the default logger
func ToolCall(toolName string, args map[string]interface{}, duration time.Duration, success bool) {
	if defaultLogger != nil {
		defaultLogger.ToolCall(toolName, args, duration, success)
	}
}

// AuthToken logs authentication token operations using the default logger
func AuthToken(tokenType string, expiresIn time.Duration, err error) {
	if defaultLogger != nil {
		defaultLogger.AuthToken(tokenType, expiresIn, err)
	}
}

// PII filtering patterns
var (
	// SSN patterns: xxx-xx-xxxx or xxxxxxxxx (9 digits)
	ssnPattern = regexp.MustCompile(`\b(\d{3}[-\s]?\d{2}[-\s]?\d{4})\b`)
	// PAN (credit card) patterns: 13-19 digit sequences, optionally with spaces/dashes
	panPattern = regexp.MustCompile(`\b(\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{1,7})\b`)
	// Additional PAN pattern for continuous digits
	panContinuousPattern = regexp.MustCompile(`\b(\d{13,19})\b`)
	// Email pattern
	emailPattern = regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)
)

// MaskSecret masks a secret value showing only the last 4 characters
// e.g., "mysecrettoken123" becomes "xxx3123"
func MaskSecret(secret string) string {
	if secret == "" {
		return ""
	}
	if len(secret) <= 4 {
		return "xxx" + secret
	}
	return "xxx" + secret[len(secret)-4:]
}

// SanitizePII removes or masks PII data from log messages
// - SSNs are replaced with [SSN-REDACTED]
// - PANs (credit card numbers) are replaced with [PAN-REDACTED]
func SanitizePII(message string) string {
	// Mask SSNs
	message = ssnPattern.ReplaceAllString(message, "[SSN-REDACTED]")
	// Mask PANs with separators
	message = panPattern.ReplaceAllString(message, "[PAN-REDACTED]")
	// Mask continuous digit PANs
	message = panContinuousPattern.ReplaceAllString(message, "[PAN-REDACTED]")
	return message
}

// SanitizeEmail masks email addresses in a message
func SanitizeEmail(message string) string {
	return emailPattern.ReplaceAllString(message, "[EMAIL-REDACTED]")
}

// SanitizeAndMaskSecrets sanitizes PII and masks known secret field values
func SanitizeAndMaskSecrets(message string, secretFields ...string) string {
	sanitized := SanitizePII(message)
	for _, field := range secretFields {
		if field != "" {
			masked := MaskSecret(field)
			sanitized = strings.ReplaceAll(sanitized, field, masked)
		}
	}
	return sanitized
}

// HTTPRequestInfo contains HTTP request details for logging
type HTTPRequestInfo struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    string
}

// HTTPResponseInfo contains HTTP response details for logging
type HTTPResponseInfo struct {
	StatusCode int
	Headers    map[string]string
	Body       string
}

// sanitizeHeaders removes sensitive header values
func sanitizeHeaders(headers map[string]string) map[string]string {
	if headers == nil {
		return nil
	}
	sanitized := make(map[string]string)
	sensitiveHeaders := []string{"authorization", "x-api-key", "api-key", "token", "secret", "password", "credential", "cookie"}
	for k, v := range headers {
		lowerKey := strings.ToLower(k)
		isSensitive := false
		for _, sensitive := range sensitiveHeaders {
			if strings.Contains(lowerKey, sensitive) {
				isSensitive = true
				break
			}
		}
		if isSensitive {
			sanitized[k] = MaskSecret(v)
		} else {
			sanitized[k] = SanitizePII(v)
		}
	}
	return sanitized
}

// formatHeaders formats headers for logging
func formatHeaders(headers map[string]string) string {
	if len(headers) == 0 {
		return "{}"
	}
	parts := make([]string, 0, len(headers))
	for k, v := range headers {
		parts = append(parts, fmt.Sprintf("%s=%q", k, v))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// LogHTTPError logs detailed HTTP error information with PII filtering
func (l *Logger) LogHTTPError(context string, req *HTTPRequestInfo, resp *HTTPResponseInfo, err error, secrets ...string) {
	if l == nil {
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("HTTP_ERROR context=%q", context))

	if req != nil {
		sb.WriteString(fmt.Sprintf(" request.method=%s request.url=%q", req.Method, req.URL))
		if len(req.Headers) > 0 {
			sanitizedHeaders := sanitizeHeaders(req.Headers)
			sb.WriteString(fmt.Sprintf(" request.headers=%s", formatHeaders(sanitizedHeaders)))
		}
		if req.Body != "" {
			sanitizedBody := SanitizeAndMaskSecrets(req.Body, secrets...)
			// Truncate long bodies
			if len(sanitizedBody) > 500 {
				sanitizedBody = sanitizedBody[:500] + "...[truncated]"
			}
			sb.WriteString(fmt.Sprintf(" request.body=%q", sanitizedBody))
		}
	}

	if resp != nil {
		sb.WriteString(fmt.Sprintf(" response.status=%d", resp.StatusCode))
		if len(resp.Headers) > 0 {
			sanitizedHeaders := sanitizeHeaders(resp.Headers)
			sb.WriteString(fmt.Sprintf(" response.headers=%s", formatHeaders(sanitizedHeaders)))
		}
		if resp.Body != "" {
			sanitizedBody := SanitizeAndMaskSecrets(resp.Body, secrets...)
			// Truncate long bodies
			if len(sanitizedBody) > 1000 {
				sanitizedBody = sanitizedBody[:1000] + "...[truncated]"
			}
			sb.WriteString(fmt.Sprintf(" response.body=%q", sanitizedBody))
		}
	}

	if err != nil {
		sanitizedErr := SanitizeAndMaskSecrets(err.Error(), secrets...)
		sb.WriteString(fmt.Sprintf(" error=%q", sanitizedErr))
	}

	l.Error(sb.String())
}

// LogTokenError logs token retrieval errors with masked credentials
func (l *Logger) LogTokenError(tokenType, authURL, username, apiToken string, statusCode int, responseBody string, err error) {
	if l == nil {
		return
	}

	maskedUsername := MaskSecret(username)
	maskedToken := MaskSecret(apiToken)
	sanitizedBody := SanitizeAndMaskSecrets(responseBody, username, apiToken)

	// Truncate long bodies
	if len(sanitizedBody) > 500 {
		sanitizedBody = sanitizedBody[:500] + "...[truncated]"
	}

	var errStr string
	if err != nil {
		errStr = SanitizeAndMaskSecrets(err.Error(), username, apiToken)
	}

	l.Error("TOKEN_ERROR type=%s auth_url=%q username=%s api_token=%s status=%d response=%q error=%q",
		tokenType, authURL, maskedUsername, maskedToken, statusCode, sanitizedBody, errStr)
}

// Global convenience functions for HTTP error logging

// LogHTTPError logs detailed HTTP error information using the default logger
func LogHTTPError(context string, req *HTTPRequestInfo, resp *HTTPResponseInfo, err error, secrets ...string) {
	if defaultLogger != nil {
		defaultLogger.LogHTTPError(context, req, resp, err, secrets...)
	}
}

// LogTokenError logs token retrieval errors using the default logger
func LogTokenError(tokenType, authURL, username, apiToken string, statusCode int, responseBody string, err error) {
	if defaultLogger != nil {
		defaultLogger.LogTokenError(tokenType, authURL, username, apiToken, statusCode, responseBody, err)
	}
}
