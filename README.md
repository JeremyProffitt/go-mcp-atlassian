# go-mcp-atlassian

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Build Status](https://github.com/JeremyProffitt/go-mcp-atlassian/actions/workflows/ci.yml/badge.svg)](https://github.com/JeremyProffitt/go-mcp-atlassian/actions/workflows/ci.yml)

A Go implementation of the Model Context Protocol (MCP) server for Atlassian products (Confluence and Jira). Supports both Cloud and Server/Data Center deployments.

This is a Go port of [mcp-atlassian](https://github.com/sooperset/mcp-atlassian) (Python), providing a single binary with no runtime dependencies.

## Features

- **Single Binary**: No Python, Node.js, or other runtime dependencies required
- **Cross-Platform**: Binaries available for Windows, macOS (Intel & Apple Silicon), and Linux
- **Full Atlassian Support**: Works with both Cloud and Server/Data Center deployments
- **Jira Integration**: Search, create, update, transition issues, manage sprints, and more
- **Confluence Integration**: Search, read, create, and update pages
- **MCP Protocol**: Implements the Model Context Protocol for AI assistant integration
- **Read-Only Mode**: Optional mode to prevent any write operations

## Quick Start

### 1. Download

Download the latest binary for your platform from the [Releases](https://github.com/JeremyProffitt/go-mcp-atlassian/releases) page.

### 2. Get Your API Token

Choose the instructions based on your deployment type:

#### Atlassian Cloud (API Token)

For Cloud deployments, you need an API token. The same token works for both Jira and Confluence.

1. Log in to your Atlassian account at https://id.atlassian.com
2. Navigate to **Security** in the left sidebar, or go directly to https://id.atlassian.com/manage-profile/security/api-tokens
3. Click **Create API token**
4. Enter a descriptive label (e.g., "MCP Server Token")
5. Click **Create**
6. Click **Copy** to copy the token to your clipboard
7. Store the token securely - you won't be able to see it again

> **Note**: Use your Atlassian account email as the username and this API token as the password in your configuration.

#### Jira Server/Data Center (Personal Access Token)

For self-hosted Jira, you need a Personal Access Token (PAT).

1. Log in to your Jira Server/Data Center instance
2. Click your profile avatar in the top-right corner
3. Select **Profile** from the dropdown
4. Click **Personal Access Tokens** in the left sidebar
5. Click **Create token**
6. Enter a descriptive name (e.g., "MCP Server Token")
7. Optionally set an expiration date (recommended for security)
8. Click **Create**
9. Copy the token immediately - it will only be shown once
10. Store the token securely

> **Note**: PATs require Jira Server/Data Center version 8.14 or later.

#### Confluence Server/Data Center (Personal Access Token)

For self-hosted Confluence, you need a Personal Access Token (PAT).

1. Log in to your Confluence Server/Data Center instance
2. Click your profile avatar in the top-right corner
3. Select **Settings** (or **Profile**)
4. Click **Personal Access Tokens** in the left sidebar
5. Click **Create token**
6. Enter a descriptive name (e.g., "MCP Server Token")
7. Optionally set an expiration date (recommended for security)
8. Click **Create**
9. Copy the token immediately - it will only be shown once
10. Store the token securely

> **Note**: PATs require Confluence Server/Data Center version 7.9 or later.

### 3. Configure Your IDE

Add to your Claude Desktop, Cursor, or other MCP-compatible client configuration.

## Configuration

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `JIRA_URL` | Jira instance URL (e.g., `https://company.atlassian.net`) | For Jira |
| `JIRA_USERNAME` | Email for Cloud, username for Server/DC | Cloud |
| `JIRA_API_TOKEN` | API token for Cloud | Cloud |
| `JIRA_PERSONAL_TOKEN` | Personal Access Token for Server/DC | Server/DC |
| `JIRA_SSL_VERIFY` | Verify SSL certificates (default: `true`) | No |
| `JIRA_PROJECTS_FILTER` | Comma-separated project keys to filter | No |
| `CONFLUENCE_URL` | Confluence instance URL | For Confluence |
| `CONFLUENCE_USERNAME` | Email for Cloud, username for Server/DC | Cloud |
| `CONFLUENCE_API_TOKEN` | API token for Cloud | Cloud |
| `CONFLUENCE_PERSONAL_TOKEN` | Personal Access Token for Server/DC | Server/DC |
| `CONFLUENCE_SSL_VERIFY` | Verify SSL certificates (default: `true`) | No |
| `CONFLUENCE_SPACES_FILTER` | Comma-separated space keys to filter | No |
| `READ_ONLY_MODE` | Disable write operations (default: `false`) | No |
| `MCP_LOG_DIR` | Log directory path | No |
| `MCP_LOG_LEVEL` | Log level: `off`, `error`, `warn`, `info`, `debug` | No |
| `MCP_LOG_QUERIES` | Log all queries to `queries/` subfolder (default: `false`) | No |

### JSON Configuration Examples

#### Claude Desktop (macOS/Linux)

Edit `~/.config/claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "atlassian": {
      "command": "/path/to/go-mcp-atlassian",
      "env": {
        "JIRA_URL": "https://your-company.atlassian.net",
        "JIRA_USERNAME": "your.email@company.com",
        "JIRA_API_TOKEN": "your_jira_api_token",
        "CONFLUENCE_URL": "https://your-company.atlassian.net/wiki",
        "CONFLUENCE_USERNAME": "your.email@company.com",
        "CONFLUENCE_API_TOKEN": "your_confluence_api_token"
      }
    }
  }
}
```

#### Claude Desktop (Windows)

Edit `%APPDATA%\Claude\claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "atlassian": {
      "command": "C:\\path\\to\\go-mcp-atlassian.exe",
      "env": {
        "JIRA_URL": "https://your-company.atlassian.net",
        "JIRA_USERNAME": "your.email@company.com",
        "JIRA_API_TOKEN": "your_jira_api_token",
        "CONFLUENCE_URL": "https://your-company.atlassian.net/wiki",
        "CONFLUENCE_USERNAME": "your.email@company.com",
        "CONFLUENCE_API_TOKEN": "your_confluence_api_token"
      }
    }
  }
}
```

#### Server/Data Center Configuration

```json
{
  "mcpServers": {
    "atlassian": {
      "command": "/path/to/go-mcp-atlassian",
      "env": {
        "JIRA_URL": "https://jira.your-internal-domain.com",
        "JIRA_PERSONAL_TOKEN": "your_jira_personal_access_token",
        "CONFLUENCE_URL": "https://confluence.your-internal-domain.com",
        "CONFLUENCE_PERSONAL_TOKEN": "your_confluence_personal_access_token",
        "JIRA_SSL_VERIFY": "true",
        "CONFLUENCE_SSL_VERIFY": "true"
      }
    }
  }
}
```

#### Read-Only Mode Configuration

```json
{
  "mcpServers": {
    "atlassian": {
      "command": "/path/to/go-mcp-atlassian",
      "env": {
        "JIRA_URL": "https://your-company.atlassian.net",
        "JIRA_USERNAME": "your.email@company.com",
        "JIRA_API_TOKEN": "your_api_token",
        "READ_ONLY_MODE": "true"
      }
    }
  }
}
```

#### With Project/Space Filtering

```json
{
  "mcpServers": {
    "atlassian": {
      "command": "/path/to/go-mcp-atlassian",
      "env": {
        "JIRA_URL": "https://your-company.atlassian.net",
        "JIRA_USERNAME": "your.email@company.com",
        "JIRA_API_TOKEN": "your_api_token",
        "JIRA_PROJECTS_FILTER": "PROJ,DEV,SUPPORT",
        "CONFLUENCE_URL": "https://your-company.atlassian.net/wiki",
        "CONFLUENCE_USERNAME": "your.email@company.com",
        "CONFLUENCE_API_TOKEN": "your_api_token",
        "CONFLUENCE_SPACES_FILTER": "DEV,TEAM,DOC"
      }
    }
  }
}
```

### YAML Configuration Examples

#### Continue.dev Configuration

Edit `~/.continue/config.yaml`:

```yaml
mcpServers:
  - name: Atlassian
    type: stdio
    command: /path/to/go-mcp-atlassian
    env:
      JIRA_URL: https://your-company.atlassian.net
      JIRA_USERNAME: your.email@company.com
      JIRA_API_TOKEN: ${{ secrets.JIRA_API_TOKEN }}
      CONFLUENCE_URL: https://your-company.atlassian.net/wiki
      CONFLUENCE_USERNAME: your.email@company.com
      CONFLUENCE_API_TOKEN: ${{ secrets.CONFLUENCE_API_TOKEN }}
```

#### Continue.dev Standalone MCP Server File

Create `~/.continue/mcpServers/atlassian.yaml`:

```yaml
name: Atlassian MCP Server
version: 0.0.1
schema: v1
mcpServers:
  - name: Atlassian
    type: stdio
    command: /path/to/go-mcp-atlassian
    env:
      JIRA_URL: https://your-company.atlassian.net
      JIRA_USERNAME: your.email@company.com
      JIRA_API_TOKEN: ${{ secrets.JIRA_API_TOKEN }}
      CONFLUENCE_URL: https://your-company.atlassian.net/wiki
      CONFLUENCE_USERNAME: your.email@company.com
      CONFLUENCE_API_TOKEN: ${{ secrets.CONFLUENCE_API_TOKEN }}
```

> **Note**: MCP servers in Continue.dev can only be used in Agent mode. Use `${{ secrets.VAR_NAME }}` syntax to reference secrets stored in Continue's secrets management.

#### Docker Compose Configuration

```yaml
version: '3.8'
services:
  mcp-atlassian:
    image: go-mcp-atlassian:latest
    build:
      context: .
      dockerfile: Dockerfile
    environment:
      - JIRA_URL=https://your-company.atlassian.net
      - JIRA_USERNAME=your.email@company.com
      - JIRA_API_TOKEN=${JIRA_API_TOKEN}
      - CONFLUENCE_URL=https://your-company.atlassian.net/wiki
      - CONFLUENCE_USERNAME=your.email@company.com
      - CONFLUENCE_API_TOKEN=${CONFLUENCE_API_TOKEN}
      - MCP_LOG_LEVEL=info
    volumes:
      - ./logs:/app/logs
```

#### Kubernetes ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: mcp-atlassian-config
data:
  JIRA_URL: "https://your-company.atlassian.net"
  CONFLUENCE_URL: "https://your-company.atlassian.net/wiki"
  MCP_LOG_LEVEL: "info"
  READ_ONLY_MODE: "false"
---
apiVersion: v1
kind: Secret
metadata:
  name: mcp-atlassian-secrets
type: Opaque
stringData:
  JIRA_USERNAME: "your.email@company.com"
  JIRA_API_TOKEN: "your_jira_api_token"
  CONFLUENCE_USERNAME: "your.email@company.com"
  CONFLUENCE_API_TOKEN: "your_confluence_api_token"
```

## Tool Reference

This section provides a comprehensive reference for all available tools, organized by category and use case.

### Tool Categories Overview

| Category | Read Tools | Write Tools | Description |
|----------|------------|-------------|-------------|
| Jira Issues | 6 | 6 | Create, read, update, delete, and search issues |
| Jira Agile | 4 | 0 | Boards, sprints, and agile workflows |
| Jira Metadata | 6 | 0 | Projects, fields, link types, versions |
| Confluence Pages | 4 | 3 | Create, read, update, and search pages |
| Confluence Metadata | 2 | 2 | Labels, comments, and users |

### Jira Tools

#### Issue Management (Core CRUD Operations)

| Tool | Description | Type | Key Parameters |
|------|-------------|------|----------------|
| `jira_get_issue` | Retrieve full issue details including fields, comments, and attachments | Read | `issue_key` (e.g., "PROJ-123") |
| `jira_search` | Search issues using JQL (Jira Query Language) | Read | `jql` query string, `max_results`, `fields` |
| `jira_create_issue` | Create a new issue with specified fields | Write | `project_key`, `issue_type`, `summary`, `description` |
| `jira_batch_create_issues` | Create multiple issues in a single operation | Write | Array of issue objects |
| `jira_update_issue` | Modify existing issue fields | Write | `issue_key`, field updates |
| `jira_delete_issue` | Permanently remove an issue | Write | `issue_key` |

#### Issue Workflow & Comments

| Tool | Description | Type | Key Parameters |
|------|-------------|------|----------------|
| `jira_get_transitions` | List available status transitions for an issue | Read | `issue_key` |
| `jira_transition_issue` | Change issue status (e.g., To Do -> In Progress -> Done) | Write | `issue_key`, `transition_id` |
| `jira_add_comment` | Add a comment to an issue | Write | `issue_key`, `body` |
| `jira_edit_comment` | Modify an existing comment | Write | `issue_key`, `comment_id`, `body` |

#### Issue Relationships & Links

| Tool | Description | Type | Key Parameters |
|------|-------------|------|----------------|
| `jira_get_link_types` | List all available issue link types (blocks, relates to, etc.) | Read | None |
| `jira_link_to_epic` | Associate an issue with an epic | Write | `issue_key`, `epic_key` |
| `jira_create_issue_link` | Create a link between two issues | Write | `inward_issue`, `outward_issue`, `link_type` |

#### Time Tracking & Attachments

| Tool | Description | Type | Key Parameters |
|------|-------------|------|----------------|
| `jira_get_worklog` | Retrieve time tracking entries for an issue | Read | `issue_key` |
| `jira_add_worklog` | Log time spent on an issue | Write | `issue_key`, `time_spent`, `started` |
| `jira_download_attachments` | Get attachment metadata and download URLs | Read | `issue_key` |

#### Agile & Sprint Management

| Tool | Description | Type | Key Parameters |
|------|-------------|------|----------------|
| `jira_get_agile_boards` | List all Scrum/Kanban boards accessible to user | Read | `project_key` (optional filter) |
| `jira_get_board_issues` | Get all issues on a specific board | Read | `board_id` |
| `jira_get_sprints_from_board` | List sprints for a board (past, active, future) | Read | `board_id`, `state` (active/closed/future) |
| `jira_get_sprint_issues` | Get issues assigned to a specific sprint | Read | `sprint_id` |

#### Project & Field Metadata

| Tool | Description | Type | Key Parameters |
|------|-------------|------|----------------|
| `jira_get_all_projects` | List all projects accessible to the user | Read | None |
| `jira_get_project_issues` | Get all issues in a project | Read | `project_key`, `max_results` |
| `jira_get_project_versions` | List fix versions/releases for a project | Read | `project_key` |
| `jira_search_fields` | Search available Jira fields by name | Read | `query` |
| `jira_batch_get_changelogs` | Get change history for multiple issues | Read | Array of `issue_keys` |
| `jira_get_user_profile` | Get current user's profile information | Read | None |

### Confluence Tools

#### Page Management (Core CRUD Operations)

| Tool | Description | Type | Key Parameters |
|------|-------------|------|----------------|
| `confluence_get_page` | Retrieve page content and metadata | Read | `page_id` or `space_key` + `title` |
| `confluence_search` | Search content using CQL (Confluence Query Language) | Read | `cql` query string, `limit` |
| `confluence_get_page_children` | List child pages of a parent page | Read | `page_id` |
| `confluence_create_page` | Create a new page in a space | Write | `space_key`, `title`, `body`, `parent_id` (optional) |
| `confluence_update_page` | Modify existing page content | Write | `page_id`, `title`, `body`, `version` |
| `confluence_delete_page` | Permanently remove a page | Write | `page_id` |

#### Page Metadata & Collaboration

| Tool | Description | Type | Key Parameters |
|------|-------------|------|----------------|
| `confluence_get_labels` | List labels attached to a page | Read | `page_id` |
| `confluence_add_label` | Add a label to a page | Write | `page_id`, `label` |
| `confluence_get_comments` | Retrieve comments on a page | Read | `page_id` |
| `confluence_add_comment` | Add a comment to a page | Write | `page_id`, `body` |
| `confluence_search_user` | Search for Confluence users by name/email | Read | `query` |

## Common Workflows

This section describes typical multi-tool workflows for common tasks.

### Jira Workflows

#### 1. Bug Triage Workflow
```
1. jira_search         -> Find bugs with JQL: "project = PROJ AND type = Bug AND status = Open"
2. jira_get_issue      -> Get details of each bug to assess priority
3. jira_update_issue   -> Set priority, assignee, and labels
4. jira_transition_issue -> Move to "In Progress" when work begins
```

#### 2. Sprint Planning Workflow
```
1. jira_get_agile_boards       -> Find the team's board
2. jira_get_sprints_from_board -> List upcoming sprints (state=future)
3. jira_search                 -> Find backlog items: "project = PROJ AND sprint is EMPTY"
4. jira_update_issue           -> Assign issues to sprint
```

#### 3. Issue Creation with Epic Link
```
1. jira_search         -> Find epic: "project = PROJ AND type = Epic AND summary ~ 'Feature Name'"
2. jira_create_issue   -> Create new story/task
3. jira_link_to_epic   -> Link new issue to the epic
```

#### 4. Time Tracking Workflow
```
1. jira_get_issue      -> Get current issue state
2. jira_transition_issue -> Move to "In Progress"
3. jira_add_worklog    -> Log time spent
4. jira_add_comment    -> Document progress
5. jira_transition_issue -> Move to "Done" when complete
```

#### 5. Release Management Workflow
```
1. jira_get_project_versions -> List all versions for the project
2. jira_search              -> Find issues for release: "fixVersion = '1.0.0'"
3. jira_batch_get_changelogs -> Get change history for release notes
```

### Confluence Workflows

#### 1. Documentation Search & Update
```
1. confluence_search      -> Find page: "space = DOC AND title ~ 'API Guide'"
2. confluence_get_page    -> Retrieve current content
3. confluence_update_page -> Update with new content (increment version)
```

#### 2. Create Documentation Hierarchy
```
1. confluence_create_page     -> Create parent page
2. confluence_create_page     -> Create child page (with parent_id)
3. confluence_add_label       -> Tag pages with labels for organization
```

#### 3. Content Review Workflow
```
1. confluence_search          -> Find pages modified recently
2. confluence_get_page        -> Review content
3. confluence_add_comment     -> Leave review feedback
4. confluence_add_label       -> Add "reviewed" label when complete
```

### Cross-Product Workflows

#### Link Jira Issues to Confluence Documentation
```
1. confluence_search  -> Find relevant documentation page
2. jira_get_issue     -> Get issue details
3. jira_add_comment   -> Add link to Confluence page in comment
```

#### Create Release Notes Page
```
1. jira_search              -> Find issues: "fixVersion = '1.0.0' AND status = Done"
2. jira_batch_get_changelogs -> Get change details
3. confluence_create_page    -> Create release notes page with issue summary
```

## JQL Quick Reference

Common JQL queries for use with `jira_search`:

| Use Case | JQL Query |
|----------|-----------|
| My open issues | `assignee = currentUser() AND status != Done` |
| Bugs in project | `project = PROJ AND type = Bug` |
| Current sprint | `sprint in openSprints()` |
| Unassigned issues | `project = PROJ AND assignee is EMPTY` |
| Recent updates | `updated >= -7d` |
| High priority | `priority in (Highest, High)` |
| Epics without stories | `type = Epic AND "Epic Link" is EMPTY` |

## CQL Quick Reference

Common CQL queries for use with `confluence_search`:

| Use Case | CQL Query |
|----------|-----------|
| Pages in space | `space = SPACENAME` |
| Search by title | `title ~ "search term"` |
| Recent pages | `lastmodified >= now("-7d")` |
| Pages with label | `label = "documentation"` |
| My pages | `creator = currentUser()` |
| Pages mentioning text | `text ~ "search term"` |

## Usage Examples

Once configured, ask your AI assistant to:

**Jira:**
- "Find issues assigned to me in PROJ project"
- "Create a bug ticket for the login issue"
- "Update the status of PROJ-123 to Done"
- "Show me the sprint backlog for the current sprint"
- "Link PROJ-456 to epic PROJ-100"

**Confluence:**
- "Search Confluence for onboarding docs"
- "Get the content of the Architecture page in DEV space"
- "Create a new meeting notes page in TEAM space"
- "Add a comment to the project plan page"

## Command Line Options

```
go-mcp-atlassian [options]

Options:
  -log-dir <path>     Log directory path
  -log-level <level>  Log level: off, error, warn, info, access, debug
  --http              Run in HTTP mode (default: stdio)
  -p, --port <port>   HTTP port (default: 3000)
  -H, --host <host>   HTTP host (default: 127.0.0.1)
  --version           Show version
  --help              Show help
```

## Building from Source

### Prerequisites

- Go 1.21 or later

### Build

```bash
# Clone the repository
git clone https://github.com/JeremyProffitt/go-mcp-atlassian.git
cd go-mcp-atlassian

# Build for current platform
go build -o go-mcp-atlassian .

# Build for all platforms
./build.sh        # macOS/Linux
.\build.ps1       # Windows PowerShell
```

### Run Tests

```bash
go test -v -race ./...
```

## Compatibility

| Product | Deployment | Support |
|---------|------------|---------|
| Confluence | Cloud | Fully supported |
| Confluence | Server/Data Center | Supported (v6.0+) |
| Jira | Cloud | Fully supported |
| Jira | Server/Data Center | Supported (v8.14+) |

## Security

- Never share API tokens or Personal Access Tokens
- Keep configuration files with credentials secure
- Use `READ_ONLY_MODE=true` to prevent accidental modifications
- Tokens are never logged (even in debug mode)

## Troubleshooting

### Connection Issues

1. Verify your Atlassian URL is correct
2. Check that your API token or PAT is valid
3. For Server/DC, ensure SSL certificates are valid or set `*_SSL_VERIFY=false`

### Authentication Errors

- **Cloud**: Ensure you're using your email as username and an API token (not password)
- **Server/DC**: Use a Personal Access Token with appropriate permissions

## Global Environment File

All go-mcp servers support loading environment variables from `~/.mcp_env`. This provides a central location to configure credentials and settings, especially useful on macOS where GUI applications don't inherit shell environment variables from `.zshrc` or `.bashrc`.

### File Format

Create `~/.mcp_env` with KEY=VALUE pairs:

```bash
# ~/.mcp_env - MCP Server Environment Variables

# Atlassian Cloud Configuration
JIRA_URL=https://your-company.atlassian.net
JIRA_USERNAME=your.email@company.com
JIRA_API_TOKEN=your_jira_api_token
CONFLUENCE_URL=https://your-company.atlassian.net/wiki
CONFLUENCE_USERNAME=your.email@company.com
CONFLUENCE_API_TOKEN=your_confluence_api_token

# Logging
MCP_LOG_DIR=~/mcp-logs
MCP_LOG_LEVEL=info
```

### Features

- Lines starting with `#` are treated as comments
- Empty lines are ignored
- Values can be quoted with single or double quotes
- **Existing environment variables are NOT overwritten** (env vars take precedence)
- Paths with `~` are automatically expanded to your home directory

### Path Expansion

All path-related settings support `~` expansion:

```bash
MCP_LOG_DIR=~/logs/atlassian
```

This works in the `~/.mcp_env` file, environment variables, and command-line flags.

### Logging

Enable debug logging for troubleshooting:

```json
{
  "mcpServers": {
    "atlassian": {
      "command": "/path/to/go-mcp-atlassian",
      "env": {
        "MCP_LOG_LEVEL": "debug",
        "MCP_LOG_DIR": "/path/to/logs"
      }
    }
  }
}
```

When `MCP_LOG_DIR` is set or `-log-dir` flag is used, logs are automatically placed in a subfolder named after the binary. This allows multiple MCP servers to share the same log directory:

```
MCP_LOG_DIR=/var/log/mcp
  └── go-mcp-atlassian/
      └── go-mcp-atlassian-2026-01-04.log
```

Logs are written to `{MCP_LOG_DIR}/go-mcp-atlassian/go-mcp-atlassian-{date}.log`

### Query Logging

Enable query logging to capture all Jira and Confluence queries for debugging or auditing:

```json
{
  "mcpServers": {
    "atlassian": {
      "command": "/path/to/go-mcp-atlassian",
      "env": {
        "MCP_LOG_QUERIES": "true",
        "MCP_LOG_DIR": "/path/to/logs"
      }
    }
  }
}
```

When enabled, queries are saved to: `{MCP_LOG_DIR}/queries/YYYYMMDD/{query_type}.YYYYMMDD.HHmmss.query`

Query files include:
- The query type (e.g., `jira_search`, `confluence_search`)
- Timestamp
- The actual query (JQL, CQL, etc.)
- All tool arguments

This feature is **off by default** and should only be enabled when needed for debugging or compliance purposes.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - See [LICENSE](LICENSE) for details.

## Acknowledgments

- [mcp-atlassian](https://github.com/sooperset/mcp-atlassian) - Original Python implementation
- [go-mcp-dynatrace](https://github.com/dynatrace/go-mcp-dynatrace) - Template for Go MCP implementation
- [Anthropic MCP Specification](https://modelcontextprotocol.io/) - Model Context Protocol
