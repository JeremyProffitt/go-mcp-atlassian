package mcp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nicholasgriffintn/go-mcp-atlassian/pkg/auth"
)

// createTestServer creates an MCP server with a test tool and returns the HTTP test server
func createTestServer(t *testing.T, authorizer auth.Authorizer) (*Server, *httptest.Server) {
	t.Helper()

	server := NewServer("test-server", "1.0.0")
	if authorizer != nil {
		server.SetAuthorizer(authorizer)
	}

	// Register a simple test tool
	testTool := Tool{
		Name:        "test_tool",
		Description: "A test tool for integration tests",
		InputSchema: JSONSchema{
			Type: "object",
			Properties: map[string]Property{
				"message": {
					Type:        "string",
					Description: "A test message",
				},
			},
		},
	}
	server.RegisterTool(testTool, func(arguments map[string]interface{}) (*CallToolResult, error) {
		msg := "default"
		if m, ok := arguments["message"].(string); ok {
			msg = m
		}
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: "Received: " + msg}},
		}, nil
	})

	// Create HTTP handler using the same logic as RunHTTP
	mux := http.NewServeMux()

	// Health check endpoint (no auth required)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"version": server.version,
		})
	})

	// MCP endpoint with authentication
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check authentication using authorizer if set
		if server.authorizer != nil {
			token := r.Header.Get("Authorization")
			authorized, err := server.authorizer.Authorize(r.Context(), token)
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

		response := server.handleMessage(body)
		if response != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	})

	httpServer := httptest.NewServer(mux)
	return server, httpServer
}

// TestHTTPHealthEndpoint tests that GET /health returns 200 with status ok
func TestHTTPHealthEndpoint(t *testing.T) {
	_, httpServer := createTestServer(t, nil)
	defer httpServer.Close()

	resp, err := http.Get(httpServer.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	// Parse and verify response body
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %v", result["status"])
	}

	if result["version"] != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %v", result["version"])
	}
}

// StrictMockAuthorizer is a mock that requires a specific token
type StrictMockAuthorizer struct {
	RequiredToken string
}

func (m *StrictMockAuthorizer) Authorize(ctx interface{}, token string) (bool, error) {
	if m.RequiredToken == "" {
		return true, nil
	}
	return token == m.RequiredToken, nil
}

