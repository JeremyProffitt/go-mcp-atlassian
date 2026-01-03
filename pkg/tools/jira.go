// Package tools provides MCP tool implementations for Atlassian services.
package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nicholasgriffintn/go-mcp-atlassian/pkg/atlassian/jira"
	"github.com/nicholasgriffintn/go-mcp-atlassian/pkg/logging"
	"github.com/nicholasgriffintn/go-mcp-atlassian/pkg/mcp"
)

// JiraRegistry holds references to the Jira client and configuration for tool registration.
type JiraRegistry struct {
	jiraClient   *jira.Client
	logger       *logging.Logger
	readOnlyMode bool
}

// NewJiraRegistry creates a new Jira tool registry.
func NewJiraRegistry(jiraClient *jira.Client, logger *logging.Logger, readOnlyMode bool) *JiraRegistry {
	return &JiraRegistry{
		jiraClient:   jiraClient,
		logger:       logger,
		readOnlyMode: readOnlyMode,
	}
}

// RegisterAll registers all Jira tools with the MCP server.
func (r *JiraRegistry) RegisterAll(server *mcp.Server) {
	// Read tools
	r.registerGetUserProfile(server)
	r.registerGetIssue(server)
	r.registerSearch(server)
	r.registerSearchFields(server)
	r.registerGetProjectIssues(server)
	r.registerGetTransitions(server)
	r.registerGetWorklog(server)
	r.registerDownloadAttachments(server)
	r.registerGetAgileBoards(server)
	r.registerGetBoardIssues(server)
	r.registerGetSprintsFromBoard(server)
	r.registerGetSprintIssues(server)
	r.registerGetLinkTypes(server)
	r.registerBatchGetChangelogs(server)
	r.registerGetProjectVersions(server)
	r.registerGetAllProjects(server)

	// Write tools (only if not in read-only mode)
	r.registerCreateIssue(server)
	r.registerBatchCreateIssues(server)
	r.registerUpdateIssue(server)
	r.registerDeleteIssue(server)
	r.registerAddComment(server)
	r.registerEditComment(server)
	r.registerAddWorklog(server)
	r.registerLinkToEpic(server)
	r.registerCreateIssueLink(server)
	r.registerTransitionIssue(server)
}

// readOnlyError returns an error result for write operations in read-only mode.
func (r *JiraRegistry) readOnlyError() *mcp.CallToolResult {
	return writeBlockedResult()
}

// ============================================================================
// READ TOOLS
// ============================================================================

// 1. jira_get_user_profile - Get current authenticated user profile
func (r *JiraRegistry) registerGetUserProfile(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_get_user_profile",
		Description: "Get the current authenticated user's profile information. Returns user details including display name, email, timezone, and account status.",
		InputSchema: mcp.JSONSchema{
			Type:       "object",
			Properties: map[string]mcp.Property{},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "Get User Profile",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_get_user_profile", args, 0, false)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		user, err := r.jiraClient.GetCurrentUser(ctx)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to get user profile: %s", err.Error())), nil
		}

		return textResult(fmt.Sprintf("**User Profile**\n\n%s", formatJSON(user))), nil
	})
}

// 2. jira_get_issue - Get issue details with fields and expand options
func (r *JiraRegistry) registerGetIssue(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_get_issue",
		Description: "Get detailed information about a Jira issue by its key or ID. Can optionally expand specific sections.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"issue_key": {
					Type:        "string",
					Description: "The issue key (e.g., 'PROJ-123') or issue ID.",
				},
				"fields": {
					Type:        "array",
					Description: "Specific fields to return. If not specified, returns all navigable fields.",
					Items:       &mcp.Property{Type: "string"},
				},
				"expand": {
					Type:        "array",
					Description: "Sections to expand (e.g., 'renderedFields', 'changelog', 'transitions', 'names').",
					Items:       &mcp.Property{Type: "string"},
				},
			},
			Required: []string{"issue_key"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "Get Issue",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_get_issue", args, 0, false)

		issueKey := getString(args, "issue_key", "")
		if issueKey == "" {
			return errorResult("issue_key is required"), nil
		}

		fields := getStringArray(args, "fields")
		expand := getStringArray(args, "expand")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		opts := &jira.GetIssueOptions{
			Fields: fields,
			Expand: expand,
		}

		issue, err := r.jiraClient.GetIssue(ctx, issueKey, opts)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to get issue: %s", err.Error())), nil
		}

		return textResult(fmt.Sprintf("**Issue: %s**\n\n%s", issueKey, formatJSON(issue))), nil
	})
}

// 3. jira_search - Search using JQL with pagination
func (r *JiraRegistry) registerSearch(server *mcp.Server) {
	minVal := float64(0)
	maxVal := float64(100)

	server.RegisterTool(mcp.Tool{
		Name:        "jira_search",
		Description: "Search for Jira issues using JQL (Jira Query Language). Supports pagination and field selection.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"jql": {
					Type:        "string",
					Description: "JQL query string (e.g., 'project = PROJ AND status = Open ORDER BY created DESC').",
				},
				"fields": {
					Type:        "array",
					Description: "Fields to return for each issue. Defaults to key, summary, status, assignee.",
					Items:       &mcp.Property{Type: "string"},
				},
				"start_at": {
					Type:        "number",
					Description: "Index of the first result to return (0-based pagination).",
					Default:     0,
					Minimum:     &minVal,
				},
				"max_results": {
					Type:        "number",
					Description: "Maximum number of results to return (default: 50, max: 100).",
					Default:     50,
					Minimum:     &minVal,
					Maximum:     &maxVal,
				},
				"expand": {
					Type:        "array",
					Description: "Sections to expand in each issue.",
					Items:       &mcp.Property{Type: "string"},
				},
			},
			Required: []string{"jql"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "Search Issues",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_search", args, 0, false)

		jql := getString(args, "jql", "")
		if jql == "" {
			return errorResult("jql is required"), nil
		}

		fields := getStringArray(args, "fields")
		if len(fields) == 0 {
			fields = []string{"key", "summary", "status", "assignee", "priority", "created", "updated"}
		}
		startAt := getInt(args, "start_at", 0)
		maxResults := getInt(args, "max_results", 50)
		expand := getStringArray(args, "expand")

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		opts := &jira.SearchOptions{
			Fields:     fields,
			StartAt:    startAt,
			MaxResults: maxResults,
			Expand:     expand,
		}

		result, err := r.jiraClient.SearchIssues(ctx, jql, opts)
		if err != nil {
			return errorResult(fmt.Sprintf("Search failed: %s", err.Error())), nil
		}

		resp := fmt.Sprintf("**Search Results** (showing %d of %d total)\n\n", len(result.Issues), result.Total)
		resp += fmt.Sprintf("JQL: `%s`\n\n", jql)
		resp += formatJSON(result)

		return textResult(resp), nil
	})
}

