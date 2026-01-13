package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/nicholasgriffintn/go-mcp-atlassian/pkg/auth"
)

// ToolHandler is a function that handles a tool call
type ToolHandler func(arguments map[string]interface{}) (*CallToolResult, error)

// ToolHandlerWithContext is a function that handles a tool call with request context.
// Use this for handlers that need access to per-request credentials from HTTP headers.
type ToolHandlerWithContext func(ctx context.Context, arguments map[string]interface{}) (*CallToolResult, error)

// Server represents an MCP server
type Server struct {
	name     string
	version  string
	tools    []Tool
	handlers map[string]ToolHandler
	ctxHandlers map[string]ToolHandlerWithContext
	mu       sync.RWMutex
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer

	// Rate limiting
	toolCallTimestamps []time.Time
	rateLimitMu        sync.Mutex

	// Callbacks
	onToolCall func(name string, args map[string]interface{}, duration time.Duration, success bool)
	onError    func(err error, context string)

	// Authentication
	authorizer auth.Authorizer

	// Request context for HTTP mode (stores current request context for tool calls)
	requestCtx context.Context
}

// NewServer creates a new MCP server
func NewServer(name, version string) *Server {
	return &Server{
		name:               name,
		version:            version,
		tools:              make([]Tool, 0),
		handlers:           make(map[string]ToolHandler),
		ctxHandlers:        make(map[string]ToolHandlerWithContext),
		stdin:              os.Stdin,
		stdout:             os.Stdout,
		stderr:             os.Stderr,
		toolCallTimestamps: make([]time.Time, 0),
	}
}

// SetToolCallCallback sets a callback for tool calls (for telemetry)
func (s *Server) SetToolCallCallback(cb func(name string, args map[string]interface{}, duration time.Duration, success bool)) {
	s.onToolCall = cb
}

// SetErrorCallback sets a callback for errors
func (s *Server) SetErrorCallback(cb func(err error, context string)) {
	s.onError = cb
}

// SetAuthorizer sets the authorizer for HTTP mode authentication.
func (s *Server) SetAuthorizer(authorizer auth.Authorizer) {
	s.authorizer = authorizer
}

// RegisterTool registers a tool with its handler
func (s *Server) RegisterTool(tool Tool, handler ToolHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools = append(s.tools, tool)
	s.handlers[tool.Name] = handler
}

// RegisterToolWithContext registers a tool with a context-aware handler.
// Use this for tools that need access to per-request credentials from HTTP headers.
func (s *Server) RegisterToolWithContext(tool Tool, handler ToolHandlerWithContext) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools = append(s.tools, tool)
	s.ctxHandlers[tool.Name] = handler
}

// GetRequestContext returns the current request context for HTTP mode.
// In stdio mode, returns context.Background().
func (s *Server) GetRequestContext() context.Context {
	if s.requestCtx != nil {
		return s.requestCtx
	}
	return context.Background()
}

// checkRateLimit returns true if the request should be rate limited
func (s *Server) checkRateLimit() bool {
	s.rateLimitMu.Lock()
	defer s.rateLimitMu.Unlock()

	now := time.Now()
	twentySecondsAgo := now.Add(-20 * time.Second)

	// Remove old timestamps
	newTimestamps := make([]time.Time, 0)
	for _, ts := range s.toolCallTimestamps {
		if ts.After(twentySecondsAgo) {
			newTimestamps = append(newTimestamps, ts)
		}
	}
	s.toolCallTimestamps = newTimestamps

	// Check if we have 5 or more calls in the past 20s
	if len(s.toolCallTimestamps) >= 5 {
		return true
	}

	// Record this call
	s.toolCallTimestamps = append(s.toolCallTimestamps, now)
	return false
}

// Run starts the server in stdio mode
func (s *Server) Run() error {
	lines := make(chan string)
	errors := make(chan error)

	go func() {
		reader := bufio.NewReader(s.stdin)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					if line != "" {
						lines <- line
					}
					errors <- io.EOF
					return
				}
				errors <- err
				return
			}
			lines <- line
		}
	}()

	receivedData := false
	initialTimeout := time.After(30 * time.Second)

	for {
		select {
		case line := <-lines:
			receivedData = true
			line = trimLine(line)
			if line == "" {
				continue
			}

			response := s.handleMessage([]byte(line))
			if response != nil {
				s.sendResponse(response)
			}

		case err := <-errors:
			if err == io.EOF {
				if receivedData {
					return nil
				}
				return fmt.Errorf("stdin closed before receiving any data")
			}
			return fmt.Errorf("read error: %w", err)

		case <-initialTimeout:
			if !receivedData {
				initialTimeout = time.After(24 * time.Hour)
			}
		}
	}
}

