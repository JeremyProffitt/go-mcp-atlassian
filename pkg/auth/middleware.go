package auth

import (
	"encoding/json"
	"net/http"
)

// Middleware provides HTTP middleware for authentication and credential injection.
type Middleware struct {
	authorizer Authorizer
}

// NewMiddleware creates a new authentication middleware.
func NewMiddleware(authorizer Authorizer) *Middleware {
	return &Middleware{authorizer: authorizer}
}

// AuthRequired wraps an http.Handler with authorization checks.
// Skips authorization for the /health endpoint.
func (m *Middleware) AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health endpoint
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		// Get authorization token
		token := r.Header.Get("Authorization")

		// Validate token
		authorized, err := m.authorizer.Authorize(r.Context(), token)
		if err != nil {
			writeUnauthorizedError(w, "Authorization error")
			return
		}

		if !authorized {
			writeUnauthorizedError(w, "Unauthorized: invalid or missing authentication token")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// writeUnauthorizedError writes a 401 Unauthorized JSON-RPC error response.
func writeUnauthorizedError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      nil,
		"error":   map[string]interface{}{"code": -32001, "message": message},
	})
}
