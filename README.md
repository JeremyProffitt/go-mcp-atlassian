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

## Available Tools

### Jira Tools

| Tool | Description | Type |
|------|-------------|------|
| `jira_get_user_profile` | Get user profile information | Read |
| `jira_get_issue` | Get issue details | Read |
| `jira_search` | Search issues using JQL | Read |
| `jira_search_fields` | Search Jira fields | Read |
| `jira_get_project_issues` | Get issues for a project | Read |
| `jira_get_transitions` | Get available status transitions | Read |
| `jira_get_worklog` | Get worklog entries | Read |
| `jira_download_attachments` | Get attachment info and URLs | Read |
| `jira_get_agile_boards` | Get agile boards | Read |
| `jira_get_board_issues` | Get issues for a board | Read |
| `jira_get_sprints_from_board` | Get sprints from a board | Read |
| `jira_get_sprint_issues` | Get issues in a sprint | Read |
| `jira_get_link_types` | Get issue link types | Read |
| `jira_batch_get_changelogs` | Get changelogs for issues | Read |
| `jira_get_project_versions` | Get project fix versions | Read |
| `jira_get_all_projects` | Get all accessible projects | Read |
| `jira_create_issue` | Create a new issue | Write |
| `jira_batch_create_issues` | Create multiple issues | Write |
| `jira_update_issue` | Update an existing issue | Write |
| `jira_delete_issue` | Delete an issue | Write |
| `jira_add_comment` | Add comment to issue | Write |
| `jira_edit_comment` | Edit existing comment | Write |
| `jira_add_worklog` | Add worklog entry | Write |
| `jira_link_to_epic` | Link issue to an epic | Write |
| `jira_create_issue_link` | Create link between issues | Write |
| `jira_transition_issue` | Transition issue status | Write |

### Confluence Tools

| Tool | Description | Type |
|------|-------------|------|
| `confluence_search` | Search content using CQL | Read |
| `confluence_get_page` | Get page content | Read |
| `confluence_get_page_children` | Get child pages | Read |
| `confluence_get_comments` | Get page comments | Read |
| `confluence_get_labels` | Get page labels | Read |
| `confluence_search_user` | Search for users | Read |
| `confluence_create_page` | Create a new page | Write |
| `confluence_update_page` | Update existing page | Write |
| `confluence_delete_page` | Delete a page | Write |
| `confluence_add_label` | Add label to page | Write |
| `confluence_add_comment` | Add comment to page | Write |

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

Logs are written to `{MCP_LOG_DIR}/go-mcp-atlassian-{date}.log`

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
