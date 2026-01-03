// Package confluence provides a client for the Confluence REST API.
package confluence

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/nicholasgriffintn/go-mcp-atlassian/pkg/atlassian"
	"github.com/nicholasgriffintn/go-mcp-atlassian/pkg/logging"
)

// API path constants
const (
	// API version paths
	apiV1Path = "/wiki/rest/api"
	apiV2Path = "/wiki/api/v2"

	// Content endpoints
	contentEndpoint    = "/content"
	spaceEndpoint      = "/space"
	searchEndpoint     = "/search"
	cqlSearchEndpoint  = "/content/search"
	labelEndpoint      = "/label"
	userEndpoint       = "/user"
	groupEndpoint      = "/group"
	pageEndpoint       = "/pages"
	attachmentEndpoint = "/attachment"
)

// Client is the Confluence API client.
type Client struct {
	*atlassian.Client
	config *atlassian.ConfluenceConfig
}

// NewClient creates a new Confluence API client.
func NewClient(config *atlassian.ConfluenceConfig, opts ...atlassian.ClientOption) (*Client, error) {
	baseClient, err := atlassian.NewClient(&config.Config, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		Client: baseClient,
		config: config,
	}, nil
}

// Config returns the Confluence configuration.
func (c *Client) Config() *atlassian.ConfluenceConfig {
	return c.config
}

// apiPath returns the appropriate API path based on version.
func (c *Client) apiV1Path() string {
	return apiV1Path
}

// apiV2Path returns the v2 API path.
func (c *Client) apiV2Path() string {
	return apiV2Path
}

// Content represents a Confluence content item (page, blog post, etc.).
type Content struct {
	ID         string                 `json:"id,omitempty"`
	Type       string                 `json:"type,omitempty"`
	Status     string                 `json:"status,omitempty"`
	Title      string                 `json:"title,omitempty"`
	Space      *Space                 `json:"space,omitempty"`
	Body       *ContentBody           `json:"body,omitempty"`
	Version    *Version               `json:"version,omitempty"`
	Ancestors  []Content              `json:"ancestors,omitempty"`
	Children   *ContentChildren       `json:"children,omitempty"`
	Container  *Container             `json:"container,omitempty"`
	Metadata   *ContentMetadata       `json:"metadata,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
	Links      *ContentLinks          `json:"_links,omitempty"`
	Expandable *ContentExpandable     `json:"_expandable,omitempty"`
	History    *ContentHistory        `json:"history,omitempty"`
}

// ContentBody represents the body of a content item.
type ContentBody struct {
	Storage         *BodyContent `json:"storage,omitempty"`
	View            *BodyContent `json:"view,omitempty"`
	ExportView      *BodyContent `json:"export_view,omitempty"`
	StyledView      *BodyContent `json:"styled_view,omitempty"`
	Editor          *BodyContent `json:"editor,omitempty"`
	Editor2         *BodyContent `json:"editor2,omitempty"`
	AnonymousExport *BodyContent `json:"anonymous_export_view,omitempty"`
	AtlasDocFormat  *BodyContent `json:"atlas_doc_format,omitempty"`
}

// BodyContent represents the content of a body format.
type BodyContent struct {
	Value          string `json:"value,omitempty"`
	Representation string `json:"representation,omitempty"`
}

// Version represents the version of a content item.
type Version struct {
	Number    int             `json:"number,omitempty"`
	When      time.Time       `json:"when,omitempty"`
	By        *atlassian.User `json:"by,omitempty"`
	Message   string          `json:"message,omitempty"`
	MinorEdit bool            `json:"minorEdit,omitempty"`
	Hidden    bool            `json:"hidden,omitempty"`
}

// Space represents a Confluence space.
type Space struct {
	ID          int                    `json:"id,omitempty"`
	Key         string                 `json:"key,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Description *SpaceDescription      `json:"description,omitempty"`
	Homepage    *Content               `json:"homepage,omitempty"`
	Icon        *Icon                  `json:"icon,omitempty"`
	Links       *SpaceLinks            `json:"_links,omitempty"`
	Expandable  *SpaceExpandable       `json:"_expandable,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SpaceDescription represents the description of a space.
type SpaceDescription struct {
	Plain *BodyContent `json:"plain,omitempty"`
	View  *BodyContent `json:"view,omitempty"`
}

// SpaceLinks represents the links in a space.
type SpaceLinks struct {
	WebUI string `json:"webui,omitempty"`
	Self  string `json:"self,omitempty"`
	Base  string `json:"base,omitempty"`
}

// SpaceExpandable represents the expandable fields in a space.
type SpaceExpandable struct {
	Metadata    string `json:"metadata,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Description string `json:"description,omitempty"`
	Homepage    string `json:"homepage,omitempty"`
}

// ContentChildren represents the children of a content item.
type ContentChildren struct {
	Page       *PagedContent `json:"page,omitempty"`
	Comment    *PagedContent `json:"comment,omitempty"`
	Attachment *PagedContent `json:"attachment,omitempty"`
}

// PagedContent represents a paged collection of content.
type PagedContent struct {
	Results []Content        `json:"results,omitempty"`
	Start   int              `json:"start,omitempty"`
	Limit   int              `json:"limit,omitempty"`
	Size    int              `json:"size,omitempty"`
	Links   *PaginationLinks `json:"_links,omitempty"`
}

// PaginationLinks represents pagination links.
type PaginationLinks struct {
	Self    string `json:"self,omitempty"`
	Next    string `json:"next,omitempty"`
	Prev    string `json:"prev,omitempty"`
	Base    string `json:"base,omitempty"`
	Context string `json:"context,omitempty"`
}