// 4. jira_search_fields - Search Jira fields by keyword
func (r *JiraRegistry) registerSearchFields(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_search_fields",
		Description: "Search for Jira field definitions by keyword. Useful for finding custom field IDs and understanding available fields.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"keyword": {
					Type:        "string",
					Description: "Keyword to search for in field names and descriptions.",
				},
				"field_type": {
					Type:        "string",
					Description: "Filter by field type ('custom' or 'system').",
					Enum:        []string{"custom", "system", "all"},
					Default:     "all",
				},
			},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "Search Fields",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_search_fields", args, 0, false)

		keyword := getString(args, "keyword", "")
		fieldType := getString(args, "field_type", "all")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		fields, err := r.jiraClient.GetFields(ctx)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to get fields: %s", err.Error())), nil
		}

		// Filter fields
		var filtered []jira.Field
		for _, f := range fields {
			// Filter by type
			if fieldType == "custom" && !f.Custom {
				continue
			}
			if fieldType == "system" && f.Custom {
				continue
			}

			// Filter by keyword
			if keyword != "" {
				keywordLower := strings.ToLower(keyword)
				if !strings.Contains(strings.ToLower(f.Name), keywordLower) &&
					!strings.Contains(strings.ToLower(f.ID), keywordLower) {
					continue
				}
			}

			filtered = append(filtered, f)
		}

		resp := fmt.Sprintf("**Found %d fields**\n\n", len(filtered))
		resp += formatJSON(filtered)

		return textResult(resp), nil
	})
}

// 5. jira_get_project_issues - Get issues for a project
func (r *JiraRegistry) registerGetProjectIssues(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_get_project_issues",
		Description: "Get all issues for a specific Jira project with optional filtering and pagination.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"project_key": {
					Type:        "string",
					Description: "The project key (e.g., 'PROJ').",
				},
				"status": {
					Type:        "string",
					Description: "Filter by status name or category (e.g., 'Open', 'In Progress', 'Done').",
				},
				"issue_type": {
					Type:        "string",
					Description: "Filter by issue type (e.g., 'Bug', 'Story', 'Task').",
				},
				"assignee": {
					Type:        "string",
					Description: "Filter by assignee account ID or 'unassigned' for unassigned issues.",
				},
				"start_at": {
					Type:        "number",
					Description: "Index of the first result to return.",
					Default:     0,
				},
				"max_results": {
					Type:        "number",
					Description: "Maximum number of results to return (default: 50).",
					Default:     50,
				},
			},
			Required: []string{"project_key"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "Get Project Issues",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_get_project_issues", args, 0, false)

		projectKey := getString(args, "project_key", "")
		if projectKey == "" {
			return errorResult("project_key is required"), nil
		}

		status := getString(args, "status", "")
		issueType := getString(args, "issue_type", "")
		assignee := getString(args, "assignee", "")
		startAt := getInt(args, "start_at", 0)
		maxResults := getInt(args, "max_results", 50)

		// Build JQL
		jqlParts := []string{fmt.Sprintf("project = %s", projectKey)}
		if status != "" {
			jqlParts = append(jqlParts, fmt.Sprintf("status = \"%s\"", status))
		}
		if issueType != "" {
			jqlParts = append(jqlParts, fmt.Sprintf("issuetype = \"%s\"", issueType))
		}
		if assignee != "" {
			if assignee == "unassigned" {
				jqlParts = append(jqlParts, "assignee IS EMPTY")
			} else {
				jqlParts = append(jqlParts, fmt.Sprintf("assignee = \"%s\"", assignee))
			}
		}
		jql := strings.Join(jqlParts, " AND ") + " ORDER BY created DESC"

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		opts := &jira.SearchOptions{
			Fields:     []string{"key", "summary", "status", "assignee", "priority", "issuetype", "created", "updated"},
			StartAt:    startAt,
			MaxResults: maxResults,
		}

		result, err := r.jiraClient.SearchIssues(ctx, jql, opts)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to get project issues: %s", err.Error())), nil
		}

		resp := fmt.Sprintf("**Project %s Issues** (showing %d of %d)\n\n", projectKey, len(result.Issues), result.Total)
		resp += formatJSON(result)

		return textResult(resp), nil
	})
}

