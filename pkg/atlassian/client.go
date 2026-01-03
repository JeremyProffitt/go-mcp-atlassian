package atlassian

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nicholasgriffintn/go-mcp-atlassian/pkg/logging"
)

// Client is the base HTTP client for Atlassian APIs.
type Client struct {
	httpClient *http.Client
	baseURL    string
	authHeader string
	authType   AuthType
	logger     *logging.Logger
	isCloud    bool
}

// ClientOption is a function that configures a Client.
type ClientOption func(*Client)

// WithLogger sets a custom logger for the client.
func WithLogger(logger *logging.Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// NewClient creates a new base Atlassian API client.
func NewClient(config *Config, opts ...ClientOption) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Normalize base URL (remove trailing slash)
	baseURL := strings.TrimSuffix(config.URL, "/")

	// Create HTTP client with optional SSL verification
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !config.SSLVerify,
		},
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	client := &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
		authType:   config.AuthType(),
		isCloud:    config.IsCloud,
		logger:     logging.GetLogger(),
	}

	// Set up authentication header
	if config.AuthType() == AuthTypeBearer {
		// Server/DC: Bearer token with Personal Access Token
		client.authHeader = "Bearer " + config.PersonalToken
	} else {
		// Cloud: Basic Auth with username + API token
		auth := config.Username + ":" + config.APIToken
		client.authHeader = "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

// BaseURL returns the base URL of the client.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// IsCloud returns whether the client is connected to a Cloud instance.
func (c *Client) IsCloud() bool {
	return c.isCloud
}

// doRequest performs an HTTP request and returns the response.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	startTime := time.Now()

	// Build full URL
	fullURL := c.baseURL + path

	// Prepare request body
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		duration := time.Since(startTime)
		if c.logger != nil {
			c.logger.APIRequest(method, path, 0, duration, err)
		}
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		if c.logger != nil {
			c.logger.APIRequest(method, path, resp.StatusCode, duration, err)
		}
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Log the request
	if c.logger != nil {
		var logErr error
		if resp.StatusCode >= 400 {
			logErr = fmt.Errorf("status %d", resp.StatusCode)
		}
		c.logger.APIRequest(method, path, resp.StatusCode, duration, logErr)
	}

	// Check for error responses
	if resp.StatusCode >= 400 {
		apiErr := &APIError{
			StatusCode: resp.StatusCode,
		}

		// Try to parse error response
		if len(respBody) > 0 {
			if err := json.Unmarshal(respBody, apiErr); err != nil {
				// If we can't parse the error, use the raw response
				apiErr.Message = string(respBody)
			}
		}

		if apiErr.Message == "" {
			apiErr.Message = fmt.Sprintf("API request failed with status %d", resp.StatusCode)
		}

		return apiErr
	}

	// Parse successful response
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	return c.doRequest(ctx, http.MethodGet, path, nil, result)
}

// GetWithQuery performs a GET request with query parameters.
func (c *Client) GetWithQuery(ctx context.Context, path string, query url.Values, result interface{}) error {
	if len(query) > 0 {
		path = path + "?" + query.Encode()
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, result)
}

// Post performs a POST request.
func (c *Client) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doRequest(ctx, http.MethodPost, path, body, result)
}

// Put performs a PUT request.
func (c *Client) Put(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doRequest(ctx, http.MethodPut, path, body, result)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

// DeleteWithBody performs a DELETE request with a body.
func (c *Client) DeleteWithBody(ctx context.Context, path string, body interface{}) error {
	return c.doRequest(ctx, http.MethodDelete, path, body, nil)
}

// Patch performs a PATCH request.
func (c *Client) Patch(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doRequest(ctx, http.MethodPatch, path, body, result)
}

// DoRaw performs a raw HTTP request and returns the response body.
// This is useful for endpoints that return non-JSON responses.
func (c *Client) DoRaw(ctx context.Context, method, path string, body io.Reader, contentType string) ([]byte, error) {
	startTime := time.Now()

	// Build full URL
	fullURL := c.baseURL + path

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", c.authHeader)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		duration := time.Since(startTime)
		if c.logger != nil {
			c.logger.APIRequest(method, path, 0, duration, err)
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		if c.logger != nil {
			c.logger.APIRequest(method, path, resp.StatusCode, duration, err)
		}
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Log the request
	if c.logger != nil {
		var logErr error
		if resp.StatusCode >= 400 {
			logErr = fmt.Errorf("status %d", resp.StatusCode)
		}
		c.logger.APIRequest(method, path, resp.StatusCode, duration, logErr)
	}

	// Check for error responses
	if resp.StatusCode >= 400 {
		apiErr := &APIError{
			StatusCode: resp.StatusCode,
		}

		// Try to parse error response as JSON
		if len(respBody) > 0 {
			if err := json.Unmarshal(respBody, apiErr); err != nil {
				apiErr.Message = string(respBody)
			}
		}

		if apiErr.Message == "" {
			apiErr.Message = fmt.Sprintf("API request failed with status %d", resp.StatusCode)
		}

		return nil, apiErr
	}

	return respBody, nil
}

// UploadFile uploads a file to the specified path.
func (c *Client) UploadFile(ctx context.Context, path string, filename string, content io.Reader, contentType string) ([]byte, error) {
	startTime := time.Now()

	// Build full URL
	fullURL := c.baseURL + path

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, content)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-Atlassian-Token", "no-check") // Required for file uploads

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		duration := time.Since(startTime)
		if c.logger != nil {
			c.logger.APIRequest(http.MethodPost, path, 0, duration, err)
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		if c.logger != nil {
			c.logger.APIRequest(http.MethodPost, path, resp.StatusCode, duration, err)
		}
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Log the request
	if c.logger != nil {
		var logErr error
		if resp.StatusCode >= 400 {
			logErr = fmt.Errorf("status %d", resp.StatusCode)
		}
		c.logger.APIRequest(http.MethodPost, path, resp.StatusCode, duration, logErr)
	}

	// Check for error responses
	if resp.StatusCode >= 400 {
		apiErr := &APIError{
			StatusCode: resp.StatusCode,
		}

		if len(respBody) > 0 {
			if err := json.Unmarshal(respBody, apiErr); err != nil {
				apiErr.Message = string(respBody)
			}
		}

		if apiErr.Message == "" {
			apiErr.Message = fmt.Sprintf("API request failed with status %d", resp.StatusCode)
		}

		return nil, apiErr
	}

	return respBody, nil
}
