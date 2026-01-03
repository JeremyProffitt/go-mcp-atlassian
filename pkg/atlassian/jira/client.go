// Package jira provides a client for the Jira REST API.
package jira

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
	apiV2Path = "/rest/api/2"
	apiV3Path = "/rest/api/3"

	// Issue endpoints
	issueEndpoint       = "/issue"
	searchEndpoint      = "/search"
	projectEndpoint     = "/project"
	fieldEndpoint       = "/field"
	transitionEndpoint  = "/transitions"
	commentEndpoint     = "/comment"
	attachmentEndpoint  = "/attachment"
	worklogEndpoint     = "/worklog"
	watchersEndpoint    = "/watchers"
	issueTypeEndpoint   = "/issuetype"
	priorityEndpoint    = "/priority"
	statusEndpoint      = "/status"
	resolutionEndpoint  = "/resolution"
	myselfEndpoint      = "/myself"
	userSearchEndpoint  = "/user/search"
	sprintEndpoint      = "/sprint"
	boardEndpoint       = "/board"
	epicEndpoint        = "/epic"
	componentEndpoint   = "/component"
	versionEndpoint     = "/version"
	serverInfoEndpoint  = "/serverInfo"
	permissionEndpoint  = "/mypermissions"
)

// Agile API path
const agileAPIPath = "/rest/agile/1.0"

// Client is the Jira API client.
type Client struct {
	*atlassian.Client
	config *atlassian.JiraConfig
}

// NewClient creates a new Jira API client.
func NewClient(config *atlassian.JiraConfig, opts ...atlassian.ClientOption) (*Client, error) {
	baseClient, err := atlassian.NewClient(&config.Config, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		Client: baseClient,
		config: config,
	}, nil
}

// Config returns the Jira configuration.
func (c *Client) Config() *atlassian.JiraConfig {
	return c.config
}

// apiPath returns the appropriate API path based on whether this is a Cloud instance.
func (c *Client) apiPath() string {
	if c.IsCloud() {
		return apiV3Path
	}
	return apiV2Path
}

// Issue represents a Jira issue.
type Issue struct {
	ID         string                 `json:"id,omitempty"`
	Key        string                 `json:"key,omitempty"`
	Self       string                 `json:"self,omitempty"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
	Changelog  *Changelog             `json:"changelog,omitempty"`
	Expand     string                 `json:"expand,omitempty"`
	Names      map[string]string      `json:"names,omitempty"`
	Renderings map[string]interface{} `json:"renderedFields,omitempty"`
}

// Changelog represents the change history of an issue.
type Changelog struct {
	StartAt    int               `json:"startAt,omitempty"`
	MaxResults int               `json:"maxResults,omitempty"`
	Total      int               `json:"total,omitempty"`
	Histories  []ChangelogEntry  `json:"histories,omitempty"`
}

// ChangelogEntry represents a single changelog entry.
type ChangelogEntry struct {
	ID      string           `json:"id,omitempty"`
	Author  *atlassian.User  `json:"author,omitempty"`
	Created time.Time        `json:"created,omitempty"`
	Items   []ChangelogItem  `json:"items,omitempty"`
}

// ChangelogItem represents a single field change in a changelog entry.
type ChangelogItem struct {
	Field      string `json:"field,omitempty"`
	FieldType  string `json:"fieldtype,omitempty"`
	FieldID    string `json:"fieldId,omitempty"`
	From       string `json:"from,omitempty"`
	FromString string `json:"fromString,omitempty"`
	To         string `json:"to,omitempty"`
	ToString   string `json:"toString,omitempty"`
}

// SearchResult represents the result of a Jira search.
type SearchResult struct {
	Expand     string   `json:"expand,omitempty"`
	StartAt    int      `json:"startAt,omitempty"`
	MaxResults int      `json:"maxResults,omitempty"`
	Total      int      `json:"total,omitempty"`
	Issues     []Issue  `json:"issues,omitempty"`
	WarningMessages []string `json:"warningMessages,omitempty"`
}

// Project represents a Jira project.
type Project struct {
	ID             string                     `json:"id,omitempty"`
	Key            string                     `json:"key,omitempty"`
	Name           string                     `json:"name,omitempty"`
	Self           string                     `json:"self,omitempty"`
	Description    string                     `json:"description,omitempty"`
	Lead           *atlassian.User            `json:"lead,omitempty"`
	ProjectTypeKey string                     `json:"projectTypeKey,omitempty"`
	Style          string                     `json:"style,omitempty"`
	IssueTypes     []IssueType                `json:"issueTypes,omitempty"`
	AvatarUrls     map[string]string          `json:"avatarUrls,omitempty"`
	Components     []Component                `json:"components,omitempty"`
	Versions       []atlassian.Version        `json:"versions,omitempty"`
	URL            string                     `json:"url,omitempty"`
}

// IssueType represents a Jira issue type.
type IssueType struct {
	ID          string `json:"id,omitempty"`
	Self        string `json:"self,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Subtask     bool   `json:"subtask,omitempty"`
	IconURL     string `json:"iconUrl,omitempty"`
	AvatarID    int    `json:"avatarId,omitempty"`
}