// 6. jira_get_transitions - Get available status transitions
func (r *JiraRegistry) registerGetTransitions(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_get_transitions",
		Description: "Get available workflow transitions for an issue. Shows what status changes are possible for the current user.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"issue_key": {
					Type:        "string",
					Description: "The issue key (e.g., 'PROJ-123').",
				},
			},
			Required: []string{"issue_key"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "Get Transitions",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_get_transitions", args, 0, false)

		issueKey := getString(args, "issue_key", "")
		if issueKey == "" {
			return errorResult("issue_key is required"), nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		transitions, err := r.jiraClient.GetTransitions(ctx, issueKey)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to get transitions: %s", err.Error())), nil
		}

		resp := fmt.Sprintf("**Available Transitions for %s**\n\n", issueKey)
		resp += formatJSON(transitions)

		return textResult(resp), nil
	})
}

// 7. jira_get_worklog - Get worklog entries
func (r *JiraRegistry) registerGetWorklog(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_get_worklog",
		Description: "Get worklog entries for a Jira issue showing time spent by users.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"issue_key": {
					Type:        "string",
					Description: "The issue key (e.g., 'PROJ-123').",
				},
				"start_at": {
					Type:        "number",
					Description: "Index of the first worklog to return.",
					Default:     0,
				},
				"max_results": {
					Type:        "number",
					Description: "Maximum number of worklogs to return (default: 50).",
					Default:     50,
				},
			},
			Required: []string{"issue_key"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "Get Worklog",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_get_worklog", args, 0, false)

		issueKey := getString(args, "issue_key", "")
		if issueKey == "" {
			return errorResult("issue_key is required"), nil
		}

		startAt := getInt(args, "start_at", 0)
		maxResults := getInt(args, "max_results", 50)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		worklogs, err := r.jiraClient.GetWorklogs(ctx, issueKey, startAt, maxResults)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to get worklogs: %s", err.Error())), nil
		}

		resp := fmt.Sprintf("**Worklogs for %s**\n\n", issueKey)
		resp += formatJSON(worklogs)

		return textResult(resp), nil
	})
}

// 8. jira_download_attachments - Download issue attachments
func (r *JiraRegistry) registerDownloadAttachments(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_download_attachments",
		Description: "Get attachment information and download URLs for a Jira issue. Returns attachment metadata including content URLs.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"issue_key": {
					Type:        "string",
					Description: "The issue key (e.g., 'PROJ-123').",
				},
			},
			Required: []string{"issue_key"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "Get Attachments",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_download_attachments", args, 0, false)

		issueKey := getString(args, "issue_key", "")
		if issueKey == "" {
			return errorResult("issue_key is required"), nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		attachments, err := r.jiraClient.GetAttachments(ctx, issueKey)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to get attachments: %s", err.Error())), nil
		}

		resp := fmt.Sprintf("**Attachments for %s** (%d files)\n\n", issueKey, len(attachments))
		resp += formatJSON(attachments)

		return textResult(resp), nil
	})
}

// 9. jira_get_agile_boards - Get agile boards
func (r *JiraRegistry) registerGetAgileBoards(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_get_agile_boards",
		Description: "Get Jira Agile boards (Scrum and Kanban). Can filter by project.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"project_key": {
					Type:        "string",
					Description: "Filter boards by project key or ID.",
				},
				"start_at": {
					Type:        "number",
					Description: "Index of the first result to return.",
					Default:     0,
				},
				"max_results": {
					Type:        "number",
					Description: "Maximum number of results to return (default: 50).",
					Default:     50,
				},
			},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "Get Agile Boards",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_get_agile_boards", args, 0, false)

		projectKey := getString(args, "project_key", "")
		startAt := getInt(args, "start_at", 0)
		maxResults := getInt(args, "max_results", 50)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		boards, err := r.jiraClient.GetBoards(ctx, projectKey, startAt, maxResults)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to get boards: %s", err.Error())), nil
		}

		resp := fmt.Sprintf("**Agile Boards** (found %d)\n\n", len(boards))
		resp += formatJSON(boards)

		return textResult(resp), nil
	})
}

// 10. jira_get_board_issues - Get issues for a board
func (r *JiraRegistry) registerGetBoardIssues(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_get_board_issues",
		Description: "Get issues on a specific Jira Agile board. Uses the board's project to search for issues.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"board_id": {
					Type:        "number",
					Description: "The board ID.",
				},
				"jql": {
					Type:        "string",
					Description: "Additional JQL filter to apply.",
				},
				"start_at": {
					Type:        "number",
					Description: "Index of the first result to return.",
					Default:     0,
				},
				"max_results": {
					Type:        "number",
					Description: "Maximum number of results to return (default: 50).",
					Default:     50,
				},
			},
			Required: []string{"board_id"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "Get Board Issues",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_get_board_issues", args, 0, false)

		boardID := getInt(args, "board_id", 0)
		if boardID == 0 {
			return errorResult("board_id is required"), nil
		}

		jqlFilter := getString(args, "jql", "")
		startAt := getInt(args, "start_at", 0)
		maxResults := getInt(args, "max_results", 50)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// Get board to find its project
		board, err := r.jiraClient.GetBoard(ctx, boardID)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to get board: %s", err.Error())), nil
		}

		// Build JQL from board's project
		jql := ""
		if board.Location != nil && board.Location.ProjectKey != "" {
			jql = fmt.Sprintf("project = %s", board.Location.ProjectKey)
			if jqlFilter != "" {
				jql += " AND " + jqlFilter
			}
		} else if jqlFilter != "" {
			jql = jqlFilter
		} else {
			return errorResult("Could not determine project for board and no JQL filter provided"), nil
		}
		jql += " ORDER BY created DESC"

		opts := &jira.SearchOptions{
			StartAt:    startAt,
			MaxResults: maxResults,
		}

		result, err := r.jiraClient.SearchIssues(ctx, jql, opts)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to get board issues: %s", err.Error())), nil
		}

		resp := fmt.Sprintf("**Board %d (%s) Issues** (showing %d of %d)\n\n", boardID, board.Name, len(result.Issues), result.Total)
		resp += formatJSON(result)

		return textResult(resp), nil
	})
}

