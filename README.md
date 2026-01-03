# go-mcp-atlassian

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Build Status](https://github.com/nicholasgriffintn/go-mcp-atlassian/actions/workflows/ci.yml/badge.svg)](https://github.com/nicholasgriffintn/go-mcp-atlassian/actions/workflows/ci.yml)

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

Download the latest binary for your platform from the [Releases](https://github.com/nicholasgriffintn/go-mcp-atlassian/releases) page.

### 2. Get Your API Token

**For Atlassian Cloud:**
- Go to https://id.atlassian.com/manage-profile/security/api-tokens
- Create and copy your API token

**For Server/Data Center:**
- Go to your profile settings → Personal Access Tokens
- Create and copy your Personal Access Token (PAT)

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
experimental:
  modelContextProtocolServers:
    - transport:
        type: stdio
        command: /path/to/go-mcp-atlassian
        env:
          JIRA_URL: https://your-company.atlassian.net
          JIRA_USERNAME: your.email@company.com
          JIRA_API_TOKEN: your_jira_api_token
          CONFLUENCE_URL: https://your-company.atlassian.net/wiki
          CONFLUENCE_USERNAME: your.email@company.com
          CONFLUENCE_API_TOKEN: your_confluence_api_token
```

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
| `jira_get_issue` | Get issue details | Read |
| `jira_search` | Search issues using JQL | Read |
| `jira_get_user_profile` | Get user profile information | Read |
| `jira_search_fields` | Search Jira fields | Read |
| `jira_get_project_issues` | Get issues for a project | Read |
| `jira_get_transitions` | Get available status transitions | Read |
| `jira_get_worklog` | Get worklog entries | Read |
| `jira_get_agile_boards` | Get agile boards | Read |
| `jira_get_board_issues` | Get issues for a board | Read |
| `jira_get_sprints_from_board` | Get sprints from a board | Read |
| `jira_get_sprint_issues` | Get issues in a sprint | Read |
| `jira_get_link_types` | Get issue link types | Read |
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
| `jira_remove_issue_link` | Remove issue link | Write |
| `jira_transition_issue` | Transition issue status | Write |
| `jira_create_sprint` | Create a new sprint | Write |
| `jira_update_sprint` | Update sprint details | Write |
| `jira_create_version` | Create project version | Write |

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
git clone https://github.com/nicholasgriffintn/go-mcp-atlassian.git
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

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - See [LICENSE](LICENSE) for details.

## Acknowledgments

- [mcp-atlassian](https://github.com/sooperset/mcp-atlassian) - Original Python implementation
- [go-mcp-dynatrace](https://github.com/dynatrace/go-mcp-dynatrace) - Template for Go MCP implementation
- [Anthropic MCP Specification](https://modelcontextprotocol.io/) - Model Context Protocol
