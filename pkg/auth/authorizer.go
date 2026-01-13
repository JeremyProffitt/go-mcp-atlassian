package auth

import (
	"context"
)

// Authorizer defines the interface for authorization logic.
type Authorizer interface {
	// Authorize validates the provided token and returns true if authorized.
	Authorize(ctx context.Context, token string) (bool, error)
}

// MockAuthorizer is a mock implementation that always authorizes.
type MockAuthorizer struct{}

// Authorize always returns true for MockAuthorizer.
func (m *MockAuthorizer) Authorize(ctx context.Context, token string) (bool, error) {
	return true, nil
}

// TokenAuthorizer validates tokens against an expected value.
type TokenAuthorizer struct {
	expectedToken string
}

// NewTokenAuthorizer creates a new TokenAuthorizer with the expected token.
func NewTokenAuthorizer(expectedToken string) *TokenAuthorizer {
	return &TokenAuthorizer{expectedToken: expectedToken}
}

// Authorize checks if the provided token matches the expected token.
func (t *TokenAuthorizer) Authorize(ctx context.Context, token string) (bool, error) {
	if t.expectedToken == "" {
		// No expected token configured, authorize all
		return true, nil
	}
	return token == t.expectedToken, nil
}