// 11. jira_get_sprints_from_board - Get sprints from board
func (r *JiraRegistry) registerGetSprintsFromBoard(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_get_sprints_from_board",
		Description: "Get all sprints for a Jira Agile board. Can filter by sprint state.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"board_id": {
					Type:        "number",
					Description: "The board ID.",
				},
				"state": {
					Type:        "string",
					Description: "Filter by sprint state.",
					Enum:        []string{"future", "active", "closed"},
				},
				"start_at": {
					Type:        "number",
					Description: "Index of the first result to return.",
					Default:     0,
				},
				"max_results": {
					Type:        "number",
					Description: "Maximum number of results to return (default: 50).",
					Default:     50,
				},
			},
			Required: []string{"board_id"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "Get Sprints",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_get_sprints_from_board", args, 0, false)

		boardID := getInt(args, "board_id", 0)
		if boardID == 0 {
			return errorResult("board_id is required"), nil
		}

		state := getString(args, "state", "")
		startAt := getInt(args, "start_at", 0)
		maxResults := getInt(args, "max_results", 50)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		sprints, err := r.jiraClient.GetSprints(ctx, boardID, state, startAt, maxResults)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to get sprints: %s", err.Error())), nil
		}

		resp := fmt.Sprintf("**Sprints for Board %d**\n\n", boardID)
		resp += formatJSON(sprints)

		return textResult(resp), nil
	})
}

// 12. jira_get_sprint_issues - Get sprint issues
func (r *JiraRegistry) registerGetSprintIssues(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_get_sprint_issues",
		Description: "Get all issues in a specific sprint.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"sprint_id": {
					Type:        "number",
					Description: "The sprint ID.",
				},
				"jql": {
					Type:        "string",
					Description: "Additional JQL filter to apply.",
				},
				"start_at": {
					Type:        "number",
					Description: "Index of the first result to return.",
					Default:     0,
				},
				"max_results": {
					Type:        "number",
					Description: "Maximum number of results to return (default: 50).",
					Default:     50,
				},
			},
			Required: []string{"sprint_id"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "Get Sprint Issues",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_get_sprint_issues", args, 0, false)

		sprintID := getInt(args, "sprint_id", 0)
		if sprintID == 0 {
			return errorResult("sprint_id is required"), nil
		}

		jql := getString(args, "jql", "")
		startAt := getInt(args, "start_at", 0)
		maxResults := getInt(args, "max_results", 50)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		issues, err := r.jiraClient.GetSprintIssues(ctx, sprintID, jql, startAt, maxResults)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to get sprint issues: %s", err.Error())), nil
		}

		resp := fmt.Sprintf("**Sprint %d Issues**\n\n", sprintID)
		resp += formatJSON(issues)

		return textResult(resp), nil
	})
}

// 13. jira_get_link_types - Get issue link types
func (r *JiraRegistry) registerGetLinkTypes(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_get_link_types",
		Description: "Get all available issue link types (e.g., 'blocks', 'is blocked by', 'relates to').",
		InputSchema: mcp.JSONSchema{
			Type:       "object",
			Properties: map[string]mcp.Property{},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "Get Link Types",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_get_link_types", args, 0, false)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		linkTypes, err := r.jiraClient.GetLinkTypes(ctx)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to get link types: %s", err.Error())), nil
		}

		resp := "**Issue Link Types**\n\n"
		resp += formatJSON(linkTypes)

		return textResult(resp), nil
	})
}

// 14. jira_batch_get_changelogs - Get changelogs for multiple issues
func (r *JiraRegistry) registerBatchGetChangelogs(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_batch_get_changelogs",
		Description: "Get change history (changelogs) for one or more Jira issues. Shows all field changes over time.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"issue_keys": {
					Type:        "array",
					Description: "List of issue keys to get changelogs for.",
					Items:       &mcp.Property{Type: "string"},
				},
			},
			Required: []string{"issue_keys"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "Batch Get Changelogs",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_batch_get_changelogs", args, 0, false)

		issueKeys := getStringArray(args, "issue_keys")
		if len(issueKeys) == 0 {
			return errorResult("issue_keys is required"), nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		results := make(map[string]interface{})
		for _, key := range issueKeys {
			// Get issue with changelog expansion
			opts := &jira.GetIssueOptions{
				Expand: []string{"changelog"},
			}
			issue, err := r.jiraClient.GetIssue(ctx, key, opts)
			if err != nil {
				results[key] = map[string]interface{}{"error": err.Error()}
			} else if issue.Changelog != nil {
				results[key] = issue.Changelog
			} else {
				results[key] = map[string]interface{}{"message": "No changelog available"}
			}
		}

		resp := fmt.Sprintf("**Changelogs for %d issues**\n\n", len(issueKeys))
		resp += formatJSON(results)

		return textResult(resp), nil
	})
}

// 15. jira_get_project_versions - Get project fix versions
func (r *JiraRegistry) registerGetProjectVersions(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_get_project_versions",
		Description: "Get all versions (fix versions/releases) for a Jira project.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"project_key": {
					Type:        "string",
					Description: "The project key (e.g., 'PROJ').",
				},
			},
			Required: []string{"project_key"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "Get Project Versions",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_get_project_versions", args, 0, false)

		projectKey := getString(args, "project_key", "")
		if projectKey == "" {
			return errorResult("project_key is required"), nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Get project with versions expanded
		project, err := r.jiraClient.GetProject(ctx, projectKey, []string{"versions"})
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to get project versions: %s", err.Error())), nil
		}

		resp := fmt.Sprintf("**Versions for Project %s**\n\n", projectKey)
		if project.Versions != nil {
			resp += formatJSON(project.Versions)
		} else {
			resp += "No versions found"
		}

		return textResult(resp), nil
	})
}

