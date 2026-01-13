package atlassian

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/nicholasgriffintn/go-mcp-atlassian/pkg/auth"
)

// AuthType represents the type of authentication to use.
type AuthType int

const (
	// AuthTypeBasic uses Basic Authentication (username + API token for Cloud).
	AuthTypeBasic AuthType = iota
	// AuthTypeBearer uses Bearer token authentication (Personal Access Token for Server/DC).
	AuthTypeBearer
)

// Config holds the base configuration for Atlassian API clients.
type Config struct {
	// URL is the base URL of the Atlassian instance.
	URL string

	// Username is the username for Basic Authentication (Cloud).
	Username string

	// APIToken is the API token for Basic Authentication (Cloud).
	APIToken string

	// PersonalToken is the Personal Access Token for Bearer Authentication (Server/DC).
	PersonalToken string

	// SSLVerify determines whether to verify SSL certificates.
	SSLVerify bool

	// IsCloud indicates whether this is a Cloud instance.
	IsCloud bool

	// ReadOnlyMode indicates whether to run in read-only mode.
	ReadOnlyMode bool
}

// AuthType returns the authentication type based on the configuration.
func (c *Config) AuthType() AuthType {
	if c.PersonalToken != "" {
		return AuthTypeBearer
	}
	return AuthTypeBasic
}

// IsCloudInstance checks if the URL is an Atlassian Cloud instance.
// Cloud instances use *.atlassian.net domains.
func IsCloudInstance(url string) bool {
	return strings.Contains(strings.ToLower(url), ".atlassian.net")
}

// JiraConfig holds the configuration for the Jira API client.
type JiraConfig struct {
	Config

	// ProjectsFilter is a comma-separated list of project keys to filter.
	ProjectsFilter []string

	// UseAPILatest uses /rest/api/latest instead of versioned API paths.
	UseAPILatest bool
}

// ConfluenceConfig holds the configuration for the Confluence API client.
type ConfluenceConfig struct {
	Config

	// SpacesFilter is a comma-separated list of space keys to filter.
	SpacesFilter []string

	// UseAPILatest uses /rest/api/latest instead of versioned API paths.
	UseAPILatest bool
}

// NewJiraConfig creates a new JiraConfig from environment variables.
// Environment variables:
//   - JIRA_URL: The base URL of the Jira instance (required)
//   - JIRA_USERNAME: Username for Basic Auth (Cloud)
//   - JIRA_API_TOKEN: API token for Basic Auth (Cloud)
//   - JIRA_PERSONAL_TOKEN: Personal Access Token for Bearer Auth (Server/DC)
//   - JIRA_SSL_VERIFY: Whether to verify SSL certificates (default: true)
//   - JIRA_PROJECTS_FILTER: Comma-separated list of project keys to filter
//   - JIRA_USE_API_LATEST: Set to 1 to use /rest/api/latest instead of versioned API paths (default: 0)
//   - READ_ONLY_MODE: Whether to run in read-only mode (default: false)
func NewJiraConfig() *JiraConfig {
	url := os.Getenv("JIRA_URL")
	sslVerify := true
	if v := os.Getenv("JIRA_SSL_VERIFY"); v != "" {
		sslVerify, _ = strconv.ParseBool(v)
	}

	readOnlyMode := false
	if v := os.Getenv("READ_ONLY_MODE"); v != "" {
		readOnlyMode, _ = strconv.ParseBool(v)
	}

	useAPILatest := false
	if v := os.Getenv("JIRA_USE_API_LATEST"); v == "1" {
		useAPILatest = true
	}

	var projectsFilter []string
	if v := os.Getenv("JIRA_PROJECTS_FILTER"); v != "" {
		projectsFilter = splitAndTrim(v, ",")
	}

	return &JiraConfig{
		Config: Config{
			URL:           url,
			Username:      os.Getenv("JIRA_USERNAME"),
			APIToken:      os.Getenv("JIRA_API_TOKEN"),
			PersonalToken: os.Getenv("JIRA_PERSONAL_TOKEN"),
			SSLVerify:     sslVerify,
			IsCloud:       IsCloudInstance(url),
			ReadOnlyMode:  readOnlyMode,
		},
		ProjectsFilter: projectsFilter,
		UseAPILatest:   useAPILatest,
	}
}