// TestHTTPAuthMiddleware_MissingHeader tests that POST / without Authorization header returns 401
func TestHTTPAuthMiddleware_MissingHeader(t *testing.T) {
	// Use a strict authorizer that requires a token
	authorizer := auth.NewTokenAuthorizer("secret-token")
	_, httpServer := createTestServer(t, authorizer)
	defer httpServer.Close()

	// Create a valid JSON-RPC request
	reqBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`

	// Make request WITHOUT Authorization header
	resp, err := http.Post(httpServer.URL+"/", "application/json", bytes.NewBufferString(reqBody))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Should return 401 Unauthorized
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}

	// Parse response to verify error structure
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["jsonrpc"] != "2.0" {
		t.Errorf("Expected jsonrpc '2.0', got %v", result["jsonrpc"])
	}

	errorObj, ok := result["error"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected error object in response")
	}

	if errorObj["code"].(float64) != -32001 {
		t.Errorf("Expected error code -32001, got %v", errorObj["code"])
	}
}

// TestHTTPAuthMiddleware_WithHeader tests that POST / with valid Authorization header proceeds
func TestHTTPAuthMiddleware_WithHeader(t *testing.T) {
	// Use MockAuthorizer which always authorizes
	authorizer := &auth.MockAuthorizer{}
	_, httpServer := createTestServer(t, authorizer)
	defer httpServer.Close()

	// Create a valid JSON-RPC request
	reqBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`

	// Create request with Authorization header
	req, err := http.NewRequest("POST", httpServer.URL+"/", bytes.NewBufferString(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer any-token")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Should return 200 OK (not 401)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	// Parse response to verify it's a valid JSON-RPC response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["jsonrpc"] != "2.0" {
		t.Errorf("Expected jsonrpc '2.0', got %v", result["jsonrpc"])
	}

	// Should have result, not error
	if result["error"] != nil {
		t.Errorf("Expected no error, got %v", result["error"])
	}

	if result["result"] == nil {
		t.Error("Expected result in response")
	}
}

// TestHTTPMCPInitialize tests that a valid JSON-RPC initialize request returns a valid response
func TestHTTPMCPInitialize(t *testing.T) {
	// No authorizer for this test
	_, httpServer := createTestServer(t, nil)
	defer httpServer.Close()

	// Create a valid initialize request
	initRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	reqBody, err := json.Marshal(initRequest)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	resp, err := http.Post(httpServer.URL+"/", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify JSON-RPC structure
	if result["jsonrpc"] != "2.0" {
		t.Errorf("Expected jsonrpc '2.0', got %v", result["jsonrpc"])
	}

	if result["id"].(float64) != 1 {
		t.Errorf("Expected id 1, got %v", result["id"])
	}

	// Verify result structure
	initResult, ok := result["result"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected result object in response")
	}

	if initResult["protocolVersion"] != "2024-11-05" {
		t.Errorf("Expected protocolVersion '2024-11-05', got %v", initResult["protocolVersion"])
	}

	serverInfo, ok := initResult["serverInfo"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected serverInfo in result")
	}

	if serverInfo["name"] != "test-server" {
		t.Errorf("Expected server name 'test-server', got %v", serverInfo["name"])
	}

	if serverInfo["version"] != "1.0.0" {
		t.Errorf("Expected server version '1.0.0', got %v", serverInfo["version"])
	}

	// Verify capabilities
	capabilities, ok := initResult["capabilities"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected capabilities in result")
	}

	if capabilities["tools"] == nil {
		t.Error("Expected tools capability to be set")
	}
}

// TestHTTPMCPToolsList tests that tools/list returns the list of registered tools
func TestHTTPMCPToolsList(t *testing.T) {
	// No authorizer for this test
	_, httpServer := createTestServer(t, nil)
	defer httpServer.Close()

	// Create a tools/list request
	listRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
		"params":  map[string]interface{}{},
	}

	reqBody, err := json.Marshal(listRequest)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	resp, err := http.Post(httpServer.URL+"/", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify JSON-RPC structure
	if result["jsonrpc"] != "2.0" {
		t.Errorf("Expected jsonrpc '2.0', got %v", result["jsonrpc"])
	}

	if result["id"].(float64) != 2 {
		t.Errorf("Expected id 2, got %v", result["id"])
	}

	// Verify result structure
	listResult, ok := result["result"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected result object in response")
	}

	tools, ok := listResult["tools"].([]interface{})
	if !ok {
		t.Fatal("Expected tools array in result")
	}

	// Should have at least the test_tool we registered
	if len(tools) < 1 {
		t.Fatal("Expected at least one tool in list")
	}

	// Find and verify the test_tool
	foundTestTool := false
	for _, toolInterface := range tools {
		tool, ok := toolInterface.(map[string]interface{})
		if !ok {
			continue
		}
		if tool["name"] == "test_tool" {
			foundTestTool = true
			if tool["description"] != "A test tool for integration tests" {
				t.Errorf("Expected test_tool description, got %v", tool["description"])
			}
			break
		}
	}

	if !foundTestTool {
		t.Error("Expected to find test_tool in tools list")
	}
}

// TestHTTPMCPToolsCall tests that tools/call executes a tool and returns the result
func TestHTTPMCPToolsCall(t *testing.T) {
	// No authorizer for this test
	_, httpServer := createTestServer(t, nil)
	defer httpServer.Close()

	// Create a tools/call request
	callRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "test_tool",
			"arguments": map[string]interface{}{
				"message": "hello world",
			},
		},
	}

	reqBody, err := json.Marshal(callRequest)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	resp, err := http.Post(httpServer.URL+"/", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify JSON-RPC structure
	if result["jsonrpc"] != "2.0" {
		t.Errorf("Expected jsonrpc '2.0', got %v", result["jsonrpc"])
	}

	// Verify result structure
	callResult, ok := result["result"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected result object in response")
	}

	content, ok := callResult["content"].([]interface{})
	if !ok {
		t.Fatal("Expected content array in result")
	}

	if len(content) < 1 {
		t.Fatal("Expected at least one content item")
	}

	contentItem, ok := content[0].(map[string]interface{})
	if !ok {
		t.Fatal("Expected content item to be an object")
	}

	if contentItem["type"] != "text" {
		t.Errorf("Expected content type 'text', got %v", contentItem["type"])
	}

	if contentItem["text"] != "Received: hello world" {
		t.Errorf("Expected content text 'Received: hello world', got %v", contentItem["text"])
	}
}

// TestHTTPMethodNotAllowed tests that non-POST requests to / return 405
func TestHTTPMethodNotAllowed(t *testing.T) {
	_, httpServer := createTestServer(t, nil)
	defer httpServer.Close()

	resp, err := http.Get(httpServer.URL + "/")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", resp.StatusCode)
	}
}

// TestHTTPInvalidJSON tests that invalid JSON returns a parse error
func TestHTTPInvalidJSON(t *testing.T) {
	_, httpServer := createTestServer(t, nil)
	defer httpServer.Close()

	// Send invalid JSON
	resp, err := http.Post(httpServer.URL+"/", "application/json", bytes.NewBufferString("not valid json"))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Should still return 200 with JSON-RPC error
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200 with JSON-RPC error, got %d. Body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify error structure
	if result["jsonrpc"] != "2.0" {
		t.Errorf("Expected jsonrpc '2.0', got %v", result["jsonrpc"])
	}

	errorObj, ok := result["error"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected error object in response")
	}

	// Parse error code is -32700
	if errorObj["code"].(float64) != -32700 {
		t.Errorf("Expected error code -32700 (parse error), got %v", errorObj["code"])
	}
}

// TestHTTPUnknownMethod tests that unknown JSON-RPC methods return method not found error
func TestHTTPUnknownMethod(t *testing.T) {
	_, httpServer := createTestServer(t, nil)
	defer httpServer.Close()

	// Send request with unknown method
	unknownRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      99,
		"method":  "unknown/method",
		"params":  map[string]interface{}{},
	}

	reqBody, err := json.Marshal(unknownRequest)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	resp, err := http.Post(httpServer.URL+"/", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify error structure
	errorObj, ok := result["error"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected error object in response")
	}

	// Method not found error code is -32601
	if errorObj["code"].(float64) != -32601 {
		t.Errorf("Expected error code -32601 (method not found), got %v", errorObj["code"])
	}
}