// 16. jira_get_all_projects - Get all accessible projects
func (r *JiraRegistry) registerGetAllProjects(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_get_all_projects",
		Description: "Get all Jira projects accessible to the current user.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"expand": {
					Type:        "array",
					Description: "Fields to expand (e.g., 'description', 'lead', 'issueTypes').",
					Items:       &mcp.Property{Type: "string"},
				},
				"recent": {
					Type:        "number",
					Description: "Return only recently accessed projects (number of projects).",
				},
				"start_at": {
					Type:        "number",
					Description: "Index of the first result to return.",
					Default:     0,
				},
				"max_results": {
					Type:        "number",
					Description: "Maximum number of results to return (default: 50).",
					Default:     50,
				},
			},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "Get All Projects",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_get_all_projects", args, 0, false)

		expand := getStringArray(args, "expand")
		recent := getInt(args, "recent", 0)
		startAt := getInt(args, "start_at", 0)
		maxResults := getInt(args, "max_results", 50)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		opts := &jira.GetProjectsOptions{
			Expand:     expand,
			Recent:     recent,
			StartAt:    startAt,
			MaxResults: maxResults,
		}

		projects, err := r.jiraClient.GetProjects(ctx, opts)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to get projects: %s", err.Error())), nil
		}

		resp := fmt.Sprintf("**All Projects** (%d found)\n\n", len(projects))
		resp += formatJSON(projects)

		return textResult(resp), nil
	})
}

// ============================================================================
// WRITE TOOLS
// ============================================================================

// 17. jira_create_issue - Create new issue
func (r *JiraRegistry) registerCreateIssue(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_create_issue",
		Description: "Create a new Jira issue with the specified fields.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"project_key": {
					Type:        "string",
					Description: "The project key (e.g., 'PROJ').",
				},
				"summary": {
					Type:        "string",
					Description: "Issue summary/title.",
				},
				"issue_type": {
					Type:        "string",
					Description: "Issue type name (e.g., 'Bug', 'Story', 'Task').",
				},
				"description": {
					Type:        "string",
					Description: "Issue description (supports Jira wiki markup or ADF for Cloud).",
				},
				"priority": {
					Type:        "string",
					Description: "Priority name (e.g., 'High', 'Medium', 'Low').",
				},
				"assignee": {
					Type:        "string",
					Description: "Assignee account ID.",
				},
				"labels": {
					Type:        "array",
					Description: "Labels to apply to the issue.",
					Items:       &mcp.Property{Type: "string"},
				},
				"components": {
					Type:        "array",
					Description: "Component names to add.",
					Items:       &mcp.Property{Type: "string"},
				},
				"fix_versions": {
					Type:        "array",
					Description: "Fix version names.",
					Items:       &mcp.Property{Type: "string"},
				},
				"custom_fields": {
					Type:        "object",
					Description: "Custom field values as key-value pairs (use field ID as key).",
				},
				"parent_key": {
					Type:        "string",
					Description: "Parent issue key for sub-tasks.",
				},
			},
			Required: []string{"project_key", "summary", "issue_type"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:           "Create Issue",
			ReadOnlyHint:    false,
			DestructiveHint: false,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_create_issue", args, 0, false)

		if r.readOnlyMode {
			return r.readOnlyError(), nil
		}

		projectKey := getString(args, "project_key", "")
		summary := getString(args, "summary", "")
		issueType := getString(args, "issue_type", "")

		if projectKey == "" || summary == "" || issueType == "" {
			return errorResult("project_key, summary, and issue_type are required"), nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Build fields map for the issue
		fields := map[string]interface{}{
			"project": map[string]interface{}{
				"key": projectKey,
			},
			"summary": summary,
			"issuetype": map[string]interface{}{
				"name": issueType,
			},
		}

		// Add optional fields
		if desc := getString(args, "description", ""); desc != "" {
			fields["description"] = desc
		}
		if priority := getString(args, "priority", ""); priority != "" {
			fields["priority"] = map[string]interface{}{"name": priority}
		}
		if assignee := getString(args, "assignee", ""); assignee != "" {
			fields["assignee"] = map[string]interface{}{"accountId": assignee}
		}
		if labels := getStringArray(args, "labels"); len(labels) > 0 {
			fields["labels"] = labels
		}
		if components := getStringArray(args, "components"); len(components) > 0 {
			componentsList := make([]map[string]interface{}, len(components))
			for i, c := range components {
				componentsList[i] = map[string]interface{}{"name": c}
			}
			fields["components"] = componentsList
		}
		if fixVersions := getStringArray(args, "fix_versions"); len(fixVersions) > 0 {
			versionsList := make([]map[string]interface{}, len(fixVersions))
			for i, v := range fixVersions {
				versionsList[i] = map[string]interface{}{"name": v}
			}
			fields["fixVersions"] = versionsList
		}
		if parentKey := getString(args, "parent_key", ""); parentKey != "" {
			fields["parent"] = map[string]interface{}{"key": parentKey}
		}

		// Add custom fields
		if customFields := getMap(args, "custom_fields"); customFields != nil {
			for k, v := range customFields {
				fields[k] = v
			}
		}

		issueData := &jira.Issue{
			Fields: fields,
		}

		issue, err := r.jiraClient.CreateIssue(ctx, issueData)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to create issue: %s", err.Error())), nil
		}

		resp := fmt.Sprintf("**Issue Created Successfully**\n\nKey: %s\nSelf: %s\n\n%s",
			issue.Key, issue.Self, formatJSON(issue))

		return textResult(resp), nil
	})
}

