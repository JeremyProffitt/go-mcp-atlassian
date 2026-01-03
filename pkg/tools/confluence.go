// Package tools provides MCP tool implementations for Atlassian services.
package tools

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/nicholasgriffintn/go-mcp-atlassian/pkg/atlassian/confluence"
	"github.com/nicholasgriffintn/go-mcp-atlassian/pkg/logging"
	"github.com/nicholasgriffintn/go-mcp-atlassian/pkg/mcp"
)

// Registry holds references to clients and configuration for tool registration.
type Registry struct {
	confluenceClient *confluence.Client
	logger           *logging.Logger
	readOnlyMode     bool
}

// NewRegistry creates a new tool registry.
func NewRegistry(confluenceClient *confluence.Client, logger *logging.Logger, readOnlyMode bool) *Registry {
	return &Registry{
		confluenceClient: confluenceClient,
		logger:           logger,
		readOnlyMode:     readOnlyMode,
	}
}

// RegisterAll registers all Confluence tools with the MCP server.
func (r *Registry) RegisterAll(server *mcp.Server) {
	r.registerConfluenceSearch(server)
	r.registerConfluenceGetPage(server)
	r.registerConfluenceGetPageChildren(server)
	r.registerConfluenceGetComments(server)
	r.registerConfluenceGetLabels(server)
	r.registerConfluenceAddLabel(server)
	r.registerConfluenceCreatePage(server)
	r.registerConfluenceUpdatePage(server)
	r.registerConfluenceDeletePage(server)
	r.registerConfluenceAddComment(server)
	r.registerConfluenceSearchUser(server)
}

// registerConfluenceSearch registers the confluence_search tool.
func (r *Registry) registerConfluenceSearch(server *mcp.Server) {
	server.RegisterTool(
		mcp.Tool{
			Name:        "confluence_search",
			Description: "Search Confluence content using CQL (Confluence Query Language). Returns pages, blog posts, and other content matching the query.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"query": {
						Type:        "string",
						Description: "CQL query string. Examples: 'text ~ \"search term\"', 'type = page AND space = KEY', 'label = \"my-label\"'",
					},
					"limit": {
						Type:        "integer",
						Description: "Maximum number of results to return (default: 10, max: 100)",
						Default:     10,
					},
					"spaces_filter": {
						Type:        "string",
						Description: "Optional: Comma-separated list of space keys to filter results",
					},
				},
				Required: []string{"query"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title:        "Confluence Search",
				ReadOnlyHint: true,
			},
		},
		func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			query := getString(args, "query", "")
			if query == "" {
				return errorResult("query parameter is required"), nil
			}

			limit := getInt(args, "limit", 10)
			if limit <= 0 {
				limit = 10
			}
			if limit > 100 {
				limit = 100
			}

			spacesFilter := getString(args, "spaces_filter", "")

			// Build CQL with space filter if provided
			cql := query
			if spacesFilter != "" {
				spaces := strings.Split(spacesFilter, ",")
				for i, s := range spaces {
					spaces[i] = strings.TrimSpace(s)
				}
				spaceFilter := fmt.Sprintf("space IN (%s)", strings.Join(spaces, ","))
				cql = fmt.Sprintf("(%s) AND %s", query, spaceFilter)
			}

			ctx := context.Background()
			results, err := r.confluenceClient.Search(ctx, cql, &confluence.SearchOptions{
				Limit: limit,
				Start: 0,
			})
			if err != nil {
				r.logger.Error("confluence_search failed: %v", err)
				return errorResult(fmt.Sprintf("Search failed: %v", err)), nil
			}

			// Format results
			type resultItem struct {
				ID      string `json:"id,omitempty"`
				Title   string `json:"title"`
				Type    string `json:"type,omitempty"`
				Space   string `json:"space,omitempty"`
				Excerpt string `json:"excerpt,omitempty"`
				URL     string `json:"url,omitempty"`
			}

			var items []resultItem
			for _, res := range results.Results {
				item := resultItem{
					Title:   res.Title,
					Excerpt: res.Excerpt,
					URL:     res.URL,
				}
				if res.Content != nil {
					item.ID = res.Content.ID
					item.Type = res.Content.Type
					if res.Content.Space != nil {
						item.Space = res.Content.Space.Key
					}
				}
				items = append(items, item)
			}

			response := map[string]interface{}{
				"total":   results.Size,
				"results": items,
			}

			return jsonResult(response), nil
		},
	)
}

