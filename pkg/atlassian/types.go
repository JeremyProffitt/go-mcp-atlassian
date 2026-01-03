// Package atlassian provides common types and utilities for Atlassian API clients.
package atlassian

import (
	"encoding/json"
	"time"
)

// User represents an Atlassian user.
type User struct {
	AccountID   string `json:"accountId,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Email       string `json:"emailAddress,omitempty"`
	Active      bool   `json:"active,omitempty"`
	AvatarURL   string `json:"avatarUrls,omitempty"`
	Self        string `json:"self,omitempty"`
	TimeZone    string `json:"timeZone,omitempty"`
}

// Label represents a label/tag used in Jira or Confluence.
type Label struct {
	ID     string `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Prefix string `json:"prefix,omitempty"`
}

// Attachment represents a file attachment.
type Attachment struct {
	ID       string    `json:"id,omitempty"`
	Filename string    `json:"filename,omitempty"`
	Size     int64     `json:"size,omitempty"`
	MimeType string    `json:"mimeType,omitempty"`
	URL      string    `json:"content,omitempty"`
	Author   *User     `json:"author,omitempty"`
	Created  time.Time `json:"created,omitempty"`
}

// Comment represents a comment on an issue or page.
type Comment struct {
	ID      string    `json:"id,omitempty"`
	Author  *User     `json:"author,omitempty"`
	Body    string    `json:"body,omitempty"`
	Created time.Time `json:"created,omitempty"`
	Updated time.Time `json:"updated,omitempty"`
}

// CommentBody represents the body content of a comment (ADF format for Cloud).
type CommentBody struct {
	Type    string          `json:"type,omitempty"`
	Version int             `json:"version,omitempty"`
	Content json.RawMessage `json:"content,omitempty"`
}

// SearchResult represents a paginated search result.
type SearchResult[T any] struct {
	Results    []T   `json:"results,omitempty"`
	StartAt    int   `json:"startAt,omitempty"`
	MaxResults int   `json:"maxResults,omitempty"`
	Total      int   `json:"total,omitempty"`
	IsLast     bool  `json:"isLast,omitempty"`
}

// PageInfo represents pagination information for cursor-based pagination.
type PageInfo struct {
	HasNextPage bool   `json:"hasNextPage,omitempty"`
	EndCursor   string `json:"endCursor,omitempty"`
}

// PaginatedResult represents a cursor-based paginated result.
type PaginatedResult[T any] struct {
	Results  []T      `json:"results,omitempty"`
	PageInfo PageInfo `json:"_links,omitempty"`
}

// APIError represents an error response from the Atlassian API.
type APIError struct {
	StatusCode int      `json:"-"`
	Message    string   `json:"message,omitempty"`
	ErrorKey   string   `json:"errorKey,omitempty"`
	Errors     []string `json:"errorMessages,omitempty"`
	Details    []struct {
		Field   string `json:"field,omitempty"`
		Message string `json:"message,omitempty"`
	} `json:"errors,omitempty"`
}

// Error implements the error interface for APIError.
func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if len(e.Errors) > 0 {
		return e.Errors[0]
	}
	if len(e.Details) > 0 {
		return e.Details[0].Message
	}
	return "unknown API error"
}

// IsNotFound returns true if the error is a 404 Not Found error.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == 404
}

// IsUnauthorized returns true if the error is a 401 Unauthorized error.
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == 401
}

// IsForbidden returns true if the error is a 403 Forbidden error.
func (e *APIError) IsForbidden() bool {
	return e.StatusCode == 403
}

// IsRateLimited returns true if the error is a 429 Too Many Requests error.
func (e *APIError) IsRateLimited() bool {
	return e.StatusCode == 429
}

// Link represents a HAL-style link.
type Link struct {
	Href string `json:"href,omitempty"`
}

// Links represents a collection of HAL-style links.
type Links struct {
	Self    string `json:"self,omitempty"`
	Base    string `json:"base,omitempty"`
	Context string `json:"context,omitempty"`
	Next    string `json:"next,omitempty"`
	Prev    string `json:"prev,omitempty"`
}

// Icon represents an icon with URL and dimensions.
type Icon struct {
	URL    string `json:"url16x16,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

// Version represents a version in Jira or Confluence.
type Version struct {
	ID          string    `json:"id,omitempty"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	Archived    bool      `json:"archived,omitempty"`
	Released    bool      `json:"released,omitempty"`
	ReleaseDate string    `json:"releaseDate,omitempty"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
}

// Space represents a Confluence space summary.
type Space struct {
	ID          string `json:"id,omitempty"`
	Key         string `json:"key,omitempty"`
	Name        string `json:"name,omitempty"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
}

// Project represents a Jira project summary.
type Project struct {
	ID          string `json:"id,omitempty"`
	Key         string `json:"key,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	ProjectType string `json:"projectTypeKey,omitempty"`
	Style       string `json:"style,omitempty"`
	Lead        *User  `json:"lead,omitempty"`
	URL         string `json:"url,omitempty"`
	AvatarURLs  struct {
		Large  string `json:"48x48,omitempty"`
		Medium string `json:"32x32,omitempty"`
		Small  string `json:"24x24,omitempty"`
		XSmall string `json:"16x16,omitempty"`
	} `json:"avatarUrls,omitempty"`
}

// AvatarURLs represents the avatar URLs for a user or project.
type AvatarURLs struct {
	Size48x48 string `json:"48x48,omitempty"`
	Size32x32 string `json:"32x32,omitempty"`
	Size24x24 string `json:"24x24,omitempty"`
	Size16x16 string `json:"16x16,omitempty"`
}