// NewConfluenceConfig creates a new ConfluenceConfig from environment variables.
// Environment variables:
//   - CONFLUENCE_URL: The base URL of the Confluence instance (required)
//   - CONFLUENCE_USERNAME: Username for Basic Auth (Cloud)
//   - CONFLUENCE_API_TOKEN: API token for Basic Auth (Cloud)
//   - CONFLUENCE_PERSONAL_TOKEN: Personal Access Token for Bearer Auth (Server/DC)
//   - CONFLUENCE_SSL_VERIFY: Whether to verify SSL certificates (default: true)
//   - CONFLUENCE_SPACES_FILTER: Comma-separated list of space keys to filter
//   - CONFLUENCE_USE_API_LATEST: Set to 1 to use /rest/api/latest instead of versioned API paths (default: 0)
//   - READ_ONLY_MODE: Whether to run in read-only mode (default: false)
func NewConfluenceConfig() *ConfluenceConfig {
	url := os.Getenv("CONFLUENCE_URL")
	sslVerify := true
	if v := os.Getenv("CONFLUENCE_SSL_VERIFY"); v != "" {
		sslVerify, _ = strconv.ParseBool(v)
	}

	readOnlyMode := false
	if v := os.Getenv("READ_ONLY_MODE"); v != "" {
		readOnlyMode, _ = strconv.ParseBool(v)
	}

	useAPILatest := false
	if v := os.Getenv("CONFLUENCE_USE_API_LATEST"); v == "1" {
		useAPILatest = true
	}

	var spacesFilter []string
	if v := os.Getenv("CONFLUENCE_SPACES_FILTER"); v != "" {
		spacesFilter = splitAndTrim(v, ",")
	}

	return &ConfluenceConfig{
		Config: Config{
			URL:           url,
			Username:      os.Getenv("CONFLUENCE_USERNAME"),
			APIToken:      os.Getenv("CONFLUENCE_API_TOKEN"),
			PersonalToken: os.Getenv("CONFLUENCE_PERSONAL_TOKEN"),
			SSLVerify:     sslVerify,
			IsCloud:       IsCloudInstance(url),
			ReadOnlyMode:  readOnlyMode,
		},
		SpacesFilter: spacesFilter,
		UseAPILatest: useAPILatest,
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.URL == "" {
		return &APIError{Message: "URL is required"}
	}

	// Check authentication credentials
	if c.PersonalToken == "" && (c.Username == "" || c.APIToken == "") {
		return &APIError{Message: "either PersonalToken or Username+APIToken is required"}
	}

	return nil
}

// Validate checks if the Jira configuration is valid.
func (c *JiraConfig) Validate() error {
	return c.Config.Validate()
}

// Validate checks if the Confluence configuration is valid.
func (c *ConfluenceConfig) Validate() error {
	return c.Config.Validate()
}

// splitAndTrim splits a string by a separator and trims whitespace from each part.
func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// HasProjectFilter checks if a project filter is set.
func (c *JiraConfig) HasProjectFilter() bool {
	return len(c.ProjectsFilter) > 0
}

// IsProjectAllowed checks if a project key is allowed by the filter.
func (c *JiraConfig) IsProjectAllowed(projectKey string) bool {
	if !c.HasProjectFilter() {
		return true
	}
	for _, p := range c.ProjectsFilter {
		if strings.EqualFold(p, projectKey) {
			return true
		}
	}
	return false
}

// HasSpaceFilter checks if a space filter is set.
func (c *ConfluenceConfig) HasSpaceFilter() bool {
	return len(c.SpacesFilter) > 0
}

// IsSpaceAllowed checks if a space key is allowed by the filter.
func (c *ConfluenceConfig) IsSpaceAllowed(spaceKey string) bool {
	if !c.HasSpaceFilter() {
		return true
	}
	for _, s := range c.SpacesFilter {
		if strings.EqualFold(s, spaceKey) {
			return true
		}
	}
	return false
}

// GetJiraPersonalToken returns the Jira personal token from context if available,
// otherwise falls back to the configured token.
func (c *JiraConfig) GetJiraPersonalToken(ctx context.Context) string {
	if token := auth.GetJiraPersonalToken(ctx); token != "" {
		return token
	}
	return c.PersonalToken
}

// GetConfluencePersonalToken returns the Confluence personal token from context if available,
// otherwise falls back to the configured token.
func (c *ConfluenceConfig) GetConfluencePersonalToken(ctx context.Context) string {
	if token := auth.GetConfluencePersonalToken(ctx); token != "" {
		return token
	}
	return c.PersonalToken
}