// registerConfluenceGetPage registers the confluence_get_page tool.
func (r *Registry) registerConfluenceGetPage(server *mcp.Server) {
	server.RegisterTool(
		mcp.Tool{
			Name:        "confluence_get_page",
			Description: "Get a Confluence page by ID or by title and space key. Returns page content, metadata, and version information.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"page_id": {
						Type:        "string",
						Description: "The ID of the page to retrieve. Either page_id or (title + space_key) is required.",
					},
					"title": {
						Type:        "string",
						Description: "The title of the page to retrieve. Requires space_key to be specified.",
					},
					"space_key": {
						Type:        "string",
						Description: "The space key where the page is located. Required when using title.",
					},
					"include_metadata": {
						Type:        "boolean",
						Description: "Whether to include page metadata like version, author, etc. (default: true)",
						Default:     true,
					},
					"convert_to_markdown": {
						Type:        "boolean",
						Description: "Whether to attempt conversion of storage format to markdown-like text (default: false)",
						Default:     false,
					},
				},
			},
			Annotations: &mcp.ToolAnnotation{
				Title:        "Get Confluence Page",
				ReadOnlyHint: true,
			},
		},
		func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			pageID := getString(args, "page_id", "")
			title := getString(args, "title", "")
			spaceKey := getString(args, "space_key", "")
			includeMetadata := getBool(args, "include_metadata", true)
			convertToMarkdown := getBool(args, "convert_to_markdown", false)

			if pageID == "" && (title == "" || spaceKey == "") {
				return errorResult("Either page_id or both title and space_key are required"), nil
			}

			ctx := context.Background()

			// Build response - we use different approaches depending on input
			response := make(map[string]interface{})

			if pageID != "" {
				// Use v2 API to get page by ID
				page, err := r.confluenceClient.GetPage(ctx, pageID, &confluence.GetPageOptions{
					BodyFormat: "storage",
				})
				if err != nil {
					r.logger.Error("confluence_get_page failed: %v", err)
					return errorResult(fmt.Sprintf("Failed to get page: %v", err)), nil
				}

				response["id"] = page.ID
				response["title"] = page.Title

				// Add content
				if page.Body != nil && page.Body.Storage != nil {
					content := page.Body.Storage.Value
					if convertToMarkdown {
						content = convertStorageToText(content)
					}
					response["content"] = content
				}

				// Add metadata if requested
				if includeMetadata {
					response["status"] = page.Status
					response["spaceId"] = page.SpaceID
					if page.Version != nil {
						response["version"] = map[string]interface{}{
							"number":  page.Version.Number,
							"message": page.Version.Message,
						}
					}
					if page.Links != nil {
						response["url"] = page.Links.WebUI
					}
				}
			} else {
				// Use v1 API to get content by space and title
				content, err := r.confluenceClient.GetContentBySpaceAndTitle(ctx, spaceKey, title, "page", []string{"body.storage", "version", "space"})
				if err != nil {
					r.logger.Error("confluence_get_page failed: %v", err)
					return errorResult(fmt.Sprintf("Failed to get page: %v", err)), nil
				}

				response["id"] = content.ID
				response["title"] = content.Title
				response["type"] = content.Type

				// Add content
				if content.Body != nil && content.Body.Storage != nil {
					pageContent := content.Body.Storage.Value
					if convertToMarkdown {
						pageContent = convertStorageToText(pageContent)
					}
					response["content"] = pageContent
				}

				// Add metadata if requested
				if includeMetadata {
					response["status"] = content.Status
					if content.Space != nil {
						response["space"] = map[string]interface{}{
							"id":   strconv.Itoa(content.Space.ID),
							"key":  content.Space.Key,
							"name": content.Space.Name,
						}
					}
					if content.Version != nil {
						response["version"] = map[string]interface{}{
							"number":  content.Version.Number,
							"message": content.Version.Message,
						}
					}
					if content.Links != nil {
						response["url"] = content.Links.WebUI
					}
				}
			}

			return jsonResult(response), nil
		},
	)
}