// 18. jira_batch_create_issues - Batch create issues
func (r *JiraRegistry) registerBatchCreateIssues(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_batch_create_issues",
		Description: "Create multiple Jira issues. Creates issues one by one and returns results for each.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"issues": {
					Type:        "array",
					Description: "Array of issue objects to create. Each object should have project_key, summary, issue_type, and optional fields.",
					Items: &mcp.Property{
						Type: "object",
						Properties: map[string]mcp.Property{
							"project_key": {Type: "string"},
							"summary":     {Type: "string"},
							"issue_type":  {Type: "string"},
							"description": {Type: "string"},
							"priority":    {Type: "string"},
							"assignee":    {Type: "string"},
							"labels":      {Type: "array", Items: &mcp.Property{Type: "string"}},
						},
					},
				},
			},
			Required: []string{"issues"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:           "Batch Create Issues",
			ReadOnlyHint:    false,
			DestructiveHint: false,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_batch_create_issues", args, 0, false)

		if r.readOnlyMode {
			return r.readOnlyError(), nil
		}

		issuesRaw, ok := args["issues"].([]interface{})
		if !ok || len(issuesRaw) == 0 {
			return errorResult("issues array is required"), nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		var results []map[string]interface{}
		var errors []map[string]interface{}

		for i, issueRaw := range issuesRaw {
			issueMap, ok := issueRaw.(map[string]interface{})
			if !ok {
				errors = append(errors, map[string]interface{}{
					"index": i,
					"error": "invalid issue format",
				})
				continue
			}

			projectKey := getString(issueMap, "project_key", "")
			summary := getString(issueMap, "summary", "")
			issueType := getString(issueMap, "issue_type", "")

			if projectKey == "" || summary == "" || issueType == "" {
				errors = append(errors, map[string]interface{}{
					"index": i,
					"error": "project_key, summary, and issue_type are required",
				})
				continue
			}

			// Build fields map
			fields := map[string]interface{}{
				"project":   map[string]interface{}{"key": projectKey},
				"summary":   summary,
				"issuetype": map[string]interface{}{"name": issueType},
			}

			if desc := getString(issueMap, "description", ""); desc != "" {
				fields["description"] = desc
			}
			if priority := getString(issueMap, "priority", ""); priority != "" {
				fields["priority"] = map[string]interface{}{"name": priority}
			}
			if assignee := getString(issueMap, "assignee", ""); assignee != "" {
				fields["assignee"] = map[string]interface{}{"accountId": assignee}
			}
			if labels := getStringArray(issueMap, "labels"); len(labels) > 0 {
				fields["labels"] = labels
			}

			issue, err := r.jiraClient.CreateIssue(ctx, &jira.Issue{Fields: fields})
			if err != nil {
				errors = append(errors, map[string]interface{}{
					"index": i,
					"error": err.Error(),
				})
				continue
			}

			results = append(results, map[string]interface{}{
				"index": i,
				"key":   issue.Key,
				"id":    issue.ID,
			})
		}

		response := map[string]interface{}{
			"created":      results,
			"created_count": len(results),
		}
		if len(errors) > 0 {
			response["errors"] = errors
			response["error_count"] = len(errors)
		}

		resp := fmt.Sprintf("**Batch Create Results** (%d created, %d errors)\n\n%s", len(results), len(errors), formatJSON(response))

		return textResult(resp), nil
	})
}

// 19. jira_update_issue - Update existing issue
func (r *JiraRegistry) registerUpdateIssue(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_update_issue",
		Description: "Update an existing Jira issue's fields.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"issue_key": {
					Type:        "string",
					Description: "The issue key (e.g., 'PROJ-123').",
				},
				"summary": {
					Type:        "string",
					Description: "New summary/title.",
				},
				"description": {
					Type:        "string",
					Description: "New description.",
				},
				"priority": {
					Type:        "string",
					Description: "New priority name.",
				},
				"assignee": {
					Type:        "string",
					Description: "New assignee account ID (use '-1' to unassign).",
				},
				"labels": {
					Type:        "array",
					Description: "New labels (replaces existing).",
					Items:       &mcp.Property{Type: "string"},
				},
				"components": {
					Type:        "array",
					Description: "New component names (replaces existing).",
					Items:       &mcp.Property{Type: "string"},
				},
				"fix_versions": {
					Type:        "array",
					Description: "New fix version names (replaces existing).",
					Items:       &mcp.Property{Type: "string"},
				},
				"custom_fields": {
					Type:        "object",
					Description: "Custom field values to update.",
				},
				"notify_users": {
					Type:        "boolean",
					Description: "Send notifications to watchers (default: true).",
					Default:     true,
				},
			},
			Required: []string{"issue_key"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:           "Update Issue",
			ReadOnlyHint:    false,
			DestructiveHint: false,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_update_issue", args, 0, false)

		if r.readOnlyMode {
			return r.readOnlyError(), nil
		}

		issueKey := getString(args, "issue_key", "")
		if issueKey == "" {
			return errorResult("issue_key is required"), nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Build update map
		fields := make(map[string]interface{})

		if summary := getString(args, "summary", ""); summary != "" {
			fields["summary"] = summary
		}
		if description := getString(args, "description", ""); description != "" {
			fields["description"] = description
		}
		if priority := getString(args, "priority", ""); priority != "" {
			fields["priority"] = map[string]interface{}{"name": priority}
		}
		if assignee, ok := args["assignee"]; ok {
			assigneeStr := getString(args, "assignee", "")
			if assigneeStr == "-1" || assigneeStr == "" {
				fields["assignee"] = nil
			} else {
				fields["assignee"] = map[string]interface{}{"accountId": assigneeStr}
			}
			_ = assignee
		}
		if labels := getStringArray(args, "labels"); labels != nil {
			fields["labels"] = labels
		}
		if components := getStringArray(args, "components"); components != nil {
			componentsList := make([]map[string]interface{}, len(components))
			for i, c := range components {
				componentsList[i] = map[string]interface{}{"name": c}
			}
			fields["components"] = componentsList
		}
		if fixVersions := getStringArray(args, "fix_versions"); fixVersions != nil {
			versionsList := make([]map[string]interface{}, len(fixVersions))
			for i, v := range fixVersions {
				versionsList[i] = map[string]interface{}{"name": v}
			}
			fields["fixVersions"] = versionsList
		}

		// Add custom fields
		if customFields := getMap(args, "custom_fields"); customFields != nil {
			for k, v := range customFields {
				fields[k] = v
			}
		}

		if len(fields) == 0 {
			return errorResult("at least one field to update is required"), nil
		}

		update := map[string]interface{}{
			"fields": fields,
		}

		err := r.jiraClient.UpdateIssue(ctx, issueKey, update)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to update issue: %s", err.Error())), nil
		}

		return textResult(fmt.Sprintf("**Issue %s Updated Successfully**", issueKey)), nil
	})
}