// Component represents a Jira project component.
type Component struct {
	ID          string          `json:"id,omitempty"`
	Self        string          `json:"self,omitempty"`
	Name        string          `json:"name,omitempty"`
	Description string          `json:"description,omitempty"`
	Lead        *atlassian.User `json:"lead,omitempty"`
	AssigneeType string         `json:"assigneeType,omitempty"`
	Project     string          `json:"project,omitempty"`
	ProjectID   int             `json:"projectId,omitempty"`
}

// Priority represents a Jira priority.
type Priority struct {
	ID          string `json:"id,omitempty"`
	Self        string `json:"self,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	IconURL     string `json:"iconUrl,omitempty"`
	StatusColor string `json:"statusColor,omitempty"`
}

// Status represents a Jira status.
type Status struct {
	ID             string          `json:"id,omitempty"`
	Self           string          `json:"self,omitempty"`
	Name           string          `json:"name,omitempty"`
	Description    string          `json:"description,omitempty"`
	IconURL        string          `json:"iconUrl,omitempty"`
	StatusCategory *StatusCategory `json:"statusCategory,omitempty"`
}

// StatusCategory represents a Jira status category.
type StatusCategory struct {
	ID        int    `json:"id,omitempty"`
	Self      string `json:"self,omitempty"`
	Key       string `json:"key,omitempty"`
	ColorName string `json:"colorName,omitempty"`
	Name      string `json:"name,omitempty"`
}

// Resolution represents a Jira resolution.
type Resolution struct {
	ID          string `json:"id,omitempty"`
	Self        string `json:"self,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// Transition represents a Jira workflow transition.
type Transition struct {
	ID         string                 `json:"id,omitempty"`
	Name       string                 `json:"name,omitempty"`
	To         *Status                `json:"to,omitempty"`
	HasScreen  bool                   `json:"hasScreen,omitempty"`
	IsGlobal   bool                   `json:"isGlobal,omitempty"`
	IsInitial  bool                   `json:"isInitial,omitempty"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
}

// TransitionsResponse represents the response from the transitions endpoint.
type TransitionsResponse struct {
	Expand      string       `json:"expand,omitempty"`
	Transitions []Transition `json:"transitions,omitempty"`
}

// Comment represents a Jira comment.
type Comment struct {
	ID           string          `json:"id,omitempty"`
	Self         string          `json:"self,omitempty"`
	Author       *atlassian.User `json:"author,omitempty"`
	Body         interface{}     `json:"body,omitempty"` // Can be string (v2) or ADF object (v3)
	UpdateAuthor *atlassian.User `json:"updateAuthor,omitempty"`
	Created      time.Time       `json:"created,omitempty"`
	Updated      time.Time       `json:"updated,omitempty"`
	JSDPublic    bool            `json:"jsdPublic,omitempty"`
}

// CommentsResponse represents the response from the comments endpoint.
type CommentsResponse struct {
	StartAt    int       `json:"startAt,omitempty"`
	MaxResults int       `json:"maxResults,omitempty"`
	Total      int       `json:"total,omitempty"`
	Comments   []Comment `json:"comments,omitempty"`
}

// Worklog represents a Jira worklog entry.
type Worklog struct {
	ID               string          `json:"id,omitempty"`
	Self             string          `json:"self,omitempty"`
	Author           *atlassian.User `json:"author,omitempty"`
	UpdateAuthor     *atlassian.User `json:"updateAuthor,omitempty"`
	Comment          interface{}     `json:"comment,omitempty"` // Can be string (v2) or ADF object (v3)
	Created          time.Time       `json:"created,omitempty"`
	Updated          time.Time       `json:"updated,omitempty"`
	Started          string          `json:"started,omitempty"`
	TimeSpent        string          `json:"timeSpent,omitempty"`
	TimeSpentSeconds int             `json:"timeSpentSeconds,omitempty"`
	IssueID          string          `json:"issueId,omitempty"`
}

// WorklogsResponse represents the response from the worklogs endpoint.
type WorklogsResponse struct {
	StartAt    int       `json:"startAt,omitempty"`
	MaxResults int       `json:"maxResults,omitempty"`
	Total      int       `json:"total,omitempty"`
	Worklogs   []Worklog `json:"worklogs,omitempty"`
}

// Field represents a Jira field.
type Field struct {
	ID          string        `json:"id,omitempty"`
	Key         string        `json:"key,omitempty"`
	Name        string        `json:"name,omitempty"`
	Custom      bool          `json:"custom,omitempty"`
	Orderable   bool          `json:"orderable,omitempty"`
	Navigable   bool          `json:"navigable,omitempty"`
	Searchable  bool          `json:"searchable,omitempty"`
	ClauseNames []string      `json:"clauseNames,omitempty"`
	Schema      *FieldSchema  `json:"schema,omitempty"`
}

// FieldSchema represents the schema of a Jira field.
type FieldSchema struct {
	Type     string `json:"type,omitempty"`
	Items    string `json:"items,omitempty"`
	System   string `json:"system,omitempty"`
	Custom   string `json:"custom,omitempty"`
	CustomID int    `json:"customId,omitempty"`
}

// Board represents a Jira Agile board.
type Board struct {
	ID       int           `json:"id,omitempty"`
	Self     string        `json:"self,omitempty"`
	Name     string        `json:"name,omitempty"`
	Type     string        `json:"type,omitempty"`
	Location *BoardLocation `json:"location,omitempty"`
}

// BoardLocation represents the location of a board.
type BoardLocation struct {
	ProjectID      int    `json:"projectId,omitempty"`
	ProjectName    string `json:"projectName,omitempty"`
	ProjectKey     string `json:"projectKey,omitempty"`
	ProjectTypeKey string `json:"projectTypeKey,omitempty"`
	DisplayName    string `json:"displayName,omitempty"`
	AvatarURI      string `json:"avatarURI,omitempty"`
}

// Sprint represents a Jira Agile sprint.
type Sprint struct {
	ID            int       `json:"id,omitempty"`
	Self          string    `json:"self,omitempty"`
	Name          string    `json:"name,omitempty"`
	State         string    `json:"state,omitempty"`
	StartDate     time.Time `json:"startDate,omitempty"`
	EndDate       time.Time `json:"endDate,omitempty"`
	CompleteDate  time.Time `json:"completeDate,omitempty"`
	OriginBoardID int       `json:"originBoardId,omitempty"`
	Goal          string    `json:"goal,omitempty"`
}

// Epic represents a Jira Agile epic.
type Epic struct {
	ID      int    `json:"id,omitempty"`
	Key     string `json:"key,omitempty"`
	Self    string `json:"self,omitempty"`
	Name    string `json:"name,omitempty"`
	Summary string `json:"summary,omitempty"`
	Done    bool   `json:"done,omitempty"`
	Color   struct {
		Key string `json:"key,omitempty"`
	} `json:"color,omitempty"`
}

// ServerInfo represents Jira server information.
type ServerInfo struct {
	BaseURL        string    `json:"baseUrl,omitempty"`
	Version        string    `json:"version,omitempty"`
	VersionNumbers []int     `json:"versionNumbers,omitempty"`
	DeploymentType string    `json:"deploymentType,omitempty"`
	BuildNumber    int       `json:"buildNumber,omitempty"`
	BuildDate      time.Time `json:"buildDate,omitempty"`
	ServerTime     time.Time `json:"serverTime,omitempty"`
	ScmInfo        string    `json:"scmInfo,omitempty"`
	ServerTitle    string    `json:"serverTitle,omitempty"`
}

// User represents a Jira user.
type User struct {
	atlassian.User
	Key         string `json:"key,omitempty"` // Used in Server/DC
	Name        string `json:"name,omitempty"` // Used in Server/DC
}

// Attachment represents a Jira attachment.
type Attachment struct {
	ID        string          `json:"id,omitempty"`
	Self      string          `json:"self,omitempty"`
	Filename  string          `json:"filename,omitempty"`
	Author    *atlassian.User `json:"author,omitempty"`
	Created   time.Time       `json:"created,omitempty"`
	Size      int64           `json:"size,omitempty"`
	MimeType  string          `json:"mimeType,omitempty"`
	Content   string          `json:"content,omitempty"`
	Thumbnail string          `json:"thumbnail,omitempty"`
}

// GetIssue retrieves an issue by its key or ID.
func (c *Client) GetIssue(ctx context.Context, issueKeyOrID string, opts *GetIssueOptions) (*Issue, error) {
	path := c.apiPath() + issueEndpoint + "/" + issueKeyOrID

	query := url.Values{}
	if opts != nil {
		if len(opts.Fields) > 0 {
			query.Set("fields", strings.Join(opts.Fields, ","))
		}
		if len(opts.Expand) > 0 {
			query.Set("expand", strings.Join(opts.Expand, ","))
		}
		if opts.Properties != "" {
			query.Set("properties", opts.Properties)
		}
	}

	var issue Issue
	if err := c.GetWithQuery(ctx, path, query, &issue); err != nil {
		return nil, err
	}

	return &issue, nil
}

// GetIssueOptions represents options for getting an issue.
type GetIssueOptions struct {
	Fields     []string
	Expand     []string
	Properties string
}

// SearchIssues searches for issues using JQL.
func (c *Client) SearchIssues(ctx context.Context, jql string, opts *SearchOptions) (*SearchResult, error) {
	path := c.apiPath() + searchEndpoint

	body := map[string]interface{}{
		"jql": jql,
	}

	if opts != nil {
		if opts.StartAt > 0 {
			body["startAt"] = opts.StartAt
		}
		if opts.MaxResults > 0 {
			body["maxResults"] = opts.MaxResults
		}
		if len(opts.Fields) > 0 {
			body["fields"] = opts.Fields
		}
		if len(opts.Expand) > 0 {
			body["expand"] = opts.Expand
		}
		if opts.ValidateQuery != "" {
			body["validateQuery"] = opts.ValidateQuery
		}
	}

	var result SearchResult
	if err := c.Post(ctx, path, body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// SearchOptions represents options for searching issues.
type SearchOptions struct {
	StartAt       int
	MaxResults    int
	Fields        []string
	Expand        []string
	ValidateQuery string
}

// CreateIssue creates a new issue.
func (c *Client) CreateIssue(ctx context.Context, issue *Issue) (*Issue, error) {
	if c.config.ReadOnlyMode {
		return nil, &atlassian.APIError{Message: "cannot create issue in read-only mode"}
	}

	path := c.apiPath() + issueEndpoint

	var result Issue
	if err := c.Post(ctx, path, issue, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateIssue updates an existing issue.
func (c *Client) UpdateIssue(ctx context.Context, issueKeyOrID string, update map[string]interface{}) error {
	if c.config.ReadOnlyMode {
		return &atlassian.APIError{Message: "cannot update issue in read-only mode"}
	}

	path := c.apiPath() + issueEndpoint + "/" + issueKeyOrID

	return c.Put(ctx, path, update, nil)
}

// DeleteIssue deletes an issue.
func (c *Client) DeleteIssue(ctx context.Context, issueKeyOrID string, deleteSubtasks bool) error {
	if c.config.ReadOnlyMode {
		return &atlassian.APIError{Message: "cannot delete issue in read-only mode"}
	}

	path := c.apiPath() + issueEndpoint + "/" + issueKeyOrID
	if deleteSubtasks {
		path += "?deleteSubtasks=true"
	}

	return c.Delete(ctx, path)
}

// GetTransitions retrieves available transitions for an issue.
func (c *Client) GetTransitions(ctx context.Context, issueKeyOrID string) (*TransitionsResponse, error) {
	path := c.apiPath() + issueEndpoint + "/" + issueKeyOrID + transitionEndpoint

	var result TransitionsResponse
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// TransitionIssue transitions an issue to a new status.
func (c *Client) TransitionIssue(ctx context.Context, issueKeyOrID string, transitionID string, fields map[string]interface{}) error {
	if c.config.ReadOnlyMode {
		return &atlassian.APIError{Message: "cannot transition issue in read-only mode"}
	}

	path := c.apiPath() + issueEndpoint + "/" + issueKeyOrID + transitionEndpoint

	body := map[string]interface{}{
		"transition": map[string]string{
			"id": transitionID,
		},
	}

	if len(fields) > 0 {
		body["fields"] = fields
	}

	return c.Post(ctx, path, body, nil)
}

// AssignIssue assigns an issue to a user.
func (c *Client) AssignIssue(ctx context.Context, issueKeyOrID string, accountID string) error {
	if c.config.ReadOnlyMode {
		return &atlassian.APIError{Message: "cannot assign issue in read-only mode"}
	}

	path := c.apiPath() + issueEndpoint + "/" + issueKeyOrID + "/assignee"

	body := map[string]interface{}{}
	if c.IsCloud() {
		body["accountId"] = accountID
	} else {
		body["name"] = accountID
	}

	return c.Put(ctx, path, body, nil)
}

// GetComments retrieves comments for an issue.
func (c *Client) GetComments(ctx context.Context, issueKeyOrID string, startAt, maxResults int) (*CommentsResponse, error) {
	path := c.apiPath() + issueEndpoint + "/" + issueKeyOrID + commentEndpoint

	query := url.Values{}
	if startAt > 0 {
		query.Set("startAt", strconv.Itoa(startAt))
	}
	if maxResults > 0 {
		query.Set("maxResults", strconv.Itoa(maxResults))
	}

	var result CommentsResponse
	if err := c.GetWithQuery(ctx, path, query, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// AddComment adds a comment to an issue.
func (c *Client) AddComment(ctx context.Context, issueKeyOrID string, body interface{}) (*Comment, error) {
	if c.config.ReadOnlyMode {
		return nil, &atlassian.APIError{Message: "cannot add comment in read-only mode"}
	}

	path := c.apiPath() + issueEndpoint + "/" + issueKeyOrID + commentEndpoint

	commentBody := map[string]interface{}{
		"body": body,
	}

	var result Comment
	if err := c.Post(ctx, path, commentBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateComment updates a comment.
func (c *Client) UpdateComment(ctx context.Context, issueKeyOrID, commentID string, body interface{}) (*Comment, error) {
	if c.config.ReadOnlyMode {
		return nil, &atlassian.APIError{Message: "cannot update comment in read-only mode"}
	}

	path := c.apiPath() + issueEndpoint + "/" + issueKeyOrID + commentEndpoint + "/" + commentID

	commentBody := map[string]interface{}{
		"body": body,
	}

	var result Comment
	if err := c.Put(ctx, path, commentBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteComment deletes a comment.
func (c *Client) DeleteComment(ctx context.Context, issueKeyOrID, commentID string) error {
	if c.config.ReadOnlyMode {
		return &atlassian.APIError{Message: "cannot delete comment in read-only mode"}
	}

	path := c.apiPath() + issueEndpoint + "/" + issueKeyOrID + commentEndpoint + "/" + commentID

	return c.Delete(ctx, path)
}

// GetWorklogs retrieves worklogs for an issue.
func (c *Client) GetWorklogs(ctx context.Context, issueKeyOrID string, startAt, maxResults int) (*WorklogsResponse, error) {
	path := c.apiPath() + issueEndpoint + "/" + issueKeyOrID + worklogEndpoint

	query := url.Values{}
	if startAt > 0 {
		query.Set("startAt", strconv.Itoa(startAt))
	}
	if maxResults > 0 {
		query.Set("maxResults", strconv.Itoa(maxResults))
	}

	var result WorklogsResponse
	if err := c.GetWithQuery(ctx, path, query, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// AddWorklog adds a worklog to an issue.
func (c *Client) AddWorklog(ctx context.Context, issueKeyOrID string, worklog *Worklog) (*Worklog, error) {
	if c.config.ReadOnlyMode {
		return nil, &atlassian.APIError{Message: "cannot add worklog in read-only mode"}
	}

	path := c.apiPath() + issueEndpoint + "/" + issueKeyOrID + worklogEndpoint

	var result Worklog
	if err := c.Post(ctx, path, worklog, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetWatchers retrieves watchers for an issue.
func (c *Client) GetWatchers(ctx context.Context, issueKeyOrID string) ([]User, error) {
	path := c.apiPath() + issueEndpoint + "/" + issueKeyOrID + watchersEndpoint

	var result struct {
		Self       string `json:"self"`
		IsWatching bool   `json:"isWatching"`
		WatchCount int    `json:"watchCount"`
		Watchers   []User `json:"watchers"`
	}

	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}

	return result.Watchers, nil
}

// AddWatcher adds a watcher to an issue.
func (c *Client) AddWatcher(ctx context.Context, issueKeyOrID string, accountID string) error {
	if c.config.ReadOnlyMode {
		return &atlassian.APIError{Message: "cannot add watcher in read-only mode"}
	}

	path := c.apiPath() + issueEndpoint + "/" + issueKeyOrID + watchersEndpoint

	// For Cloud, send the accountId as a quoted string in the body
	// For Server/DC, send the username
	body := fmt.Sprintf(`"%s"`, accountID)

	return c.Post(ctx, path, json.RawMessage(body), nil)
}

// RemoveWatcher removes a watcher from an issue.
func (c *Client) RemoveWatcher(ctx context.Context, issueKeyOrID string, accountID string) error {
	if c.config.ReadOnlyMode {
		return &atlassian.APIError{Message: "cannot remove watcher in read-only mode"}
	}

	path := c.apiPath() + issueEndpoint + "/" + issueKeyOrID + watchersEndpoint
	if c.IsCloud() {
		path += "?accountId=" + url.QueryEscape(accountID)
	} else {
		path += "?username=" + url.QueryEscape(accountID)
	}

	return c.Delete(ctx, path)
}

// GetProjects retrieves all projects.
func (c *Client) GetProjects(ctx context.Context, opts *GetProjectsOptions) ([]Project, error) {
	path := c.apiPath() + projectEndpoint

	query := url.Values{}
	if opts != nil {
		if opts.StartAt > 0 {
			query.Set("startAt", strconv.Itoa(opts.StartAt))
		}
		if opts.MaxResults > 0 {
			query.Set("maxResults", strconv.Itoa(opts.MaxResults))
		}
		if len(opts.Expand) > 0 {
			query.Set("expand", strings.Join(opts.Expand, ","))
		}
		if opts.Recent > 0 {
			query.Set("recent", strconv.Itoa(opts.Recent))
		}
		if opts.OrderBy != "" {
			query.Set("orderBy", opts.OrderBy)
		}
	}

	var projects []Project
	if err := c.GetWithQuery(ctx, path, query, &projects); err != nil {
		return nil, err
	}

	// Filter projects if filter is configured
	if c.config.HasProjectFilter() {
		filtered := make([]Project, 0)
		for _, p := range projects {
			if c.config.IsProjectAllowed(p.Key) {
				filtered = append(filtered, p)
			}
		}
		return filtered, nil
	}

	return projects, nil
}

// GetProjectsOptions represents options for getting projects.
type GetProjectsOptions struct {
	StartAt    int
	MaxResults int
	Expand     []string
	Recent     int
	OrderBy    string
}

// GetProject retrieves a project by its key or ID.
func (c *Client) GetProject(ctx context.Context, projectKeyOrID string, expand []string) (*Project, error) {
	// Check if project is allowed
	if c.config.HasProjectFilter() && !c.config.IsProjectAllowed(projectKeyOrID) {
		return nil, &atlassian.APIError{
			StatusCode: 403,
			Message:    "project not allowed by filter",
		}
	}

	path := c.apiPath() + projectEndpoint + "/" + projectKeyOrID

	query := url.Values{}
	if len(expand) > 0 {
		query.Set("expand", strings.Join(expand, ","))
	}

	var project Project
	if err := c.GetWithQuery(ctx, path, query, &project); err != nil {
		return nil, err
	}

	return &project, nil
}

// GetIssueTypes retrieves all issue types.
func (c *Client) GetIssueTypes(ctx context.Context) ([]IssueType, error) {
	path := c.apiPath() + issueTypeEndpoint

	var issueTypes []IssueType
	if err := c.Get(ctx, path, &issueTypes); err != nil {
		return nil, err
	}

	return issueTypes, nil
}

// GetPriorities retrieves all priorities.
func (c *Client) GetPriorities(ctx context.Context) ([]Priority, error) {
	path := c.apiPath() + priorityEndpoint

	var priorities []Priority
	if err := c.Get(ctx, path, &priorities); err != nil {
		return nil, err
	}

	return priorities, nil
}

// GetStatuses retrieves all statuses.
func (c *Client) GetStatuses(ctx context.Context) ([]Status, error) {
	path := c.apiPath() + statusEndpoint

	var statuses []Status
	if err := c.Get(ctx, path, &statuses); err != nil {
		return nil, err
	}

	return statuses, nil
}

// GetResolutions retrieves all resolutions.
func (c *Client) GetResolutions(ctx context.Context) ([]Resolution, error) {
	path := c.apiPath() + resolutionEndpoint

	var resolutions []Resolution
	if err := c.Get(ctx, path, &resolutions); err != nil {
		return nil, err
	}

	return resolutions, nil
}

// GetFields retrieves all fields.
func (c *Client) GetFields(ctx context.Context) ([]Field, error) {
	path := c.apiPath() + fieldEndpoint

	var fields []Field
	if err := c.Get(ctx, path, &fields); err != nil {
		return nil, err
	}

	return fields, nil
}

// GetCurrentUser retrieves the current authenticated user.
func (c *Client) GetCurrentUser(ctx context.Context) (*User, error) {
	path := c.apiPath() + myselfEndpoint

	var user User
	if err := c.Get(ctx, path, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// SearchUsers searches for users.
func (c *Client) SearchUsers(ctx context.Context, query string, startAt, maxResults int) ([]User, error) {
	path := c.apiPath() + userSearchEndpoint

	params := url.Values{}
	params.Set("query", query)
	if startAt > 0 {
		params.Set("startAt", strconv.Itoa(startAt))
	}
	if maxResults > 0 {
		params.Set("maxResults", strconv.Itoa(maxResults))
	}

	var users []User
	if err := c.GetWithQuery(ctx, path, params, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// GetServerInfo retrieves server information.
func (c *Client) GetServerInfo(ctx context.Context) (*ServerInfo, error) {
	path := c.apiPath() + serverInfoEndpoint

	var info ServerInfo
	if err := c.Get(ctx, path, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

// GetAttachments retrieves attachments for an issue.
func (c *Client) GetAttachments(ctx context.Context, issueKeyOrID string) ([]Attachment, error) {
	issue, err := c.GetIssue(ctx, issueKeyOrID, &GetIssueOptions{
		Fields: []string{"attachment"},
	})
	if err != nil {
		return nil, err
	}

	attachmentsRaw, ok := issue.Fields["attachment"]
	if !ok {
		return []Attachment{}, nil
	}

	// Marshal and unmarshal to convert the interface{} to []Attachment
	data, err := json.Marshal(attachmentsRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal attachments: %w", err)
	}

	var attachments []Attachment
	if err := json.Unmarshal(data, &attachments); err != nil {
		return nil, fmt.Errorf("failed to unmarshal attachments: %w", err)
	}

	return attachments, nil
}

// LinkIssues creates a link between two issues.
func (c *Client) LinkIssues(ctx context.Context, inwardIssue, outwardIssue, linkType string) error {
	if c.config.ReadOnlyMode {
		return &atlassian.APIError{Message: "cannot link issues in read-only mode"}
	}

	path := c.apiPath() + "/issueLink"

	body := map[string]interface{}{
		"type": map[string]string{
			"name": linkType,
		},
		"inwardIssue": map[string]string{
			"key": inwardIssue,
		},
		"outwardIssue": map[string]string{
			"key": outwardIssue,
		},
	}

	return c.Post(ctx, path, body, nil)
}

// GetLinkTypes retrieves all issue link types.
func (c *Client) GetLinkTypes(ctx context.Context) ([]IssueLinkType, error) {
	path := c.apiPath() + "/issueLinkType"

	var result struct {
		IssueLinkTypes []IssueLinkType `json:"issueLinkTypes"`
	}

	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}

	return result.IssueLinkTypes, nil
}

// IssueLinkType represents an issue link type.
type IssueLinkType struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Inward  string `json:"inward,omitempty"`
	Outward string `json:"outward,omitempty"`
	Self    string `json:"self,omitempty"`
}

// Agile API methods

// GetBoards retrieves all boards.
func (c *Client) GetBoards(ctx context.Context, projectKeyOrID string, startAt, maxResults int) ([]Board, error) {
	path := agileAPIPath + boardEndpoint

	query := url.Values{}
	if projectKeyOrID != "" {
		query.Set("projectKeyOrId", projectKeyOrID)
	}
	if startAt > 0 {
		query.Set("startAt", strconv.Itoa(startAt))
	}
	if maxResults > 0 {
		query.Set("maxResults", strconv.Itoa(maxResults))
	}

	var result struct {
		MaxResults int     `json:"maxResults"`
		StartAt    int     `json:"startAt"`
		Total      int     `json:"total"`
		IsLast     bool    `json:"isLast"`
		Values     []Board `json:"values"`
	}

	if err := c.GetWithQuery(ctx, path, query, &result); err != nil {
		return nil, err
	}

	return result.Values, nil
}

// GetBoard retrieves a board by ID.
func (c *Client) GetBoard(ctx context.Context, boardID int) (*Board, error) {
	path := agileAPIPath + boardEndpoint + "/" + strconv.Itoa(boardID)

	var board Board
	if err := c.Get(ctx, path, &board); err != nil {
		return nil, err
	}

	return &board, nil
}

// GetSprints retrieves sprints for a board.
func (c *Client) GetSprints(ctx context.Context, boardID int, state string, startAt, maxResults int) ([]Sprint, error) {
	path := agileAPIPath + boardEndpoint + "/" + strconv.Itoa(boardID) + sprintEndpoint

	query := url.Values{}
	if state != "" {
		query.Set("state", state)
	}
	if startAt > 0 {
		query.Set("startAt", strconv.Itoa(startAt))
	}
	if maxResults > 0 {
		query.Set("maxResults", strconv.Itoa(maxResults))
	}

	var result struct {
		MaxResults int      `json:"maxResults"`
		StartAt    int      `json:"startAt"`
		Total      int      `json:"total"`
		IsLast     bool     `json:"isLast"`
		Values     []Sprint `json:"values"`
	}

	if err := c.GetWithQuery(ctx, path, query, &result); err != nil {
		return nil, err
	}

	return result.Values, nil
}

// GetSprint retrieves a sprint by ID.
func (c *Client) GetSprint(ctx context.Context, sprintID int) (*Sprint, error) {
	path := agileAPIPath + sprintEndpoint + "/" + strconv.Itoa(sprintID)

	var sprint Sprint
	if err := c.Get(ctx, path, &sprint); err != nil {
		return nil, err
	}

	return &sprint, nil
}

// GetSprintIssues retrieves issues in a sprint.
func (c *Client) GetSprintIssues(ctx context.Context, sprintID int, jql string, startAt, maxResults int) (*SearchResult, error) {
	path := agileAPIPath + sprintEndpoint + "/" + strconv.Itoa(sprintID) + issueEndpoint

	query := url.Values{}
	if jql != "" {
		query.Set("jql", jql)
	}
	if startAt > 0 {
		query.Set("startAt", strconv.Itoa(startAt))
	}
	if maxResults > 0 {
		query.Set("maxResults", strconv.Itoa(maxResults))
	}

	var result SearchResult
	if err := c.GetWithQuery(ctx, path, query, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetEpics retrieves epics for a board.
func (c *Client) GetEpics(ctx context.Context, boardID int, done bool, startAt, maxResults int) ([]Epic, error) {
	path := agileAPIPath + boardEndpoint + "/" + strconv.Itoa(boardID) + epicEndpoint

	query := url.Values{}
	query.Set("done", strconv.FormatBool(done))
	if startAt > 0 {
		query.Set("startAt", strconv.Itoa(startAt))
	}
	if maxResults > 0 {
		query.Set("maxResults", strconv.Itoa(maxResults))
	}

	var result struct {
		MaxResults int    `json:"maxResults"`
		StartAt    int    `json:"startAt"`
		Total      int    `json:"total"`
		IsLast     bool   `json:"isLast"`
		Values     []Epic `json:"values"`
	}

	if err := c.GetWithQuery(ctx, path, query, &result); err != nil {
		return nil, err
	}

	return result.Values, nil
}

// GetEpicIssues retrieves issues in an epic.
func (c *Client) GetEpicIssues(ctx context.Context, epicID string, jql string, startAt, maxResults int) (*SearchResult, error) {
	path := agileAPIPath + epicEndpoint + "/" + epicID + issueEndpoint

	query := url.Values{}
	if jql != "" {
		query.Set("jql", jql)
	}
	if startAt > 0 {
		query.Set("startAt", strconv.Itoa(startAt))
	}
	if maxResults > 0 {
		query.Set("maxResults", strconv.Itoa(maxResults))
	}

	var result SearchResult
	if err := c.GetWithQuery(ctx, path, query, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// MoveIssuesToSprint moves issues to a sprint.
func (c *Client) MoveIssuesToSprint(ctx context.Context, sprintID int, issueKeys []string) error {
	if c.config.ReadOnlyMode {
		return &atlassian.APIError{Message: "cannot move issues to sprint in read-only mode"}
	}

	path := agileAPIPath + sprintEndpoint + "/" + strconv.Itoa(sprintID) + issueEndpoint

	body := map[string]interface{}{
		"issues": issueKeys,
	}

	return c.Post(ctx, path, body, nil)
}

// MoveIssuesToBacklog moves issues to the backlog.
func (c *Client) MoveIssuesToBacklog(ctx context.Context, issueKeys []string) error {
	if c.config.ReadOnlyMode {
		return &atlassian.APIError{Message: "cannot move issues to backlog in read-only mode"}
	}

	path := agileAPIPath + "/backlog/issue"

	body := map[string]interface{}{
		"issues": issueKeys,
	}

	return c.Post(ctx, path, body, nil)
}

// Helper function to log with the logger
func (c *Client) log() *logging.Logger {
	return logging.GetLogger()
}