// registerConfluenceGetPageChildren registers the confluence_get_page_children tool.
func (r *Registry) registerConfluenceGetPageChildren(server *mcp.Server) {
	server.RegisterTool(
		mcp.Tool{
			Name:        "confluence_get_page_children",
			Description: "Get child pages of a Confluence page. Returns a list of direct child pages.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"parent_id": {
						Type:        "string",
						Description: "The ID of the parent page",
					},
					"limit": {
						Type:        "integer",
						Description: "Maximum number of results to return (default: 25)",
						Default:     25,
					},
					"cursor": {
						Type:        "string",
						Description: "Cursor for pagination (optional)",
					},
				},
				Required: []string{"parent_id"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title:        "Get Page Children",
				ReadOnlyHint: true,
			},
		},
		func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			parentID := getString(args, "parent_id", "")
			if parentID == "" {
				return errorResult("parent_id parameter is required"), nil
			}

			limit := getInt(args, "limit", 25)
			cursor := getString(args, "cursor", "")

			ctx := context.Background()
			result, err := r.confluenceClient.GetPageChildren(ctx, parentID, cursor, limit)
			if err != nil {
				r.logger.Error("confluence_get_page_children failed: %v", err)
				return errorResult(fmt.Sprintf("Failed to get child pages: %v", err)), nil
			}

			// Format results
			type childPage struct {
				ID     string `json:"id"`
				Title  string `json:"title"`
				Status string `json:"status"`
			}

			var children []childPage
			for _, page := range result.Results {
				child := childPage{
					ID:     page.ID,
					Title:  page.Title,
					Status: page.Status,
				}
				children = append(children, child)
			}

			response := map[string]interface{}{
				"parent_id": parentID,
				"children":  children,
				"count":     len(children),
			}

			return jsonResult(response), nil
		},
	)
}

// registerConfluenceGetComments registers the confluence_get_comments tool.
func (r *Registry) registerConfluenceGetComments(server *mcp.Server) {
	server.RegisterTool(
		mcp.Tool{
			Name:        "confluence_get_comments",
			Description: "Get comments on a Confluence page. Returns footer comments with their content.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"page_id": {
						Type:        "string",
						Description: "The ID of the page to get comments for",
					},
				},
				Required: []string{"page_id"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title:        "Get Page Comments",
				ReadOnlyHint: true,
			},
		},
		func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			pageID := getString(args, "page_id", "")
			if pageID == "" {
				return errorResult("page_id parameter is required"), nil
			}

			ctx := context.Background()
			result, err := r.confluenceClient.GetComments(ctx, pageID, &confluence.GetCommentsOptions{
				Expand: []string{"body.storage", "version"},
			})
			if err != nil {
				r.logger.Error("confluence_get_comments failed: %v", err)
				return errorResult(fmt.Sprintf("Failed to get comments: %v", err)), nil
			}

			// Format results
			type commentInfo struct {
				ID      string `json:"id"`
				Content string `json:"content,omitempty"`
				Version int    `json:"version,omitempty"`
			}

			var comments []commentInfo
			for _, c := range result.Results {
				comment := commentInfo{
					ID: c.ID,
				}
				if c.Body != nil && c.Body.Storage != nil {
					comment.Content = c.Body.Storage.Value
				}
				if c.Version != nil {
					comment.Version = c.Version.Number
				}
				comments = append(comments, comment)
			}

			response := map[string]interface{}{
				"page_id":  pageID,
				"comments": comments,
				"count":    len(comments),
			}

			return jsonResult(response), nil
		},
	)
}

// registerConfluenceGetLabels registers the confluence_get_labels tool.
func (r *Registry) registerConfluenceGetLabels(server *mcp.Server) {
	server.RegisterTool(
		mcp.Tool{
			Name:        "confluence_get_labels",
			Description: "Get labels attached to a Confluence page.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"page_id": {
						Type:        "string",
						Description: "The ID of the page to get labels for",
					},
				},
				Required: []string{"page_id"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title:        "Get Page Labels",
				ReadOnlyHint: true,
			},
		},
		func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			pageID := getString(args, "page_id", "")
			if pageID == "" {
				return errorResult("page_id parameter is required"), nil
			}

			ctx := context.Background()
			result, err := r.confluenceClient.GetLabels(ctx, pageID, "", 0, 100)
			if err != nil {
				r.logger.Error("confluence_get_labels failed: %v", err)
				return errorResult(fmt.Sprintf("Failed to get labels: %v", err)), nil
			}

			// Format results
			type labelInfo struct {
				ID     string `json:"id,omitempty"`
				Name   string `json:"name"`
				Prefix string `json:"prefix,omitempty"`
			}

			var labels []labelInfo
			for _, l := range result.Results {
				labels = append(labels, labelInfo{
					ID:     l.ID,
					Name:   l.Name,
					Prefix: l.Prefix,
				})
			}

			response := map[string]interface{}{
				"page_id": pageID,
				"labels":  labels,
				"count":   len(labels),
			}

			return jsonResult(response), nil
		},
	)
}

