package auth

import (
	"context"
	"net/http"
)

// Context keys for credentials
type contextKey string

const (
	// JiraPersonalTokenKey is the context key for Jira personal token.
	JiraPersonalTokenKey contextKey = "jira_personal_token"
	// ConfluencePersonalTokenKey is the context key for Confluence personal token.
	ConfluencePersonalTokenKey contextKey = "confluence_personal_token"
)

// Header names for credentials
const (
	// JiraPersonalTokenHeader is the HTTP header for Jira personal token.
	JiraPersonalTokenHeader = "X-Jira-Personal-Token"
	// ConfluencePersonalTokenHeader is the HTTP header for Confluence personal token.
	ConfluencePersonalTokenHeader = "X-Confluence-Personal-Token"
)

// CredentialsMiddleware extracts credentials from HTTP headers and adds them to the request context.
type CredentialsMiddleware struct{}

// NewCredentialsMiddleware creates a new credentials middleware.
func NewCredentialsMiddleware() *CredentialsMiddleware {
	return &CredentialsMiddleware{}
}

// InjectCredentials wraps an http.Handler to extract credentials from headers.
func (m *CredentialsMiddleware) InjectCredentials(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract Jira personal token from header
		if token := r.Header.Get(JiraPersonalTokenHeader); token != "" {
			ctx = context.WithValue(ctx, JiraPersonalTokenKey, token)
		}

		// Extract Confluence personal token from header
		if token := r.Header.Get(ConfluencePersonalTokenHeader); token != "" {
			ctx = context.WithValue(ctx, ConfluencePersonalTokenKey, token)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetJiraPersonalToken retrieves the Jira personal token from context.
// Returns empty string if not present.
func GetJiraPersonalToken(ctx context.Context) string {
	if token, ok := ctx.Value(JiraPersonalTokenKey).(string); ok {
		return token
	}
	return ""
}

// GetConfluencePersonalToken retrieves the Confluence personal token from context.
// Returns empty string if not present.
func GetConfluencePersonalToken(ctx context.Context) string {
	if token, ok := ctx.Value(ConfluencePersonalTokenKey).(string); ok {
		return token
	}
	return ""
}
