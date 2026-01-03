// Package tools provides MCP tool implementations for Atlassian services.
package tools

import (
	"encoding/json"
	"fmt"

	"github.com/nicholasgriffintn/go-mcp-atlassian/pkg/mcp"
)

// getString extracts a string value from args with a default.
func getString(args map[string]interface{}, key string, defaultVal string) string {
	if val, ok := args[key].(string); ok {
		return val
	}
	return defaultVal
}

// getInt extracts an int value from args with a default.
func getInt(args map[string]interface{}, key string, defaultVal int) int {
	if val, ok := args[key].(float64); ok {
		return int(val)
	}
	return defaultVal
}

// getBool extracts a bool value from args with a default.
func getBool(args map[string]interface{}, key string, defaultVal bool) bool {
	if val, ok := args[key].(bool); ok {
		return val
	}
	return defaultVal
}

// getStringArray extracts a string array from args.
func getStringArray(args map[string]interface{}, key string) []string {
	if val, ok := args[key].([]interface{}); ok {
		result := make([]string, 0, len(val))
		for _, v := range val {
			if s, ok := v.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}

// getMap extracts a map from args.
func getMap(args map[string]interface{}, key string) map[string]interface{} {
	if val, ok := args[key].(map[string]interface{}); ok {
		return val
	}
	return nil
}

// textResult creates a successful text result.
func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.ContentItem{{Type: "text", Text: text}},
	}
}

// jsonResult creates a successful JSON result.
func jsonResult(v interface{}) *mcp.CallToolResult {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return errorResult(fmt.Sprintf("Failed to marshal result: %s", err.Error()))
	}
	return &mcp.CallToolResult{
		Content: []mcp.ContentItem{{Type: "text", Text: string(data)}},
	}
}

// errorResult creates an error result.
func errorResult(message string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.ContentItem{{Type: "text", Text: message}},
		IsError: true,
	}
}

// writeBlockedResult creates a result indicating write operations are blocked.
func writeBlockedResult() *mcp.CallToolResult {
	return errorResult("This operation is not allowed in read-only mode. Set READ_ONLY_MODE=false to enable write operations.")
}

// formatJSON formats a value as indented JSON.
func formatJSON(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(data)
}