// 20. jira_delete_issue - Delete issue
func (r *JiraRegistry) registerDeleteIssue(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_delete_issue",
		Description: "Delete a Jira issue. This action cannot be undone.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"issue_key": {
					Type:        "string",
					Description: "The issue key (e.g., 'PROJ-123').",
				},
				"delete_subtasks": {
					Type:        "boolean",
					Description: "Also delete sub-tasks (default: false).",
					Default:     false,
				},
			},
			Required: []string{"issue_key"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:           "Delete Issue",
			ReadOnlyHint:    false,
			DestructiveHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_delete_issue", args, 0, false)

		if r.readOnlyMode {
			return r.readOnlyError(), nil
		}

		issueKey := getString(args, "issue_key", "")
		if issueKey == "" {
			return errorResult("issue_key is required"), nil
		}

		deleteSubtasks := getBool(args, "delete_subtasks", false)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := r.jiraClient.DeleteIssue(ctx, issueKey, deleteSubtasks)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to delete issue: %s", err.Error())), nil
		}

		return textResult(fmt.Sprintf("**Issue %s Deleted Successfully**", issueKey)), nil
	})
}

// 21. jira_add_comment - Add comment to issue
func (r *JiraRegistry) registerAddComment(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_add_comment",
		Description: "Add a comment to a Jira issue.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"issue_key": {
					Type:        "string",
					Description: "The issue key (e.g., 'PROJ-123').",
				},
				"body": {
					Type:        "string",
					Description: "Comment body text (supports Jira wiki markup or ADF for Cloud).",
				},
				"visibility": {
					Type:        "object",
					Description: "Visibility restrictions (e.g., {\"type\": \"role\", \"value\": \"Developers\"}).",
				},
			},
			Required: []string{"issue_key", "body"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:           "Add Comment",
			ReadOnlyHint:    false,
			DestructiveHint: false,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_add_comment", args, 0, false)

		if r.readOnlyMode {
			return r.readOnlyError(), nil
		}

		issueKey := getString(args, "issue_key", "")
		body := getString(args, "body", "")
		if issueKey == "" || body == "" {
			return errorResult("issue_key and body are required"), nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		comment, err := r.jiraClient.AddComment(ctx, issueKey, body)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to add comment: %s", err.Error())), nil
		}

		resp := fmt.Sprintf("**Comment Added to %s**\n\n%s", issueKey, formatJSON(comment))

		return textResult(resp), nil
	})
}

// 22. jira_edit_comment - Edit existing comment
func (r *JiraRegistry) registerEditComment(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_edit_comment",
		Description: "Edit an existing comment on a Jira issue.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"issue_key": {
					Type:        "string",
					Description: "The issue key (e.g., 'PROJ-123').",
				},
				"comment_id": {
					Type:        "string",
					Description: "The comment ID to edit.",
				},
				"body": {
					Type:        "string",
					Description: "New comment body text.",
				},
			},
			Required: []string{"issue_key", "comment_id", "body"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:           "Edit Comment",
			ReadOnlyHint:    false,
			DestructiveHint: false,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_edit_comment", args, 0, false)

		if r.readOnlyMode {
			return r.readOnlyError(), nil
		}

		issueKey := getString(args, "issue_key", "")
		commentID := getString(args, "comment_id", "")
		body := getString(args, "body", "")
		if issueKey == "" || commentID == "" || body == "" {
			return errorResult("issue_key, comment_id, and body are required"), nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		comment, err := r.jiraClient.UpdateComment(ctx, issueKey, commentID, body)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to edit comment: %s", err.Error())), nil
		}

		resp := fmt.Sprintf("**Comment %s Updated**\n\n%s", commentID, formatJSON(comment))

		return textResult(resp), nil
	})
}

// 23. jira_add_worklog - Add worklog entry
func (r *JiraRegistry) registerAddWorklog(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_add_worklog",
		Description: "Add a worklog entry to track time spent on a Jira issue.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"issue_key": {
					Type:        "string",
					Description: "The issue key (e.g., 'PROJ-123').",
				},
				"time_spent": {
					Type:        "string",
					Description: "Time spent in Jira format (e.g., '1h 30m', '2d', '3h').",
				},
				"comment": {
					Type:        "string",
					Description: "Work description.",
				},
				"started": {
					Type:        "string",
					Description: "When the work started (ISO 8601 format). Defaults to now.",
				},
			},
			Required: []string{"issue_key", "time_spent"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:           "Add Worklog",
			ReadOnlyHint:    false,
			DestructiveHint: false,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_add_worklog", args, 0, false)

		if r.readOnlyMode {
			return r.readOnlyError(), nil
		}

		issueKey := getString(args, "issue_key", "")
		timeSpent := getString(args, "time_spent", "")
		if issueKey == "" || timeSpent == "" {
			return errorResult("issue_key and time_spent are required"), nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		worklogData := &jira.Worklog{
			TimeSpent: timeSpent,
		}

		if comment := getString(args, "comment", ""); comment != "" {
			worklogData.Comment = comment
		}
		if started := getString(args, "started", ""); started != "" {
			worklogData.Started = started
		}

		worklog, err := r.jiraClient.AddWorklog(ctx, issueKey, worklogData)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to add worklog: %s", err.Error())), nil
		}

		resp := fmt.Sprintf("**Worklog Added to %s**\n\n%s", issueKey, formatJSON(worklog))

		return textResult(resp), nil
	})
}