// registerConfluenceAddLabel registers the confluence_add_label tool.
func (r *Registry) registerConfluenceAddLabel(server *mcp.Server) {
	server.RegisterTool(
		mcp.Tool{
			Name:        "confluence_add_label",
			Description: "Add a label to a Confluence page. This is a write operation.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"page_id": {
						Type:        "string",
						Description: "The ID of the page to add the label to",
					},
					"name": {
						Type:        "string",
						Description: "The label name to add",
					},
				},
				Required: []string{"page_id", "name"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title:           "Add Page Label",
				ReadOnlyHint:    false,
				DestructiveHint: false,
				IdempotentHint:  true,
			},
		},
		func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			if r.readOnlyMode {
				return writeBlockedResult(), nil
			}

			pageID := getString(args, "page_id", "")
			name := getString(args, "name", "")

			if pageID == "" {
				return errorResult("page_id parameter is required"), nil
			}
			if name == "" {
				return errorResult("name parameter is required"), nil
			}

			ctx := context.Background()
			labels := []confluence.Label{{Name: name, Prefix: "global"}}
			result, err := r.confluenceClient.AddLabels(ctx, pageID, labels)
			if err != nil {
				r.logger.Error("confluence_add_label failed: %v", err)
				return errorResult(fmt.Sprintf("Failed to add label: %v", err)), nil
			}

			// Find the added label in results
			var addedLabel *confluence.Label
			for i := range result.Results {
				if result.Results[i].Name == name {
					addedLabel = &result.Results[i]
					break
				}
			}

			response := map[string]interface{}{
				"success": true,
				"page_id": pageID,
			}
			if addedLabel != nil {
				response["label"] = map[string]string{
					"name": addedLabel.Name,
				}
			} else {
				response["label"] = map[string]string{
					"name": name,
				}
			}

			return jsonResult(response), nil
		},
	)
}

// registerConfluenceCreatePage registers the confluence_create_page tool.
func (r *Registry) registerConfluenceCreatePage(server *mcp.Server) {
	server.RegisterTool(
		mcp.Tool{
			Name:        "confluence_create_page",
			Description: "Create a new Confluence page. This is a write operation.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"space_key": {
						Type:        "string",
						Description: "The key of the space where the page will be created",
					},
					"title": {
						Type:        "string",
						Description: "The title of the new page",
					},
					"content": {
						Type:        "string",
						Description: "The content of the page in storage format (XHTML)",
					},
					"parent_id": {
						Type:        "string",
						Description: "Optional: The ID of the parent page for hierarchical placement",
					},
				},
				Required: []string{"space_key", "title", "content"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title:           "Create Confluence Page",
				ReadOnlyHint:    false,
				DestructiveHint: false,
				IdempotentHint:  false,
			},
		},
		func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			if r.readOnlyMode {
				return writeBlockedResult(), nil
			}

			spaceKey := getString(args, "space_key", "")
			title := getString(args, "title", "")
			content := getString(args, "content", "")
			parentID := getString(args, "parent_id", "")

			if spaceKey == "" {
				return errorResult("space_key parameter is required"), nil
			}
			if title == "" {
				return errorResult("title parameter is required"), nil
			}
			if content == "" {
				return errorResult("content parameter is required"), nil
			}

			ctx := context.Background()
			createdContent, err := r.confluenceClient.CreatePageWithContent(ctx, spaceKey, title, content, parentID)
			if err != nil {
				r.logger.Error("confluence_create_page failed: %v", err)
				return errorResult(fmt.Sprintf("Failed to create page: %v", err)), nil
			}

			response := map[string]interface{}{
				"success": true,
				"page": map[string]interface{}{
					"id":    createdContent.ID,
					"title": createdContent.Title,
					"type":  createdContent.Type,
				},
			}

			if createdContent.Links != nil {
				response["url"] = createdContent.Links.WebUI
			}

			return jsonResult(response), nil
		},
	)
}