// Container represents the container of a content item.
type Container struct {
	ID    int    `json:"id,omitempty"`
	Key   string `json:"key,omitempty"`
	Name  string `json:"name,omitempty"`
	Type  string `json:"type,omitempty"`
	Links struct {
		Self string `json:"self,omitempty"`
	} `json:"_links,omitempty"`
}

// ContentMetadata represents metadata of a content item.
type ContentMetadata struct {
	Labels      *Labels                `json:"labels,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	Frontend    map[string]interface{} `json:"frontend,omitempty"`
	CurrentUser map[string]interface{} `json:"currentuser,omitempty"`
}

// Labels represents a collection of labels.
type Labels struct {
	Results []Label `json:"results,omitempty"`
	Start   int     `json:"start,omitempty"`
	Limit   int     `json:"limit,omitempty"`
	Size    int     `json:"size,omitempty"`
}

// Label represents a content label.
type Label struct {
	ID     string `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Prefix string `json:"prefix,omitempty"`
}

// ContentLinks represents the links in a content item.
type ContentLinks struct {
	Self       string `json:"self,omitempty"`
	Base       string `json:"base,omitempty"`
	Context    string `json:"context,omitempty"`
	WebUI      string `json:"webui,omitempty"`
	Edit       string `json:"edit,omitempty"`
	TinyUI     string `json:"tinyui,omitempty"`
	Collection string `json:"collection,omitempty"`
}

// ContentExpandable represents the expandable fields in a content item.
type ContentExpandable struct {
	ChildTypes   string `json:"childTypes,omitempty"`
	Container    string `json:"container,omitempty"`
	Metadata     string `json:"metadata,omitempty"`
	Operations   string `json:"operations,omitempty"`
	Children     string `json:"children,omitempty"`
	Restrictions string `json:"restrictions,omitempty"`
	History      string `json:"history,omitempty"`
	Ancestors    string `json:"ancestors,omitempty"`
	Body         string `json:"body,omitempty"`
	Version      string `json:"version,omitempty"`
	Descendants  string `json:"descendants,omitempty"`
	Space        string `json:"space,omitempty"`
	Extensions   string `json:"extensions,omitempty"`
}

// ContentHistory represents the history of a content item.
type ContentHistory struct {
	Latest          bool            `json:"latest,omitempty"`
	CreatedBy       *atlassian.User `json:"createdBy,omitempty"`
	CreatedDate     time.Time       `json:"createdDate,omitempty"`
	LastUpdated     *Version        `json:"lastUpdated,omitempty"`
	PreviousVersion *Version        `json:"previousVersion,omitempty"`
	Contributors    *Contributors   `json:"contributors,omitempty"`
	NextVersion     *Version        `json:"nextVersion,omitempty"`
}

// Contributors represents the contributors to a content item.
type Contributors struct {
	Publishers *Publishers `json:"publishers,omitempty"`
}

// Publishers represents the publishers of a content item.
type Publishers struct {
	Users    []atlassian.User `json:"users,omitempty"`
	UserKeys []string         `json:"userKeys,omitempty"`
}

// Icon represents an icon.
type Icon struct {
	Path      string `json:"path,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	IsDefault bool   `json:"isDefault,omitempty"`
}

// SearchResult represents the result of a Confluence search.
type SearchResult struct {
	Results        []SearchResultItem `json:"results,omitempty"`
	Start          int                `json:"start,omitempty"`
	Limit          int                `json:"limit,omitempty"`
	Size           int                `json:"size,omitempty"`
	TotalSize      int                `json:"totalSize,omitempty"`
	CqlQuery       string             `json:"cqlQuery,omitempty"`
	SearchDuration int                `json:"searchDuration,omitempty"`
	Links          *PaginationLinks   `json:"_links,omitempty"`
}

// SearchResultItem represents a single search result item.
type SearchResultItem struct {
	Content              *Content `json:"content,omitempty"`
	Space                *Space   `json:"space,omitempty"`
	Title                string   `json:"title,omitempty"`
	Excerpt              string   `json:"excerpt,omitempty"`
	URL                  string   `json:"url,omitempty"`
	EntityType           string   `json:"entityType,omitempty"`
	LastModified         string   `json:"lastModified,omitempty"`
	FriendlyLastModified string   `json:"friendlyLastModified,omitempty"`
}

// Comment represents a Confluence comment.
type Comment struct {
	ID         string                 `json:"id,omitempty"`
	Type       string                 `json:"type,omitempty"`
	Status     string                 `json:"status,omitempty"`
	Title      string                 `json:"title,omitempty"`
	Body       *ContentBody           `json:"body,omitempty"`
	Version    *Version               `json:"version,omitempty"`
	Container  *Container             `json:"container,omitempty"`
	Ancestors  []Comment              `json:"ancestors,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
	Links      *ContentLinks          `json:"_links,omitempty"`
}

// Attachment represents a Confluence attachment.
type Attachment struct {
	ID         string                `json:"id,omitempty"`
	Type       string                `json:"type,omitempty"`
	Status     string                `json:"status,omitempty"`
	Title      string                `json:"title,omitempty"`
	Metadata   *AttachmentMetadata   `json:"metadata,omitempty"`
	Extensions *AttachmentExtensions `json:"extensions,omitempty"`
	Links      *ContentLinks         `json:"_links,omitempty"`
	Version    *Version              `json:"version,omitempty"`
	Container  *Container            `json:"container,omitempty"`
}