// RunHTTP starts the server in HTTP mode with optional authentication
func (s *Server) RunHTTP(addr string) error {
	mux := http.NewServeMux()

	// Health check endpoint (no auth required)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"version": s.version,
		})
	})

	// MCP endpoint with authentication
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check authentication using authorizer if set
		if s.authorizer != nil {
			token := r.Header.Get("Authorization")
			authorized, err := s.authorizer.Authorize(r.Context(), token)
			if err != nil || !authorized {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      nil,
					"error":   map[string]interface{}{"code": -32001, "message": "Unauthorized: invalid or missing authentication token"},
				})
				return
			}
		} else if auth.IsAuthEnabled() {
			// Fallback to legacy auth check for backward compatibility
			token := r.Header.Get(auth.AuthHeaderName)
			if !auth.ValidateAgainstExpected(token) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      nil,
					"error":   map[string]interface{}{"code": -32001, "message": "Unauthorized: invalid or missing authentication token"},
				})
				return
			}
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      nil,
				"error":   map[string]interface{}{"code": -32700, "message": "Parse error"},
			})
			return
		}

		// Inject credentials from headers into context
		ctx := r.Context()
		if token := r.Header.Get(auth.JiraPersonalTokenHeader); token != "" {
			ctx = context.WithValue(ctx, auth.JiraPersonalTokenKey, token)
		}
		if token := r.Header.Get(auth.ConfluencePersonalTokenHeader); token != "" {
			ctx = context.WithValue(ctx, auth.ConfluencePersonalTokenKey, token)
		}

		// Store context for tool handlers
		s.requestCtx = ctx
		defer func() { s.requestCtx = nil }()

		response := s.handleMessage(body)
		if response != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	})

	authEnabled := s.authorizer != nil || auth.IsAuthEnabled()
	if authEnabled {
		fmt.Fprintf(s.stderr, "Atlassian MCP Server running on HTTP at %s (authentication enabled)\n", addr)
	} else {
		fmt.Fprintf(s.stderr, "Atlassian MCP Server running on HTTP at %s (authentication disabled)\n", addr)
	}
	return http.ListenAndServe(addr, mux)
}

func trimLine(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

func (s *Server) handleMessage(data []byte) *JSONRPCResponse {
	var request JSONRPCRequest
	if err := json.Unmarshal(data, &request); err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    ParseError,
				Message: "Parse error",
				Data:    err.Error(),
			},
		}
	}

	// Handle notifications (no ID)
	if request.ID == nil {
		s.handleNotification(&request)
		return nil
	}

	return s.handleRequest(&request)
}

func (s *Server) handleNotification(request *JSONRPCRequest) {
	switch request.Method {
	case "notifications/initialized":
		fmt.Fprintln(s.stderr, "Client initialized")
	case "notifications/cancelled":
		// Request cancellation
	}
}

func (s *Server) handleRequest(request *JSONRPCRequest) *JSONRPCResponse {
	response := &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
	}

	switch request.Method {
	case "initialize":
		response.Result = s.handleInitialize(request.Params)
	case "tools/list":
		response.Result = s.handleListTools()
	case "tools/call":
		result, err := s.handleCallTool(request.Params)
		if err != nil {
			response.Error = &JSONRPCError{
				Code:    InternalError,
				Message: err.Error(),
			}
		} else {
			response.Result = result
		}
	case "ping":
		response.Result = map[string]interface{}{}
	default:
		response.Error = &JSONRPCError{
			Code:    MethodNotFound,
			Message: fmt.Sprintf("Method not found: %s", request.Method),
		}
	}

	return response
}

func (s *Server) handleInitialize(params interface{}) *InitializeResult {
	return &InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: false,
			},
		},
		ServerInfo: ServerInfo{
			Name:    s.name,
			Version: s.version,
		},
	}
}

func (s *Server) handleListTools() *ListToolsResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &ListToolsResult{
		Tools: s.tools,
	}
}

func (s *Server) handleCallTool(params interface{}) (*CallToolResult, error) {
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid params type")
	}

	name, ok := paramsMap["name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing tool name")
	}

	arguments, _ := paramsMap["arguments"].(map[string]interface{})

	// Check rate limit
	if s.checkRateLimit() {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: "Rate limit exceeded: Maximum 5 tool calls per 20 seconds. Please try again later."}},
			IsError: true,
		}, nil
	}

	s.mu.RLock()
	handler, handlerExists := s.handlers[name]
	ctxHandler, ctxHandlerExists := s.ctxHandlers[name]
	s.mu.RUnlock()

	if !handlerExists && !ctxHandlerExists {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Unknown tool: %s", name)}},
			IsError: true,
		}, nil
	}

	startTime := time.Now()
	var result *CallToolResult
	var err error

	// Use context-aware handler if available, otherwise use legacy handler
	if ctxHandlerExists {
		result, err = ctxHandler(s.GetRequestContext(), arguments)
	} else {
		result, err = handler(arguments)
	}
	duration := time.Since(startTime)

	success := err == nil && (result == nil || !result.IsError)

	// Call telemetry callback
	if s.onToolCall != nil {
		s.onToolCall(name, arguments, duration, success)
	}

	if err != nil {
		if s.onError != nil {
			s.onError(err, fmt.Sprintf("tool_%s", name))
		}
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Error: %s", err.Error())}},
			IsError: true,
		}, nil
	}

	return result, nil
}

func (s *Server) sendResponse(response *JSONRPCResponse) {
	data, err := json.Marshal(response)
	if err != nil {
		fmt.Fprintf(s.stderr, "Error marshaling response: %v\n", err)
		return
	}
	fmt.Fprintln(s.stdout, string(data))
}

// Log writes a message to stderr for debugging
func (s *Server) Log(format string, args ...interface{}) {
	fmt.Fprintf(s.stderr, format+"\n", args...)
}