// registerConfluenceUpdatePage registers the confluence_update_page tool.
func (r *Registry) registerConfluenceUpdatePage(server *mcp.Server) {
	server.RegisterTool(
		mcp.Tool{
			Name:        "confluence_update_page",
			Description: "Update an existing Confluence page. This is a write operation that increments the page version.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"page_id": {
						Type:        "string",
						Description: "The ID of the page to update",
					},
					"title": {
						Type:        "string",
						Description: "The title for the page (required)",
					},
					"content": {
						Type:        "string",
						Description: "The new content for the page in storage format (XHTML)",
					},
				},
				Required: []string{"page_id", "title", "content"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title:           "Update Confluence Page",
				ReadOnlyHint:    false,
				DestructiveHint: false,
				IdempotentHint:  false,
			},
		},
		func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			if r.readOnlyMode {
				return writeBlockedResult(), nil
			}

			pageID := getString(args, "page_id", "")
			title := getString(args, "title", "")
			content := getString(args, "content", "")

			if pageID == "" {
				return errorResult("page_id parameter is required"), nil
			}
			if title == "" {
				return errorResult("title parameter is required"), nil
			}
			if content == "" {
				return errorResult("content parameter is required"), nil
			}

			ctx := context.Background()

			// First get current page to get its version
			currentContent, err := r.confluenceClient.GetContent(ctx, pageID, &confluence.GetContentOptions{
				Expand: []string{"version"},
			})
			if err != nil {
				r.logger.Error("confluence_update_page failed to get current version: %v", err)
				return errorResult(fmt.Sprintf("Failed to get current page version: %v", err)), nil
			}

			currentVersion := 1
			if currentContent.Version != nil {
				currentVersion = currentContent.Version.Number
			}

			updatedContent, err := r.confluenceClient.UpdatePageContent(ctx, pageID, title, content, currentVersion)
			if err != nil {
				r.logger.Error("confluence_update_page failed: %v", err)
				return errorResult(fmt.Sprintf("Failed to update page: %v", err)), nil
			}

			response := map[string]interface{}{
				"success": true,
				"page": map[string]interface{}{
					"id":    updatedContent.ID,
					"title": updatedContent.Title,
					"type":  updatedContent.Type,
				},
			}

			if updatedContent.Version != nil {
				response["version"] = updatedContent.Version.Number
			}

			if updatedContent.Links != nil {
				response["url"] = updatedContent.Links.WebUI
			}

			return jsonResult(response), nil
		},
	)
}

// registerConfluenceDeletePage registers the confluence_delete_page tool.
func (r *Registry) registerConfluenceDeletePage(server *mcp.Server) {
	server.RegisterTool(
		mcp.Tool{
			Name:        "confluence_delete_page",
			Description: "Delete a Confluence page. This is a destructive write operation.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"page_id": {
						Type:        "string",
						Description: "The ID of the page to delete",
					},
				},
				Required: []string{"page_id"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title:           "Delete Confluence Page",
				ReadOnlyHint:    false,
				DestructiveHint: true,
				IdempotentHint:  true,
			},
		},
		func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			if r.readOnlyMode {
				return writeBlockedResult(), nil
			}

			pageID := getString(args, "page_id", "")
			if pageID == "" {
				return errorResult("page_id parameter is required"), nil
			}

			ctx := context.Background()
			err := r.confluenceClient.DeletePage(ctx, pageID)
			if err != nil {
				r.logger.Error("confluence_delete_page failed: %v", err)
				return errorResult(fmt.Sprintf("Failed to delete page: %v", err)), nil
			}

			response := map[string]interface{}{
				"success": true,
				"page_id": pageID,
				"message": "Page deleted successfully",
			}

			return jsonResult(response), nil
		},
	)
}