// AttachmentMetadata represents attachment metadata.
type AttachmentMetadata struct {
	MediaType string `json:"mediaType,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

// AttachmentExtensions represents attachment extensions.
type AttachmentExtensions struct {
	MediaType            string `json:"mediaType,omitempty"`
	FileSize             int64  `json:"fileSize,omitempty"`
	Comment              string `json:"comment,omitempty"`
	MediaTypeDescription string `json:"mediaTypeDescription,omitempty"`
}

// Page represents a Confluence page (v2 API).
type Page struct {
	ID         string       `json:"id,omitempty"`
	Status     string       `json:"status,omitempty"`
	Title      string       `json:"title,omitempty"`
	SpaceID    string       `json:"spaceId,omitempty"`
	ParentID   string       `json:"parentId,omitempty"`
	ParentType string       `json:"parentType,omitempty"`
	Position   int          `json:"position,omitempty"`
	AuthorID   string       `json:"authorId,omitempty"`
	OwnerID    string       `json:"ownerId,omitempty"`
	CreatedAt  time.Time    `json:"createdAt,omitempty"`
	Version    *PageVersion `json:"version,omitempty"`
	Body       *PageBody    `json:"body,omitempty"`
	Labels     *PageLabels  `json:"labels,omitempty"`
	Links      *PageLinks   `json:"_links,omitempty"`
}

// PageVersion represents the version of a page.
type PageVersion struct {
	Number    int       `json:"number,omitempty"`
	Message   string    `json:"message,omitempty"`
	MinorEdit bool      `json:"minorEdit,omitempty"`
	AuthorID  string    `json:"authorId,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

// PageBody represents the body of a page.
type PageBody struct {
	Storage        *PageBodyContent `json:"storage,omitempty"`
	AtlasDocFormat *PageBodyContent `json:"atlas_doc_format,omitempty"`
}

// PageBodyContent represents the content of a page body.
type PageBodyContent struct {
	Representation string `json:"representation,omitempty"`
	Value          string `json:"value,omitempty"`
}

// PageLabels represents the labels of a page.
type PageLabels struct {
	Results []Label `json:"results,omitempty"`
	Meta    struct {
		HasMore bool   `json:"hasMore,omitempty"`
		Cursor  string `json:"cursor,omitempty"`
	} `json:"meta,omitempty"`
	Links *PaginationLinks `json:"_links,omitempty"`
}

// PageLinks represents the links of a page.
type PageLinks struct {
	WebUI  string `json:"webui,omitempty"`
	EditUI string `json:"editui,omitempty"`
	TinyUI string `json:"tinyui,omitempty"`
}

// PagesResult represents the result of getting pages.
type PagesResult struct {
	Results []Page           `json:"results,omitempty"`
	Links   *PaginationLinks `json:"_links,omitempty"`
}

// User represents a Confluence user.
type User struct {
	atlassian.User
	Username       string          `json:"username,omitempty"` // Used in Server/DC
	UserKey        string          `json:"userKey,omitempty"`
	ProfilePicture *ProfilePicture `json:"profilePicture,omitempty"`
}

// ProfilePicture represents a user's profile picture.
type ProfilePicture struct {
	Path      string `json:"path,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	IsDefault bool   `json:"isDefault,omitempty"`
}

// GetContent retrieves content by ID.
func (c *Client) GetContent(ctx context.Context, contentID string, opts *GetContentOptions) (*Content, error) {
	path := c.apiV1Path() + contentEndpoint + "/" + contentID

	query := url.Values{}
	if opts != nil {
		if len(opts.Expand) > 0 {
			query.Set("expand", strings.Join(opts.Expand, ","))
		}
		if opts.Status != "" {
			query.Set("status", opts.Status)
		}
		if opts.Version > 0 {
			query.Set("version", strconv.Itoa(opts.Version))
		}
	}

	var content Content
	if err := c.GetWithQuery(ctx, path, query, &content); err != nil {
		return nil, err
	}

	return &content, nil
}

// GetContentOptions represents options for getting content.
type GetContentOptions struct {
	Expand  []string
	Status  string
	Version int
}

// GetContentBySpaceAndTitle retrieves content by space key and title.
func (c *Client) GetContentBySpaceAndTitle(ctx context.Context, spaceKey, title string, contentType string, expand []string) (*Content, error) {
	path := c.apiV1Path() + contentEndpoint

	query := url.Values{}
	query.Set("spaceKey", spaceKey)
	query.Set("title", title)
	if contentType != "" {
		query.Set("type", contentType)
	}
	if len(expand) > 0 {
		query.Set("expand", strings.Join(expand, ","))
	}

	var result PagedContent
	if err := c.GetWithQuery(ctx, path, query, &result); err != nil {
		return nil, err
	}

	if len(result.Results) == 0 {
		return nil, &atlassian.APIError{
			StatusCode: 404,
			Message:    "content not found",
		}
	}

	return &result.Results[0], nil
}

// SearchContent searches for content using CQL.
func (c *Client) SearchContent(ctx context.Context, cql string, opts *SearchOptions) (*SearchResult, error) {
	path := c.apiV1Path() + cqlSearchEndpoint

	query := url.Values{}
	query.Set("cql", cql)
	if opts != nil {
		if opts.Start > 0 {
			query.Set("start", strconv.Itoa(opts.Start))
		}
		if opts.Limit > 0 {
			query.Set("limit", strconv.Itoa(opts.Limit))
		}
		if len(opts.Expand) > 0 {
			query.Set("expand", strings.Join(opts.Expand, ","))
		}
		if opts.Excerpt != "" {
			query.Set("excerpt", opts.Excerpt)
		}
	}

	var result SearchResult
	if err := c.GetWithQuery(ctx, path, query, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// SearchOptions represents options for searching content.
type SearchOptions struct {
	Start   int
	Limit   int
	Expand  []string
	Excerpt string
}

// Search performs a search using the search API.
func (c *Client) Search(ctx context.Context, cql string, opts *SearchOptions) (*SearchResult, error) {
	path := c.apiV1Path() + searchEndpoint

	query := url.Values{}
	query.Set("cql", cql)
	if opts != nil {
		if opts.Start > 0 {
			query.Set("start", strconv.Itoa(opts.Start))
		}
		if opts.Limit > 0 {
			query.Set("limit", strconv.Itoa(opts.Limit))
		}
		if len(opts.Expand) > 0 {
			query.Set("expand", strings.Join(opts.Expand, ","))
		}
		if opts.Excerpt != "" {
			query.Set("excerpt", opts.Excerpt)
		}
	}

	var result SearchResult
	if err := c.GetWithQuery(ctx, path, query, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateContent creates new content.
func (c *Client) CreateContent(ctx context.Context, content *Content) (*Content, error) {
	if c.config.ReadOnlyMode {
		return nil, &atlassian.APIError{Message: "cannot create content in read-only mode"}
	}

	// Check if space is allowed
	if content.Space != nil && c.config.HasSpaceFilter() && !c.config.IsSpaceAllowed(content.Space.Key) {
		return nil, &atlassian.APIError{
			StatusCode: 403,
			Message:    "space not allowed by filter",
		}
	}

	path := c.apiV1Path() + contentEndpoint

	var result Content
	if err := c.Post(ctx, path, content, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateContent updates existing content.
func (c *Client) UpdateContent(ctx context.Context, contentID string, content *Content) (*Content, error) {
	if c.config.ReadOnlyMode {
		return nil, &atlassian.APIError{Message: "cannot update content in read-only mode"}
	}

	path := c.apiV1Path() + contentEndpoint + "/" + contentID

	var result Content
	if err := c.Put(ctx, path, content, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteContent deletes content.
func (c *Client) DeleteContent(ctx context.Context, contentID string) error {
	if c.config.ReadOnlyMode {
		return &atlassian.APIError{Message: "cannot delete content in read-only mode"}
	}

	path := c.apiV1Path() + contentEndpoint + "/" + contentID

	return c.Delete(ctx, path)
}

// GetSpaces retrieves all spaces.
func (c *Client) GetSpaces(ctx context.Context, opts *GetSpacesOptions) ([]Space, error) {
	path := c.apiV1Path() + spaceEndpoint

	query := url.Values{}
	if opts != nil {
		if opts.Start > 0 {
			query.Set("start", strconv.Itoa(opts.Start))
		}
		if opts.Limit > 0 {
			query.Set("limit", strconv.Itoa(opts.Limit))
		}
		if len(opts.Expand) > 0 {
			query.Set("expand", strings.Join(opts.Expand, ","))
		}
		if opts.Type != "" {
			query.Set("type", opts.Type)
		}
		if opts.Status != "" {
			query.Set("status", opts.Status)
		}
		if len(opts.SpaceKey) > 0 {
			query.Set("spaceKey", strings.Join(opts.SpaceKey, ","))
		}
	}

	var result struct {
		Results []Space          `json:"results"`
		Start   int              `json:"start"`
		Limit   int              `json:"limit"`
		Size    int              `json:"size"`
		Links   *PaginationLinks `json:"_links"`
	}

	if err := c.GetWithQuery(ctx, path, query, &result); err != nil {
		return nil, err
	}

	// Filter spaces if filter is configured
	if c.config.HasSpaceFilter() {
		filtered := make([]Space, 0)
		for _, s := range result.Results {
			if c.config.IsSpaceAllowed(s.Key) {
				filtered = append(filtered, s)
			}
		}
		return filtered, nil
	}

	return result.Results, nil
}

// GetSpacesOptions represents options for getting spaces.
type GetSpacesOptions struct {
	Start    int
	Limit    int
	Expand   []string
	Type     string
	Status   string
	SpaceKey []string
}

// GetSpace retrieves a space by key.
func (c *Client) GetSpace(ctx context.Context, spaceKey string, expand []string) (*Space, error) {
	// Check if space is allowed
	if c.config.HasSpaceFilter() && !c.config.IsSpaceAllowed(spaceKey) {
		return nil, &atlassian.APIError{
			StatusCode: 403,
			Message:    "space not allowed by filter",
		}
	}

	path := c.apiV1Path() + spaceEndpoint + "/" + spaceKey

	query := url.Values{}
	if len(expand) > 0 {
		query.Set("expand", strings.Join(expand, ","))
	}

	var space Space
	if err := c.GetWithQuery(ctx, path, query, &space); err != nil {
		return nil, err
	}

	return &space, nil
}

// GetSpaceContent retrieves content in a space.
func (c *Client) GetSpaceContent(ctx context.Context, spaceKey string, opts *GetSpaceContentOptions) (*PagedContent, error) {
	// Check if space is allowed
	if c.config.HasSpaceFilter() && !c.config.IsSpaceAllowed(spaceKey) {
		return nil, &atlassian.APIError{
			StatusCode: 403,
			Message:    "space not allowed by filter",
		}
	}

	path := c.apiV1Path() + spaceEndpoint + "/" + spaceKey + contentEndpoint

	query := url.Values{}
	if opts != nil {
		if opts.Start > 0 {
			query.Set("start", strconv.Itoa(opts.Start))
		}
		if opts.Limit > 0 {
			query.Set("limit", strconv.Itoa(opts.Limit))
		}
		if len(opts.Expand) > 0 {
			query.Set("expand", strings.Join(opts.Expand, ","))
		}
		if opts.Depth != "" {
			query.Set("depth", opts.Depth)
		}
	}

	var result PagedContent
	if err := c.GetWithQuery(ctx, path, query, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetSpaceContentOptions represents options for getting space content.
type GetSpaceContentOptions struct {
	Start  int
	Limit  int
	Expand []string
	Depth  string
}

// GetContentChildren retrieves children of content.
func (c *Client) GetContentChildren(ctx context.Context, contentID string, opts *GetChildrenOptions) (*ContentChildren, error) {
	path := c.apiV1Path() + contentEndpoint + "/" + contentID + "/child"

	query := url.Values{}
	if opts != nil {
		if len(opts.Expand) > 0 {
			query.Set("expand", strings.Join(opts.Expand, ","))
		}
		if opts.ParentVersion > 0 {
			query.Set("parentVersion", strconv.Itoa(opts.ParentVersion))
		}
	}

	var result ContentChildren
	if err := c.GetWithQuery(ctx, path, query, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetChildrenOptions represents options for getting children.
type GetChildrenOptions struct {
	Expand        []string
	ParentVersion int
}

// GetContentChildrenByType retrieves children of a specific type.
func (c *Client) GetContentChildrenByType(ctx context.Context, contentID, childType string, start, limit int, expand []string) (*PagedContent, error) {
	path := c.apiV1Path() + contentEndpoint + "/" + contentID + "/child/" + childType

	query := url.Values{}
	if start > 0 {
		query.Set("start", strconv.Itoa(start))
	}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	if len(expand) > 0 {
		query.Set("expand", strings.Join(expand, ","))
	}

	var result PagedContent
	if err := c.GetWithQuery(ctx, path, query, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetLabels retrieves labels for content.
func (c *Client) GetLabels(ctx context.Context, contentID string, prefix string, start, limit int) (*Labels, error) {
	path := c.apiV1Path() + contentEndpoint + "/" + contentID + labelEndpoint

	query := url.Values{}
	if prefix != "" {
		query.Set("prefix", prefix)
	}
	if start > 0 {
		query.Set("start", strconv.Itoa(start))
	}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}

	var result Labels
	if err := c.GetWithQuery(ctx, path, query, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// AddLabels adds labels to content.
func (c *Client) AddLabels(ctx context.Context, contentID string, labels []Label) (*Labels, error) {
	if c.config.ReadOnlyMode {
		return nil, &atlassian.APIError{Message: "cannot add labels in read-only mode"}
	}

	path := c.apiV1Path() + contentEndpoint + "/" + contentID + labelEndpoint

	var result Labels
	if err := c.Post(ctx, path, labels, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// RemoveLabel removes a label from content.
func (c *Client) RemoveLabel(ctx context.Context, contentID, labelName string) error {
	if c.config.ReadOnlyMode {
		return &atlassian.APIError{Message: "cannot remove label in read-only mode"}
	}

	path := c.apiV1Path() + contentEndpoint + "/" + contentID + labelEndpoint + "/" + labelName

	return c.Delete(ctx, path)
}

// GetComments retrieves comments for content.
func (c *Client) GetComments(ctx context.Context, contentID string, opts *GetCommentsOptions) (*PagedContent, error) {
	path := c.apiV1Path() + contentEndpoint + "/" + contentID + "/child/comment"

	query := url.Values{}
	if opts != nil {
		if opts.Start > 0 {
			query.Set("start", strconv.Itoa(opts.Start))
		}
		if opts.Limit > 0 {
			query.Set("limit", strconv.Itoa(opts.Limit))
		}
		if len(opts.Expand) > 0 {
			query.Set("expand", strings.Join(opts.Expand, ","))
		}
		if opts.Location != "" {
			query.Set("location", opts.Location)
		}
		if opts.Depth != "" {
			query.Set("depth", opts.Depth)
		}
	}

	var result PagedContent
	if err := c.GetWithQuery(ctx, path, query, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetCommentsOptions represents options for getting comments.
type GetCommentsOptions struct {
	Start    int
	Limit    int
	Expand   []string
	Location string
	Depth    string
}

// AddComment adds a comment to content.
func (c *Client) AddComment(ctx context.Context, contentID string, comment *Content) (*Content, error) {
	if c.config.ReadOnlyMode {
		return nil, &atlassian.APIError{Message: "cannot add comment in read-only mode"}
	}

	path := c.apiV1Path() + contentEndpoint

	// Set the container to the parent content
	comment.Type = "comment"
	comment.Container = &Container{
		ID:   0,
		Type: "page",
	}
	// Parse contentID as int if possible
	if id, err := strconv.Atoi(contentID); err == nil {
		comment.Container.ID = id
	}
	comment.Ancestors = []Content{{ID: contentID}}

	var result Content
	if err := c.Post(ctx, path, comment, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetAttachments retrieves attachments for content.
func (c *Client) GetAttachments(ctx context.Context, contentID string, start, limit int, expand []string) (*PagedContent, error) {
	path := c.apiV1Path() + contentEndpoint + "/" + contentID + "/child/attachment"

	query := url.Values{}
	if start > 0 {
		query.Set("start", strconv.Itoa(start))
	}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	if len(expand) > 0 {
		query.Set("expand", strings.Join(expand, ","))
	}

	var result PagedContent
	if err := c.GetWithQuery(ctx, path, query, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetContentHistory retrieves the history of content.
func (c *Client) GetContentHistory(ctx context.Context, contentID string, expand []string) (*ContentHistory, error) {
	path := c.apiV1Path() + contentEndpoint + "/" + contentID + "/history"

	query := url.Values{}
	if len(expand) > 0 {
		query.Set("expand", strings.Join(expand, ","))
	}

	var result ContentHistory
	if err := c.GetWithQuery(ctx, path, query, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetContentVersions retrieves versions of content.
func (c *Client) GetContentVersions(ctx context.Context, contentID string, start, limit int, expand []string) ([]Version, error) {
	path := c.apiV1Path() + contentEndpoint + "/" + contentID + "/version"

	query := url.Values{}
	if start > 0 {
		query.Set("start", strconv.Itoa(start))
	}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	if len(expand) > 0 {
		query.Set("expand", strings.Join(expand, ","))
	}

	var result struct {
		Results []Version        `json:"results"`
		Start   int              `json:"start"`
		Limit   int              `json:"limit"`
		Size    int              `json:"size"`
		Links   *PaginationLinks `json:"_links"`
	}

	if err := c.GetWithQuery(ctx, path, query, &result); err != nil {
		return nil, err
	}

	return result.Results, nil
}

// ConvertContentBody converts content body from one format to another.
func (c *Client) ConvertContentBody(ctx context.Context, from, to, value string, expand []string) (*BodyContent, error) {
	path := c.apiV1Path() + "/contentbody/convert/" + to

	query := url.Values{}
	if len(expand) > 0 {
		query.Set("expand", strings.Join(expand, ","))
	}

	body := map[string]interface{}{
		"value":          value,
		"representation": from,
	}

	var result BodyContent
	if err := c.Post(ctx, path+"?"+query.Encode(), body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetCurrentUser retrieves the current authenticated user.
func (c *Client) GetCurrentUser(ctx context.Context, expand []string) (*User, error) {
	path := c.apiV1Path() + userEndpoint + "/current"

	query := url.Values{}
	if len(expand) > 0 {
		query.Set("expand", strings.Join(expand, ","))
	}

	var user User
	if err := c.GetWithQuery(ctx, path, query, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// SearchUsers searches for users.
func (c *Client) SearchUsers(ctx context.Context, cql string, start, limit int, expand []string) ([]User, error) {
	path := c.apiV1Path() + searchEndpoint + userEndpoint

	query := url.Values{}
	query.Set("cql", cql)
	if start > 0 {
		query.Set("start", strconv.Itoa(start))
	}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	if len(expand) > 0 {
		query.Set("expand", strings.Join(expand, ","))
	}

	var result struct {
		Results []User           `json:"results"`
		Start   int              `json:"start"`
		Limit   int              `json:"limit"`
		Size    int              `json:"size"`
		Links   *PaginationLinks `json:"_links"`
	}

	if err := c.GetWithQuery(ctx, path, query, &result); err != nil {
		return nil, err
	}

	return result.Results, nil
}

// v2 API methods

// GetPage retrieves a page by ID using the v2 API.
func (c *Client) GetPage(ctx context.Context, pageID string, opts *GetPageOptions) (*Page, error) {
	path := c.apiV2Path() + pageEndpoint + "/" + pageID

	query := url.Values{}
	if opts != nil {
		if opts.BodyFormat != "" {
			query.Set("body-format", opts.BodyFormat)
		}
		if opts.GetDraft {
			query.Set("get-draft", "true")
		}
		if opts.Version > 0 {
			query.Set("version", strconv.Itoa(opts.Version))
		}
		if opts.IncludeLabels {
			query.Set("include-labels", "true")
		}
	}

	var page Page
	if err := c.GetWithQuery(ctx, path, query, &page); err != nil {
		return nil, err
	}

	return &page, nil
}

// GetPageOptions represents options for getting a page.
type GetPageOptions struct {
	BodyFormat    string
	GetDraft      bool
	Version       int
	IncludeLabels bool
}

// GetPages retrieves pages using the v2 API.
func (c *Client) GetPages(ctx context.Context, opts *GetPagesOptions) (*PagesResult, error) {
	path := c.apiV2Path() + pageEndpoint

	query := url.Values{}
	if opts != nil {
		if opts.SpaceID != "" {
			query.Set("space-id", opts.SpaceID)
		}
		if opts.Title != "" {
			query.Set("title", opts.Title)
		}
		if opts.Status != "" {
			query.Set("status", opts.Status)
		}
		if opts.BodyFormat != "" {
			query.Set("body-format", opts.BodyFormat)
		}
		if opts.Cursor != "" {
			query.Set("cursor", opts.Cursor)
		}
		if opts.Limit > 0 {
			query.Set("limit", strconv.Itoa(opts.Limit))
		}
		if opts.Sort != "" {
			query.Set("sort", opts.Sort)
		}
	}

	var result PagesResult
	if err := c.GetWithQuery(ctx, path, query, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetPagesOptions represents options for getting pages.
type GetPagesOptions struct {
	SpaceID    string
	Title      string
	Status     string
	BodyFormat string
	Cursor     string
	Limit      int
	Sort       string
}

// CreatePage creates a new page using the v2 API.
func (c *Client) CreatePage(ctx context.Context, page *Page) (*Page, error) {
	if c.config.ReadOnlyMode {
		return nil, &atlassian.APIError{Message: "cannot create page in read-only mode"}
	}

	path := c.apiV2Path() + pageEndpoint

	var result Page
	if err := c.Post(ctx, path, page, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdatePage updates a page using the v2 API.
func (c *Client) UpdatePage(ctx context.Context, pageID string, page *Page) (*Page, error) {
	if c.config.ReadOnlyMode {
		return nil, &atlassian.APIError{Message: "cannot update page in read-only mode"}
	}

	path := c.apiV2Path() + pageEndpoint + "/" + pageID

	var result Page
	if err := c.Put(ctx, path, page, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeletePage deletes a page using the v2 API.
func (c *Client) DeletePage(ctx context.Context, pageID string) error {
	if c.config.ReadOnlyMode {
		return &atlassian.APIError{Message: "cannot delete page in read-only mode"}
	}

	path := c.apiV2Path() + pageEndpoint + "/" + pageID

	return c.Delete(ctx, path)
}

// GetPageChildren retrieves children of a page using the v2 API.
func (c *Client) GetPageChildren(ctx context.Context, pageID string, cursor string, limit int) (*PagesResult, error) {
	path := c.apiV2Path() + pageEndpoint + "/" + pageID + "/children"

	query := url.Values{}
	if cursor != "" {
		query.Set("cursor", cursor)
	}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}

	var result PagesResult
	if err := c.GetWithQuery(ctx, path, query, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetPageAncestors retrieves ancestors of a page using the v2 API.
func (c *Client) GetPageAncestors(ctx context.Context, pageID string, limit int) ([]Page, error) {
	path := c.apiV2Path() + pageEndpoint + "/" + pageID + "/ancestors"

	query := url.Values{}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}

	var result struct {
		Results []Page           `json:"results"`
		Links   *PaginationLinks `json:"_links"`
	}

	if err := c.GetWithQuery(ctx, path, query, &result); err != nil {
		return nil, err
	}

	return result.Results, nil
}

// Helper function for extracting text from storage format content
func ExtractTextFromStorage(storage string) string {
	// Simple HTML tag removal - for a more robust solution, use an HTML parser
	// This is a basic implementation that removes HTML tags
	result := storage

	// Remove script and style contents
	for {
		start := strings.Index(strings.ToLower(result), "<script")
		if start == -1 {
			break
		}
		end := strings.Index(strings.ToLower(result[start:]), "</script>")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+9:]
	}

	for {
		start := strings.Index(strings.ToLower(result), "<style")
		if start == -1 {
			break
		}
		end := strings.Index(strings.ToLower(result[start:]), "</style>")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+8:]
	}

	// Remove remaining tags
	for {
		start := strings.Index(result, "<")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], ">")
		if end == -1 {
			break
		}
		result = result[:start] + " " + result[start+end+1:]
	}

	// Clean up whitespace
	result = strings.Join(strings.Fields(result), " ")

	return strings.TrimSpace(result)
}

// ConvertWikiToStorage converts wiki markup to storage format.
func (c *Client) ConvertWikiToStorage(ctx context.Context, wiki string) (string, error) {
	result, err := c.ConvertContentBody(ctx, "wiki", "storage", wiki, nil)
	if err != nil {
		return "", err
	}
	return result.Value, nil
}

// ConvertStorageToView converts storage format to view format.
func (c *Client) ConvertStorageToView(ctx context.Context, storage string) (string, error) {
	result, err := c.ConvertContentBody(ctx, "storage", "view", storage, nil)
	if err != nil {
		return "", err
	}
	return result.Value, nil
}

// Helper function to log with the logger
func (c *Client) log() *logging.Logger {
	return logging.GetLogger()
}

// CreateContentRequest represents the request to create content
type CreateContentRequest struct {
	Type      string        `json:"type"`
	Title     string        `json:"title"`
	Space     *SpaceKey     `json:"space"`
	Body      *ContentBody  `json:"body"`
	Ancestors []AncestorRef `json:"ancestors,omitempty"`
}

// SpaceKey represents a space reference by key
type SpaceKey struct {
	Key string `json:"key"`
}

// AncestorRef represents an ancestor reference
type AncestorRef struct {
	ID string `json:"id"`
}

// UpdateContentRequest represents the request to update content
type UpdateContentRequest struct {
	Version *VersionUpdate `json:"version"`
	Type    string         `json:"type"`
	Title   string         `json:"title"`
	Body    *ContentBody   `json:"body,omitempty"`
}

// VersionUpdate represents a version update
type VersionUpdate struct {
	Number int `json:"number"`
}

// CreatePageWithContent creates a new page with the given content
func (c *Client) CreatePageWithContent(ctx context.Context, spaceKey, title, content string, parentID string) (*Content, error) {
	if c.config.ReadOnlyMode {
		return nil, &atlassian.APIError{Message: "cannot create page in read-only mode"}
	}

	// Check if space is allowed
	if c.config.HasSpaceFilter() && !c.config.IsSpaceAllowed(spaceKey) {
		return nil, &atlassian.APIError{
			StatusCode: 403,
			Message:    "space not allowed by filter",
		}
	}

	req := &CreateContentRequest{
		Type:  "page",
		Title: title,
		Space: &SpaceKey{Key: spaceKey},
		Body: &ContentBody{
			Storage: &BodyContent{
				Value:          content,
				Representation: "storage",
			},
		},
	}

	if parentID != "" {
		req.Ancestors = []AncestorRef{{ID: parentID}}
	}

	path := c.apiV1Path() + contentEndpoint

	var result Content
	if err := c.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdatePageContent updates an existing page's content
func (c *Client) UpdatePageContent(ctx context.Context, contentID, title, content string, version int) (*Content, error) {
	if c.config.ReadOnlyMode {
		return nil, &atlassian.APIError{Message: "cannot update page in read-only mode"}
	}

	req := &UpdateContentRequest{
		Version: &VersionUpdate{Number: version + 1},
		Type:    "page",
		Title:   title,
		Body: &ContentBody{
			Storage: &BodyContent{
				Value:          content,
				Representation: "storage",
			},
		},
	}

	path := c.apiV1Path() + contentEndpoint + "/" + contentID

	var result Content
	if err := c.Put(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetContentProperty retrieves a content property
func (c *Client) GetContentProperty(ctx context.Context, contentID, propertyKey string) (json.RawMessage, error) {
	path := c.apiV1Path() + contentEndpoint + "/" + contentID + "/property/" + propertyKey

	var result struct {
		ID      string          `json:"id"`
		Key     string          `json:"key"`
		Value   json.RawMessage `json:"value"`
		Version struct {
			Number int `json:"number"`
		} `json:"version"`
	}

	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}

	return result.Value, nil
}

// SetContentProperty sets a content property
func (c *Client) SetContentProperty(ctx context.Context, contentID, propertyKey string, value interface{}, version int) error {
	if c.config.ReadOnlyMode {
		return &atlassian.APIError{Message: "cannot set content property in read-only mode"}
	}

	path := c.apiV1Path() + contentEndpoint + "/" + contentID + "/property/" + propertyKey

	body := map[string]interface{}{
		"key":   propertyKey,
		"value": value,
	}

	if version > 0 {
		body["version"] = map[string]int{"number": version}
		return c.Put(ctx, path, body, nil)
	}

	return c.Post(ctx, path, body, nil)
}

// DeleteContentProperty deletes a content property
func (c *Client) DeleteContentProperty(ctx context.Context, contentID, propertyKey string) error {
	if c.config.ReadOnlyMode {
		return &atlassian.APIError{Message: "cannot delete content property in read-only mode"}
	}

	path := c.apiV1Path() + contentEndpoint + "/" + contentID + "/property/" + propertyKey

	return c.Delete(ctx, path)
}

// GetContentRestrictions retrieves restrictions for content
func (c *Client) GetContentRestrictions(ctx context.Context, contentID string, expand []string) (map[string]interface{}, error) {
	path := c.apiV1Path() + contentEndpoint + "/" + contentID + "/restriction"

	query := url.Values{}
	if len(expand) > 0 {
		query.Set("expand", strings.Join(expand, ","))
	}

	var result map[string]interface{}
	if err := c.GetWithQuery(ctx, path, query, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// WatchContent adds the current user as a watcher of content
func (c *Client) WatchContent(ctx context.Context, contentID string) error {
	if c.config.ReadOnlyMode {
		return &atlassian.APIError{Message: "cannot watch content in read-only mode"}
	}

	path := c.apiV1Path() + userEndpoint + "/watch/content/" + contentID

	return c.Post(ctx, path, nil, nil)
}

// UnwatchContent removes the current user as a watcher of content
func (c *Client) UnwatchContent(ctx context.Context, contentID string) error {
	if c.config.ReadOnlyMode {
		return &atlassian.APIError{Message: "cannot unwatch content in read-only mode"}
	}

	path := c.apiV1Path() + userEndpoint + "/watch/content/" + contentID

	return c.Delete(ctx, path)
}

// IsWatchingContent checks if the current user is watching content
func (c *Client) IsWatchingContent(ctx context.Context, contentID string) (bool, error) {
	path := c.apiV1Path() + userEndpoint + "/watch/content/" + contentID

	var result struct {
		Watching bool `json:"watching"`
	}

	if err := c.Get(ctx, path, &result); err != nil {
		// If we get a 404, the user is not watching
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.IsNotFound() {
			return false, nil
		}
		return false, err
	}

	return result.Watching, nil
}

// MoveContent moves content to a new location
func (c *Client) MoveContent(ctx context.Context, contentID, targetSpaceKey, targetParentID string) (*Content, error) {
	if c.config.ReadOnlyMode {
		return nil, &atlassian.APIError{Message: "cannot move content in read-only mode"}
	}

	// Check if target space is allowed
	if c.config.HasSpaceFilter() && !c.config.IsSpaceAllowed(targetSpaceKey) {
		return nil, &atlassian.APIError{
			StatusCode: 403,
			Message:    "target space not allowed by filter",
		}
	}

	// First get current content to get version
	current, err := c.GetContent(ctx, contentID, &GetContentOptions{
		Expand: []string{"version", "space"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get current content: %w", err)
	}

	// Build update request
	update := &Content{
		Type:  current.Type,
		Title: current.Title,
		Space: &Space{Key: targetSpaceKey},
		Version: &Version{
			Number: current.Version.Number + 1,
		},
	}

	if targetParentID != "" {
		update.Ancestors = []Content{{ID: targetParentID}}
	}

	return c.UpdateContent(ctx, contentID, update)
}

// CopyContent copies content to a new location
func (c *Client) CopyContent(ctx context.Context, contentID, targetSpaceKey, targetParentID, newTitle string) (*Content, error) {
	if c.config.ReadOnlyMode {
		return nil, &atlassian.APIError{Message: "cannot copy content in read-only mode"}
	}

	// Check if target space is allowed
	if c.config.HasSpaceFilter() && !c.config.IsSpaceAllowed(targetSpaceKey) {
		return nil, &atlassian.APIError{
			StatusCode: 403,
			Message:    "target space not allowed by filter",
		}
	}

	path := c.apiV1Path() + contentEndpoint + "/" + contentID + "/copy"

	body := map[string]interface{}{
		"copyAttachments":    true,
		"copyPermissions":    false,
		"copyProperties":     true,
		"copyLabels":         true,
		"copyCustomContents": true,
		"destination": map[string]interface{}{
			"type":  "space",
			"value": targetSpaceKey,
		},
		"pageTitle": newTitle,
	}

	if targetParentID != "" {
		body["destination"] = map[string]interface{}{
			"type":  "parent_page",
			"value": targetParentID,
		}
	}

	var result struct {
		ID string `json:"id"`
	}

	if err := c.Post(ctx, path, body, &result); err != nil {
		return nil, err
	}

	// Get the full content
	return c.GetContent(ctx, result.ID, &GetContentOptions{
		Expand: []string{"body.storage", "version", "space"},
	})
}