// 24. jira_link_to_epic - Link issue to epic
func (r *JiraRegistry) registerLinkToEpic(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_link_to_epic",
		Description: "Link an issue to an Epic (add issue to Epic's scope). Uses the epic link custom field.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"issue_key": {
					Type:        "string",
					Description: "The issue key to link (e.g., 'PROJ-123').",
				},
				"epic_key": {
					Type:        "string",
					Description: "The Epic key to link to (e.g., 'PROJ-100').",
				},
				"epic_link_field": {
					Type:        "string",
					Description: "The custom field ID for epic link (default: 'customfield_10014'). Use jira_search_fields to find the correct field.",
					Default:     "customfield_10014",
				},
			},
			Required: []string{"issue_key", "epic_key"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:           "Link to Epic",
			ReadOnlyHint:    false,
			DestructiveHint: false,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_link_to_epic", args, 0, false)

		if r.readOnlyMode {
			return r.readOnlyError(), nil
		}

		issueKey := getString(args, "issue_key", "")
		epicKey := getString(args, "epic_key", "")
		if issueKey == "" || epicKey == "" {
			return errorResult("issue_key and epic_key are required"), nil
		}

		epicLinkField := getString(args, "epic_link_field", "customfield_10014")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Update the issue with the epic link field
		update := map[string]interface{}{
			"fields": map[string]interface{}{
				epicLinkField: epicKey,
			},
		}

		err := r.jiraClient.UpdateIssue(ctx, issueKey, update)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to link to epic: %s", err.Error())), nil
		}

		return textResult(fmt.Sprintf("**Issue %s linked to Epic %s**", issueKey, epicKey)), nil
	})
}

// 25. jira_create_issue_link - Create link between issues
func (r *JiraRegistry) registerCreateIssueLink(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_create_issue_link",
		Description: "Create a link between two Jira issues (e.g., 'blocks', 'is blocked by', 'relates to').",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"link_type": {
					Type:        "string",
					Description: "Link type name (e.g., 'Blocks', 'Relates', 'Clones'). Use jira_get_link_types to see available types.",
				},
				"inward_issue": {
					Type:        "string",
					Description: "The inward issue key (e.g., 'PROJ-123 is blocked by PROJ-456').",
				},
				"outward_issue": {
					Type:        "string",
					Description: "The outward issue key (e.g., 'PROJ-456 blocks PROJ-123').",
				},
			},
			Required: []string{"link_type", "inward_issue", "outward_issue"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:           "Create Issue Link",
			ReadOnlyHint:    false,
			DestructiveHint: false,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_create_issue_link", args, 0, false)

		if r.readOnlyMode {
			return r.readOnlyError(), nil
		}

		linkType := getString(args, "link_type", "")
		inwardIssue := getString(args, "inward_issue", "")
		outwardIssue := getString(args, "outward_issue", "")
		if linkType == "" || inwardIssue == "" || outwardIssue == "" {
			return errorResult("link_type, inward_issue, and outward_issue are required"), nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := r.jiraClient.LinkIssues(ctx, inwardIssue, outwardIssue, linkType)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to create issue link: %s", err.Error())), nil
		}

		return textResult(fmt.Sprintf("**Link Created**: %s -> %s (%s)", outwardIssue, inwardIssue, linkType)), nil
	})
}

// 26. jira_transition_issue - Transition issue status
func (r *JiraRegistry) registerTransitionIssue(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "jira_transition_issue",
		Description: "Transition a Jira issue to a new status. Use jira_get_transitions to see available transitions.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"issue_key": {
					Type:        "string",
					Description: "The issue key (e.g., 'PROJ-123').",
				},
				"transition_id": {
					Type:        "string",
					Description: "The transition ID to execute.",
				},
				"comment": {
					Type:        "string",
					Description: "Optional comment to add during transition.",
				},
				"resolution": {
					Type:        "string",
					Description: "Resolution name if required (e.g., 'Done', 'Won't Do').",
				},
				"fields": {
					Type:        "object",
					Description: "Additional fields to set during transition.",
				},
			},
			Required: []string{"issue_key", "transition_id"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:           "Transition Issue",
			ReadOnlyHint:    false,
			DestructiveHint: false,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		logging.ToolCall("jira_transition_issue", args, 0, false)

		if r.readOnlyMode {
			return r.readOnlyError(), nil
		}

		issueKey := getString(args, "issue_key", "")
		transitionID := getString(args, "transition_id", "")
		if issueKey == "" || transitionID == "" {
			return errorResult("issue_key and transition_id are required"), nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Build fields map for the transition
		fields := getMap(args, "fields")
		if fields == nil {
			fields = make(map[string]interface{})
		}
		if resolution := getString(args, "resolution", ""); resolution != "" {
			fields["resolution"] = map[string]interface{}{"name": resolution}
		}

		err := r.jiraClient.TransitionIssue(ctx, issueKey, transitionID, fields)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to transition issue: %s", err.Error())), nil
		}

		return textResult(fmt.Sprintf("**Issue %s Transitioned Successfully**", issueKey)), nil
	})
}