// registerConfluenceAddComment registers the confluence_add_comment tool.
func (r *Registry) registerConfluenceAddComment(server *mcp.Server) {
	server.RegisterTool(
		mcp.Tool{
			Name:        "confluence_add_comment",
			Description: "Add a comment to a Confluence page. This is a write operation.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"page_id": {
						Type:        "string",
						Description: "The ID of the page to add the comment to",
					},
					"content": {
						Type:        "string",
						Description: "The comment content in storage format (XHTML)",
					},
				},
				Required: []string{"page_id", "content"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title:           "Add Page Comment",
				ReadOnlyHint:    false,
				DestructiveHint: false,
				IdempotentHint:  false,
			},
		},
		func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			if r.readOnlyMode {
				return writeBlockedResult(), nil
			}

			pageID := getString(args, "page_id", "")
			content := getString(args, "content", "")

			if pageID == "" {
				return errorResult("page_id parameter is required"), nil
			}
			if content == "" {
				return errorResult("content parameter is required"), nil
			}

			ctx := context.Background()
			comment := &confluence.Content{
				Body: &confluence.ContentBody{
					Storage: &confluence.BodyContent{
						Value:          content,
						Representation: "storage",
					},
				},
			}
			createdComment, err := r.confluenceClient.AddComment(ctx, pageID, comment)
			if err != nil {
				r.logger.Error("confluence_add_comment failed: %v", err)
				return errorResult(fmt.Sprintf("Failed to add comment: %v", err)), nil
			}

			response := map[string]interface{}{
				"success": true,
				"page_id": pageID,
				"comment": map[string]interface{}{
					"id": createdComment.ID,
				},
			}

			return jsonResult(response), nil
		},
	)
}

// registerConfluenceSearchUser registers the confluence_search_user tool.
func (r *Registry) registerConfluenceSearchUser(server *mcp.Server) {
	server.RegisterTool(
		mcp.Tool{
			Name:        "confluence_search_user",
			Description: "Search for Confluence users by name or email.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"query": {
						Type:        "string",
						Description: "Search query (name or email)",
					},
					"limit": {
						Type:        "integer",
						Description: "Maximum number of results to return (default: 10)",
						Default:     10,
					},
				},
				Required: []string{"query"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title:        "Search Confluence Users",
				ReadOnlyHint: true,
			},
		},
		func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			query := getString(args, "query", "")
			if query == "" {
				return errorResult("query parameter is required"), nil
			}

			limit := getInt(args, "limit", 10)
			if limit <= 0 {
				limit = 10
			}

			ctx := context.Background()
			// Build CQL for user search
			cql := fmt.Sprintf("user.fullname ~ \"%s\" OR user.email ~ \"%s\"", query, query)
			users, err := r.confluenceClient.SearchUsers(ctx, cql, 0, limit, nil)
			if err != nil {
				r.logger.Error("confluence_search_user failed: %v", err)
				return errorResult(fmt.Sprintf("Failed to search users: %v", err)), nil
			}

			// Format results
			type userInfo struct {
				AccountID   string `json:"account_id"`
				DisplayName string `json:"display_name"`
				Email       string `json:"email,omitempty"`
				Type        string `json:"type,omitempty"`
			}

			var userList []userInfo
			for _, u := range users {
				userList = append(userList, userInfo{
					AccountID:   u.AccountID,
					DisplayName: u.DisplayName,
					Email:       u.Email,
				})
			}

			response := map[string]interface{}{
				"users": userList,
				"count": len(userList),
			}

			return jsonResult(response), nil
		},
	)
}

// convertStorageToText performs a basic conversion of Confluence storage format to plain text.
// This is a simplified conversion and may not handle all cases perfectly.
func convertStorageToText(storage string) string {
	// Remove CDATA sections
	storage = strings.ReplaceAll(storage, "<![CDATA[", "")
	storage = strings.ReplaceAll(storage, "]]>", "")

	// Convert common HTML tags to text equivalents
	replacements := []struct {
		old, new string
	}{
		{"<br/>", "\n"},
		{"<br />", "\n"},
		{"<br>", "\n"},
		{"</p>", "\n\n"},
		{"</div>", "\n"},
		{"</li>", "\n"},
		{"<ul>", "\n"},
		{"</ul>", "\n"},
		{"<ol>", "\n"},
		{"</ol>", "\n"},
		{"</h1>", "\n\n"},
		{"</h2>", "\n\n"},
		{"</h3>", "\n\n"},
		{"</h4>", "\n\n"},
		{"</h5>", "\n\n"},
		{"</h6>", "\n\n"},
		{"<li>", "  - "},
		{"&nbsp;", " "},
		{"&amp;", "&"},
		{"&lt;", "<"},
		{"&gt;", ">"},
		{"&quot;", "\""},
		{"&#39;", "'"},
	}

	for _, r := range replacements {
		storage = strings.ReplaceAll(storage, r.old, r.new)
	}

	// Remove remaining HTML tags using a simple approach
	result := storage
	for {
		start := strings.Index(result, "<")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], ">")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+1:]
	}

	// Clean up multiple newlines
	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}

	return strings.TrimSpace(result)
}
